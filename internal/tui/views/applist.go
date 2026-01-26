package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// appItem represents an app in the list
type appItem struct {
	key         string
	name        string
	author      string
	description string
	installed   bool
	missing     bool
}

func (i appItem) Title() string {
	// Format: {APP_NAME} by {AUTHOR}
	title := i.name
	if i.author != "" {
		title = fmt.Sprintf("%s by %s", title, i.author)
	}
	if i.missing {
		title += styles.WarningStyle.Render(" (missing)")
	}
	return title
}

func (i appItem) Description() string {
	return i.description
}

func (i appItem) FilterValue() string { return i.name + " " + i.author + " " + i.description }

// AppListModel is the model for the app list view
type AppListModel struct {
	list     list.Model
	width    int
	height   int
	keys     tui.KeyMap
	index    *appindex.Index
	selected *appItem
	loading  bool
	err      error
}

// NewAppListModel creates a new app list model
func NewAppListModel() AppListModel {
	// Create a custom delegate
	delegate := list.NewDefaultDelegate()
	// Keep text colors consistent - only use indicator for selection
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(styles.Foreground).
		Padding(0, 0, 0, 2) // indent to align with selected items
	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(styles.Muted).
		Padding(0, 0, 0, 2)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(styles.Primary).
		Foreground(styles.Foreground).
		Padding(0, 0, 0, 1)
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, false, true).
		BorderForeground(styles.Primary).
		Foreground(styles.Muted).
		Padding(0, 0, 0, 1)
	delegate.SetSpacing(0)
	delegate.SetHeight(3) // Title + Description + blank line for spacing

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "My Apps"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(styles.Primary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(styles.Secondary)

	return AppListModel{
		list:    l,
		keys:    tui.DefaultKeyMap(),
		loading: true,
	}
}

// SetSize updates the view dimensions
func (m *AppListModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	// Account for header and status bar
	m.list.SetSize(width, height-4)
}

// Init initializes the app list model
func (m *AppListModel) Init() tea.Cmd {
	return m.loadApps
}

func (m *AppListModel) loadApps() tea.Msg {
	idx, err := appindex.Load()
	if err != nil {
		return tui.AppsLoadedMsg{Err: err}
	}
	return tui.AppsLoadedMsg{Index: idx}
}

// Update handles messages for the app list view
func (m *AppListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if item, ok := m.list.SelectedItem().(appItem); ok {
				m.selected = &item
				return m, func() tea.Msg {
					return tui.AppSelectedMsg{
						Key: item.key,
						Entry: &appindex.AppEntry{
							Name:        item.name,
							Description: item.description,
						},
					}
				}
			}
		}

	case tui.AppsLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.index = msg.Index
		m.updateListItems()
	}

	// Update the list
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m *AppListModel) updateListItems() {
	if m.index == nil {
		return
	}

	keys := m.index.List()
	sort.Strings(keys)

	// Validate filesystem
	exists := m.index.ValidateFilesystem()

	items := make([]list.Item, 0, len(keys))
	for _, k := range keys {
		entry := m.index.Get(k)
		author, name := splitAppKey(k)

		item := appItem{
			key:         k,
			name:        name,
			author:      author,
			description: entry.Description,
			installed:   true,
			missing:     !exists[k],
		}
		items = append(items, item)
	}

	m.list.SetItems(items)

	if len(items) == 0 {
		m.list.SetShowStatusBar(false)
	} else {
		m.list.SetShowStatusBar(true)
	}
}

func splitAppKey(key string) (author, name string) {
	parts := strings.SplitN(key, "/", 2)
	author = parts[0]
	name = parts[0]
	if len(parts) == 2 {
		name = parts[1]
	} else {
		author = ""
	}
	return author, name
}

// View renders the app list view
func (m *AppListModel) View() string {
	if m.loading {
		return styles.MutedStyle.Render("Loading apps...")
	}

	if m.err != nil {
		return styles.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	if m.index == nil || m.index.Count() == 0 {
		return m.emptyView()
	}

	return m.list.View()
}

func (m *AppListModel) emptyView() string {
	var b strings.Builder

	// Calculate available width for content
	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("My Apps"))
	b.WriteString("\n\n")

	emptyStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Italic(true).
		MarginLeft(2).
		MaxWidth(contentWidth)

	b.WriteString(emptyStyle.Render("No apps installed yet."))
	b.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		MaxWidth(contentWidth)
	b.WriteString(hintStyle.Render("  Run an app with: kiosk run <app>"))
	b.WriteString("\n")
	b.WriteString(hintStyle.Render("  Example: kiosk run anthropic/claude-starter"))
	b.WriteString("\n\n")

	helpStyle := lipgloss.NewStyle().Foreground(styles.Muted).MaxWidth(contentWidth)
	b.WriteString(helpStyle.Render("  Press esc to go back"))

	return b.String()
}
