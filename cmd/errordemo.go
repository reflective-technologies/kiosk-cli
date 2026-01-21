package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/reflective-technologies/kiosk-cli/internal/errors"
	"github.com/spf13/cobra"
)

var errorDemoCmd = &cobra.Command{
	Use:    "error-demo",
	Short:  "Display example error messages for testing",
	Long:   `Shows how different error types are displayed in the terminal. Useful for testing and development.`,
	Hidden: true, // Hide from main help output
	RunE: func(cmd *cobra.Command, args []string) error {
		errorType, _ := cmd.Flags().GetString("type")

		// If a specific type is requested, show only that
		if errorType != "" {
			return showErrorByType(errorType)
		}

		// Otherwise show all errors
		showAllErrors()
		return nil
	},
}

func showAllErrors() {
	fmt.Fprintln(os.Stderr, "=== Kiosk CLI Error Display Demo ===")
	fmt.Fprintln(os.Stderr)

	// API Errors
	fmt.Fprintln(os.Stderr, "--- API Errors ---")
	fmt.Fprintln(os.Stderr)

	showError("Private Repository", errors.NewAPIError(http.StatusBadRequest, []byte(`{"error":"Repository must be public to publish to Kiosk"}`)))
	showError("App Not Found (404)", errors.NewAPIError(http.StatusNotFound, []byte(`{"error":"App not found"}`)))
	showError("Unauthorized (401)", errors.NewAPIError(http.StatusUnauthorized, []byte(`{"error":"Invalid or expired token"}`)))
	showError("Forbidden (403)", errors.NewAPIError(http.StatusForbidden, []byte(`{"error":"You don't have permission to modify this app"}`)))
	showError("Invalid Git URL", errors.NewAPIError(http.StatusBadRequest, []byte(`{"error":"Invalid Git URL provided"}`)))
	showError("Missing Name", errors.NewAPIError(http.StatusBadRequest, []byte(`{"error":"Name is required"}`)))
	showError("Missing Description", errors.NewAPIError(http.StatusBadRequest, []byte(`{"error":"Description is required"}`)))
	showError("Missing KIOSK.md", errors.NewAPIError(http.StatusBadRequest, []byte(`{"error":"KIOSK.md file not found in repository"}`)))
	showError("Rate Limited", errors.NewAPIError(http.StatusTooManyRequests, []byte(`{"error":"Rate limit exceeded. Too many requests"}`)))
	showError("Server Error (500)", errors.NewAPIError(http.StatusInternalServerError, []byte(`{"error":"Internal server error"}`)))
	showError("Server Error (503)", errors.NewAPIError(http.StatusServiceUnavailable, []byte(`{"error":"Service temporarily unavailable"}`)))

	// Auth Errors
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "--- Authentication Errors ---")
	fmt.Fprintln(os.Stderr)

	showError("Not Logged In", errors.NewAuthError("Authentication required"))
	showError("Token Expired", errors.NewAuthError("Your session has expired"))

	// Network Errors
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "--- Network Errors ---")
	fmt.Fprintln(os.Stderr)

	showError("DNS Failure", errors.NewNetworkError("Could not reach the Kiosk API (DNS lookup failed)", nil))
	showError("Connection Refused", errors.NewNetworkError("Could not connect to the Kiosk API (connection refused)", nil))
	showError("Timeout", errors.NewNetworkError("Request to Kiosk API timed out", nil))
	showError("SSL Error", errors.NewNetworkError("SSL/TLS certificate error when connecting to Kiosk API", nil))

	// Generic Errors
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "--- Generic Errors ---")
	fmt.Fprintln(os.Stderr)

	showError("Generic Error", fmt.Errorf("something went wrong"))
	showError("Login Hint", fmt.Errorf("not logged in"))
	showError("Not Found Hint", fmt.Errorf("app \"my-app\" not found"))
}

func showError(name string, err error) {
	fmt.Fprintf(os.Stderr, "[%s]\n", name)
	errors.PrintError(err)
	fmt.Fprintln(os.Stderr)
}

func showErrorByType(errorType string) error {
	switch errorType {
	case "private-repo":
		errors.PrintError(errors.NewAPIError(http.StatusBadRequest, []byte(`{"error":"Repository must be public to publish to Kiosk"}`)))
	case "not-found":
		errors.PrintError(errors.NewAPIError(http.StatusNotFound, []byte(`{"error":"App not found"}`)))
	case "unauthorized":
		errors.PrintError(errors.NewAPIError(http.StatusUnauthorized, []byte(`{"error":"Invalid or expired token"}`)))
	case "forbidden":
		errors.PrintError(errors.NewAPIError(http.StatusForbidden, []byte(`{"error":"You don't have permission to modify this app"}`)))
	case "server-error":
		errors.PrintError(errors.NewAPIError(http.StatusInternalServerError, []byte(`{"error":"Internal server error"}`)))
	case "rate-limit":
		errors.PrintError(errors.NewAPIError(http.StatusTooManyRequests, []byte(`{"error":"Rate limit exceeded"}`)))
	case "auth":
		errors.PrintError(errors.NewAuthError("Authentication required"))
	case "network":
		errors.PrintError(errors.NewNetworkError("Could not reach the Kiosk API", nil))
	case "timeout":
		errors.PrintError(errors.NewNetworkError("Request to Kiosk API timed out", nil))
	default:
		return fmt.Errorf("unknown error type: %s\n\nAvailable types: private-repo, not-found, unauthorized, forbidden, server-error, rate-limit, auth, network, timeout", errorType)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(errorDemoCmd)
	errorDemoCmd.Flags().StringP("type", "t", "", "Show specific error type (private-repo, not-found, unauthorized, forbidden, server-error, rate-limit, auth, network, timeout)")
}
