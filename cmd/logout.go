package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from GitHub",
	Long:  `Remove stored GitHub credentials from this machine.`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	if !auth.IsLoggedIn() {
		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("  You are not logged in."))
		fmt.Println()
		return nil
	}

	// Get current user info for display
	user, _ := auth.GetUser()

	// Run interactive confirmation
	m := newLogoutModel(user)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error: %w", err)
	}

	if model, ok := finalModel.(*logoutModel); ok && model.confirmed {
		if err := auth.DeleteCredentials(); err != nil {
			return fmt.Errorf("failed to logout: %w", err)
		}
		fmt.Println()
		fmt.Println(styles.MutedStyle.Render("  Successfully logged out."))
		fmt.Println()
	}

	return nil
}

// logoutModel is the bubbletea model for logout confirmation
type logoutModel struct {
	user      *auth.UserInfo
	cursor    int // 0 = Cancel, 1 = Logout
	confirmed bool
	quitting  bool
}

func newLogoutModel(user *auth.UserInfo) *logoutModel {
	return &logoutModel{
		user:   user,
		cursor: 0, // Default to Cancel for safety
	}
}

func (m *logoutModel) Init() tea.Cmd {
	return nil
}

func (m *logoutModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 1 {
				m.confirmed = true
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *logoutModel) View() string {
	var b strings.Builder

	b.WriteString("\n")

	// Title
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
	b.WriteString("  ")
	b.WriteString(titleStyle.Render("Log Out"))
	b.WriteString("\n\n")

	// Current user info
	if m.user != nil {
		b.WriteString("  ")
		b.WriteString("You are logged in as ")
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("@" + m.user.Username))
		b.WriteString("\n\n")
	}

	// Confirmation message
	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render("Are you sure you want to log out?"))
	b.WriteString("\n\n")

	// Buttons
	b.WriteString("  ")
	b.WriteString(m.renderButtons())
	b.WriteString("\n\n")

	// Help
	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render("left/right select | enter confirm | esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m *logoutModel) renderButtons() string {
	cancelStyle := lipgloss.NewStyle().Padding(0, 2)
	logoutStyle := lipgloss.NewStyle().Padding(0, 2)

	if m.cursor == 0 {
		cancelStyle = cancelStyle.
			Background(styles.Muted).
			Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		cancelStyle = cancelStyle.Foreground(styles.Muted)
	}

	if m.cursor == 1 {
		logoutStyle = logoutStyle.
			Background(styles.Error).
			Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		logoutStyle = logoutStyle.Foreground(styles.Muted)
	}

	return cancelStyle.Render("Cancel") + "  " + logoutStyle.Render("Log Out")
}
