package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// browseItem represents an app in the browse list
type browseItem struct {
	app api.App
}

func (i browseItem) Title() string {
	// Format: {APP_NAME} by {CREATOR}  | # of Installs
	title := i.app.Name
	if i.app.Creator != nil && i.app.Creator.Username != "" {
		title = fmt.Sprintf("%s by %s", title, i.app.Creator.Username)
	}
	if i.app.InstallCount > 0 {
		installText := "install"
		if i.app.InstallCount != 1 {
			installText = "installs"
		}
		title = fmt.Sprintf("%s  | %d %s", title, i.app.InstallCount, installText)
	}
	return title
}

func (i browseItem) Description() string {
	return i.app.Description
}

func (i browseItem) FilterValue() string {
	filterStr := i.app.Name + " " + i.app.Description
	if i.app.Creator != nil {
		filterStr += " " + i.app.Creator.Username
	}
	return filterStr
}

// BrowseModel is the model for the browse apps view
type BrowseModel struct {
	list    list.Model
	spinner spinner.Model
	width   int
	height  int
	keys    tui.KeyMap
	loading bool
	err     error
	apps    []api.App
}

// NewBrowseModel creates a new browse model
func NewBrowseModel() BrowseModel {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	// Create a custom delegate with multi-line description support
	delegate := NewAppItemDelegate()

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Browse Apps"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(styles.Primary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(styles.Secondary)

	return BrowseModel{
		list:    l,
		spinner: s,
		keys:    tui.DefaultKeyMap(),
		loading: true,
	}
}

// SetSize updates the view dimensions
func (m *BrowseModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.list.SetSize(width, height-2)
}

// Init initializes the browse model
func (m *BrowseModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		m.fetchApps,
	)
}

func (m *BrowseModel) fetchApps() tea.Msg {
	cfg, err := config.Load()
	if err != nil {
		return tui.BrowseAppsLoadedMsg{Err: err}
	}

	client := api.NewClient(cfg.APIUrl)
	apps, err := client.ListApps()
	if err != nil {
		return tui.BrowseAppsLoadedMsg{Err: err}
	}

	return tui.BrowseAppsLoadedMsg{Apps: apps}
}

// Update handles messages for the browse view
func (m *BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Don't process key events when filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }

		case key.Matches(msg, m.keys.Enter):
			if !m.loading && m.err == nil {
				if item, ok := m.list.SelectedItem().(browseItem); ok {
					app := item.app // capture for closure
					return m, func() tea.Msg {
						return tui.ShowAppDetailMsg{
							App:         &app,
							IsInstalled: false,
							AppKey:      app.ID,
						}
					}
				}
			}
		}

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tui.BrowseAppsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil
		m.apps = msg.Apps
		m.updateListItems()
	}

	// Update the list
	if !m.loading && m.err == nil {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *BrowseModel) updateListItems() {
	items := make([]list.Item, 0, len(m.apps))
	for _, app := range m.apps {
		items = append(items, browseItem{app: app})
	}
	m.list.SetItems(items)
}

// View renders the browse view
func (m *BrowseModel) View() string {
	if m.loading {
		return m.loadingView()
	}

	if m.err != nil {
		return m.errorView()
	}

	if len(m.apps) == 0 {
		return m.emptyView()
	}

	return m.list.View()
}

func (m *BrowseModel) loadingView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Browse Apps"))
	b.WriteString("\n\n")

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(styles.MutedStyle.Render("Loading apps from Kiosk..."))

	return b.String()
}

func (m *BrowseModel) errorView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Browse Apps"))
	b.WriteString("\n\n")

	b.WriteString(styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err)))
	b.WriteString("\n\n")

	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("esc go back"))

	return b.String()
}

func (m *BrowseModel) emptyView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Browse Apps"))
	b.WriteString("\n\n")

	contentStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		MaxWidth(contentWidth)
	b.WriteString(contentStyle.Render("No apps available yet."))
	b.WriteString("\n\n")

	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("esc go back"))

	return b.String()
}
