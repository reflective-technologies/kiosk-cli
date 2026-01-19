package api

import (
	"bytes"
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

// CreateAppRequest represents the payload for creating an app
type CreateAppRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	GitUrl       string `json:"gitUrl"`
	Branch       string `json:"branch,omitempty"`
	Subdirectory string `json:"subdirectory,omitempty"`
	Screenshot   string `json:"screenshot,omitempty"`
	Instructions string `json:"instructions,omitempty"`
}

// UpdateAppRequest represents the payload for updating an app
type UpdateAppRequest CreateAppRequest

// NewClient creates a new API client
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetApp fetches app metadata by ID.
// ID can be either "appId" or "org/repo" format.
// When "org/repo" format is used, only the repo name is extracted as the appId.
// Note: This means different orgs with same-named repos would resolve to the same app.
// This matches the Kiosk API behavior where apps are identified by repo name alone.
func (c *Client) GetApp(id string) (*App, error) {
	appId := id
	if strings.Contains(id, "/") {
		parts := strings.SplitN(id, "/", 2)
		if len(parts) == 2 {
			appId = parts[1]
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

// GetInstallPrompt fetches the installation prompt for an app.
// ID can be either "appId" or "org/repo" format (see GetApp for details).
func (c *Client) GetInstallPrompt(id string) (string, error) {
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

// GetCreatePrompt fetches the publish/create prompt
func (c *Client) GetCreatePrompt() (string, error) {
	url := fmt.Sprintf("%s/api/create", c.BaseURL)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch create prompt: %w", err)
	}
	defer resp.Body.Close()

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

// ListApps fetches all published apps
func (c *Client) ListApps() ([]App, error) {
	url := fmt.Sprintf("%s/api/kiosk", c.BaseURL)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch apps: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	var apps []App
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apps, nil
}

// CreateApp publishes a new app
func (c *Client) CreateApp(req CreateAppRequest) (*App, error) {
	url := fmt.Sprintf("%s/api/kiosk", c.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var app App
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

// UpdateApp updates an existing app
func (c *Client) UpdateApp(id string, req UpdateAppRequest) (*App, error) {
	url := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, id)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	reqHTTP, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(reqHTTP)
	if err != nil {
		return nil, fmt.Errorf("failed to update app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var app App
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

// DeleteApp removes an app
func (c *Client) DeleteApp(id string) error {
	url := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, id)

	reqHTTP, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient.Do(reqHTTP)
	if err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// RefreshApp triggers a refresh of the app's Kiosk.md from the repository
func (c *Client) RefreshApp(id string) error {
	url := fmt.Sprintf("%s/api/kiosk/%s/refresh", c.BaseURL, id)

	resp, err := c.HTTPClient.Post(url, "application/json", nil)
	if err != nil {
		return fmt.Errorf("failed to refresh app: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}
