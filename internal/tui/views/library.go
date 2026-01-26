package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// LibraryModel is the model for the library view
type LibraryModel struct {
	width  int
	height int
	keys   tui.KeyMap
}

// NewLibraryModel creates a new library model
func NewLibraryModel() LibraryModel {
	return LibraryModel{
		keys: tui.DefaultKeyMap(),
	}
}

// SetSize updates the view dimensions
func (m *LibraryModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the library model
func (m LibraryModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the library view
func (m LibraryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		}
	}

	return m, nil
}

// View renders the library view
func (m LibraryModel) View() string {
	var b strings.Builder

	titleStyle := styles.Title.Copy().MarginBottom(1)
	b.WriteString(titleStyle.Render("Library"))
	b.WriteString("\n\n")

	// Placeholder content
	contentStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	b.WriteString(contentStyle.Render("Your app collection and history."))
	b.WriteString("\n\n")

	b.WriteString(contentStyle.Render("Coming soon..."))
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Render("esc go back"))

	return b.String()
}
