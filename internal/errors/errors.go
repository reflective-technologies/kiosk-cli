// Package errors provides structured error types for the kiosk CLI.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Kiosk API.
// It parses the JSON error response and provides structured access to the error details.
type APIError struct {
	StatusCode int    // HTTP status code
	Message    string // Human-readable error message
	RawBody    string // Raw response body for debugging
}

func (e *APIError) Error() string {
	return e.Message
}

// IsNotFound returns true if this is a 404 error.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsUnauthorized returns true if this is a 401 error.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if this is a 403 error.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsBadRequest returns true if this is a 400 error.
func (e *APIError) IsBadRequest() bool {
	return e.StatusCode == http.StatusBadRequest
}

// IsServerError returns true if this is a 5xx error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// NewAPIError creates an APIError from an HTTP response.
// It attempts to parse the JSON error response, falling back to the raw body.
func NewAPIError(statusCode int, body []byte) *APIError {
	message := parseErrorMessage(body)
	if message == "" {
		message = fmt.Sprintf("API request failed with status %d", statusCode)
	}

	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		RawBody:    string(body),
	}
}

// parseErrorMessage attempts to extract a user-friendly error message from the API response.
// It handles multiple common response formats.
func parseErrorMessage(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	// Try parsing as {"error": "message"}
	var errResp struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
		return errResp.Error
	}

	// Try parsing as {"message": "..."}
	var msgResp struct {
		Message string `json:"message"`
	}
	if json.Unmarshal(body, &msgResp) == nil && msgResp.Message != "" {
		return msgResp.Message
	}

	// Try parsing as {"errors": ["..."]}
	var errsResp struct {
		Errors []string `json:"errors"`
	}
	if json.Unmarshal(body, &errsResp) == nil && len(errsResp.Errors) > 0 {
		return errsResp.Errors[0]
	}

	// Fall back to raw body if it's reasonably short and looks like text
	if len(body) < 200 && isPrintable(body) {
		return string(body)
	}

	return ""
}

// isPrintable returns true if the byte slice contains only printable ASCII characters.
func isPrintable(b []byte) bool {
	for _, c := range b {
		if c < 32 || c > 126 {
			if c != '\n' && c != '\r' && c != '\t' {
				return false
			}
		}
	}
	return true
}

// AuthError represents an authentication-related error.
type AuthError struct {
	Message    string
	Suggestion string
}

func (e *AuthError) Error() string {
	return e.Message
}

// NewAuthError creates a new authentication error.
func NewAuthError(message string) *AuthError {
	return &AuthError{
		Message:    message,
		Suggestion: "Run 'kiosk login' to authenticate",
	}
}

// NetworkError represents a network-related error (connection failures, timeouts, etc.)
type NetworkError struct {
	Message string
	Cause   error
}

func (e *NetworkError) Error() string {
	return e.Message
}

func (e *NetworkError) Unwrap() error {
	return e.Cause
}

// NewNetworkError creates a new network error.
func NewNetworkError(message string, cause error) *NetworkError {
	return &NetworkError{
		Message: message,
		Cause:   cause,
	}
}

// Helper functions for checking error types

// IsAPIError checks if the error is an APIError and returns it.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}

// IsAuthError checks if the error is an AuthError and returns it.
func IsAuthError(err error) (*AuthError, bool) {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr, true
	}
	return nil, false
}

// IsNetworkError checks if the error is a NetworkError and returns it.
func IsNetworkError(err error) (*NetworkError, bool) {
	var netErr *NetworkError
	if errors.As(err, &netErr) {
		return netErr, true
	}
	return nil, false
}
