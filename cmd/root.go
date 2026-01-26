package cmd

import (
	"os"

	"github.com/reflective-technologies/kiosk-cli/internal/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "kiosk",
	Short:         "Kiosk CLI",
	Long:          logo + `The app store for Claude Code apps.`,
	Version:       Version,
	SilenceErrors: true, // We handle error display ourselves
	SilenceUsage:  true, // Don't show usage on errors
	Run: func(cmd *cobra.Command, args []string) {
		// Launch interactive TUI when no subcommand is given
		if err := runInteractiveTUI(); err != nil {
			errors.PrintError(err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		errors.PrintError(err)
		os.Exit(1)
	}
}

func init() {
	// Enable verbose error logging in dev mode
	errors.DevMode = Version == "dev"

	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kiosk.yaml)")
}
