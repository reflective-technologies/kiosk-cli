package cmd

import (
	"fmt"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/style"
)

// appItem represents an app in the TUI list
type appItem struct {
	key    string // org/repo
	name   string
	author string
}

// tuiState represents the current state of the TUI
type tuiState int

const (
	stateSelecting tuiState = iota
	stateChecking
	stateConfirming
	stateRunning
)

// model is the Bubble Tea model for the interactive TUI
type model struct {
	apps       []appItem
	cursor     int
	state      tuiState
	hasUpdate  bool
	behind     int
	err        error
	quitting   bool
	styler     *style.Styler
	appPath    string // path to selected app for running
	selectedKey string
}

// checkUpdateMsg is sent when update check completes
type checkUpdateMsg struct {
	hasUpdate bool
	behind    int
	err       error
}

// updateAppliedMsg is sent when update is applied
type updateAppliedMsg struct {
	info *updateInfo
	err  error
}

// runInteractiveTUI launches the interactive app selector
func runInteractiveTUI() error {
	idx, err := appindex.Load()
	if err != nil {
		return fmt.Errorf("failed to load app index: %w", err)
	}

	if idx.Count() == 0 {
		fmt.Println("No apps installed.")
		fmt.Println()
		fmt.Println("Install an app with:")
		fmt.Println("  kiosk run <app>")
		fmt.Println()
		fmt.Println("Browse apps at: https://kiosk.app")
		return nil
	}

	// Build app list
	keys := idx.List()
	sort.Strings(keys)

	apps := make([]appItem, 0, len(keys))
	for _, key := range keys {
		author, name := splitAppKey(key)
		apps = append(apps, appItem{
			key:    key,
			name:   name,
			author: author,
		})
	}

	m := model{
		apps:   apps,
		cursor: 0,
		state:  stateSelecting,
		styler: style.Stdout(),
	}

	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	// Check if we need to run an app after TUI exits
	fm := finalModel.(model)
	if fm.state == stateRunning && fm.appPath != "" {
		fmt.Printf("Running %s...\n", fm.selectedKey)
		fmt.Print(logo)
		return execClaude(fm.appPath, runPrompt, false)
	}

	return nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case checkUpdateMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateSelecting
			return m, nil
		}
		m.hasUpdate = msg.hasUpdate
		m.behind = msg.behind
		if m.hasUpdate {
			m.state = stateConfirming
		} else {
			// No updates, run immediately
			m.state = stateRunning
			m.quitting = true
			return m, tea.Quit
		}
		return m, nil

	case updateAppliedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = stateSelecting
			return m, nil
		}
		// Update applied, now run
		m.state = stateRunning
		m.quitting = true
		return m, tea.Quit
	}

	return m, nil
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateSelecting:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.apps)-1 {
				m.cursor++
			}
		case "enter":
			if len(m.apps) > 0 {
				selected := m.apps[m.cursor]
				m.selectedKey = selected.key
				parts := strings.SplitN(selected.key, "/", 2)
				if len(parts) == 2 {
					m.appPath = config.AppPath(parts[0], parts[1])
					m.state = stateChecking
					m.err = nil
					return m, m.checkForUpdates()
				}
			}
		}

	case stateConfirming:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "y", "Y", "enter":
			// Apply update and run
			m.state = stateChecking
			return m, m.applyUpdates()
		case "n", "N":
			// Run without updating
			m.state = stateRunning
			m.quitting = true
			return m, tea.Quit
		}

	case stateChecking:
		// Allow quitting during check
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

func (m model) View() string {
	if m.quitting && m.state != stateRunning {
		return ""
	}

	var b strings.Builder

	switch m.state {
	case stateSelecting:
		b.WriteString(m.viewSelecting())
	case stateChecking:
		b.WriteString(m.viewChecking())
	case stateConfirming:
		b.WriteString(m.viewConfirming())
	case stateRunning:
		// Will be handled after TUI exits
		return ""
	}

	return b.String()
}

func (m model) viewSelecting() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.styler.Apply(style.Bold, "  Select an app:"))
	b.WriteString("\n\n")

	// Calculate max widths for alignment
	maxName := 0
	for _, app := range m.apps {
		if len(app.name) > maxName {
			maxName = len(app.name)
		}
	}

	for i, app := range m.apps {
		cursor := "  "
		if i == m.cursor {
			cursor = m.styler.Apply(style.Cyan, "> ")
		}

		name := app.name
		if i == m.cursor {
			name = m.styler.Apply(style.Bold, name)
		}

		author := m.styler.Apply(style.Dim, "("+app.author+")")

		b.WriteString(fmt.Sprintf("%s%s  %s\n", cursor, padRight(name, maxName), author))
	}

	b.WriteString("\n")
	b.WriteString(m.styler.Apply(style.Dim, "  ↑/↓ navigate  enter select  q quit"))
	b.WriteString("\n")

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(m.styler.Apply(style.Red, "  Error: "+m.err.Error()))
		b.WriteString("\n")
	}

	return b.String()
}

func (m model) viewChecking() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString("  Checking for updates...")
	b.WriteString("\n")
	return b.String()
}

func (m model) viewConfirming() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.styler.Apply(style.Bold, "  Update available"))
	b.WriteString(fmt.Sprintf(" for %s\n", m.selectedKey))
	b.WriteString(fmt.Sprintf("  %s commit(s) behind upstream\n", m.styler.Apply(style.Cyan, fmt.Sprintf("%d", m.behind))))
	b.WriteString("\n")
	b.WriteString("  Update and run? ")
	b.WriteString(m.styler.Apply(style.Dim, "[Y/n]"))
	b.WriteString("\n")
	return b.String()
}

// checkForUpdates returns a command that checks for updates
func (m model) checkForUpdates() tea.Cmd {
	return func() tea.Msg {
		hasUpdate, behind, err := CheckForUpdates(m.appPath)
		return checkUpdateMsg{
			hasUpdate: hasUpdate,
			behind:    behind,
			err:       err,
		}
	}
}

// applyUpdates returns a command that applies updates
func (m model) applyUpdates() tea.Cmd {
	return func() tea.Msg {
		info, err := ApplyUpdates(m.appPath)
		return updateAppliedMsg{
			info: info,
			err:  err,
		}
	}
}

// CheckForUpdates checks if an app has updates without applying them
func CheckForUpdates(appPath string) (hasUpdate bool, behind int, err error) {
	if _, err := exec.LookPath("git"); err != nil {
		return false, 0, nil
	}

	inside, err := gitOutput(appPath, "rev-parse", "--is-inside-work-tree")
	if err != nil || inside != "true" {
		return false, 0, nil
	}

	if err := gitRun(appPath, "fetch", "--quiet"); err != nil {
		return false, 0, nil
	}

	counts, err := gitOutput(appPath, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return false, 0, nil
	}

	parts := strings.Fields(counts)
	if len(parts) != 2 {
		return false, 0, nil
	}

	ahead, err := strconv.Atoi(parts[0])
	if err != nil {
		return false, 0, nil
	}
	behindCount, err := strconv.Atoi(parts[1])
	if err != nil {
		return false, 0, nil
	}

	if behindCount == 0 {
		return false, 0, nil
	}

	if ahead > 0 {
		return false, 0, fmt.Errorf("local branch has diverged from upstream; resolve manually before running")
	}

	return true, behindCount, nil
}

// ApplyUpdates pulls updates for an app
func ApplyUpdates(appPath string) (*updateInfo, error) {
	oldCommit, err := gitOutput(appPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, err
	}

	hasChanges := false
	status, err := gitOutput(appPath, "status", "--porcelain")
	if err == nil && strings.TrimSpace(status) != "" {
		hasChanges = true
		if err := gitRun(appPath, "stash", "push", "-u", "-m", "kiosk: pre-update stash"); err != nil {
			return nil, fmt.Errorf("failed to stash local changes: %w", err)
		}
	}

	if err := gitRun(appPath, "pull", "--ff-only"); err != nil {
		if hasChanges {
			_ = gitRun(appPath, "stash", "pop")
		}
		return nil, err
	}

	newCommit, err := gitOutput(appPath, "rev-parse", "HEAD")
	if err != nil {
		if hasChanges {
			_ = gitRun(appPath, "stash", "pop")
		}
		return nil, err
	}

	unstashConflicts := false
	if hasChanges {
		if err := gitRun(appPath, "stash", "pop"); err != nil {
			unstashConflicts = true
		}
	}

	return &updateInfo{
		updated:          true,
		oldCommit:        oldCommit,
		newCommit:        newCommit,
		hadStash:         hasChanges,
		unstashConflicts: unstashConflicts,
	}, nil
}
