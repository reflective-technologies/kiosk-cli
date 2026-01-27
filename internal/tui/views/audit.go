package views

import (
	"bytes"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	kioskexec "github.com/reflective-technologies/kiosk-cli/internal/exec"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// AuditState represents the current state of the audit
type AuditState int

const (
	AuditStateInitial AuditState = iota
	AuditStateRunning
	AuditStateComplete
	AuditStateError
)

// AuditModel is the model for the audit view
type AuditModel struct {
	width    int
	height   int
	keys     tui.KeyMap
	state    AuditState
	spinner  spinner.Model
	viewport viewport.Model
	result   string
	error    error
	ready    bool
}

// NewAuditModel creates a new audit model
func NewAuditModel() AuditModel {
	s := spinner.New()
	s.Spinner = spinner.Globe
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return AuditModel{
		keys:    tui.DefaultKeyMap(),
		state:   AuditStateInitial,
		spinner: s,
	}
}

// SetSize updates the view dimensions
func (m *AuditModel) SetSize(width, height int) {
	m.width = width
	m.height = height

	headerHeight := 4
	footerHeight := 2
	verticalMargins := headerHeight + footerHeight

	if !m.ready {
		m.viewport = viewport.New(width, height-verticalMargins)
		m.viewport.HighPerformanceRendering = false
		m.ready = true
	} else {
		m.viewport.Width = width
		m.viewport.Height = height - verticalMargins
	}
}

// Init initializes the audit model
func (m *AuditModel) Init() tea.Cmd {
	// Reset state for re-entry (view models are reused across navigations)
	m.state = AuditStateInitial
	m.result = ""
	m.error = nil

	return tea.Batch(
		m.spinner.Tick,
		m.runAudit,
	)
}

func (m *AuditModel) runAudit() tea.Msg {
	cwd, err := os.Getwd()
	if err != nil {
		return tui.AuditCompleteMsg{Err: err}
	}

	cmd := kioskexec.ClaudeCmd("-p", kioskexec.AuditPrompt)
	cmd.Dir = cwd

	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return tui.AuditCompleteMsg{Err: err}
	}

	return tui.AuditCompleteMsg{Result: stdout.String()}
}

// Update handles messages for the audit view
func (m *AuditModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		case key.Matches(msg, m.keys.Quit):
			return m, func() tea.Msg { return tui.GoBackMsg{} }
		}

	case spinner.TickMsg:
		if m.state == AuditStateInitial || m.state == AuditStateRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case tui.AuditStartedMsg:
		m.state = AuditStateRunning

	case tui.AuditCompleteMsg:
		if msg.Err != nil {
			m.state = AuditStateError
			m.error = msg.Err
		} else {
			m.state = AuditStateComplete
			m.result = msg.Result

			// Render markdown
			rendered, err := m.renderMarkdown(msg.Result)
			if err == nil {
				m.result = rendered
			}

			m.viewport.SetContent(m.result)
		}
	}

	// Update viewport for scrolling
	if m.state == AuditStateComplete {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AuditModel) renderMarkdown(content string) (string, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(m.width-4),
	)
	if err != nil {
		return content, err
	}

	return renderer.Render(content)
}

// View renders the audit view
func (m *AuditModel) View() string {
	var b strings.Builder

	titleStyle := styles.Title.Copy().MarginBottom(1)
	b.WriteString(titleStyle.Render("Security Audit"))
	b.WriteString("\n")

	switch m.state {
	case AuditStateInitial, AuditStateRunning:
		b.WriteString("\n")
		b.WriteString(m.spinner.View())
		b.WriteString(" ")
		b.WriteString(styles.MutedStyle.Render("Running security audit..."))
		b.WriteString("\n\n")
		b.WriteString(styles.MutedStyle.Render("This may take a minute. Scanning for secrets, credentials, and sensitive data."))

	case AuditStateComplete:
		b.WriteString("\n")
		b.WriteString(m.viewport.View())

	case AuditStateError:
		b.WriteString("\n")
		errorIcon := styles.ErrorStyle.Render("✗")
		b.WriteString(errorIcon)
		b.WriteString(" ")
		b.WriteString(styles.ErrorStyle.Render("Audit failed"))
		b.WriteString("\n\n")
		if m.error != nil {
			b.WriteString("  ")
			b.WriteString(styles.MutedStyle.Render(m.error.Error()))
		}
	}

	b.WriteString("\n\n")
	helpStyle := styles.HelpStyle
	if m.state == AuditStateComplete {
		scrollPercent := int(m.viewport.ScrollPercent() * 100)
		b.WriteString(helpStyle.Render("↑/↓ scroll • esc back • " + strconv.Itoa(scrollPercent) + "%"))
	} else {
		b.WriteString(helpStyle.Render("Press esc to cancel"))
	}

	return b.String()
}
