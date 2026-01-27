package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
	"github.com/spf13/cobra"
)

var loginTimeout time.Duration

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub",
	Long: `Authenticate with GitHub using the device flow.

This will open your browser to authorize the Kiosk CLI with your GitHub account.
The CLI will wait for you to complete the authorization in your browser.`,
	RunE: runLogin,
}

func init() {
	loginCmd.Flags().DurationVar(&loginTimeout, "timeout", auth.DefaultPollTimeout, "timeout for waiting for authorization")
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	// Check if already logged in
	if auth.IsLoggedIn() {
		user, _ := auth.GetUser()
		fmt.Println()
		fmt.Println("  You are already logged in as " + lipgloss.NewStyle().Bold(true).Render("@"+user.Username))
		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("  Run 'kiosk logout' first if you want to switch accounts."))
		fmt.Println()
		return nil
	}

	// Load config to get API URL
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create device flow handler pointing to Kiosk API
	flow := auth.NewDeviceFlow(cfg.APIUrl)

	deviceCode, err := flow.RequestDeviceCode()
	if err != nil {
		return fmt.Errorf("failed to initiate login: %w", err)
	}

	// Try to open browser automatically
	openBrowser(deviceCode.VerificationURI)

	// Run interactive login UI
	m := newLoginModel(deviceCode, flow, loginTimeout)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	model, ok := finalModel.(*loginModel)
	if !ok {
		return nil
	}

	if model.err != nil {
		return fmt.Errorf("authentication failed: %w", model.err)
	}

	if model.authResp == nil {
		// User cancelled
		return nil
	}

	// Save credentials with user info
	creds := &auth.Credentials{
		AccessToken: model.authResp.AccessToken,
		TokenType:   model.authResp.TokenType,
		Scope:       model.authResp.Scope,
		CreatedAt:   time.Now(),
	}

	// Copy user info if available
	if model.authResp.User != nil {
		creds.User = &auth.UserInfo{
			ID:        model.authResp.User.ID,
			Username:  model.authResp.User.Username,
			Name:      model.authResp.User.Name,
			Email:     model.authResp.User.Email,
			AvatarURL: model.authResp.User.AvatarURL,
		}
	}

	if err := auth.SaveCredentials(creds); err != nil {
		return fmt.Errorf("failed to save credentials: %w", err)
	}

	fmt.Println()
	successStyle := lipgloss.NewStyle().Foreground(styles.Success)
	if creds.User != nil && creds.User.Username != "" {
		fmt.Println("  " + successStyle.Render("Successfully authenticated as @"+creds.User.Username))
	} else {
		fmt.Println("  " + successStyle.Render("Successfully authenticated!"))
	}
	fmt.Println()
	return nil
}

// loginModel is the bubbletea model for the login flow
type loginModel struct {
	deviceCode        *auth.DeviceCodeResponse
	flow              *auth.DeviceFlow
	timeout           time.Duration
	authResp          *auth.AuthResponse
	err               error
	polling           bool
	frame             int
	quitting          bool
	copiedToClipboard bool
}

func newLoginModel(deviceCode *auth.DeviceCodeResponse, flow *auth.DeviceFlow, timeout time.Duration) *loginModel {
	// Try to copy code to clipboard
	copied := false
	if err := copyToClipboard(deviceCode.UserCode); err == nil {
		copied = true
	}

	return &loginModel{
		deviceCode:        deviceCode,
		flow:              flow,
		timeout:           timeout,
		polling:           true,
		copiedToClipboard: copied,
	}
}

type pollTickMsg struct{}
type pollResultMsg struct {
	resp *auth.AuthResponse
	err  error
}
type spinnerTickMsg struct{}

func (m *loginModel) Init() tea.Cmd {
	return tea.Batch(
		m.pollForAuth(),
		m.spinnerTick(),
	)
}

func (m *loginModel) pollForAuth() tea.Cmd {
	return func() tea.Msg {
		resp, err := m.flow.PollForAuth(m.deviceCode.DeviceCode, m.deviceCode.Interval, m.timeout)
		return pollResultMsg{resp: resp, err: err}
	}
}

func (m *loginModel) spinnerTick() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
}

func (m *loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		}

	case spinnerTickMsg:
		if m.polling {
			m.frame++
			return m, m.spinnerTick()
		}

	case pollResultMsg:
		m.polling = false
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.authResp = msg.resp
		}
		return m, tea.Quit
	}

	return m, nil
}

func (m *loginModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
	b.WriteString("  ")
	b.WriteString(titleStyle.Render("GitHub Authentication"))
	b.WriteString("\n\n")

	// Show clipboard status first if copied
	if m.copiedToClipboard {
		b.WriteString("  ")
		b.WriteString("Code copied to clipboard")
		b.WriteString("\n\n")
	}

	// Instructions
	b.WriteString("  ")
	b.WriteString("Visit ")
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(styles.Secondary).Render(m.deviceCode.VerificationURI))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString("and enter this code:")
	b.WriteString("\n\n")

	// Code box - make it prominent
	codeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 3).
		MarginLeft(2)

	// Center the code with spacing between characters for readability
	code := m.deviceCode.UserCode
	spacedCode := strings.Join(strings.Split(code, ""), " ")

	b.WriteString(codeStyle.Render(spacedCode))
	b.WriteString("\n\n")

	// Status
	if m.polling {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		spinner := lipgloss.NewStyle().Foreground(styles.Primary).Render(frames[m.frame%len(frames)])
		b.WriteString("  ")
		b.WriteString(spinner)
		b.WriteString(" ")
		b.WriteString(styles.MutedStyle.Render("Waiting for authorization..."))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Help
	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render("Press esc or q to cancel"))
	b.WriteString("\n")

	return b.String()
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

// copyToClipboard copies text to the system clipboard
func copyToClipboard(text string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux":
		// Try xclip first, fall back to xsel
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (install xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("cmd", "/c", "clip")
	default:
		return fmt.Errorf("unsupported platform")
	}

	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}
