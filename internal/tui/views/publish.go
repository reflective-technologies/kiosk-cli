package views

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// PublishState represents the current state of the publish view
type PublishState int

const (
	PublishStateChecking PublishState = iota
	PublishStatePublishable
	PublishStatePickDirectory
	PublishStateNotPublishable
)

// directoryItem represents a directory in the picker
type directoryItem struct {
	name     string
	path     string
	isParent bool
}

// PublishModel is the model for the publish app view
type PublishModel struct {
	width   int
	height  int
	keys    tui.KeyMap
	state   PublishState
	spinner spinner.Model

	// Current directory info
	currentDir    string
	startDir      string // The directory we started in
	projectName   string
	hasKioskMd    bool
	hasGit        bool
	isPublishable bool

	// Directory picker
	directories []directoryItem
	cursor      int
	dirHistory  []string // Stack of directories for back navigation

	// Confirmation
	confirmCursor int
}

// NewPublishModel creates a new publish model
func NewPublishModel() PublishModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return PublishModel{
		keys:    tui.DefaultKeyMap(),
		state:   PublishStateChecking,
		spinner: s,
	}
}

// SetSize updates the view dimensions
func (m *PublishModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the publish model
func (m *PublishModel) Init() tea.Cmd {
	m.state = PublishStateChecking
	return tea.Batch(
		m.spinner.Tick,
		m.checkCurrentDirectory,
	)
}

// checkCurrentDirectory checks if the current directory is publishable
func (m *PublishModel) checkCurrentDirectory() tea.Msg {
	cwd, err := os.Getwd()
	if err != nil {
		return publishCheckResultMsg{err: err}
	}

	return publishCheckResultMsg{
		dir:             cwd,
		isUnpublishable: isUnpublishableDirectory(cwd),
	}
}

// isUnpublishableDirectory checks if a directory is a common non-project directory
func isUnpublishableDirectory(dir string) bool {
	home, _ := os.UserHomeDir()

	// Check if it's the home directory or root
	if dir == home || dir == "/" {
		return true
	}

	// Get the base name of the directory
	baseName := strings.ToLower(filepath.Base(dir))

	// Common non-project directories
	unpublishableDirs := []string{
		"desktop",
		"documents",
		"downloads",
		"developer",
		"development",
		"projects",
		"repos",
		"repositories",
		"workspace",
		"workspaces",
		"code",
		"src",
		"applications",
		"library",
		"pictures",
		"movies",
		"music",
	}

	for _, d := range unpublishableDirs {
		if baseName == d {
			return true
		}
	}

	return false
}

// checkIfPublishable checks if a directory can be published
func checkIfPublishable(dir string) (hasKioskMd, hasGit bool) {
	// Check for KIOSK.md
	kioskPath := filepath.Join(dir, "KIOSK.md")
	if _, err := os.Stat(kioskPath); err == nil {
		hasKioskMd = true
	}

	// Check for .git directory
	gitPath := filepath.Join(dir, ".git")
	if info, err := os.Stat(gitPath); err == nil && info.IsDir() {
		hasGit = true
	}

	return hasKioskMd, hasGit
}

// loadDirectories loads subdirectories for the picker
func loadDirectories(dir string) []directoryItem {
	var items []directoryItem

	// Add parent directory option if not at root
	if dir != "/" {
		items = append(items, directoryItem{
			name:     "..",
			path:     filepath.Dir(dir),
			isParent: true,
		})
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return items
	}

	var dirs []directoryItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden directories
		if strings.HasPrefix(name, ".") {
			continue
		}
		dirs = append(dirs, directoryItem{
			name: name,
			path: filepath.Join(dir, name),
		})
	}

	// Sort directories alphabetically
	sort.Slice(dirs, func(i, j int) bool {
		return strings.ToLower(dirs[i].name) < strings.ToLower(dirs[j].name)
	})

	items = append(items, dirs...)
	return items
}

// Messages
type publishCheckResultMsg struct {
	dir             string
	isUnpublishable bool
	err             error
}

// Update handles messages for the publish view
func (m *PublishModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case PublishStatePublishable:
			switch {
			case key.Matches(msg, m.keys.Back):
				// If we came from directory picker, go back to it
				if len(m.dirHistory) > 0 {
					// Pop from history and go back to directory picker
					prevDir := m.dirHistory[len(m.dirHistory)-1]
					m.dirHistory = m.dirHistory[:len(m.dirHistory)-1]
					m.currentDir = prevDir
					m.directories = loadDirectories(prevDir)
					m.cursor = 0
					m.state = PublishStatePickDirectory
					m.isPublishable = false
				} else {
					return m, func() tea.Msg { return tui.GoBackMsg{} }
				}
			case key.Matches(msg, m.keys.Up), msg.String() == "left":
				if m.confirmCursor > 0 {
					m.confirmCursor--
				}
			case key.Matches(msg, m.keys.Down), msg.String() == "right":
				if m.confirmCursor < 1 {
					m.confirmCursor++
				}
			case key.Matches(msg, m.keys.Enter):
				if m.confirmCursor == 0 {
					// Yes - publish
					// TODO: Implement actual publish flow
					return m, func() tea.Msg { return tui.GoBackMsg{} }
				} else {
					// No - go back to directory picker if we have history
					if len(m.dirHistory) > 0 {
						prevDir := m.dirHistory[len(m.dirHistory)-1]
						m.dirHistory = m.dirHistory[:len(m.dirHistory)-1]
						m.currentDir = prevDir
						m.directories = loadDirectories(prevDir)
						m.cursor = 0
						m.state = PublishStatePickDirectory
						m.isPublishable = false
					} else {
						return m, func() tea.Msg { return tui.GoBackMsg{} }
					}
				}
			}

		case PublishStatePickDirectory:
			switch {
			case key.Matches(msg, m.keys.Back):
				// If we have history, go back to previous directory
				if len(m.dirHistory) > 0 {
					// Pop from history
					prevDir := m.dirHistory[len(m.dirHistory)-1]
					m.dirHistory = m.dirHistory[:len(m.dirHistory)-1]
					m.currentDir = prevDir
					m.directories = loadDirectories(prevDir)
					m.cursor = 0
				} else {
					// No history, exit the view
					return m, func() tea.Msg { return tui.GoBackMsg{} }
				}
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.directories)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.Enter):
				if m.cursor < len(m.directories) {
					selected := m.directories[m.cursor]

					// Handle parent directory navigation
					if selected.isParent {
						// Going up doesn't add to history, just navigate
						m.currentDir = selected.path
						m.directories = loadDirectories(selected.path)
						m.cursor = 0
					} else {
						// Push current directory to history before navigating
						m.dirHistory = append(m.dirHistory, m.currentDir)
						m.currentDir = selected.path
						m.directories = loadDirectories(selected.path)
						m.cursor = 0

						// Check if the selected directory is publishable
						hasKioskMd, hasGit := checkIfPublishable(selected.path)
						if hasKioskMd || hasGit {
							m.hasKioskMd = hasKioskMd
							m.hasGit = hasGit
							m.projectName = filepath.Base(selected.path)
							m.isPublishable = true
							m.state = PublishStatePublishable
							m.confirmCursor = 0
						}
					}
				}
			}

		case PublishStateNotPublishable:
			switch {
			case key.Matches(msg, m.keys.Back), key.Matches(msg, m.keys.Enter):
				return m, func() tea.Msg { return tui.GoBackMsg{} }
			}
		}

	case spinner.TickMsg:
		if m.state == PublishStateChecking {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case publishCheckResultMsg:
		if msg.err != nil {
			m.state = PublishStateNotPublishable
			return m, nil
		}

		m.currentDir = msg.dir
		m.startDir = msg.dir
		m.projectName = filepath.Base(msg.dir)
		m.dirHistory = []string{} // Reset history

		if msg.isUnpublishable {
			// Show directory picker
			m.state = PublishStatePickDirectory
			m.directories = loadDirectories(msg.dir)
			m.cursor = 0
		} else {
			// Check if current directory is publishable
			m.hasKioskMd, m.hasGit = checkIfPublishable(msg.dir)
			if m.hasKioskMd || m.hasGit {
				m.isPublishable = true
				m.state = PublishStatePublishable
				m.confirmCursor = 0
			} else {
				m.state = PublishStateNotPublishable
			}
		}
	}

	return m, nil
}

// View renders the publish view
func (m *PublishModel) View() string {
	switch m.state {
	case PublishStateChecking:
		return m.checkingView()
	case PublishStatePublishable:
		return m.publishableView()
	case PublishStatePickDirectory:
		return m.pickDirectoryView()
	case PublishStateNotPublishable:
		return m.notPublishableView()
	default:
		return ""
	}
}

func (m *PublishModel) checkingView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Publish App"))
	b.WriteString("\n\n")

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(styles.MutedStyle.Render("Checking current directory..."))

	return b.String()
}

func (m *PublishModel) publishableView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Publish App"))
	b.WriteString("\n")

	// Show project info
	contentStyle := lipgloss.NewStyle().
		Foreground(styles.Foreground).
		MaxWidth(contentWidth)

	b.WriteString(contentStyle.Render("Found project: "))
	b.WriteString(styles.Highlight.Render(m.projectName))
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render(m.currentDir))
	b.WriteString("\n\n")

	// Show status - align checkmarks on left
	if m.hasKioskMd {
		b.WriteString(styles.SuccessStyle.Render("✓ "))
		b.WriteString("KIOSK.md found")
		b.WriteString("\n")
	} else {
		b.WriteString(styles.WarningStyle.Render("○ "))
		b.WriteString(styles.MutedStyle.Render("No KIOSK.md (will be created)"))
		b.WriteString("\n")
	}

	if m.hasGit {
		b.WriteString(styles.SuccessStyle.Render("✓ "))
		b.WriteString("Git repository")
		b.WriteString("\n")
	}

	b.WriteString("\n")

	// Confirmation prompt
	b.WriteString(contentStyle.Render("Publish this app to Kiosk?"))
	b.WriteString("\n\n")

	// Yes/No buttons
	yesStyle := lipgloss.NewStyle().Padding(0, 2)
	noStyle := lipgloss.NewStyle().Padding(0, 2)

	if m.confirmCursor == 0 {
		yesStyle = yesStyle.
			Background(styles.Primary).
			Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		yesStyle = yesStyle.Foreground(styles.Muted)
	}

	if m.confirmCursor == 1 {
		noStyle = noStyle.
			Background(styles.Primary).
			Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		noStyle = noStyle.Foreground(styles.Muted)
	}

	b.WriteString("  ")
	b.WriteString(yesStyle.Render("Yes"))
	b.WriteString("  ")
	b.WriteString(noStyle.Render("No"))
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("←/→ select • enter confirm • esc cancel"))

	return b.String()
}

func (m *PublishModel) pickDirectoryView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Publish App"))
	b.WriteString("\n\n")

	// Current location
	b.WriteString(styles.MutedStyle.Render("Select a project directory:"))
	b.WriteString("\n")
	b.WriteString(styles.MutedStyle.Render(m.currentDir))
	b.WriteString("\n\n")

	// Directory list
	visibleItems := m.height - 10
	if visibleItems < 5 {
		visibleItems = 5
	}

	startIdx := 0
	if m.cursor >= visibleItems {
		startIdx = m.cursor - visibleItems + 1
	}

	endIdx := startIdx + visibleItems
	if endIdx > len(m.directories) {
		endIdx = len(m.directories)
	}

	for i := startIdx; i < endIdx; i++ {
		dir := m.directories[i]

		cursor := "  "
		itemStyle := lipgloss.NewStyle().Foreground(styles.Foreground)

		if i == m.cursor {
			cursor = styles.Highlight.Render("> ")
			itemStyle = itemStyle.Foreground(styles.Primary)
		}

		displayName := dir.name
		if dir.isParent {
			displayName = ".. (parent directory)"
		} else {
			displayName = "📁 " + displayName
			// Check if this directory looks like a project
			hasKiosk, hasGit := checkIfPublishable(dir.path)
			if hasKiosk {
				displayName += styles.SuccessStyle.Render(" [KIOSK.md]")
			} else if hasGit {
				displayName += styles.MutedStyle.Render(" [git]")
			}
		}

		b.WriteString(cursor)
		b.WriteString(itemStyle.Render(displayName))
		b.WriteString("\n")
	}

	// Scroll indicator
	if len(m.directories) > visibleItems {
		b.WriteString("\n")
		b.WriteString(styles.MutedStyle.Render(
			"  ... and more (scroll with ↑/↓)"))
	}

	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("↑/↓ navigate • enter select • esc cancel"))

	return b.String()
}

func (m *PublishModel) notPublishableView() string {
	var b strings.Builder

	contentWidth := m.width
	if contentWidth <= 0 {
		contentWidth = 80
	}

	titleStyle := styles.Title.Copy().MaxWidth(contentWidth)
	b.WriteString(titleStyle.Render("Publish App"))
	b.WriteString("\n\n")

	b.WriteString(styles.WarningStyle.Render("This directory doesn't appear to be a publishable project."))
	b.WriteString("\n\n")

	b.WriteString(styles.MutedStyle.Render(m.currentDir))
	b.WriteString("\n\n")

	contentStyle := lipgloss.NewStyle().
		Foreground(styles.Muted).
		MaxWidth(contentWidth)

	b.WriteString(contentStyle.Render("To publish an app, navigate to a project directory that contains:"))
	b.WriteString("\n")
	b.WriteString(contentStyle.Render("  • A KIOSK.md file, or"))
	b.WriteString("\n")
	b.WriteString(contentStyle.Render("  • A .git directory"))
	b.WriteString("\n\n")

	hintStyle := lipgloss.NewStyle().
		Foreground(styles.Secondary).
		MaxWidth(contentWidth)
	b.WriteString(hintStyle.Render("Tip: Run 'kiosk publish' from your project directory"))
	b.WriteString("\n\n")

	// Help
	b.WriteString(styles.HelpStyle.Copy().MaxWidth(contentWidth).Render("enter or esc to go back"))

	return b.String()
}
