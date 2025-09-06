package cmd

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/handlers"
	"github.com/nbrglm/nexeres/internal/cache"
	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/nbrglm/nexeres/internal/metrics"
	"github.com/nbrglm/nexeres/internal/middlewares"
	"github.com/nbrglm/nexeres/internal/notifications"
	"github.com/nbrglm/nexeres/internal/notifications/templates"
	"github.com/nbrglm/nexeres/internal/store"
	"github.com/nbrglm/nexeres/internal/tokens"
	"github.com/nbrglm/nexeres/internal/tracing"
	_ "github.com/nbrglm/nexeres/oapispec"
	"github.com/nbrglm/nexeres/opts"
	"github.com/nbrglm/nexeres/utils"
	"github.com/spf13/cobra"
	swaggerFiles "github.com/swaggo/files"     // swagger embed files
	ginSwagger "github.com/swaggo/gin-swagger" // gin-swagger middleware
	"go.uber.org/zap"
)

func initServeCommand(migrationsFS embed.FS) {
	var serveCommand = &cobra.Command{
		Use:   "serve",
		Short: "Start the server and listen for incoming requests",
		Run: func(cmd *cobra.Command, args []string) {
			// Start the server
			runServer(cmd, migrationsFS)
		},
	}

	serveCommand.Flags().StringVar(opts.ConfigPath, "config", "/etc/nbrglm/workspace/nexeres/config.yaml", "Path to the config file")
	serveCommand.MarkPersistentFlagFilename("config", "yaml", "yml")

	rootCmd.AddCommand(serveCommand)
}

func runServer(cmd *cobra.Command, migrationsFS embed.FS) {
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("Starting %s v%s...\n", opts.Name, opts.Version)
	fmt.Printf("Config file: %s\n", *opts.ConfigPath)

	// Initialize the validator before everything else, since validation is used by the config file loader.
	utils.InitValidator()

	// Load the configuration file
	if err := config.LoadConfigOptions(*opts.ConfigPath); err != nil {
		fatal(cmd, "Error loading config file: %v\n", err)
	}

	// Initialize the logger
	logging.InitLogger()

	// Before doing ANYTHING else, we migrate the DB to the latest version.
	// This ensures that the DB is always in the latest state, and we don't run into issues
	// due to missing tables, columns, etc.
	m, err := getMigrations(migrationsFS)
	if err != nil {
		// We don't wrap the error here, since getMigrations already wraps it.
		fatalLogger("Error initializing migrations: %v\n", zap.Error(err))
	}
	if opts.Debug {
		cmd.Println("Running database migrations...")
		err = runMigrations(true, false, cmd, m)
		if err != nil {
			fatalLogger("Error running migrations: %v\n", zap.Error(err))
		}
		cmd.Println("Database migrations completed.")
	} else {
		pending, err := hasPendingMigrations(m, migrationsFS)
		if err != nil {
			fatalLogger("Error checking for pending migrations: %v\n", zap.Error(err))
		}
		if pending {
			fatalLogger("Database is not up-to-date, please run the following command to migrate the database.", zap.String("command", fmt.Sprintf("%s migrate --up --config %s", opts.Name, *opts.ConfigPath)))
		}
	}

	// Close the migration instance, as we don't need it anymore.
	if err := errors.Join(m.Close()); err != nil {
		fatalLogger("Error closing migration instance", zap.Error(err))
	}

	engine := gin.Default()
	if opts.Debug {
		gin.SetMode(gin.DebugMode)
		logging.Logger.Warn("Debug mode is enabled! This is not recommended for production environments. Use with caution. The following behaviour is used.", zap.String("Debug Mode", "Enabled"), zap.String("API Docs", fmt.Sprintf("%s/docs", config.Public.GetBaseURL())), zap.String("CSRF Protection", "Disabled"))
		// Setup docs
		engine.GET("/docs", func(ctx *gin.Context) {
			ctx.Header("Content-Type", "text/html")
			ctx.String(200, `<!doctype html>
	<html>
		<head>
			<title>API Reference</title>
			<meta charset="utf-8" />
			<meta
				name="viewport"
				content="width=device-width, initial-scale=1" />
		</head>
		<body>
			<script
				id="api-reference"
				data-url="/swagger/doc.json"></script>
			<script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
		</body>
	</html>`)
		})
		engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Add CORS middleware
	middlewares.InitCORS(engine)

	// Add the API Key middleware, before rate limiting middlewares,
	// since those need access to keys for rate limiting
	engine.Use(middlewares.APIKeyMiddleware())

	// Initialize the rate limiter, before adding the handler routes.
	if err := middlewares.InitRateLimitStore(); err != nil {
		fatalLogger("Failed to initialize rate limit store", zap.Error(err))
	}

	// Add the rate limit middleware AFTER the API Key middleware,
	// since it needs access to the API key to apply rate limits.
	engine.Use(middlewares.RateLimitMiddleware())

	// Register the routes
	handlers.RegisterAPIRoutes(engine)

	// Health check endpoint
	engine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	// Initialize the metrics collection system
	//
	// NOTE: Always do this after registering the API routes.
	//
	// This is because the collectors need to be registered with the Prometheus registry
	// before the metrics route is added to the engine.
	// And the collectors are only assigned in the Register() methods of each handler, hence we need
	// to call this after registering the API routes.
	// This will also register the /metrics route to serve the metrics in Prometheus format.
	// This is done to ensure that the metrics are collected and reported correctly.
	metrics.InitMetrics()
	metrics.AddMetricsRoute(engine)

	// Initialize the OpenTelemetry Tracer
	err = tracing.InitTracer(context.Background())
	if err != nil {
		// If OTEL tracer provider initialization fails, log the error and exit
		fatalLogger("Failed to initialize OTEL tracer provider", zap.Error(err))
	}

	tracing.AddTracingMiddleware(engine)

	// Parse the notification templates
	if err := templates.ParseEmailTemplates(); err != nil {
		fatalLogger("Failed to parse email templates", zap.Error(err))
	}
	if err := templates.ParseMessageTemplates(); err != nil {
		fatalLogger("Failed to parse message templates", zap.Error(err))
	}

	// Setup Notifications senders
	notifications.InitEmail()
	notifications.InitSMS()

	// Initialize the cache
	if err := cache.InitCache(); err != nil {
		fatalLogger("Failed to initialize cache", zap.Error(err))
	}

	// Initialize the token generation and keys
	if err := tokens.InitTokens(); err != nil {
		fatalLogger("Failed to initialize tokens", zap.Error(err))
	}

	// Connect with the database
	if err := store.InitDB(context.Background()); err != nil {
		fatalLogger("Failed to initialize database connection pool", zap.Error(err))
	}

	// Initialize the s3 store
	if err := store.InitS3Store(context.Background()); err != nil {
		fatalLogger("Failed to initialize S3 store", zap.Error(err))
	}

	// Start the server
	serverAddress := fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port)
	srv := &http.Server{
		Addr:    serverAddress,
		Handler: engine.Handler(),
	}

	logging.Logger.Info("Starting server", zap.String("address", serverAddress))
	fmt.Printf("Starting server at %v", serverAddress)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fatalLogger("Failed to start server", zap.Error(err))
	}

	fmt.Printf("Started server at %v", serverAddress)
	logging.Logger.Info("Started server", zap.String("Address", serverAddress))

	// Wait for OS signals to gracefully shutdown the server
	<-osSignals

	logging.Logger.Info("Received shutdown signal, shutting down server gracefully...")

	logging.Logger.Info("Closing database connection pool")
	if err := store.CloseDB(); err != nil {
		logging.Logger.Error("Failed to close database connection pool", zap.Error(err))
	}

	logging.Logger.Info("Shutting down OTEL tracer provider")
	if err := tracing.ShutdownTracer(context.Background()); err != nil {
		logging.Logger.Error("Failed to shutdown OTEL tracer provider", zap.Error(err))
	}

	// Metrics collector shutdown is not needed as it is handled by the Prometheus registry

	// Shut down server
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logging.Logger.Error("Failed to shutdown server gracefully", zap.Error(err))
	}

	// Wait for the context to be done before exiting
	<-ctx.Done()

	logging.Logger.Info("Shutting down logger.")
	if err := logging.ShutdownLogger(context.Background()); err != nil {
		fmt.Printf("Failed to shutdown logger, %v", err)
	}

	// No logging.* calls after this point, as the logger is shut down.
}
