package cmd

import (
	"fmt"
	"os"

	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish the current repo to kiosk.app",
	Long: `Publish the current repository to the kiosk.app marketplace.

Run this command from within a git repository that has a GitHub remote.
Claude Code will guide you through the publishing process, including:
  - Detecting project information
  - Creating a Kiosk.md if needed
  - Publishing to kiosk.app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		client := api.NewClient(cfg.APIUrl)

		// Fetch the create/publish prompt
		fmt.Println("Fetching publish instructions...")
		prompt, err := client.GetCreatePrompt()
		if err != nil {
			return err
		}

		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Exec claude with the prompt in the current directory
		fmt.Println("Starting Claude Code...")
		return execClaude(cwd, prompt)
	},
}

func init() {
	rootCmd.AddCommand(publishCmd)
}
