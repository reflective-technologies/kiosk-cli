package config

import (
	"os"
	"path/filepath"
)

const (
	kioskDirName          = ".kiosk"
	appsDirName           = "apps"
	workspaceDirName      = "workspace"
	configFileName        = "config.json"
	sessionsFile          = "sessions.json"
	workspaceSessionsFile = "sessions.workspace.json"
)

// KioskDir returns the path to ~/.kiosk
func KioskDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory if home can't be determined
		return kioskDirName
	}
	return filepath.Join(home, kioskDirName)
}

// AppsDir returns the path to ~/.kiosk/apps
func AppsDir() string {
	return filepath.Join(KioskDir(), appsDirName)
}

// WorkspaceDir returns the path to ~/.kiosk/workspace
func WorkspaceDir() string {
	return filepath.Join(KioskDir(), workspaceDirName)
}

// AppPath returns the path to a specific app: ~/.kiosk/apps/org/repo
func AppPath(org, repo string) string {
	return filepath.Join(AppsDir(), org, repo)
}

// WorkspacePath returns the path to a workspace project: ~/.kiosk/workspace/project
func WorkspacePath(project string) string {
	return filepath.Join(WorkspaceDir(), project)
}

// ConfigPath returns the path to ~/.kiosk/config.json
func ConfigPath() string {
	return filepath.Join(KioskDir(), configFileName)
}

// SessionsPath returns the path to ~/.kiosk/sessions.json
func SessionsPath() string {
	return filepath.Join(KioskDir(), sessionsFile)
}

// WorkspaceSessionsPath returns the path to ~/.kiosk/sessions.workspace.json
func WorkspaceSessionsPath() string {
	return filepath.Join(KioskDir(), workspaceSessionsFile)
}
