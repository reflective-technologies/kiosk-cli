package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const (
	repoOwner = "reflective-technologies"
	repoName  = "kiosk-cli"
)

type githubRelease struct {
	TagName string `json:"tag_name"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update kiosk to the latest version",
	Long:  `Downloads and installs the latest version of kiosk from GitHub releases.`,
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	fmt.Printf("Current version: %s\n", Version)

	// Fetch latest version
	latest, err := fetchLatestVersion()
	if err != nil {
		return fmt.Errorf("failed to fetch latest version: %w", err)
	}

	fmt.Printf("Latest version: %s\n", latest)

	// Compare versions
	if Version == latest || "v"+Version == latest {
		fmt.Println("Already up to date!")
		return nil
	}

	if Version != "dev" {
		fmt.Printf("Updating %s -> %s\n", Version, latest)
	} else {
		fmt.Printf("Installing %s\n", latest)
	}

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	// Download and install
	if err := downloadAndInstall(latest, execPath); err != nil {
		return fmt.Errorf("failed to update: %w", err)
	}

	fmt.Printf("Successfully updated to %s!\n", latest)
	return nil
}

func fetchLatestVersion() (string, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", repoOwner, repoName)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", err
	}

	return release.TagName, nil
}

func downloadAndInstall(version, execPath string) error {
	// Determine OS and arch
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Version without 'v' prefix for asset name
	versionNum := strings.TrimPrefix(version, "v")

	assetName := fmt.Sprintf("kiosk_%s_%s_%s.tar.gz", versionNum, goos, goarch)
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s",
		repoOwner, repoName, version, assetName)

	fmt.Printf("Downloading %s...\n", assetName)

	// Download to temp file
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "kiosk-update-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// Extract binary from tarball
	newBinaryPath, err := extractBinary(resp.Body, tmpDir)
	if err != nil {
		return fmt.Errorf("failed to extract: %w", err)
	}

	// Replace current binary
	if err := replaceBinary(newBinaryPath, execPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

func extractBinary(r io.Reader, destDir string) (string, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Look for the kiosk binary
		if header.Typeflag == tar.TypeReg && filepath.Base(header.Name) == "kiosk" {
			outPath := filepath.Join(destDir, "kiosk")
			outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return "", err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()

			return outPath, nil
		}
	}

	return "", fmt.Errorf("kiosk binary not found in archive")
}

func replaceBinary(newPath, oldPath string) error {
	// Get permissions from old binary
	info, err := os.Stat(oldPath)
	if err != nil {
		return err
	}

	// On Unix, we can't write to a running binary, but we can rename it
	// Rename old binary to .old
	oldBackup := oldPath + ".old"
	if err := os.Rename(oldPath, oldBackup); err != nil {
		return fmt.Errorf("failed to backup old binary: %w", err)
	}

	// Copy new binary to target location
	newFile, err := os.Open(newPath)
	if err != nil {
		// Restore old binary
		os.Rename(oldBackup, oldPath)
		return err
	}
	defer newFile.Close()

	outFile, err := os.OpenFile(oldPath, os.O_CREATE|os.O_WRONLY, info.Mode())
	if err != nil {
		// Restore old binary
		os.Rename(oldBackup, oldPath)
		return err
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, newFile); err != nil {
		// Restore old binary
		os.Rename(oldBackup, oldPath)
		return err
	}

	// Remove backup
	os.Remove(oldBackup)

	return nil
}
