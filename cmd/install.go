package cmd

import (
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <app>",
	Short: "Install and run a kiosk app (alias for 'run')",
	Long: `Install and run a kiosk app from the marketplace. This is an alias for the 'run' command.

If the app is not already installed, it will be fetched and installed first, then run.

The app can be specified as:
  - org/repo (e.g., anthropic/claude-starter)
  - appId (e.g., claude-starter)`,
	Args: cobra.ExactArgs(1),
	RunE: runCmd.RunE, // Use the same function as the run command
}

func init() {
	rootCmd.AddCommand(installCmd)
	// Add the same flags as run command
	installCmd.Flags().StringVar(&sandboxFlag, "sandbox", "", "sandbox mode: comma-separated list of 'default', 'fs', 'net'")
	installCmd.Flags().BoolVar(&safeFlag, "safe", false, "run with default permission mode (prompts for permissions)")
}
