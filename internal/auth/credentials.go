package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Credentials stores the user's authentication credentials
type Credentials struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	Scope       string    `json:"scope"`
	CreatedAt   time.Time `json:"created_at"`
}

// CredentialsPath returns the path to the credentials file
func CredentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".kiosk", "credentials.json")
	}
	return filepath.Join(home, ".kiosk", "credentials.json")
}

// SaveCredentials saves the credentials to disk with secure permissions
func SaveCredentials(creds *Credentials) error {
	// Ensure the directory exists
	dir := filepath.Dir(CredentialsPath())
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal credentials: %w", err)
	}

	// Write with restricted permissions (owner read/write only)
	if err := os.WriteFile(CredentialsPath(), data, 0600); err != nil {
		return fmt.Errorf("failed to write credentials: %w", err)
	}

	return nil
}

// LoadCredentials reads the credentials from disk
func LoadCredentials() (*Credentials, error) {
	data, err := os.ReadFile(CredentialsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No credentials file, user not logged in
		}
		return nil, fmt.Errorf("failed to read credentials: %w", err)
	}

	var creds Credentials
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("failed to parse credentials: %w", err)
	}

	return &creds, nil
}

// DeleteCredentials removes the credentials file
func DeleteCredentials() error {
	err := os.Remove(CredentialsPath())
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete credentials: %w", err)
	}
	return nil
}

// IsLoggedIn checks if valid credentials exist
func IsLoggedIn() bool {
	creds, err := LoadCredentials()
	return err == nil && creds != nil && creds.AccessToken != ""
}

// GetToken returns the current access token or an error if not logged in
func GetToken() (string, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return "", err
	}
	if creds == nil || creds.AccessToken == "" {
		return "", fmt.Errorf("not logged in, run 'kiosk login' first")
	}
	return creds.AccessToken, nil
}
