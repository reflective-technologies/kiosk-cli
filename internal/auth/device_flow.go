package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// DeviceCodeResponse represents the response from requesting a device code
// Field names match the Kiosk API response (camelCase)
type DeviceCodeResponse struct {
	DeviceCode      string `json:"deviceCode"`
	UserCode        string `json:"userCode"`
	VerificationURI string `json:"verificationUri"`
	ExpiresIn       int    `json:"expiresIn"`
	Interval        int    `json:"interval"`
}

// User represents the authenticated user info returned by the API
// Field names match the Kiosk API response (camelCase)
type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatarUrl"`
}

// AuthResponse represents the response from polling for auth completion
type AuthResponse struct {
	Status      string `json:"status"` // "pending" or "complete"
	AccessToken string `json:"access_token,omitempty"`
	TokenType   string `json:"token_type,omitempty"`
	Scope       string `json:"scope,omitempty"`
	User        *User  `json:"user,omitempty"`
}

// TokenErrorResponse represents an error response when polling for token
type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Interval         int    `json:"interval,omitempty"`
}

// DeviceFlow handles the OAuth device flow authentication via Kiosk API
type DeviceFlow struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewDeviceFlow creates a new DeviceFlow instance
func NewDeviceFlow(baseURL string) *DeviceFlow {
	return &DeviceFlow{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RequestDeviceCode initiates the device flow by requesting a device code from Kiosk API
func (d *DeviceFlow) RequestDeviceCode() (*DeviceCodeResponse, error) {
	url := fmt.Sprintf("%s/api/auth/github/device", d.BaseURL)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp TokenErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("%s: %s", errResp.Error, errResp.ErrorDescription)
	}

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// PollForAuth polls Kiosk API for auth completion until the user authorizes or an error occurs
func (d *DeviceFlow) PollForAuth(deviceCode string, interval int) (*AuthResponse, error) {
	pollInterval := time.Duration(interval) * time.Second

	for {
		authResp, err := d.checkAuth(deviceCode)
		if err != nil {
			// Check if it's a polling error we should handle
			if pollErr, ok := err.(*PollError); ok {
				switch pollErr.Code {
				case "authorization_pending":
					// User hasn't authorized yet, keep polling
					time.Sleep(pollInterval)
					continue
				case "slow_down":
					// We're polling too fast, increase interval
					pollInterval += 5 * time.Second
					time.Sleep(pollInterval)
					continue
				case "expired_token":
					return nil, fmt.Errorf("device code expired, please run login again")
				case "access_denied":
					return nil, fmt.Errorf("authorization denied by user")
				default:
					return nil, fmt.Errorf("%s: %s", pollErr.Code, pollErr.Description)
				}
			}
			return nil, err
		}

		// Check response status
		if authResp.Status == "pending" {
			time.Sleep(pollInterval)
			continue
		}

		if authResp.Status == "complete" {
			return authResp, nil
		}

		return nil, fmt.Errorf("unexpected auth status: %s", authResp.Status)
	}
}

// PollError represents a polling error from the token endpoint
type PollError struct {
	Code        string
	Description string
}

func (e *PollError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

func (d *DeviceFlow) checkAuth(deviceCode string) (*AuthResponse, error) {
	url := fmt.Sprintf("%s/api/auth/github/device?device_code=%s", d.BaseURL, deviceCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to check auth: %w", err)
	}
	defer resp.Body.Close()

	var rawResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if there's an error in the response
	if errCode, ok := rawResponse["error"].(string); ok {
		errDesc, _ := rawResponse["error_description"].(string)
		return nil, &PollError{
			Code:        errCode,
			Description: errDesc,
		}
	}

	// Parse the auth response
	authResp := &AuthResponse{}

	if status, ok := rawResponse["status"].(string); ok {
		authResp.Status = status
	}

	// Try camelCase first (Kiosk API), then snake_case as fallback
	if accessToken, ok := rawResponse["accessToken"].(string); ok {
		authResp.AccessToken = accessToken
	} else if accessToken, ok := rawResponse["access_token"].(string); ok {
		authResp.AccessToken = accessToken
	}

	if tokenType, ok := rawResponse["tokenType"].(string); ok {
		authResp.TokenType = tokenType
	} else if tokenType, ok := rawResponse["token_type"].(string); ok {
		authResp.TokenType = tokenType
	}

	if scope, ok := rawResponse["scope"].(string); ok {
		authResp.Scope = scope
	}

	// Parse user if present
	if userData, ok := rawResponse["user"].(map[string]interface{}); ok {
		authResp.User = &User{}
		if id, ok := userData["id"].(string); ok {
			authResp.User.ID = id
		}
		if username, ok := userData["username"].(string); ok {
			authResp.User.Username = username
		}
		if name, ok := userData["name"].(string); ok {
			authResp.User.Name = name
		}
		if email, ok := userData["email"].(string); ok {
			authResp.User.Email = email
		}
		if avatarURL, ok := userData["avatarUrl"].(string); ok {
			authResp.User.AvatarURL = avatarURL
		}
	}

	return authResp, nil
}
