package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// GitHub OAuth endpoints
	deviceCodeURL = "https://github.com/login/device/code"
	tokenURL      = "https://github.com/login/oauth/access_token"
)

// DeviceCodeResponse represents the response from requesting a device code
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// TokenResponse represents a successful token response
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// TokenErrorResponse represents an error response when polling for token
type TokenErrorResponse struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Interval         int    `json:"interval,omitempty"`
}

// DeviceFlow handles the OAuth device flow authentication
type DeviceFlow struct {
	ClientID   string
	HTTPClient *http.Client
}

// NewDeviceFlow creates a new DeviceFlow instance
func NewDeviceFlow(clientID string) *DeviceFlow {
	return &DeviceFlow{
		ClientID: clientID,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// RequestDeviceCode initiates the device flow by requesting a device code
func (d *DeviceFlow) RequestDeviceCode(scopes []string) (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", d.ClientID)
	if len(scopes) > 0 {
		data.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequest("POST", deviceCodeURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

// PollForToken polls GitHub for the access token until the user authorizes or an error occurs
func (d *DeviceFlow) PollForToken(deviceCode string, interval int) (*TokenResponse, error) {
	pollInterval := time.Duration(interval) * time.Second

	for {
		token, err := d.requestToken(deviceCode)
		if err == nil {
			return token, nil
		}

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
}

// PollError represents a polling error from the token endpoint
type PollError struct {
	Code        string
	Description string
}

func (e *PollError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Description)
}

func (d *DeviceFlow) requestToken(deviceCode string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", d.ClientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	// GitHub returns 200 even for errors during polling, we need to check the response body
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

	// Parse as successful token response
	accessToken, ok := rawResponse["access_token"].(string)
	if !ok || accessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	return &TokenResponse{
		AccessToken: accessToken,
		TokenType:   rawResponse["token_type"].(string),
		Scope:       rawResponse["scope"].(string),
	}, nil
}
