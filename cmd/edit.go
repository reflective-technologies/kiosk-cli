package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit [project-name]",
	Short: "Resume a kiosk workspace project",
	Long: `Select a project from ~/.kiosk/workspace and resume its Claude session.

If a project name is provided, it will open that project directly.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.EnsureInitialized(); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		workspaceDir := config.WorkspaceDir()
		if err := os.MkdirAll(workspaceDir, 0755); err != nil {
			return fmt.Errorf("failed to create workspace directory: %w", err)
		}

		projectName := ""
		if len(args) == 1 {
			projectName = strings.TrimSpace(args[0])
		}

		if projectName == "" {
			selected, err := selectWorkspaceProject(workspaceDir)
			if err != nil {
				switch {
				case errors.Is(err, errNoWorkspaceProjects):
					fmt.Println()
					fmt.Println(styles.MutedStyle.Render("  No workspace projects found."))
					fmt.Println()
					return nil
				case errors.Is(err, errUserCanceled):
					return nil
				default:
					return err
				}
			}
			projectName = selected
		}

		projectDir := filepath.Join(workspaceDir, projectName)
		info, err := os.Stat(projectDir)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("workspace project %q not found", projectName)
			}
			return fmt.Errorf("failed to stat project directory: %w", err)
		}
		if !info.IsDir() {
			return fmt.Errorf("workspace project path is not a directory: %s", projectDir)
		}

		fmt.Printf("Resuming %s...\n", projectName)
		return runWorkspaceClaude(projectDir, projectName, "")
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
}
