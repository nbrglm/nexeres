package cmd

import (
	"context"
	"embed"
	"fmt"
	"os"

	"github.com/nbrglm/nexeres/internal/logging"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use:   "nexeres",
	Short: "NBRGLM Nexeres CLI application",
	Long:  "A command line interface for the NBRGLM Nexeres",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cmd.PrintErrln("Please specify a command to run!")
		cmd.Help()
	},
}

// Exec starts the command application
// This is the entry point for the command line application.
// It is responsible for setting up the command line interface and executing the commands.
// Only supposed to be called once, when the application is started, by the main function.
func Exec(migrationsFS embed.FS) {
	initServeCommand(migrationsFS)
	initKeygenCommand()
	initMigrationCommand(migrationsFS)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		// Exit with a non-zero status code to indicate an error
		// This is important for CI/CD pipelines and other automated systems.
		os.Exit(1)
	}
}

func fatal(cmd *cobra.Command, format string, args ...any) {
	cmd.PrintErrf(format, args...)
	os.Exit(1)
}

func fatalLogger(msg string, fields ...zap.Field) {
	logging.Logger.Error(msg, fields...)
	logging.ShutdownLogger(context.Background())
	os.Exit(1)
}
