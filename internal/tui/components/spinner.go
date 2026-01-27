package components

import (
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// SpinnerModel is a wrapper around the bubbles spinner
type SpinnerModel struct {
	spinner spinner.Model
	message string
	style   lipgloss.Style
}

// NewSpinner creates a new spinner with the given message
func NewSpinner(message string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return SpinnerModel{
		spinner: s,
		message: message,
		style:   lipgloss.NewStyle().Foreground(styles.Muted),
	}
}

// NewSpinnerWithStyle creates a spinner with custom spinner style
func NewSpinnerWithStyle(message string, spinnerStyle spinner.Spinner) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinnerStyle
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return SpinnerModel{
		spinner: s,
		message: message,
		style:   lipgloss.NewStyle().Foreground(styles.Muted),
	}
}

// SetMessage updates the spinner message
func (m *SpinnerModel) SetMessage(msg string) {
	m.message = msg
}

// Init initializes the spinner
func (m SpinnerModel) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

// View renders the spinner
func (m SpinnerModel) View() string {
	return m.spinner.View() + " " + m.style.Render(m.message)
}

// Tick returns a tick command for the spinner
func (m SpinnerModel) Tick() tea.Cmd {
	return m.spinner.Tick
}

// Common spinner styles
var (
	SpinnerDot     = spinner.Dot
	SpinnerLine    = spinner.Line
	SpinnerMiniDot = spinner.MiniDot
	SpinnerJump    = spinner.Jump
	SpinnerPulse   = spinner.Pulse
	SpinnerPoints  = spinner.Points
	SpinnerGlobe   = spinner.Globe
	SpinnerMoon    = spinner.Moon
	SpinnerMonkey  = spinner.Monkey
)
