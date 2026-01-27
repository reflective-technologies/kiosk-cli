package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// HelpModel is the model for the help view
type HelpModel struct {
	width  int
	height int
	keys   tui.KeyMap
}

// NewHelpModel creates a new help model
func NewHelpModel() HelpModel {
	return HelpModel{
		keys: tui.DefaultKeyMap(),
	}
}

// SetSize updates the view dimensions
func (m *HelpModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the help model
func (m *HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help view
func (m *HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		}
	}

	return m, nil
}

// View renders the help view
func (m *HelpModel) View() string {
	var b strings.Builder

	titleStyle := styles.Title.Copy().MarginBottom(1)
	b.WriteString(titleStyle.Render("Help"))
	b.WriteString("\n\n")

	// Keyboard shortcuts section
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Foreground)
	b.WriteString(sectionStyle.Render("Keyboard Shortcuts"))
	b.WriteString("\n\n")

	keyStyle := lipgloss.NewStyle().Foreground(styles.Secondary).Width(15)
	descStyle := lipgloss.NewStyle().Foreground(styles.Muted)

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"↑/k", "Move up"},
		{"↓/j", "Move down"},
		{"enter", "Select / Confirm"},
		{"esc", "Go back"},
		{"q", "Quit (from home)"},
		{"/", "Filter list"},
		{"?", "Toggle help"},
	}

	for _, s := range shortcuts {
		b.WriteString("  ")
		b.WriteString(keyStyle.Render(s.key))
		b.WriteString(descStyle.Render(s.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Commands section
	b.WriteString(sectionStyle.Render("CLI Commands"))
	b.WriteString("\n\n")

	commands := []struct {
		cmd  string
		desc string
	}{
		{"kiosk run <app>", "Run an app (installs if needed)"},
		{"kiosk ls", "List installed apps"},
		{"kiosk rm <app>", "Remove an installed app"},
		{"kiosk login", "Authenticate with GitHub"},
		{"kiosk audit", "Security audit current directory"},
	}

	cmdStyle := lipgloss.NewStyle().Foreground(styles.Secondary).Width(20)

	for _, c := range commands {
		b.WriteString("  ")
		b.WriteString(cmdStyle.Render(c.cmd))
		b.WriteString(descStyle.Render(c.desc))
		b.WriteString("\n")
	}

	b.WriteString("\n\n")

	// Links
	b.WriteString(sectionStyle.Render("Links"))
	b.WriteString("\n\n")

	linkStyle := lipgloss.NewStyle().Foreground(styles.Primary).Underline(true)
	b.WriteString("  Documentation: ")
	b.WriteString(linkStyle.Render("https://kiosk.dev/docs"))
	b.WriteString("\n")
	b.WriteString("  Report Issues: ")
	b.WriteString(linkStyle.Render("https://github.com/kiosk-dev/cli/issues"))
	b.WriteString("\n\n")

	// Help footer
	b.WriteString(styles.HelpStyle.Render("esc go back"))

	return b.String()
}
