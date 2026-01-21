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
	token      string // GitHub access token for authenticated requests
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

// NewClient creates a new API client without authentication
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// NewAuthenticatedClient creates a new API client with GitHub token authentication
func NewAuthenticatedClient(baseURL, token string) *Client {
	return &Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		token: token,
	}
}

// SetToken sets the authentication token for the client
func (c *Client) SetToken(token string) {
	c.token = token
}

// doRequest performs an HTTP request with optional authentication
func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.HTTPClient.Do(req)
}

// doAuthenticatedRequest performs an HTTP request that requires authentication
// Returns an error if no token is set
func (c *Client) doAuthenticatedRequest(req *http.Request) (*http.Response, error) {
	if c.token == "" {
		return nil, fmt.Errorf("authentication required: no token set")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	return c.HTTPClient.Do(req)
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

// GetInitPrompt fetches the KIOSK.md creation prompt
func (c *Client) GetInitPrompt() (string, error) {
	url := fmt.Sprintf("%s/api/prompts/init", c.BaseURL)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch init prompt: %w", err)
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

// GetPublishPrompt fetches the publish prompt
func (c *Client) GetPublishPrompt() (string, error) {
	url := fmt.Sprintf("%s/api/prompts/publish", c.BaseURL)
	resp, err := c.HTTPClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch publish prompt: %w", err)
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

// CreateApp publishes a new app (requires authentication)
func (c *Client) CreateApp(req CreateAppRequest) (*App, error) {
	url := fmt.Sprintf("%s/api/kiosk", c.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.doAuthenticatedRequest(httpReq)
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

// UpdateApp updates an existing app (requires authentication)
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

	resp, err := c.doAuthenticatedRequest(reqHTTP)
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

// DeleteApp removes an app (requires authentication)
func (c *Client) DeleteApp(id string) error {
	url := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, id)

	reqHTTP, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doAuthenticatedRequest(reqHTTP)
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

// RefreshApp triggers a refresh of the app's Kiosk.md from the repository (requires authentication)
func (c *Client) RefreshApp(id string) error {
	url := fmt.Sprintf("%s/api/kiosk/%s/refresh", c.BaseURL, id)

	reqHTTP, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	reqHTTP.Header.Set("Content-Type", "application/json")

	resp, err := c.doAuthenticatedRequest(reqHTTP)
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
