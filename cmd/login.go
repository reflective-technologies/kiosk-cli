package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using the device flow.

This will open your browser to authorize the Kiosk CLI with your GitHub account.
The CLI will wait for you to complete the authorization in your browser.`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Check if already logged in
	if auth.IsLoggedIn() {
		fmt.Println("You are already logged in.")
		fmt.Println("Run 'kiosk logout' first if you want to switch accounts.")
		return nil
	}

	// Create device flow handler
	flow := auth.NewDeviceFlow(GitHubClientID)

	// Request device code - read:user for identity, public_repo for publishing
	scopes := []string{"read:user", "public_repo"}

	fmt.Println("Initiating GitHub authentication...")

	deviceCode, err := flow.RequestDeviceCode(scopes)
	if err != nil {
		return fmt.Errorf("failed to initiate login: %w", err)
	}

	// Display instructions to user
	fmt.Println()
	fmt.Printf("Please visit: %s\n", deviceCode.VerificationURI)
	fmt.Printf("and enter code: %s\n", deviceCode.UserCode)
	fmt.Println()

	// Try to open browser automatically
	if err := openBrowser(deviceCode.VerificationURI); err == nil {
		fmt.Println("(Browser opened automatically)")
	}

	fmt.Println("Waiting for authorization...")

	// Poll for token
	token, err := flow.PollForToken(deviceCode.DeviceCode, deviceCode.Interval)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials
	creds := &auth.Credentials{
		AccessToken: token.AccessToken,
		TokenType:   token.TokenType,
		Scope:       token.Scope,
		CreatedAt:   time.Now(),
	}

	if err := auth.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println()
	fmt.Println("Successfully authenticated!")
	return nil
}

// openBrowser opens the specified URL in the default browser
func openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform")
	}

	return cmd.Start()
}
