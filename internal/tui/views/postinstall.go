package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// PostInstallState represents the state of the post-install flow
type PostInstallState int

const (
	PostInstallStateCloning PostInstallState = iota
	PostInstallStateInstalling
	PostInstallStateReady
	PostInstallStateRunning
	PostInstallStateError
)

// PostInstallOption represents an option in the post-install menu
type PostInstallOption struct {
	Title       string
	Description string
	Command     string
	Icon        string
}

// PostInstallModel is the model for the post-install view
type PostInstallModel struct {
	width    int
	height   int
	keys     tui.KeyMap
	state    PostInstallState
	spinner  spinner.Model
	progress progress.Model
	appName  string
	appKey   string
	appPath  string
	error    error

	// Cloning progress
	cloneProgress float64
	cloneMessage  string

	// Post-install options
	cursor  int
	options []PostInstallOption
}

// NewPostInstallModel creates a new post-install model
func NewPostInstallModel(appName, appKey, appPath string) PostInstallModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)

	options := []PostInstallOption{
		{
			Title:       "Start Development Server",
			Description: "Run the app in development mode with hot reload",
			Command:     "dev",
			Icon:        "🚀",
		},
		{
			Title:       "Build for Production",
			Description: "Create an optimized production build",
			Command:     "build",
			Icon:        "📦",
		},
		{
			Title:       "Run Tests",
			Description: "Execute the test suite",
			Command:     "test",
			Icon:        "🧪",
		},
		{
			Title:       "Open in Editor",
			Description: "Open the project in your default editor",
			Command:     "edit",
			Icon:        "✏️",
		},
		{
			Title:       "Explore with AI",
			Description: "Let Claude help you understand and modify the code",
			Command:     "claude",
			Icon:        "🤖",
		},
	}

	return PostInstallModel{
		keys:     tui.DefaultKeyMap(),
		state:    PostInstallStateCloning,
		spinner:  s,
		progress: p,
		appName:  appName,
		appKey:   appKey,
		appPath:  appPath,
		options:  options,
	}
}

// SetSize updates the view dimensions
func (m *PostInstallModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// Init initializes the post-install model
func (m *PostInstallModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages for the post-install view
func (m *PostInstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case PostInstallStateReady:
			switch {
			case key.Matches(msg, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}
			case key.Matches(msg, m.keys.Down):
				if m.cursor < len(m.options)-1 {
					m.cursor++
				}
			case key.Matches(msg, m.keys.Enter):
				// Execute the selected option
				m.state = PostInstallStateRunning
				return m, m.executeOption(m.options[m.cursor])
			case key.Matches(msg, m.keys.Back):
				return m, func() tea.Msg { return tui.GoBackMsg{} }
			}
		case PostInstallStateError:
			if key.Matches(msg, m.keys.Back) || key.Matches(msg, m.keys.Enter) {
				return m, func() tea.Msg { return tui.GoBackMsg{} }
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		cmds = append(cmds, cmd)

	case tui.CloneProgressMsg:
		m.cloneProgress = float64(msg.Percent) / 100.0
		m.cloneMessage = msg.Message

	case tui.CloneCompleteMsg:
		if msg.Err != nil {
			m.state = PostInstallStateError
			m.error = msg.Err
		} else {
			m.state = PostInstallStateInstalling
		}

	case tui.AppInstalledMsg:
		if msg.Err != nil {
			m.state = PostInstallStateError
			m.error = msg.Err
		} else {
			m.state = PostInstallStateReady
		}

	case tui.ErrorMsg:
		m.state = PostInstallStateError
		m.error = msg.Err

	case tui.SuccessMsg:
		// Option executed successfully, return to ready state
		m.state = PostInstallStateReady
	}

	return m, tea.Batch(cmds...)
}

func (m *PostInstallModel) executeOption(opt PostInstallOption) tea.Cmd {
	// This would normally execute the command
	// For now, we just return a success message
	return func() tea.Msg {
		return tui.SuccessMsg{Message: fmt.Sprintf("Executing: %s", opt.Command)}
	}
}

// SetState allows external code to set the state
func (m *PostInstallModel) SetState(state PostInstallState) {
	m.state = state
}

// SetCloneProgress updates the clone progress
func (m *PostInstallModel) SetCloneProgress(percent int, message string) {
	m.cloneProgress = float64(percent) / 100.0
	m.cloneMessage = message
}

// View renders the post-install view
func (m *PostInstallModel) View() string {
	var b strings.Builder

	// App name header
	titleStyle := styles.Title.Copy().MarginBottom(1)
	b.WriteString(titleStyle.Render(fmt.Sprintf("Installing %s", m.appName)))
	b.WriteString("\n\n")

	switch m.state {
	case PostInstallStateCloning:
		b.WriteString(m.cloningView())

	case PostInstallStateInstalling:
		b.WriteString(m.installingView())

	case PostInstallStateReady:
		b.WriteString(m.readyView())

	case PostInstallStateRunning:
		b.WriteString(m.runningView())

	case PostInstallStateError:
		b.WriteString(m.errorView())
	}

	return b.String()
}

func (m PostInstallModel) cloningView() string {
	var b strings.Builder

	// Step indicator
	stepStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	b.WriteString(stepStyle.Render("Step 1 of 2"))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(m.spinner.View())
	b.WriteString(" Cloning repository...")
	b.WriteString("\n\n")

	// Progress bar
	b.WriteString(m.progress.ViewAs(m.cloneProgress))
	b.WriteString("\n")

	if m.cloneMessage != "" {
		b.WriteString(styles.MutedStyle.Render(m.cloneMessage))
	}

	return b.String()
}

func (m PostInstallModel) installingView() string {
	var b strings.Builder

	// Step indicator
	stepStyle := lipgloss.NewStyle().Foreground(styles.Muted)
	b.WriteString(stepStyle.Render("Step 2 of 2"))
	b.WriteString("\n\n")

	// Progress
	b.WriteString(m.spinner.View())
	b.WriteString(" Installing dependencies...")
	b.WriteString("\n\n")

	b.WriteString(styles.MutedStyle.Render("Claude is setting up your environment..."))

	return b.String()
}

func (m PostInstallModel) readyView() string {
	var b strings.Builder

	// Success indicator
	successIcon := styles.SuccessStyle.Render("✓")
	b.WriteString(successIcon)
	b.WriteString(" ")
	b.WriteString(styles.SuccessStyle.Render("Installation complete!"))
	b.WriteString("\n\n")

	// App location
	b.WriteString(styles.MutedStyle.Render("Location: "))
	b.WriteString(styles.Code.Render(m.appPath))
	b.WriteString("\n\n")

	// Divider
	divider := lipgloss.NewStyle().
		Foreground(styles.Muted).
		Render(strings.Repeat("─", 40))
	b.WriteString(divider)
	b.WriteString("\n\n")

	// Question
	questionStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.Foreground)
	b.WriteString(questionStyle.Render("What would you like to do next?"))
	b.WriteString("\n\n")

	// Options
	for i, opt := range m.options {
		cursor := "  "
		titleStyle := lipgloss.NewStyle().Foreground(styles.Foreground)
		descStyle := lipgloss.NewStyle().Foreground(styles.Muted)

		if i == m.cursor {
			cursor = styles.Highlight.Render("> ")
			titleStyle = titleStyle.Bold(true).Foreground(styles.Primary)
		}

		b.WriteString(cursor)
		b.WriteString(opt.Icon)
		b.WriteString(" ")
		b.WriteString(titleStyle.Render(opt.Title))
		b.WriteString("\n")
		b.WriteString("      ")
		b.WriteString(descStyle.Render(opt.Description))
		b.WriteString("\n\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(styles.HelpStyle.Render("↑/↓ navigate • enter select • esc go back"))

	return b.String()
}

func (m PostInstallModel) runningView() string {
	var b strings.Builder

	b.WriteString(m.spinner.View())
	b.WriteString(" ")
	b.WriteString(fmt.Sprintf("Running %s...", m.options[m.cursor].Title))

	return b.String()
}

func (m PostInstallModel) errorView() string {
	var b strings.Builder

	errorIcon := styles.ErrorStyle.Render("✗")
	b.WriteString(errorIcon)
	b.WriteString(" ")
	b.WriteString(styles.ErrorStyle.Render("Installation failed"))
	b.WriteString("\n\n")

	if m.error != nil {
		b.WriteString("  ")
		b.WriteString(styles.MutedStyle.Render(m.error.Error()))
	}

	b.WriteString("\n\n")
	b.WriteString(styles.HelpStyle.Render("Press enter or esc to go back"))

	return b.String()
}
