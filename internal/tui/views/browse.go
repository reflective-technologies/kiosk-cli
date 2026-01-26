package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// BrowseModel is the model for the browse apps view
type BrowseModel struct {
	width  int
	height int
	keys   tui.KeyMap
}

// NewBrowseModel creates a new browse model
func NewBrowseModel() BrowseModel {
	return BrowseModel{
		keys: tui.DefaultKeyMap(),
	}
}

// SetSize updates the view dimensions
func (m *BrowseModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the browse model
func (m BrowseModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the browse view
func (m BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		}
	}

	return m, nil
}

// View renders the browse view
func (m BrowseModel) View() string {
	var b strings.Builder

	titleStyle := styles.Title.Copy().MarginBottom(1)
	b.WriteString(titleStyle.Render("Browse Apps"))
	b.WriteString("\n\n")

	// Placeholder content
	contentStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	b.WriteString(contentStyle.Render("Discover and install apps from the Kiosk marketplace."))
	b.WriteString("\n\n")

	b.WriteString(contentStyle.Render("Coming soon..."))
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Render("esc go back"))

	return b.String()
}
