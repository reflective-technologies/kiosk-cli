package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/skills"
	"github.com/spf13/cobra"
)

const defaultClaudeTemplate = `# Claude Project Context

This repository is intended to be published on Kiosk, the app store for Claude Code apps.

Before publishing to Kiosk, make sure the project lives in a public GitHub repository
and all changes are pushed. Kiosk requires a publicly accessible GitHub repo so the
registry can point users to your code.

Use the kiosk skill for publishing and updating apps in the Kiosk registry. It
includes the end-to-end init and publish workflows, guidance for metadata, and how
to update existing listings.
`

var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new kiosk workspace project",
	Long: `Create a new project in ~/.kiosk/workspace and start a Claude session
for initial requirements gathering.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Ensure working directory is initialized
		if err := config.EnsureInitialized(); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		workspaceDir := config.WorkspaceDir()
		if err := os.MkdirAll(workspaceDir, 0755); err != nil {
			return fmt.Errorf("failed to create workspace directory: %w", err)
		}

		initialName := ""
		if len(args) == 1 {
			initialName = strings.TrimSpace(args[0])
		}

		projectName, err := resolveProjectName(initialName, workspaceDir)
		if err != nil {
			if errors.Is(err, errUserCanceled) {
				return nil
			}
			return err
		}

		projectDir := filepath.Join(workspaceDir, projectName)
		if _, err := os.Stat(projectDir); err == nil {
			return fmt.Errorf("project already exists at %s", projectDir)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to check project directory: %w", err)
		}

		if err := os.MkdirAll(projectDir, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %w", err)
		}

		if err := initWorkspaceGitRepo(projectDir); err != nil {
			return err
		}

		if err := installKioskSkill(projectDir); err != nil {
			return err
		}

		if err := writeDefaultClaude(projectDir); err != nil {
			return err
		}

		prompt := fmt.Sprintf("This is a new project directory for the %s project. Please guide me through initial requirements gathering for my project so we can implement it as soon as possible.", projectName)
		fmt.Printf("Starting Claude Code for %s...\n", projectName)
		return runWorkspaceClaude(projectDir, projectName, prompt)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}

func initWorkspaceGitRepo(projectDir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git is required to initialize the repository: %w", err)
	}

	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = projectDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to initialize git repo: %w", err)
	}
	return nil
}

func installKioskSkill(projectDir string) error {
	dirs := []string{
		filepath.Join(projectDir, ".skills", "kiosk"),
		filepath.Join(projectDir, ".claude", "skills", "kiosk"),
	}

	files := []struct {
		name    string
		content string
	}{
		{name: "SKILL.md", content: skills.KioskSkill},
		{name: "init-prompt.md", content: skills.KioskInitPrompt},
		{name: "publish-prompt.md", content: skills.KioskPublishPrompt},
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create skills directory: %w", err)
		}
		for _, file := range files {
			path := filepath.Join(dir, file.name)
			if err := os.WriteFile(path, []byte(file.content), 0644); err != nil {
				return fmt.Errorf("failed to write kiosk skill file %s: %w", file.name, err)
			}
		}
	}

	return nil
}

func writeDefaultClaude(projectDir string) error {
	path := filepath.Join(projectDir, "CLAUDE.md")
	if err := os.WriteFile(path, []byte(defaultClaudeTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write CLAUDE.md: %w", err)
	}
	return nil
}
