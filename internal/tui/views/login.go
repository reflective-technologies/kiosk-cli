package views

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// LoginState represents the current state of the login flow
type LoginState int

const (
	LoginStateInitial LoginState = iota
	LoginStateRequestingCode
	LoginStateWaitingForAuth
	LoginStateSuccess
	LoginStateError
)

// LoginModel is the model for the login view
type LoginModel struct {
	width           int
	height          int
	keys            tui.KeyMap
	state           LoginState
	spinner         spinner.Model
	userCode        string
	verificationURI string
	deviceCode      string
	interval        int
	error           error
	user            *auth.UserInfo
	flow            *auth.DeviceFlow
	pollTicker      *time.Ticker
}

// NewLoginModel creates a new login model
func NewLoginModel() LoginModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return LoginModel{
		keys:    tui.DefaultKeyMap(),
		state:   LoginStateInitial,
		spinner: s,
	}
}

// SetSize updates the view dimensions
func (m *LoginModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the login model
func (m *LoginModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.requestDeviceCode,
	)
}

func (m *LoginModel) requestDeviceCode() tea.Msg {
	cfg, err := config.Load()
	if err != nil {
		return tui.ErrorMsg{Err: fmt.Errorf("failed to load config: %w", err)}
	}

	flow := auth.NewDeviceFlow(cfg.APIUrl)
	deviceCode, err := flow.RequestDeviceCode()
	if err != nil {
		return tui.ErrorMsg{Err: fmt.Errorf("failed to initiate login: %w", err)}
	}

	return tui.LoginStartedMsg{
		DeviceCode:      deviceCode.DeviceCode,
		UserCode:        deviceCode.UserCode,
		VerificationURI: deviceCode.VerificationURI,
	}
}

// pollForAuth is a command that polls for auth completion
func (m *LoginModel) pollForAuth() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return tui.LoginCompleteMsg{Err: err}
		}

		flow := auth.NewDeviceFlow(cfg.APIUrl)
		authResp, err := flow.PollForAuth(m.deviceCode, m.interval, auth.DefaultPollTimeout)
		if err != nil {
			return tui.LoginCompleteMsg{Err: err}
		}

		// Save credentials
		creds := &auth.Credentials{
			AccessToken: authResp.AccessToken,
			TokenType:   authResp.TokenType,
			Scope:       authResp.Scope,
			CreatedAt:   time.Now(),
		}

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
			return tui.LoginCompleteMsg{Err: err}
		}

		return tui.LoginCompleteMsg{User: creds.User}
	}
}

// Update handles messages for the login view
func (m *LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		case key.Matches(msg, m.keys.Enter):
			if m.state == LoginStateSuccess || m.state == LoginStateError {
				return m, func() tea.Msg { return tui.GoBackMsg{} }
			}
			// If waiting for auth, try to open browser again
			if m.state == LoginStateWaitingForAuth && m.verificationURI != "" {
				openBrowser(m.verificationURI)
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case tui.LoginStartedMsg:
		m.state = LoginStateWaitingForAuth
		m.deviceCode = msg.DeviceCode
		m.userCode = msg.UserCode
		m.verificationURI = msg.VerificationURI
		m.interval = 5 // Default interval

		// Try to open browser
		openBrowser(m.verificationURI)

		// Start polling for auth completion
		cmds = append(cmds, m.pollForAuth())

	case tui.LoginCompleteMsg:
		if msg.Err != nil {
			m.state = LoginStateError
			m.error = msg.Err
		} else {
			m.state = LoginStateSuccess
			m.user = msg.User
		}

	case tui.ErrorMsg:
		m.state = LoginStateError
		m.error = msg.Err
	}

	return m, tea.Batch(cmds...)
}

// View renders the login view
func (m *LoginModel) View() string {
	var b strings.Builder

	titleStyle := styles.Title.Copy().MarginBottom(1)
	b.WriteString(titleStyle.Render("GitHub Authentication"))
	b.WriteString("\n\n")

	switch m.state {
	case LoginStateInitial, LoginStateRequestingCode:
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(styles.MutedStyle.Render("Initiating GitHub authentication..."))

	case LoginStateWaitingForAuth:
		b.WriteString(m.waitingView())

	case LoginStateSuccess:
		b.WriteString(m.successView())

	case LoginStateError:
		b.WriteString(m.errorView())
	}

	b.WriteString("\n\n")
	helpStyle := styles.HelpStyle
	if m.state == LoginStateSuccess || m.state == LoginStateError {
		b.WriteString(helpStyle.Render("Press enter or esc to continue"))
	} else {
		b.WriteString(helpStyle.Render("Press esc to cancel"))
	}

	return b.String()
}

func (m LoginModel) waitingView() string {
	var b strings.Builder

	// Instructions box
	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(1, 2).
		MarginBottom(1)

	var instructions strings.Builder
	instructions.WriteString(styles.Bold.Render("1."))
	instructions.WriteString(" Visit: ")
	instructions.WriteString(styles.Highlight.Render(m.verificationURI))
	instructions.WriteString("\n\n")
	instructions.WriteString(styles.Bold.Render("2."))
	instructions.WriteString(" Enter code: ")
	instructions.WriteString(styles.Code.Render(m.userCode))

	b.WriteString(boxStyle.Render(instructions.String()))
	b.WriteString("\n\n")

	// Status
	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(styles.MutedStyle.Render("Waiting for authorization..."))
	b.WriteString("\n\n")

	// Hint
	b.WriteString(styles.MutedStyle.Render("(Press enter to open browser again)"))

	return b.String()
}

func (m LoginModel) successView() string {
	var b strings.Builder

	successIcon := styles.SuccessStyle.Render("✓")
	b.WriteString(successIcon)
	b.WriteString(" ")
	b.WriteString(styles.SuccessStyle.Render("Successfully authenticated!"))
	b.WriteString("\n\n")

	if m.user != nil && m.user.Username != "" {
		b.WriteString("  Logged in as: ")
		b.WriteString(styles.Bold.Render(m.user.Username))
		if m.user.Name != "" {
			b.WriteString(" (")
			b.WriteString(m.user.Name)
			b.WriteString(")")
		}
	}

	return b.String()
}

func (m LoginModel) errorView() string {
	var b strings.Builder

	errorIcon := styles.ErrorStyle.Render("✗")
	b.WriteString(errorIcon)
	b.WriteString(" ")
	b.WriteString(styles.ErrorStyle.Render("Authentication failed"))
	b.WriteString("\n\n")

	if m.error != nil {
		b.WriteString("  ")
		b.WriteString(styles.MutedStyle.Render(m.error.Error()))
	}

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
