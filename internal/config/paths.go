package config

import (
	"os"
	"path/filepath"
)

const (
	kioskDirName   = ".kiosk"
	appsDirName    = "apps"
	configFileName = "config.json"
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

// AppPath returns the path to a specific app: ~/.kiosk/apps/org/repo
func AppPath(org, repo string) string {
	return filepath.Join(AppsDir(), org, repo)
}

// ConfigPath returns the path to ~/.kiosk/config.json
func ConfigPath() string {
	return filepath.Join(KioskDir(), configFileName)
}
