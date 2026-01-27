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
	"github.com/reflective-technologies/kiosk-cli/internal/prefetch"
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

	// Pagination state
	nextCursor      *string // cursor for next page, nil if no more pages
	loadingMore     bool    // true when loading additional pages
	fetchGeneration uint64  // incremented on Init() to invalidate in-flight fetches
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
	// Increment generation to invalidate any in-flight pagination fetches
	m.fetchGeneration++

	// Reset pagination state
	m.loadingMore = false
	m.nextCursor = nil

	// Check if we have prefetched data available
	cache := prefetch.GetCache()
	result := cache.GetBrowseApps()

	if result.Loaded && result.Err == nil {
		// Data was prefetched successfully - apply it immediately (no loading state)
		m.loading = false
		m.err = nil
		m.apps = result.Apps
		m.nextCursor = result.NextCursor
		m.updateListItems()
		return nil
	}

	// If there was a cached error, reset and start a fresh prefetch
	if result.Loaded && result.Err != nil {
		cache.ResetBrowseApps()
		cache.StartBrowseAppsPrefetch()
	}

	// Data not ready yet (or retrying after error) - show spinner and wait for prefetch to complete
	m.loading = true
	m.apps = nil
	m.err = nil
	return tea.Batch(
		m.spinner.Tick,
		m.waitForPrefetch,
	)
}

// waitForPrefetch waits for the prefetch to complete and returns the result
func (m *BrowseModel) waitForPrefetch() tea.Msg {
	cache := prefetch.GetCache()
	result := cache.WaitForBrowseApps()

	return tui.BrowseAppsLoadedMsg{
		Apps:       result.Apps,
		NextCursor: result.NextCursor,
		Err:        result.Err,
	}
}

func (m *BrowseModel) fetchMoreApps() tea.Cmd {
	if m.nextCursor == nil {
		return nil
	}

	cursor := *m.nextCursor
	generation := m.fetchGeneration // capture current generation
	return func() tea.Msg {
		cfg, err := config.Load()
		if err != nil {
			return tui.BrowseAppsPageLoadedMsg{Err: err, Generation: generation}
		}

		client := api.NewClient(cfg.APIUrl)
		result, err := client.ListAppsPaginated(prefetch.DefaultPageSize, cursor)
		if err != nil {
			return tui.BrowseAppsPageLoadedMsg{Err: err, Generation: generation}
		}

		return tui.BrowseAppsPageLoadedMsg{
			Apps:       result.Apps,
			NextCursor: result.NextCursor,
			Generation: generation,
		}
	}
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
		if m.loading || m.loadingMore {
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
		m.nextCursor = msg.NextCursor
		m.updateListItems()

	case tui.BrowseAppsPageLoadedMsg:
		// Ignore stale messages from previous sessions (e.g., user navigated away and back)
		if msg.Generation != m.fetchGeneration {
			return m, nil
		}
		m.loadingMore = false
		if msg.Err != nil {
			// Don't show error for pagination failures, just stop loading
			return m, nil
		}
		// Append new apps to existing list
		m.apps = append(m.apps, msg.Apps...)
		m.nextCursor = msg.NextCursor
		m.updateListItems()
	}

	// Update the list
	if !m.loading && m.err == nil {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)

		// Check if we should load more apps (when near the bottom of the list)
		if m.shouldLoadMore() {
			m.loadingMore = true
			cmds = append(cmds, m.spinner.Tick, m.fetchMoreApps())
		}
	}

	return m, tea.Batch(cmds...)
}

// shouldLoadMore returns true if we should fetch the next page of apps
func (m *BrowseModel) shouldLoadMore() bool {
	// Don't load more if already loading or no more pages
	if m.loadingMore || m.nextCursor == nil {
		return false
	}

	// Don't load more when filtering
	if m.list.FilterState() == list.Filtering {
		return false
	}

	// Load more when user is within 3 items of the bottom
	totalItems := len(m.list.Items())
	if totalItems == 0 {
		return false
	}

	currentIndex := m.list.Index()
	threshold := 3

	return currentIndex >= totalItems-threshold
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

	// Show list with optional loading indicator for pagination
	view := m.list.View()
	if m.loadingMore {
		view += "\n" + m.spinner.View() + " " + styles.MutedStyle.Render("Loading more...")
	}
	return view
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
