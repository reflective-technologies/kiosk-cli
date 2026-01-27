package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// AppDetailModel is the model for the app detail view
type AppDetailModel struct {
	width  int
	height int
	keys   tui.KeyMap

	// App info
	app         *api.App
	isInstalled bool
	appKey      string

	// Button selection (0 = Run, 1 = Delete for installed; 0 = Install for browse)
	cursor int
}

// NewAppDetailModel creates a new app detail model
func NewAppDetailModel() AppDetailModel {
	return AppDetailModel{
		keys: tui.DefaultKeyMap(),
	}
}

// SetApp sets the app to display
func (m *AppDetailModel) SetApp(app *api.App, isInstalled bool, appKey string) {
	m.app = app
	m.appKey = appKey
	m.cursor = 0

	// Check if the app is installed by looking at the app index
	if isInstalled {
		m.isInstalled = true
	} else {
		m.isInstalled = m.checkIfInstalled(app)
	}
}

// checkIfInstalled checks if the app is in the local app index
func (m *AppDetailModel) checkIfInstalled(app *api.App) bool {
	if app == nil {
		return false
	}

	idx, err := appindex.Load()
	if err != nil {
		return false
	}

	// Check by app ID
	if idx.Get(app.ID) != nil {
		return true
	}

	// Also check by extracting repo name from git URL if available
	if app.GitUrl != "" {
		// Try to extract owner/repo from git URL
		// e.g., https://github.com/owner/repo -> owner/repo
		gitUrl := app.GitUrl
		gitUrl = strings.TrimSuffix(gitUrl, ".git")
		if strings.Contains(gitUrl, "github.com/") {
			parts := strings.Split(gitUrl, "github.com/")
			if len(parts) == 2 {
				repoPath := parts[1]
				if idx.Get(repoPath) != nil {
					return true
				}
			}
		}
	}

	return false
}

// SetSize updates the view dimensions
func (m *AppDetailModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the app detail model
func (m *AppDetailModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the app detail view
func (m *AppDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		case key.Matches(msg, m.keys.Up), msg.String() == "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case key.Matches(msg, m.keys.Down), msg.String() == "right":
			maxCursor := 0
			if m.isInstalled {
				maxCursor = 1 // Run and Delete
			}
			if m.cursor < maxCursor {
				m.cursor++
			}
		case key.Matches(msg, m.keys.Enter):
			return m, m.handleAction()
		}

	case tui.ShowAppDetailMsg:
		m.SetApp(msg.App, msg.IsInstalled, msg.AppKey)
	}

	return m, nil
}

func (m *AppDetailModel) handleAction() tea.Cmd {
	if m.app == nil {
		return nil
	}

	if m.isInstalled {
		if m.cursor == 0 {
			// Run
			return func() tea.Msg {
				return tui.RunAppMsg{
					AppKey: m.appKey,
					GitURL: m.app.GitUrl,
				}
			}
		} else {
			// Delete
			return func() tea.Msg {
				return tui.DeleteAppMsg{
					AppKey: m.appKey,
				}
			}
		}
	} else {
		// Install (for browse apps)
		return func() tea.Msg {
			return tui.RunAppMsg{
				AppKey: m.app.ID,
				GitURL: m.app.GitUrl,
			}
		}
	}
}

// View renders the app detail view
func (m *AppDetailModel) View() string {
	if m.app == nil {
		return styles.MutedStyle.Render("No app selected")
	}

	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	// Left padding to align with browse list items
	indent := "  " // 2 spaces to match list item indentation

	// App name
	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(indent)
	b.WriteString(titleStyle.Render(m.app.Name))
	b.WriteString("\n")

	// Author/Creator as subheader with install count
	var subheaderParts []string
	if m.app.Creator != nil && m.app.Creator.Username != "" {
		subheaderParts = append(subheaderParts, fmt.Sprintf("by %s", m.app.Creator.Username))
	}
	if m.app.InstallCount > 0 {
		installText := "install"
		if m.app.InstallCount != 1 {
			installText = "installs"
		}
		subheaderParts = append(subheaderParts, fmt.Sprintf("%d %s", m.app.InstallCount, installText))
	}
	if m.isInstalled {
		subheaderParts = append(subheaderParts, styles.SuccessStyle.Render("Installed"))
	}
	if len(subheaderParts) > 0 {
		b.WriteString(indent)
		b.WriteString(styles.MutedStyle.Render(subheaderParts[0]))
		for i := 1; i < len(subheaderParts); i++ {
			b.WriteString(styles.MutedStyle.Render("  |  "))
			b.WriteString(subheaderParts[i])
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Description
	if m.app.Description != "" {
		descStyle := lipgloss.NewStyle().
			Foreground(styles.Foreground).
			MaxWidth(contentWidth - 3) // account for indent
		b.WriteString(indent)
		b.WriteString(descStyle.Render(m.app.Description))
		b.WriteString("\n\n")
	}

	// Action buttons
	b.WriteString(indent)
	b.WriteString(m.renderButtons())
	b.WriteString("\n\n")

	// Help
	b.WriteString(indent)
	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("←/→ select • enter confirm • esc go back"))

	return b.String()
}

func (m *AppDetailModel) renderButtons() string {
	if m.isInstalled {
		// Run and Delete buttons
		runStyle := lipgloss.NewStyle().Padding(0, 2)
		deleteStyle := lipgloss.NewStyle().Padding(0, 2)

		if m.cursor == 0 {
			runStyle = runStyle.
				Background(styles.Primary).
				Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			runStyle = runStyle.Foreground(styles.Muted)
		}

		if m.cursor == 1 {
			deleteStyle = deleteStyle.
				Background(styles.Error).
				Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			deleteStyle = deleteStyle.Foreground(styles.Muted)
		}

		return runStyle.Render("Run") + "  " + deleteStyle.Render("Delete")
	} else {
		// Install button only
		installStyle := lipgloss.NewStyle().Padding(0, 2)

		if m.cursor == 0 {
			installStyle = installStyle.
				Background(styles.Primary).
				Foreground(lipgloss.Color("#FFFFFF"))
		} else {
			installStyle = installStyle.Foreground(styles.Muted)
		}

		return installStyle.Render("Install")
	}
}
