package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client is a kiosk API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// App represents an app from the API
type App struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	GitUrl      string `json:"gitUrl"`
	Branch      string `json:"branch,omitempty"`
	KioskMd     string `json:"kioskMd,omitempty"`
}

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetApp fetches app metadata by ID
// ID can be either "appId" or "org/repo" format
func (c *Client) GetApp(id string) (*App, error) {
	// Handle org/repo format - extract repo name as appId
	appId := id
	if strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			appId = parts[1] // Use repo name as appId
		}
	}

	url := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, appId)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("app %q not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var app App
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

// GetInstallPrompt fetches the installation prompt for an app
func (c *Client) GetInstallPrompt(id string) (string, error) {
	// Handle org/repo format
	appId := id
	if strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			appId = parts[1]
		}
	}

	url := fmt.Sprintf("%s/api/kiosk/%s/install", c.BaseURL, appId)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch install prompt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("app %q not found", id)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}
