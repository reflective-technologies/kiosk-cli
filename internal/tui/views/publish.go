package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// PublishModel is the model for the publish app view
type PublishModel struct {
	width  int
	height int
	keys   tui.KeyMap
}

// NewPublishModel creates a new publish model
func NewPublishModel() PublishModel {
	return PublishModel{
		keys: tui.DefaultKeyMap(),
	}
}

// SetSize updates the view dimensions
func (m *PublishModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the publish model
func (m *PublishModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the publish view
func (m *PublishModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		}
	}

	return m, nil
}

// View renders the publish view
func (m *PublishModel) View() string {
	var b strings.Builder

	// Calculate available width for content
	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Publish App"))
	b.WriteString("\n\n")

	// Instructions
	contentStyle := lipgloss.NewStyle().
		Foreground(styles.Foreground).
		MaxWidth(contentWidth)

	b.WriteString(contentStyle.Render("To publish your app to Kiosk:"))
	b.WriteString("\n\n")

	stepStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		MaxWidth(contentWidth)

	steps := []string{
		"1. Create a KIOSK.md file in your repository root",
		"2. Run 'kiosk publish' from your project directory",
		"3. Follow the prompts to authenticate and publish",
	}

	for _, step := range steps {
		b.WriteString(stepStyle.Render("  " + step))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Hint about init command
	hintStyle := lipgloss.NewStyle().
		Foreground(styles.Secondary).
		MaxWidth(contentWidth)
	b.WriteString(hintStyle.Render("Tip: Run 'kiosk init' to generate a KIOSK.md template"))
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("esc go back"))

	return b.String()
}
