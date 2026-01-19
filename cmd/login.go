package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
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

	// Load config to get API URL
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create device flow handler pointing to Kiosk API
	flow := auth.NewDeviceFlow(cfg.APIUrl)

	fmt.Println("Initiating GitHub authentication...")

	deviceCode, err := flow.RequestDeviceCode()
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

	// Poll for auth completion
	authResp, err := flow.PollForAuth(deviceCode.DeviceCode, deviceCode.Interval)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save credentials with user info
	creds := &auth.Credentials{
		AccessToken: authResp.AccessToken,
		TokenType:   authResp.TokenType,
		Scope:       authResp.Scope,
		CreatedAt:   time.Now(),
	}

	// Copy user info if available
	if authResp.User != nil {
		creds.User = &auth.UserInfo{
			ID:        authResp.User.ID,
			Username:  authResp.User.Username,
			Name:      authResp.User.Name,
			Email:     authResp.User.Email,
			AvatarURL: authResp.User.AvatarURL,
		}
	}

	if err := auth.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println()
	if creds.User != nil && creds.User.Username != "" {
		fmt.Printf("Successfully authenticated as %s!\n", creds.User.Username)
	} else {
		fmt.Println("Successfully authenticated!")
	}
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
