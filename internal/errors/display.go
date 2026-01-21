package errors

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// DevMode enables verbose error logging when true.
// Set this from the cmd package based on the Version variable.
var DevMode bool

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorDim    = "\033[2m"
	colorBold   = "\033[1m"
)

// useColor determines if we should use ANSI colors based on terminal capabilities.
func useColor() bool {
	// Check if stdout is a terminal
	if !term.IsTerminal(int(os.Stderr.Fd())) {
		return false
	}

	// Check for NO_COLOR environment variable
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Check TERM for dumb terminal
	if os.Getenv("TERM") == "dumb" {
		return false
	}

	return true
}

// color wraps text in ANSI color codes if colors are enabled.
func color(c, text string) string {
	if !useColor() {
		return text
	}
	return c + text + colorReset
}

// FormatError formats an error for display to the user with colors and helpful suggestions.
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var sb strings.Builder

	// Check for specific error types and format accordingly
	if apiErr, ok := IsAPIError(err); ok {
		formatAPIError(&sb, apiErr)
	} else if authErr, ok := IsAuthError(err); ok {
		formatAuthError(&sb, authErr)
	} else if netErr, ok := IsNetworkError(err); ok {
		formatNetworkError(&sb, netErr)
	} else {
		formatGenericError(&sb, err)
	}

	return sb.String()
}

func formatAPIError(sb *strings.Builder, err *APIError) {
	sb.WriteString(color(colorRed+colorBold, "Error: "))
	sb.WriteString(getNarrativeMessage(err))
	sb.WriteString("\n")

	if DevMode {
		sb.WriteString("\n")
		sb.WriteString(color(colorDim, "--- Debug Info ---\n"))
		sb.WriteString(color(colorDim, fmt.Sprintf("Status Code: %d\n", err.StatusCode)))
		sb.WriteString(color(colorDim, fmt.Sprintf("Raw Message: %s\n", err.Message)))
		if err.RawBody != "" && err.RawBody != err.Message {
			sb.WriteString(color(colorDim, fmt.Sprintf("Raw Body: %s\n", err.RawBody)))
		}
	}
}

func formatAuthError(sb *strings.Builder, err *AuthError) {
	sb.WriteString(color(colorRed+colorBold, "Error: "))
	sb.WriteString(getAuthNarrativeMessage(err))
	sb.WriteString("\n")

	if DevMode {
		sb.WriteString("\n")
		sb.WriteString(color(colorDim, "--- Debug Info ---\n"))
		sb.WriteString(color(colorDim, fmt.Sprintf("Raw Message: %s\n", err.Message)))
	}
}

func formatNetworkError(sb *strings.Builder, err *NetworkError) {
	sb.WriteString(color(colorRed+colorBold, "Error: "))
	sb.WriteString(getNetworkNarrativeMessage(err))
	sb.WriteString("\n")

	if DevMode {
		sb.WriteString("\n")
		sb.WriteString(color(colorDim, "--- Debug Info ---\n"))
		sb.WriteString(color(colorDim, fmt.Sprintf("Raw Message: %s\n", err.Message)))
		if err.Cause != nil {
			sb.WriteString(color(colorDim, fmt.Sprintf("Cause: %v\n", err.Cause)))
		}
	}
}

func formatGenericError(sb *strings.Builder, err error) {
	sb.WriteString(color(colorRed+colorBold, "Error: "))
	sb.WriteString(getGenericNarrativeMessage(err))
	sb.WriteString("\n")

	if DevMode {
		sb.WriteString("\n")
		sb.WriteString(color(colorDim, "--- Debug Info ---\n"))
		sb.WriteString(color(colorDim, fmt.Sprintf("Error Type: %T\n", err)))
		sb.WriteString(color(colorDim, fmt.Sprintf("Raw Error: %v\n", err)))
	}
}

// getNarrativeMessage returns a user-friendly narrative message for API errors.
func getNarrativeMessage(err *APIError) string {
	msg := strings.ToLower(err.Message)

	switch {
	// Authentication/authorization errors - check status codes first
	case err.IsUnauthorized():
		return "Your session has expired or is invalid. Please run 'kiosk login' to authenticate."
	case err.IsForbidden():
		return "You don't have permission to perform this action. Please verify that you own this repository."

	// Not found errors
	case err.IsNotFound():
		return "The requested app could not be found. Please run 'kiosk list' to see your available apps."

	// Server errors
	case err.IsServerError():
		return "The Kiosk server encountered an error. Please try again later."

	// Repository visibility errors (400 Bad Request from server)
	// Only match on 400s to avoid false positives from other errors containing "public" or "private"
	case err.IsBadRequest() && (strings.Contains(msg, "public") && strings.Contains(msg, "repo")):
		return "You are attempting to publish a private repository. Please make your repository public on GitHub to continue."

	// Git URL errors
	case strings.Contains(msg, "git url") || strings.Contains(msg, "invalid url"):
		return "The Git URL provided is invalid. Please run this command from a git repository with a valid GitHub remote."

	// Validation errors
	case strings.Contains(msg, "name") && (strings.Contains(msg, "required") || strings.Contains(msg, "missing")):
		return "A name is required for your app. Please provide one in your KIOSK.md file."

	case strings.Contains(msg, "description") && (strings.Contains(msg, "required") || strings.Contains(msg, "missing")):
		return "A description is required for your app. Please provide one in your KIOSK.md file."

	// Rate limiting
	case strings.Contains(msg, "rate limit") || strings.Contains(msg, "too many requests"):
		return "You've made too many requests. Please wait a moment before trying again."

	default:
		return err.Message
	}
}

// getAuthNarrativeMessage returns a user-friendly narrative message for auth errors.
func getAuthNarrativeMessage(err *AuthError) string {
	msg := strings.ToLower(err.Message)

	switch {
	case strings.Contains(msg, "expired"):
		return "Your session has expired. Please run 'kiosk login' to authenticate again."
	case strings.Contains(msg, "required"):
		return "This action requires authentication. Please run 'kiosk login' to continue."
	default:
		return "You are not authenticated. Please run 'kiosk login' to continue."
	}
}

// getNetworkNarrativeMessage returns a user-friendly narrative message for network errors.
func getNetworkNarrativeMessage(err *NetworkError) string {
	msg := strings.ToLower(err.Message)

	switch {
	case strings.Contains(msg, "dns"):
		return "Unable to resolve the Kiosk API server. Please check your internet connection and try again."
	case strings.Contains(msg, "refused"):
		return "The connection to the Kiosk API was refused. Please check your internet connection and try again."
	case strings.Contains(msg, "timeout"):
		return "The request to the Kiosk API timed out. Please check your internet connection and try again."
	case strings.Contains(msg, "ssl") || strings.Contains(msg, "tls") || strings.Contains(msg, "certificate"):
		return "There was a security certificate error connecting to the Kiosk API. Please check your network configuration."
	default:
		return "Unable to connect to the Kiosk API. Please check your internet connection and try again."
	}
}

// getGenericNarrativeMessage returns a user-friendly narrative message for generic errors.
func getGenericNarrativeMessage(err error) string {
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, "not logged in") || strings.Contains(errStr, "authentication required"):
		return "You are not logged in. Please run 'kiosk login' to authenticate."
	case strings.Contains(errStr, "not found"):
		return "The requested resource was not found. Please run 'kiosk list' to see available apps."
	case strings.Contains(errStr, "connection refused") || strings.Contains(errStr, "no such host"):
		return "Unable to connect to the server. Please check your internet connection and try again."
	default:
		return err.Error()
	}
}

// PrintError prints a formatted error to stderr.
func PrintError(err error) {
	if err == nil {
		return
	}
	fmt.Fprint(os.Stderr, FormatError(err))
}
