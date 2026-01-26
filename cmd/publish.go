package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish the current repo to kiosk.app",
	Long: `Publish the current repository to the kiosk.app marketplace.

Run this command from within a git repository that has a GitHub remote.
Claude Code will guide you through the publishing process.

Note: Run 'kiosk init' first to create a KIOSK.md file if you don't have one.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if audit flag is set
		runAudit, _ := cmd.Flags().GetBool("audit")
		if runAudit {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			if err := execClaudeAudit(cwd, auditPrompt); err != nil {
				return fmt.Errorf("audit failed: %w", err)
			}

			fmt.Print("\nContinue with publish? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Publish cancelled.")
				return nil
			}
			fmt.Println()
		}

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

		// Check for KIOSK.md and warn if missing
		if !kioskMdExists(cwd) {
			fmt.Println("Warning: No KIOSK.md found. Consider running 'kiosk init' first to create one.")
		}

		client := api.NewClient(cfg.APIUrl)

		// Fetch the publish prompt
		fmt.Println("Fetching publish instructions...")
		prompt, err := client.GetPublishPrompt()
		if err != nil {
			return err
		}

		// Get safe flag
		safe, _ := cmd.Flags().GetBool("safe")

		// Exec claude with the prompt in the current directory
		fmt.Println("Starting Claude Code...")
		return execClaude(cwd, prompt, safe)
	},
}

// kioskMdExists checks if a KIOSK.md file exists in the given directory
func kioskMdExists(dir string) bool {
	variants := []string{"KIOSK.md", "Kiosk.md", "kiosk.md"}
	for _, name := range variants {
		if _, err := os.Stat(dir + "/" + name); err == nil {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(publishCmd)
	publishCmd.Flags().Bool("safe", false, "Run Claude Code in safe mode (prompts for permissions)")
	publishCmd.Flags().Bool("audit", false, "Run security audit before publishing")
}
