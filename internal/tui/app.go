package tui

import (
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/prefetch"
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

	// App to execute after TUI exits (set when user clicks Run)
	ExecApp string

	// View models - these will be set by the cmd package
	// to avoid circular imports
	HomeView        tea.Model
	AppListView     tea.Model
	AppDetailView   tea.Model
	BrowseView      tea.Model
	PublishView     tea.Model
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

// SetAppDetailView sets the app detail view model
func (m *Model) SetAppDetailView(v tea.Model) {
	m.AppDetailView = v
}

// SetLoginView sets the login view model
func (m *Model) SetLoginView(v tea.Model) {
	m.LoginView = v
}

// SetBrowseView sets the browse view model
func (m *Model) SetBrowseView(v tea.Model) {
	m.BrowseView = v
}

// SetPublishView sets the publish view model
func (m *Model) SetPublishView(v tea.Model) {
	m.PublishView = v
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
func (m *Model) Init() tea.Cmd {
	var cmds []tea.Cmd

	// Start prefetching browse apps in the background
	// so they're ready when the user navigates to Browse Apps
	prefetch.GetCache().StartBrowseAppsPrefetch()

	// Initialize the home view
	if m.HomeView != nil {
		cmds = append(cmds, m.HomeView.Init())
	}

	return tea.Batch(cmds...)
}

// Update handles messages for the TUI application
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case ShowAppDetailMsg:
		// Navigate to app detail and pass the app data
		m.navigateTo(ViewAppDetail)
		if m.AppDetailView != nil {
			// Update the detail view with the app info
			m.AppDetailView, _ = m.AppDetailView.Update(msg)
		}
		cmds = append(cmds, m.initCurrentView())

	case RunAppMsg:
		// Store the app key to execute after TUI exits
		m.ExecApp = msg.AppKey
		return m, tea.Quit

	case DeleteAppMsg:
		// Delete the app and go back to previous view
		cmds = append(cmds, m.deleteApp(msg.AppKey))

	case AppRemovedMsg:
		if msg.Err != nil {
			m.err = msg.Err
		} else {
			m.status = "App removed successfully"
			// Go back to previous view and refresh
			m.goBack()
			cmds = append(cmds, m.initCurrentView())
		}

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
	// Calculate content area (accounting for padding applied in View())
	// Padding(1, 2) means 1 line top/bottom, 2 chars left/right
	contentWidth := m.width - 4   // Account for horizontal padding (2 left + 2 right)
	contentHeight := m.height - 4 // Account for vertical padding (1 top + 1 bottom) + status/help

	if contentWidth < 20 {
		contentWidth = 20
	}
	if contentHeight < 10 {
		contentHeight = 10
	}

	views := []tea.Model{
		m.HomeView,
		m.AppListView,
		m.AppDetailView,
		m.BrowseView,
		m.PublishView,
		m.HelpView,
		m.LoginView,
		m.AuditView,
		m.PostInstallView,
	}

	for _, v := range views {
		if v != nil {
			if sizer, ok := v.(Sizer); ok {
				sizer.SetSize(contentWidth, contentHeight)
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
	case ViewAppDetail:
		if m.AppDetailView != nil {
			return m.AppDetailView.Init()
		}
	case ViewBrowse:
		if m.BrowseView != nil {
			return m.BrowseView.Init()
		}
	case ViewPublish:
		if m.PublishView != nil {
			return m.PublishView.Init()
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
	case ViewAppDetail:
		if m.AppDetailView != nil {
			m.AppDetailView, cmd = m.AppDetailView.Update(msg)
		}
	case ViewBrowse:
		if m.BrowseView != nil {
			m.BrowseView, cmd = m.BrowseView.Update(msg)
		}
	case ViewPublish:
		if m.PublishView != nil {
			m.PublishView, cmd = m.PublishView.Update(msg)
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
func (m *Model) View() string {
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
	case ViewAppDetail:
		if m.AppDetailView != nil {
			content = m.AppDetailView.View()
		}
	case ViewBrowse:
		if m.BrowseView != nil {
			content = m.BrowseView.View()
		}
	case ViewPublish:
		if m.PublishView != nil {
			content = m.PublishView.View()
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

// deleteApp removes an app from the index and filesystem
func (m *Model) deleteApp(key string) tea.Cmd {
	return func() tea.Msg {
		// Load index
		idx, err := appindex.Load()
		if err != nil {
			return AppRemovedMsg{Key: key, Err: err}
		}

		// Check if app is in index
		if !idx.Has(key) {
			return AppRemovedMsg{Key: key, Err: nil} // Already removed
		}

		// Remove directory if it exists
		parts := strings.SplitN(key, "/", 2)
		if len(parts) == 2 {
			appPath := config.AppPath(parts[0], parts[1])
			if _, err := os.Stat(appPath); err == nil {
				if err := os.RemoveAll(appPath); err != nil {
					return AppRemovedMsg{Key: key, Err: err}
				}
			}
		}

		// Remove from index
		idx.Remove(key)
		if err := appindex.Save(idx); err != nil {
			return AppRemovedMsg{Key: key, Err: err}
		}

		return AppRemovedMsg{Key: key, Err: nil}
	}
}
