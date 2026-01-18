package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

const kioskMdTemplate = `# My Kiosk App

> Brief description of what this app does.

## Installation

<!-- Describe how Claude Code should install this app -->

### Files to Copy

Copy the following files to your project:

- ` + "`src/example.ts`" + ` -> ` + "`src/example.ts`" + `

### Dependencies

Install required dependencies:

` + "```bash" + `
npm install <your-dependencies>
` + "```" + `

## Usage

<!-- Describe how to use the installed app -->

` + "```typescript" + `
import { example } from './example';

// Usage example
` + "```" + `

## Configuration

<!-- Any configuration needed -->

## Notes

<!-- Additional notes for the installing agent -->
`

var newCmd = &cobra.Command{
	Use:   "new <org/repo>",
	Short: "Initialize a new kiosk app project",
	Long: `Create a new kiosk app project with a KIOSK.md template.

Creates a new project at ~/.kiosk/apps/<org>/<repo> with a starter KIOSK.md.

Example:
  kiosk new myorg/myapp`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Validate org/repo format
		parts := strings.SplitN(key, "/", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return fmt.Errorf("invalid format: expected org/repo (e.g., myorg/myapp)")
		}

		// Ensure working directory is initialized
		if err := config.EnsureInitialized(); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		// Create project directory in ~/.kiosk/apps/org/repo
		projectDir := config.AppPath(parts[0], parts[1])
		if _, err := os.Stat(projectDir); err == nil {
			return fmt.Errorf("project already exists at %s", projectDir)
		}

		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		fmt.Printf("Created %s\n", projectDir)

		// Write KIOSK.md template
		kioskPath := filepath.Join(projectDir, "KIOSK.md")
		if err := os.WriteFile(kioskPath, []byte(kioskMdTemplate), 0644); err != nil {
			return fmt.Errorf("failed to write KIOSK.md: %w", err)
		}
		fmt.Println("Created KIOSK.md")

		// Initialize git repo
		gitCmd := exec.Command("git", "init")
		gitCmd.Dir = projectDir
		if err := gitCmd.Run(); err != nil {
			fmt.Println("Warning: failed to initialize git repo")
		} else {
			fmt.Println("Initialized git repository")
		}

		fmt.Println("\nNext steps:")
		fmt.Printf("  1. cd %s\n", projectDir)
		fmt.Println("  2. Edit KIOSK.md with your app's installation instructions")
		fmt.Println("  3. Add your app code")
		fmt.Println("  4. Push to GitHub")
		fmt.Println("  5. Run 'kiosk publish' to publish to kiosk.app")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
