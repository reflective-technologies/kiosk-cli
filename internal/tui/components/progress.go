package components

import (
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// ProgressModel is a wrapper around the bubbles progress bar
type ProgressModel struct {
	progress progress.Model
	message  string
	percent  float64
}

// NewProgress creates a new progress bar
func NewProgress() ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(40),
	)
	p.FullColor = string(styles.Primary)
	p.EmptyColor = string(styles.Muted)

	return ProgressModel{
		progress: p,
		message:  "",
		percent:  0,
	}
}

// NewProgressWithWidth creates a progress bar with custom width
func NewProgressWithWidth(width int) ProgressModel {
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(width),
	)
	p.FullColor = string(styles.Primary)
	p.EmptyColor = string(styles.Muted)

	return ProgressModel{
		progress: p,
		message:  "",
		percent:  0,
	}
}

// SetPercent updates the progress percentage (0-1)
func (m *ProgressModel) SetPercent(percent float64) {
	m.percent = percent
}

// SetMessage updates the progress message
func (m *ProgressModel) SetMessage(msg string) {
	m.message = msg
}

// Init initializes the progress bar
func (m ProgressModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	var cmd tea.Cmd
	progressModel, cmd := m.progress.Update(msg)
	m.progress = progressModel.(progress.Model)
	return m, cmd
}

// View renders the progress bar
func (m ProgressModel) View() string {
	bar := m.progress.ViewAs(m.percent)
	if m.message != "" {
		msgStyle := lipgloss.NewStyle().Foreground(styles.Muted)
		return bar + "\n" + msgStyle.Render(m.message)
	}
	return bar
}

// ViewCompact renders just the progress bar without the message
func (m ProgressModel) ViewCompact() string {
	return m.progress.ViewAs(m.percent)
}
