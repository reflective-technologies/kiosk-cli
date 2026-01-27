package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	apierrors "github.com/reflective-technologies/kiosk-cli/internal/errors"
)

// Client is a kiosk API client
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	token      string // GitHub access token for authenticated requests
}

// Creator represents the app creator from the API
type Creator struct {
	ID        string `json:"id"`
	GithubID  int    `json:"githubId"`
	Username  string `json:"username"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

// App represents an app from the API
type App struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	GitUrl       string   `json:"gitUrl"`
	Branch       string   `json:"branch,omitempty"`
	KioskMd      string   `json:"kioskMd,omitempty"`
	Creator      *Creator `json:"creator,omitempty"`
	InstallCount int      `json:"installCount,omitempty"`
	CreatedAt    string   `json:"createdAt,omitempty"`
	UpdatedAt    string   `json:"updatedAt,omitempty"`
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
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, wrapNetworkError(err)
	}
	return resp, nil
}

// wrapNetworkError wraps network-related errors with user-friendly messages
func wrapNetworkError(err error) error {
	if err == nil {
		return nil
	}

	// Check for URL errors (DNS, connection refused, etc.)
	if urlErr, ok := err.(*url.Error); ok {
		switch {
		case strings.Contains(urlErr.Error(), "no such host"):
			return apierrors.NewNetworkError("Could not reach the Kiosk API (DNS lookup failed)", err)
		case strings.Contains(urlErr.Error(), "connection refused"):
			return apierrors.NewNetworkError("Could not connect to the Kiosk API (connection refused)", err)
		case strings.Contains(urlErr.Error(), "timeout"):
			return apierrors.NewNetworkError("Request to Kiosk API timed out", err)
		case strings.Contains(urlErr.Error(), "certificate"):
			return apierrors.NewNetworkError("SSL/TLS certificate error when connecting to Kiosk API", err)
		}
	}

	return apierrors.NewNetworkError("Network error while connecting to Kiosk API", err)
}

// handleAPIError creates an appropriate error from an HTTP response
func handleAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	return apierrors.NewAPIError(resp.StatusCode, body)
}

// doAuthenticatedRequest performs an HTTP request that requires authentication
// Returns an error if no token is set
func (c *Client) doAuthenticatedRequest(req *http.Request) (*http.Response, error) {
	if c.token == "" {
		return nil, apierrors.NewAuthError("Authentication required")
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, wrapNetworkError(err)
	}
	return resp, nil
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

	reqURL := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, appId)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
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

	reqURL := fmt.Sprintf("%s/api/kiosk/%s/install", c.BaseURL, appId)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", handleAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// GetInitPrompt fetches the KIOSK.md creation prompt
func (c *Client) GetInitPrompt() (string, error) {
	reqURL := fmt.Sprintf("%s/api/prompts/init", c.BaseURL)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", handleAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// GetPublishPrompt fetches the publish prompt
func (c *Client) GetPublishPrompt() (string, error) {
	reqURL := fmt.Sprintf("%s/api/prompts/publish", c.BaseURL)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", handleAPIError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(body), nil
}

// ListApps fetches all published apps
func (c *Client) ListApps() ([]App, error) {
	reqURL := fmt.Sprintf("%s/api/kiosk", c.BaseURL)
	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	var apps []App
	if err := json.NewDecoder(resp.Body).Decode(&apps); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return apps, nil
}

// CreateApp publishes a new app (requires authentication)
func (c *Client) CreateApp(req CreateAppRequest) (*App, error) {
	reqURL := fmt.Sprintf("%s/api/kiosk", c.BaseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.doAuthenticatedRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	var app App
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

// UpdateApp updates an existing app (requires authentication)
func (c *Client) UpdateApp(id string, req UpdateAppRequest) (*App, error) {
	reqURL := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, id)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.doAuthenticatedRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, handleAPIError(resp)
	}

	var app App
	if err := json.NewDecoder(resp.Body).Decode(&app); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &app, nil
}

// DeleteApp removes an app (requires authentication)
func (c *Client) DeleteApp(id string) error {
	reqURL := fmt.Sprintf("%s/api/kiosk/%s", c.BaseURL, id)

	httpReq, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.doAuthenticatedRequest(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return handleAPIError(resp)
	}

	return nil
}

// RefreshApp triggers a refresh of the app's Kiosk.md from the repository (requires authentication)
func (c *Client) RefreshApp(id string) error {
	reqURL := fmt.Sprintf("%s/api/kiosk/%s/refresh", c.BaseURL, id)

	httpReq, err := http.NewRequest(http.MethodPost, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.doAuthenticatedRequest(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return handleAPIError(resp)
	}

	return nil
}
