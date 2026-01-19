package cmd

import (
	"fmt"
	"os"

	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a KIOSK.md file for your repository",
	Long: `Initialize your repository for publishing to kiosk.app.

Claude Code will analyze your project and create a KIOSK.md file with
installation instructions for users who install your app.

Run this command from within a git repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check authentication
		if !auth.IsLoggedIn() {
			return fmt.Errorf("not logged in, run 'kiosk login' first")
		}

		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		client := api.NewClient(cfg.APIUrl)

		// Fetch the init prompt
		fmt.Println("Fetching init instructions...")
		prompt, err := client.GetInitPrompt()
		if err != nil {
			return err
		}

		// Exec claude with the prompt in safe mode (prompts for permissions)
		fmt.Println("Starting Claude Code...")
		return execClaude(cwd, prompt, true)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
