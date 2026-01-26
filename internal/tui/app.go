package tui

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// Model is the main application model for the TUI
type Model struct {
	width       int
	height      int
	currentView ViewType
	viewStack   []ViewType
	keys        KeyMap
	help        help.Model
	showHelp    bool
	status      string
	err         error

	// View models - these will be set by the cmd package
	// to avoid circular imports
	HomeView        tea.Model
	AppListView     tea.Model
	BrowseView      tea.Model
	LibraryView     tea.Model
	HelpView        tea.Model
	LoginView       tea.Model
	AuditView       tea.Model
	PostInstallView tea.Model
}

// New creates a new TUI application model
func New() Model {
	return Model{
		currentView: ViewHome,
		viewStack:   []ViewType{},
		keys:        DefaultKeyMap(),
		help:        help.New(),
	}
}

// SetHomeView sets the home view model
func (m *Model) SetHomeView(v tea.Model) {
	m.HomeView = v
}

// SetAppListView sets the app list view model
func (m *Model) SetAppListView(v tea.Model) {
	m.AppListView = v
}

// SetLoginView sets the login view model
func (m *Model) SetLoginView(v tea.Model) {
	m.LoginView = v
}

// SetBrowseView sets the browse view model
func (m *Model) SetBrowseView(v tea.Model) {
	m.BrowseView = v
}

// SetLibraryView sets the library view model
func (m *Model) SetLibraryView(v tea.Model) {
	m.LibraryView = v
}

// SetHelpView sets the help view model
func (m *Model) SetHelpView(v tea.Model) {
	m.HelpView = v
}

// SetAuditView sets the audit view model
func (m *Model) SetAuditView(v tea.Model) {
	m.AuditView = v
}

// SetPostInstallView sets the post-install view model
func (m *Model) SetPostInstallView(v tea.Model) {
	m.PostInstallView = v
}

// Init initializes the TUI application
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Initialize the home view
	if m.HomeView != nil {
		cmds = append(cmds, m.HomeView.Init())
	}

	return tea.Batch(cmds...)
}

// Update handles messages for the TUI application
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width

		// Update size for all views using the Sizer interface
		m.updateViewSizes()

	case tea.KeyMsg:
		// Global key handling
		switch msg.String() {
		case "ctrl+c", "q":
			// Only quit from home view
			if m.currentView == ViewHome {
				return m, tea.Quit
			}
		case "?":
			m.showHelp = !m.showHelp
		}

	case NavigateMsg:
		m.navigateTo(msg.View)
		cmds = append(cmds, m.initCurrentView())

	case GoBackMsg:
		m.goBack()
		cmds = append(cmds, m.initCurrentView())

	case ErrorMsg:
		m.err = msg.Err

	case StatusMsg:
		m.status = msg.Message

	case ClearStatusMsg:
		m.status = ""
	}

	// Update the current view
	cmd := m.updateCurrentView(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Sizer is an interface for views that can be resized
type Sizer interface {
	SetSize(width, height int)
}

func (m *Model) updateViewSizes() {
	// Calculate content area (accounting for any chrome)
	contentHeight := m.height - 2 // Leave room for status/help

	views := []tea.Model{
		m.HomeView,
		m.AppListView,
		m.BrowseView,
		m.LibraryView,
		m.HelpView,
		m.LoginView,
		m.AuditView,
		m.PostInstallView,
	}

	for _, v := range views {
		if v != nil {
			if sizer, ok := v.(Sizer); ok {
				sizer.SetSize(m.width, contentHeight)
			}
		}
	}
}

func (m *Model) navigateTo(view ViewType) {
	// Push current view to stack
	m.viewStack = append(m.viewStack, m.currentView)
	m.currentView = view
}

func (m *Model) goBack() {
	if len(m.viewStack) > 0 {
		// Pop from stack
		m.currentView = m.viewStack[len(m.viewStack)-1]
		m.viewStack = m.viewStack[:len(m.viewStack)-1]
	}
}

func (m Model) initCurrentView() tea.Cmd {
	switch m.currentView {
	case ViewHome:
		if m.HomeView != nil {
			return m.HomeView.Init()
		}
	case ViewAppList:
		if m.AppListView != nil {
			return m.AppListView.Init()
		}
	case ViewBrowse:
		if m.BrowseView != nil {
			return m.BrowseView.Init()
		}
	case ViewLibrary:
		if m.LibraryView != nil {
			return m.LibraryView.Init()
		}
	case ViewHelp:
		if m.HelpView != nil {
			return m.HelpView.Init()
		}
	case ViewLogin:
		if m.LoginView != nil {
			return m.LoginView.Init()
		}
	case ViewAudit:
		if m.AuditView != nil {
			return m.AuditView.Init()
		}
	case ViewPostInstall:
		if m.PostInstallView != nil {
			return m.PostInstallView.Init()
		}
	}
	return nil
}

func (m *Model) updateCurrentView(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd

	switch m.currentView {
	case ViewHome:
		if m.HomeView != nil {
			m.HomeView, cmd = m.HomeView.Update(msg)
		}
	case ViewAppList:
		if m.AppListView != nil {
			m.AppListView, cmd = m.AppListView.Update(msg)
		}
	case ViewBrowse:
		if m.BrowseView != nil {
			m.BrowseView, cmd = m.BrowseView.Update(msg)
		}
	case ViewLibrary:
		if m.LibraryView != nil {
			m.LibraryView, cmd = m.LibraryView.Update(msg)
		}
	case ViewHelp:
		if m.HelpView != nil {
			m.HelpView, cmd = m.HelpView.Update(msg)
		}
	case ViewLogin:
		if m.LoginView != nil {
			m.LoginView, cmd = m.LoginView.Update(msg)
		}
	case ViewAudit:
		if m.AuditView != nil {
			m.AuditView, cmd = m.AuditView.Update(msg)
		}
	case ViewPostInstall:
		if m.PostInstallView != nil {
			m.PostInstallView, cmd = m.PostInstallView.Update(msg)
		}
	}

	return cmd
}

// View renders the TUI application
func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// Main content
	var content string
	switch m.currentView {
	case ViewHome:
		if m.HomeView != nil {
			content = m.HomeView.View()
		}
	case ViewAppList:
		if m.AppListView != nil {
			content = m.AppListView.View()
		}
	case ViewBrowse:
		if m.BrowseView != nil {
			content = m.BrowseView.View()
		}
	case ViewLibrary:
		if m.LibraryView != nil {
			content = m.LibraryView.View()
		}
	case ViewHelp:
		if m.HelpView != nil {
			content = m.HelpView.View()
		}
	case ViewLogin:
		if m.LoginView != nil {
			content = m.LoginView.View()
		}
	case ViewAudit:
		if m.AuditView != nil {
			content = m.AuditView.View()
		}
	case ViewPostInstall:
		if m.PostInstallView != nil {
			content = m.PostInstallView.View()
		}
	default:
		content = "Unknown view"
	}

	// Apply padding
	paddedContent := lipgloss.NewStyle().
		Padding(1, 2).
		Render(content)

	// Show help if enabled
	if m.showHelp {
		helpView := m.help.View(m.keys)
		paddedContent += "\n" + helpView
	}

	// Show error if any
	if m.err != nil {
		errorView := styles.ErrorStyle.Render("Error: " + m.err.Error())
		paddedContent += "\n" + errorView
	}

	// Show status if any
	if m.status != "" {
		statusView := styles.MutedStyle.Render(m.status)
		paddedContent += "\n" + statusView
	}

	return paddedContent
}
