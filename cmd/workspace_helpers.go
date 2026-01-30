package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/sessions"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
	"golang.org/x/term"
)

var (
	errUserCanceled        = errors.New("user canceled")
	errNoWorkspaceProjects = errors.New("no workspace projects")
)

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdin.Fd()))
}

func resolveProjectName(initial, workspaceDir string) (string, error) {
	name := strings.TrimSpace(initial)
	errMsg := ""

	for {
		if name == "" {
			var err error
			name, err = promptProjectName(errMsg)
			if err != nil {
				return "", err
			}
			errMsg = ""
		}

		if name == "" {
			errMsg = "Project name is required."
			continue
		}

		suggested := kebabCase(name)
		if suggested == "" {
			errMsg = "Project name must include letters or numbers."
			name = ""
			continue
		}

		if suggested != name {
			ok, err := confirmKebabCase(name, suggested)
			if err != nil {
				return "", err
			}
			if !ok {
				name = ""
				continue
			}
			name = suggested
		}

		if projectExists(workspaceDir, name) {
			errMsg = fmt.Sprintf("A project named %q already exists.", name)
			name = ""
			continue
		}

		return name, nil
	}
}

func projectExists(workspaceDir, name string) bool {
	path := filepath.Join(workspaceDir, name)
	_, err := os.Stat(path)
	return err == nil
}

func kebabCase(input string) string {
	input = strings.TrimSpace(input)
	if input == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(input) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if b.Len() == 0 || lastDash {
			continue
		}
		b.WriteByte('-')
		lastDash = true
	}

	result := strings.Trim(b.String(), "-")
	return result
}

func promptProjectName(errMsg string) (string, error) {
	if !isInteractive() {
		return "", fmt.Errorf("project name required when not running interactively")
	}

	m := newProjectNameModel(errMsg)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	model, ok := finalModel.(*projectNameModel)
	if !ok {
		return "", fmt.Errorf("unexpected model")
	}
	if model.canceled {
		return "", errUserCanceled
	}

	return strings.TrimSpace(model.value), nil
}

func confirmKebabCase(original, suggested string) (bool, error) {
	if !isInteractive() {
		return false, fmt.Errorf("approval required when not running interactively")
	}

	m := newConfirmModel(
		"Use suggested name?",
		fmt.Sprintf("We converted %q to %q for a directory-friendly name.", original, suggested),
		fmt.Sprintf("Use %s", suggested),
		"Enter a new name",
	)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return false, fmt.Errorf("confirmation failed: %w", err)
	}

	model, ok := finalModel.(*confirmModel)
	if !ok {
		return false, fmt.Errorf("unexpected model")
	}
	if model.canceled {
		return false, errUserCanceled
	}
	return model.confirmed, nil
}

func selectWorkspaceProject(workspaceDir string) (string, error) {
	projects, err := listWorkspaceProjects(workspaceDir)
	if err != nil {
		return "", err
	}
	if len(projects) == 0 {
		return "", errNoWorkspaceProjects
	}
	if !isInteractive() {
		return "", fmt.Errorf("project name required when not running interactively")
	}

	m := newWorkspaceSelectModel(projects)
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("selection failed: %w", err)
	}

	model, ok := finalModel.(*workspaceSelectModel)
	if !ok {
		return "", fmt.Errorf("unexpected model")
	}
	if model.canceled {
		return "", errUserCanceled
	}
	if model.selected == "" {
		return "", errUserCanceled
	}
	return model.selected, nil
}

func listWorkspaceProjects(workspaceDir string) ([]string, error) {
	entries, err := os.ReadDir(workspaceDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read workspace: %w", err)
	}

	projects := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			projects = append(projects, entry.Name())
		}
	}

	sort.Strings(projects)
	return projects, nil
}

func runWorkspaceClaude(projectDir, projectName, prompt string) error {
	store, err := sessions.LoadWorkspace()
	if err != nil {
		return fmt.Errorf("load workspace sessions: %w", err)
	}

	sessionCfg := &claudeSessionConfig{
		Store: store,
	}

	return execClaudeSession(projectDir, prompt, false, projectName, sessionCfg)
}

type projectNameModel struct {
	input    textinput.Model
	errMsg   string
	value    string
	canceled bool
}

func newProjectNameModel(errMsg string) *projectNameModel {
	ti := textinput.New()
	ti.Placeholder = "my-kiosk-app"
	ti.Prompt = "> "
	ti.Focus()
	ti.CharLimit = 64
	ti.Width = 32

	return &projectNameModel{
		input:  ti,
		errMsg: errMsg,
	}
}

func (m *projectNameModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *projectNameModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			if value == "" {
				m.errMsg = "Project name is required."
				return m, nil
			}
			m.value = value
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *projectNameModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)

	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render("New Kiosk Project"))
	b.WriteString("\n\n")

	if m.errMsg != "" {
		b.WriteString("  ")
		b.WriteString(styles.ErrorStyle.Render(m.errMsg))
		b.WriteString("\n\n")
	}

	b.WriteString("  Project name\n")
	b.WriteString("  ")
	b.WriteString(m.input.View())
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render("enter submit | esc cancel"))
	b.WriteString("\n")

	return b.String()
}

type confirmModel struct {
	title     string
	message   string
	yesLabel  string
	noLabel   string
	cursor    int
	confirmed bool
	canceled  bool
}

func newConfirmModel(title, message, yesLabel, noLabel string) *confirmModel {
	return &confirmModel{
		title:    title,
		message:  message,
		yesLabel: yesLabel,
		noLabel:  noLabel,
		cursor:   0,
	}
}

func (m *confirmModel) Init() tea.Cmd {
	return nil
}

func (m *confirmModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.canceled = true
			return m, tea.Quit
		case "left", "h":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right", "l":
			if m.cursor < 1 {
				m.cursor++
			}
		case "enter":
			if m.cursor == 0 {
				m.confirmed = true
			}
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *confirmModel) View() string {
	var b strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)

	b.WriteString("\n  ")
	b.WriteString(titleStyle.Render(m.title))
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render(m.message))
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(m.renderButtons())
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(styles.MutedStyle.Render("left/right select | enter confirm | esc cancel"))
	b.WriteString("\n")

	return b.String()
}

func (m *confirmModel) renderButtons() string {
	yesStyle := lipgloss.NewStyle().Padding(0, 2)
	noStyle := lipgloss.NewStyle().Padding(0, 2)

	if m.cursor == 0 {
		yesStyle = yesStyle.Background(styles.Primary).Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		yesStyle = yesStyle.Foreground(styles.Muted)
	}

	if m.cursor == 1 {
		noStyle = noStyle.Background(styles.Muted).Foreground(lipgloss.Color("#FFFFFF"))
	} else {
		noStyle = noStyle.Foreground(styles.Muted)
	}

	return yesStyle.Render(m.yesLabel) + "  " + noStyle.Render(m.noLabel)
}

type workspaceItem struct {
	name string
}

func (i workspaceItem) Title() string       { return i.name }
func (i workspaceItem) Description() string { return "" }
func (i workspaceItem) FilterValue() string { return i.name }

type workspaceSelectModel struct {
	list     list.Model
	selected string
	canceled bool
}

func newWorkspaceSelectModel(projects []string) *workspaceSelectModel {
	items := make([]list.Item, 0, len(projects))
	for _, project := range projects {
		items = append(items, workspaceItem{name: project})
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0)
	l.Title = "Workspace Projects"
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(true)
	l.Styles.Title = lipgloss.NewStyle().Bold(true).Foreground(styles.Primary)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(styles.Primary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(styles.Secondary)

	return &workspaceSelectModel{list: l}
}

func (m *workspaceSelectModel) Init() tea.Cmd {
	return nil
}

func (m *workspaceSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc", "q":
			m.canceled = true
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(workspaceItem); ok {
				m.selected = item.name
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *workspaceSelectModel) View() string {
	return "\n" + m.list.View()
}
