package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/views"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List installed apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := appindex.Load()
		if err != nil {
			return fmt.Errorf("failed to load app index: %w", err)
		}

		if idx.Count() == 0 {
			fmt.Println()
			fmt.Println(styles.MutedStyle.Render("  No apps installed."))
			fmt.Println()
			fmt.Println("  Run " + lipgloss.NewStyle().Bold(true).Render("kiosk run <app>") + " to install an app.")
			fmt.Println()
			return nil
		}

		// Run interactive list
		m := newLsModel(idx)
		p := tea.NewProgram(m, tea.WithAltScreen())

		finalModel, err := p.Run()
		if err != nil {
			return fmt.Errorf("error running list: %w", err)
		}

		// Check if user selected an app to run
		if model, ok := finalModel.(*lsModel); ok && model.runApp != "" {
			return runInstalledApp(model.runApp, nil, false)
		}

		return nil
	},
}

// View state for ls command
type lsView int

const (
	lsViewList lsView = iota
	lsViewDetail
)

// lsModel is the bubbletea model for the ls command
type lsModel struct {
	list         list.Model
	index        *appindex.Index
	currentView  lsView
	selectedItem *lsItem
	detailCursor int // 0 = Run, 1 = Delete
	runApp       string
	width        int
	height       int
}

// lsItem represents an app in the list
type lsItem struct {
	key         string
	name        string
	author      string
	description string
	gitUrl      string
	missing     bool
}

func (i lsItem) Title() string {
	title := i.name
	if i.author != "" {
		title = fmt.Sprintf("%s by %s", title, i.author)
	}
	if i.missing {
		title += styles.WarningStyle.Render(" (missing)")
	}
	return title
}

func (i lsItem) Description() string {
	return i.description
}

func (i lsItem) FilterValue() string {
	return i.name + " " + i.author + " " + i.description
}

func newLsModel(idx *appindex.Index) *lsModel {
	// Create delegate with same styling as TUI
	delegate := views.NewAppItemDelegate()

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "My Apps"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(styles.Primary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(styles.Secondary)

	// Set help text
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		}
	}

	m := &lsModel{
		list:        l,
		index:       idx,
		currentView: lsViewList,
	}

	m.loadItems()

	return m
}

func (m *lsModel) loadItems() {
	keys := m.index.List()
	sort.Strings(keys)

	exists := m.index.ValidateFilesystem()

	items := make([]list.Item, 0, len(keys))
	for _, k := range keys {
		entry := m.index.Get(k)
		author, name := splitAppKey(k)

		item := lsItem{
			key:     k,
			name:    name,
			author:  author,
			missing: !exists[k],
		}

		if entry != nil {
			item.description = entry.Description
			item.gitUrl = entry.GitUrl
		}

		items = append(items, item)
	}

	m.list.SetItems(items)
}

func (m *lsModel) Init() tea.Cmd {
	return nil
}

func (m *lsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		if m.currentView == lsViewDetail {
			return m.updateDetailView(msg)
		}
		return m.updateListView(msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *lsModel) updateListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Don't handle keys when filtering
	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter":
		if item, ok := m.list.SelectedItem().(lsItem); ok {
			m.selectedItem = &item
			m.currentView = lsViewDetail
			m.detailCursor = 0
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *lsModel) updateDetailView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.currentView = lsViewList
		m.selectedItem = nil
		return m, nil
	case "left", "h":
		if m.detailCursor > 0 {
			m.detailCursor--
		}
	case "right", "l":
		if m.detailCursor < 1 {
			m.detailCursor++
		}
	case "enter":
		if m.selectedItem == nil {
			return m, nil
		}
		if m.detailCursor == 0 {
			// Run
			if !m.selectedItem.missing {
				m.runApp = m.selectedItem.key
				return m, tea.Quit
			}
		} else {
			// Delete
			m.deleteApp(m.selectedItem.key)
			m.currentView = lsViewList
			m.selectedItem = nil
			m.loadItems()
			return m, nil
		}
	}
	return m, nil
}

func (m *lsModel) deleteApp(key string) {
	// Remove from filesystem
	parts := strings.SplitN(key, "/", 2)
	if len(parts) == 2 {
		appPath := config.AppPath(parts[0], parts[1])
		os.RemoveAll(appPath)
	}

	// Remove from index
	m.index.Remove(key)
	appindex.Save(m.index)
}

func (m *lsModel) View() string {
	if m.currentView == lsViewDetail {
		return m.viewDetail()
	}
	return m.list.View()
}

func (m *lsModel) viewDetail() string {
	if m.selectedItem == nil {
		return ""
	}

	item := m.selectedItem
	var b strings.Builder

	// Padding
	b.WriteString("\n")

	// App name
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
	b.WriteString("  ")
	b.WriteString(titleStyle.Render(item.name))
	b.WriteString("\n")

	// Author
	if item.author != "" {
		b.WriteString("  ")
		b.WriteString(styles.MutedStyle.Render("by " + item.author))
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Description
	if item.description != "" {
		descStyle := lipgloss.NewStyle().Width(m.width - 4)
		b.WriteString("  ")
		b.WriteString(descStyle.Render(item.description))
		b.WriteString("\n\n")
	}

	// Buttons
	b.WriteString("  ")
	b.WriteString(m.renderButtons())
	b.WriteString("\n\n")

	// Help
	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render("left/right select | enter confirm | esc back | q quit"))
	b.WriteString("\n")

	return b.String()
}

func (m *lsModel) renderButtons() string {
	runStyle := lipgloss.NewStyle().Padding(0, 2)
	deleteStyle := lipgloss.NewStyle().Padding(0, 2)

	if m.detailCursor == 0 {
		runStyle = runStyle.
			Background(styles.Primary).
			Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		runStyle = runStyle.Foreground(styles.Muted)
	}

	if m.detailCursor == 1 {
		deleteStyle = deleteStyle.
			Background(styles.Error).
			Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		deleteStyle = deleteStyle.Foreground(styles.Muted)
	}

	runLabel := "Run"
	if m.selectedItem != nil && m.selectedItem.missing {
		runLabel = "Run (missing)"
		runStyle = runStyle.Foreground(styles.Muted)
	}

	return runStyle.Render(runLabel) + "  " + deleteStyle.Render("Delete")
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

func init() {
	rootCmd.AddCommand(lsCmd)
}
