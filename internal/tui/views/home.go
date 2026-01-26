package views

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// HomeModel is the model for the home/main menu view
type HomeModel struct {
	width  int
	height int
	cursor int
	items  []menuItem
	keys   tui.KeyMap
}

type menuItem struct {
	title       string
	description string
	action      func() tea.Msg
}

// NewHomeModel creates a new home model
func NewHomeModel() HomeModel {
	m := HomeModel{
		cursor: 0,
		keys:   tui.DefaultKeyMap(),
	}
	m.updateMenuItems()
	return m
}

func (m *HomeModel) updateMenuItems() {
	m.items = []menuItem{
		{
			title:       "My Apps",
			description: "View and manage your installed apps",
			action:      func() tea.Msg { return tui.NavigateMsg{View: tui.ViewAppList} },
		},
		{
			title:       "Browse Apps",
			description: "Discover and install new apps",
			action:      func() tea.Msg { return tui.NavigateMsg{View: tui.ViewBrowse} },
		},
		{
			title:       "Publish App",
			description: "Publish your app to Kiosk",
			action:      func() tea.Msg { return tui.NavigateMsg{View: tui.ViewPublish} },
		},
		{
			title:       "Help",
			description: "Documentation and support",
			action:      func() tea.Msg { return tui.NavigateMsg{View: tui.ViewHelp} },
		},
	}
}

// SetSize updates the view dimensions
func (m *HomeModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the home model
func (m *HomeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the home view
func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			if m.cursor < len(m.items) {
				return m, func() tea.Msg { return m.items[m.cursor].action() }
			}
		}
	}

	return m, nil
}

// View renders the home view
func (m *HomeModel) View() string {
	var b strings.Builder

	// Calculate available width for content
	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80 // reasonable default
	}

	// Logo - only show if terminal is wide enough
	if contentWidth >= 90 {
		b.WriteString(styles.LogoStyled())
		b.WriteString("\n\n")
	}

	// Menu items - use MaxWidth to truncate if needed
	for i, item := range m.items {
		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(styles.Foreground)
		descStyle := lipgloss.NewStyle().Foreground(styles.Muted)

		if i == m.cursor {
			cursor = styles.Highlight.Render("> ")
			itemStyle = itemStyle.Bold(true).Foreground(styles.Primary)
		}

		// Calculate remaining width for description
		titleLen := len(item.title) + 5 // cursor (2) + title + " - " (3)
		descWidth := contentWidth - titleLen
		if descWidth < 10 {
			descWidth = 10
		}

		line := cursor + itemStyle.Render(item.title) + " " + descStyle.Render("- "+item.description)

		// Truncate line if it exceeds content width
		lineStyle := lipgloss.NewStyle().MaxWidth(contentWidth)
		b.WriteString(lineStyle.Render(line))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(styles.Muted).MaxWidth(contentWidth)
	b.WriteString(helpStyle.Render("↑/↓ navigate • enter select • q quit"))

	return b.String()
}
