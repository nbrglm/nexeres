package cmd

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/nbrglm/nexeres/config"
	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/nbrglm/nexeres/opts"
	"github.com/nbrglm/nexeres/utils"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func initMigrationCommand(migrationsFS embed.FS) {
	var migrateCommand = &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
		Run: func(cmd *cobra.Command, args []string) {
			// Start the migration
			up, _ := cmd.Flags().GetBool("up")
			down, _ := cmd.Flags().GetBool("down")
			version, _ := cmd.Flags().GetBool("version")
			numericVersion, _ := cmd.Flags().GetBool("numeric-version")

			// Initialize the validator before everything else, since validation is used by the config file loader.
			utils.InitValidator()

			// Load the configuration file
			if err := config.LoadConfigOptions(*opts.ConfigPath); err != nil {
				fatal(cmd, "error loading config file: %v\n", err)
			}

			m, err := getMigrations(migrationsFS) // We ignore the error here, since getMigrations already prints it.
			if err != nil {
				fatal(cmd, "Error initializing migrations: %v\n", err)
			}

			if version || numericVersion {
				v, dirty, err := m.Version()
				if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
					fatal(cmd, "Error getting migration version: %v\n", err)
				}
				if errors.Is(err, migrate.ErrNilVersion) {
					if numericVersion {
						cmd.Println("0")
						return
					}
					cmd.Println("Current migration version: none (database is not migrated)")
					return
				}

				if dirty && !numericVersion {
					cmd.Printf("%d (dirty)\n", v)
				} else {
					cmd.Printf("%d\n", v)
				}
				return
			}

			err = runMigrations(up, down, cmd, m)
			if err != nil {
				fatal(cmd, "Migration error: %v\n", err)
			}
			cmd.Println("Migration completed successfully.")
		},
	}

	migrateCommand.Flags().StringVar(opts.ConfigPath, "config", "/etc/nbrglm/workspace/nexeres/config.yaml", "Path to the config file")
	migrateCommand.MarkPersistentFlagFilename("config", "yaml", "yml")
	migrateCommand.Flags().Bool("up", false, "Run ALL the UP migrations, to the latest version.")
	migrateCommand.Flags().Bool("down", false, "Run ONE DOWN migration. Useful to rollback to the last version.")
	migrateCommand.Flags().Bool("force", false, "Force the migration to run. This is just a safeguard, ONLY affects the DOWN migration.")
	migrateCommand.Flags().Bool("version", false, "Print the current migration version and exit.")
	migrateCommand.Flags().Bool("numeric-version", false, "Print the current migration version as a number and exit, regardless of dirty state. If the database is not migrated, prints '0'.")
	migrateCommand.MarkFlagsMutuallyExclusive("up", "down", "version", "numeric-version")
	migrateCommand.MarkFlagsRequiredTogether("down", "force")
	migrateCommand.MarkFlagRequired("config")

	rootCmd.AddCommand(migrateCommand)
}

// runMigrations runs the migrations based on the provided flags.
// If 'up' is true, it runs all UP migrations to the latest version.
// If 'down' is true, it runs one DOWN migration.
// It returns an error if any operation fails.
//
// Note: Only call this AFTER the config file has been loaded, as it uses the DSN from the config.
func runMigrations(up bool, down bool, cmd *cobra.Command, m *migrate.Migrate) error {
	if up {
		cmd.Println("Running UP migrations on database...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running UP migrations: %w", err)
		}
		cmd.Println("UP migrations completed successfully.")
	} else if down {
		cmd.Println("Running ONE DOWN migration on database...")
		if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running ONE DOWN migration: %w", err)
		}
		cmd.Println("ONE DOWN migration completed successfully.")
	} else {
		return fmt.Errorf("either up or down must be true")
	}

	if err := errors.Join(m.Close()); err != nil {
		return fmt.Errorf("error closing migration resources: %w", err)
	}

	return nil
}

// getMigrations initializes the migration instance with the provided filesystem.
//
// Note: Only call this AFTER the config file has been loaded, as it uses the DSN from the config.
func getMigrations(migrationsFS embed.FS) (*migrate.Migrate, error) {
	migrationDSN := config.C.Stores.Postgres.DSN
	if strings.HasPrefix(migrationDSN, "postgres://") {
		migrationDSN = strings.Replace(migrationDSN, "postgres://", "pgx5://", 1)
	} else {
		return nil, fmt.Errorf("only 'postgres://' DSN is supported for migrations, \"%v\"", config.C.Stores.Postgres.DSN)
	}

	migrationSource, err := iofs.New(migrationsFS, "sqlc/migrations")
	if err != nil {
		return nil, fmt.Errorf("error initializing migration source: %w", err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", migrationSource, migrationDSN)
	if err != nil {
		return nil, fmt.Errorf("error initializing migrations: %w", err)
	}
	m.Log = ZapWrapperLogger{logger: logging.Logger}
	return m, nil
}

type ZapWrapperLogger struct {
	logger *zap.Logger
}

func (z ZapWrapperLogger) Printf(format string, v ...interface{}) {
	if opts.Debug {
		fmt.Printf(format, v...)
	} else {
		z.logger.Sugar().Infof(format, v...)
	}
}

func (z ZapWrapperLogger) Verbose() bool {
	return true
}

func hasPendingMigrations(m *migrate.Migrate, migrationsFS embed.FS) (bool, error) {
	version, dirty, err := m.Version()
	if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
		return false, fmt.Errorf("error getting migration version: %w", err)
	}
	if dirty {
		return false, fmt.Errorf("database is in a dirty state at version %v, please fix it manually", version)
	}
	if errors.Is(err, migrate.ErrNilVersion) {
		// Database is not migrated at all, so there are definitely pending migrations.
		return true, nil
	}

	// Initialize a migration source
	migrationSource, err := iofs.New(migrationsFS, "sqlc/migrations")
	if err != nil {
		return false, fmt.Errorf("error initializing migration source: %w", err)
	}
	_, err = migrationSource.Next(version)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return false, fmt.Errorf("error getting next migration version: %w", err)
	}
	if errors.Is(err, os.ErrNotExist) {
		// No next migration, so we are at the latest version.
		return false, nil
	}

	return true, nil
}
