package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

var sandboxFlag string
var safeFlag bool

const runPrompt = `Run the app in this directory. Check KIOSK.md for instructions on how to start and use this app.`

var runCmd = &cobra.Command{
	Use:   "run <app>",
	Short: "Run a kiosk app (install if needed)",
	Long: `Run a kiosk app from the marketplace. If the app is not installed,
it will be fetched and installed first.

The app can be specified as:
  - org/repo (e.g., anthropic/claude-starter)
  - appId (e.g., claude-starter)`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appArg := args[0]

		// Parse and transform sandbox values
		sandboxValues, err := parseSandboxValues(sandboxFlag)
		if err != nil {
			return err
		}
		sandboxValues = transformSandboxValues(sandboxValues)

		// Ensure working directory is initialized
		if err := config.EnsureInitialized(); err != nil {
			return fmt.Errorf("failed to initialize: %w", err)
		}

		// Load config and index
		cfg, err := config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		idx, err := appindex.Load()
		if err != nil {
			return fmt.Errorf("failed to load app index: %w", err)
		}

		// Normalize key to org/repo format for index lookup
		key := normalizeAppKey(appArg)

		// Check if app is installed
		if idx.Has(key) {
			return runInstalledApp(key, sandboxValues, safeFlag)
		}

		// App not installed - fetch from API and install
		return installAndRunApp(cfg, idx, appArg, key, sandboxValues, safeFlag)
	},
}

// normalizeAppKey ensures we have an org/repo format for the index
// If only appId is provided, we'll update this after fetching from API
func normalizeAppKey(input string) string {
	// If already has slash, assume it's org/repo
	if strings.Contains(input, "/") {
		return input
	}
	// Otherwise, return as-is (will be updated after API fetch)
	return input
}

// runInstalledApp runs an already-installed app
func runInstalledApp(key string, sandboxValues []string, safe bool) error {
	parts := strings.SplitN(key, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid app key: %s", key)
	}

	appPath := config.AppPath(parts[0], parts[1])

	// Verify directory exists
	if _, err := os.Stat(appPath); os.IsNotExist(err) {
		return fmt.Errorf("app directory missing: %s (try removing and reinstalling)", appPath)
	}

	prompt := runPrompt
	updateInfo, err := updateRepoIfNeeded(appPath)
	if err != nil {
		return err
	}
	if updateInfo != nil && updateInfo.updated {
		prompt = buildUpdatePrompt(updateInfo)
	}

	// Apply sandbox settings if specified
	if len(sandboxValues) > 0 {
		fmt.Printf("Configuring sandbox mode...\n")
		if err := writeSandboxSettings(appPath, sandboxValues); err != nil {
			return fmt.Errorf("failed to configure sandbox: %w", err)
		}
	}

	fmt.Printf("Running %s...\n", key)
	fmt.Print(logo)
	return execClaude(appPath, prompt, safe)
}

// installAndRunApp fetches an app from the API and installs it
func installAndRunApp(cfg *config.Config, idx *appindex.Index, appArg, key string, sandboxValues []string, safe bool) error {
	client := api.NewClient(cfg.APIUrl)

	// Fetch app metadata
	fmt.Printf("Fetching %s...\n", appArg)
	app, err := client.GetApp(appArg)
	if err != nil {
		return err
	}

	// Get installation prompt
	prompt, err := client.GetInstallPrompt(appArg)
	if err != nil {
		return err
	}

	// Determine the key (org/repo) from git URL if we only had appId
	if !strings.Contains(key, "/") {
		key = extractOrgRepo(app.GitUrl)
		if key == "" {
			key = app.ID // Fallback to just appId
		}
	}

	parts := strings.SplitN(key, "/", 2)
	if len(parts) != 2 {
		return fmt.Errorf("could not determine org/repo for app")
	}

	appPath := config.AppPath(parts[0], parts[1])

	parentDir := filepath.Dir(appPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("failed to create app parent directory: %w", err)
	}

	if _, err := os.Stat(appPath); err == nil {
		return fmt.Errorf("app already exists at %s (try removing it first)", appPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to check app path: %w", err)
	}

	fmt.Printf("Cloning %s...\n", app.GitUrl)
	if err := cloneRepo(app.GitUrl, appPath); err != nil {
		return err
	}

	// Apply sandbox settings if specified
	if len(sandboxValues) > 0 {
		fmt.Printf("Configuring sandbox mode...\n")
		if err := writeSandboxSettings(appPath, sandboxValues); err != nil {
			return fmt.Errorf("failed to configure sandbox: %w", err)
		}
	}

	// Register in index
	idx.Add(key, &appindex.AppEntry{
		Name:        app.Name,
		Description: app.Description,
		GitUrl:      app.GitUrl,
	})
	if err := appindex.Save(idx); err != nil {
		return fmt.Errorf("failed to save app index: %w", err)
	}

	fmt.Printf("Installing %s...\n", app.Name)
	fmt.Print(logo)
	return execClaude(appPath, prompt, safe)
}

// extractOrgRepo extracts org/repo from a GitHub URL
func extractOrgRepo(gitUrl string) string {
	// Handle https://github.com/org/repo or https://github.com/org/repo.git
	gitUrl = strings.TrimSuffix(gitUrl, ".git")

	for _, prefix := range []string{
		"https://github.com/",
		"https://gitlab.com/",
		"https://bitbucket.org/",
		"git@github.com:",
		"git@gitlab.com:",
		"git@bitbucket.org:",
	} {
		if strings.HasPrefix(gitUrl, prefix) {
			return strings.TrimPrefix(gitUrl, prefix)
		}
	}
	return ""
}

type updateInfo struct {
	updated          bool
	oldCommit        string
	newCommit        string
	hadStash         bool
	unstashConflicts bool
}

func updateRepoIfNeeded(appPath string) (*updateInfo, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, nil
	}

	inside, err := gitOutput(appPath, "rev-parse", "--is-inside-work-tree")
	if err != nil || inside != "true" {
		return nil, nil
	}

	oldCommit, err := gitOutput(appPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, nil
	}

	if err := gitRun(appPath, "fetch", "--quiet"); err != nil {
		fmt.Printf("Warning: failed to fetch updates in %s: %v\n", appPath, err)
		return nil, nil
	}

	counts, err := gitOutput(appPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return nil, nil
	}

	parts := strings.Fields(counts)
	if len(parts) != 2 {
		return nil, nil
	}

	ahead, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, nil
	}
	behind, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil
	}

	if behind == 0 {
		return nil, nil
	}
	if ahead > 0 {
		return nil, fmt.Errorf("local branch has diverged from upstream in %s; resolve manually before running", appPath)
	}

	hasChanges := false
	status, err := gitOutput(appPath, "status", "--porcelain")
	if err == nil && strings.TrimSpace(status) != "" {
		hasChanges = true
		if err := gitRun(appPath, "stash", "push", "-u", "-m", "kiosk: pre-update stash"); err != nil {
			return nil, fmt.Errorf("failed to stash local changes: %w", err)
		}
	}

	if err := gitRun(appPath, "pull", "--ff-only"); err != nil {
		if hasChanges {
			_ = gitRun(appPath, "stash", "pop")
		}
		return nil, err
	}

	newCommit, err := gitOutput(appPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	unstashConflicts := false
	if hasChanges {
		if err := gitRun(appPath, "stash", "pop"); err != nil {
			unstashConflicts = true
		}
	}

	return &updateInfo{
		updated:          true,
		oldCommit:        oldCommit,
		newCommit:        newCommit,
		hadStash:         hasChanges,
		unstashConflicts: unstashConflicts,
	}, nil
}

func buildUpdatePrompt(info *updateInfo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "You are resuming an app that was previously set up and run at commit %s.\n", info.oldCommit)
	fmt.Fprintf(&b, "The repository has been updated to commit %s on the current branch.\n", info.newCommit)
	fmt.Fprintf(&b, "Review changes between %s and %s (git log --oneline %s..%s or git diff %s..%s).\n", info.oldCommit, info.newCommit, info.oldCommit, info.newCommit, info.oldCommit, info.newCommit)
	if info.hadStash {
		if info.unstashConflicts {
			b.WriteString("Local changes were stashed and re-applied; resolve any merge conflicts from the unstash and drop the stash if it remains.\n")
		} else {
			b.WriteString("Local changes were stashed and re-applied; verify they still apply cleanly.\n")
		}
	}
	b.WriteString("Apply any configuration fixes or updates needed to get the app running again for the user.\n")
	b.WriteString("Prompt the user once installation is complete: ask what they'd like to do next via multiple choice. Tailor options to the app—some are runnable apps (dev server, production build), others are workflow-oriented (scripts, generators, automation). For workflows, offer to help run them interactively.\n")
	b.WriteString(runPrompt)
	return b.String()
}

func cloneRepo(gitURL, dest string) error {
	if gitURL == "" {
		return fmt.Errorf("app has no git URL to clone")
	}

	cmd := exec.Command("git", "clone", gitURL, dest)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repo: %w", err)
	}
	return nil
}

func runCommand(cmd *exec.Cmd, dir string) error {
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func gitOutput(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		out := strings.TrimSpace(string(output))
		if out != "" {
			return "", fmt.Errorf("git %s failed: %s", strings.Join(args, " "), out)
		}
		return "", fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(output)), nil
}

func gitRun(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		out := strings.TrimSpace(string(output))
		if out != "" {
			return fmt.Errorf("git %s failed: %s", strings.Join(args, " "), out)
		}
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

// execClaude runs claude in the given directory with the given prompt
func execClaude(dir, prompt string, safe bool) error {
	permissionMode := "bypassPermissions"
	if safe {
		permissionMode = "default"
	}

	if _, err := exec.LookPath("claude"); err == nil {
		cmd := exec.Command("claude", "--permission-mode", permissionMode, prompt)
		return runCommand(cmd, dir)
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	cmd := exec.Command(shell, "-i", "-c", "claude --permission-mode \"$1\" \"$2\"", "claude", permissionMode, prompt)
	return runCommand(cmd, dir)
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&sandboxFlag, "sandbox", "", "sandbox mode: comma-separated list of 'default', 'fs', 'net'")
	runCmd.Flags().BoolVar(&safeFlag, "safe", false, "run with default permission mode (prompts for permissions)")
}

// parseSandboxValues parses and validates the sandbox flag value
func parseSandboxValues(input string) ([]string, error) {
	if input == "" {
		return nil, nil
	}

	validValues := map[string]bool{"default": true, "fs": true, "net": true}
	seen := make(map[string]bool)
	var values []string

	for _, v := range strings.Split(input, ",") {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if !validValues[v] {
			return nil, fmt.Errorf("invalid sandbox value: %q (valid: default, fs, net)", v)
		}
		if !seen[v] {
			seen[v] = true
			values = append(values, v)
		}
	}

	return values, nil
}

// transformSandboxValues expands 'default' to include 'fs' and deduplicates
func transformSandboxValues(values []string) []string {
	if len(values) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var result []string

	for _, v := range values {
		if v == "default" {
			// 'default' expands to include 'fs'
			if !seen["fs"] {
				seen["fs"] = true
				result = append(result, "fs")
			}
		}
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}

	return result
}

// writeSandboxSettings creates or updates .claude/settings.json with sandbox config
func writeSandboxSettings(appPath string, sandboxValues []string) error {
	if len(sandboxValues) == 0 {
		return nil
	}

	claudeDir := filepath.Join(appPath, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// Ensure .claude directory exists
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .claude directory: %w", err)
	}

	// Load existing settings or create empty object
	settings := make(map[string]interface{})
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("failed to parse existing settings.json: %w", err)
		}
	}

	// Build sandbox config
	sandboxConfig := map[string]interface{}{
		"enabled": true,
	}

	hasFS := false
	hasNet := false
	for _, v := range sandboxValues {
		if v == "fs" {
			hasFS = true
		}
		if v == "net" {
			hasNet = true
		}
	}

	if hasFS {
		absPath, err := filepath.Abs(appPath)
		if err != nil {
			absPath = appPath
		}
		sandboxConfig["allowedDirectories"] = []string{absPath}
	}

	if hasNet {
		sandboxConfig["allowedDomains"] = []string{}
	}

	settings["sandbox"] = sandboxConfig

	// Write settings back
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write settings.json: %w", err)
	}

	return nil
}
