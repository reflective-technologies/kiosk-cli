package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// StatusBar renders a status bar at the bottom of the screen
type StatusBar struct {
	width       int
	leftText    string
	rightText   string
	centerText  string
	showSpinner bool
	spinner     SpinnerModel
}

// NewStatusBar creates a new status bar
func NewStatusBar(width int) StatusBar {
	return StatusBar{
		width:   width,
		spinner: NewSpinner(""),
	}
}

// SetWidth updates the status bar width
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// SetLeft sets the left-aligned text
func (s *StatusBar) SetLeft(text string) {
	s.leftText = text
}

// SetRight sets the right-aligned text
func (s *StatusBar) SetRight(text string) {
	s.rightText = text
}

// SetCenter sets the center-aligned text
func (s *StatusBar) SetCenter(text string) {
	s.centerText = text
}

// SetSpinner enables/disables the spinner
func (s *StatusBar) SetSpinner(show bool, message string) {
	s.showSpinner = show
	s.spinner.SetMessage(message)
}

// View renders the status bar
func (s StatusBar) View() string {
	leftStyle := lipgloss.NewStyle().
		Foreground(styles.Foreground).
		Bold(true)

	rightStyle := lipgloss.NewStyle().
		Foreground(styles.Muted)

	centerStyle := lipgloss.NewStyle().
		Foreground(styles.Primary)

	barStyle := lipgloss.NewStyle().
		Background(styles.Background).
		Width(s.width).
		Padding(0, 1)

	left := leftStyle.Render(s.leftText)
	right := rightStyle.Render(s.rightText)
	center := centerStyle.Render(s.centerText)

	if s.showSpinner {
		center = s.spinner.View()
	}

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	centerWidth := lipgloss.Width(center)

	// Create spacing between sections
	totalContentWidth := leftWidth + centerWidth + rightWidth
	if totalContentWidth < s.width-2 {
		leftPadding := (s.width - 2 - totalContentWidth) / 2
		rightPadding := s.width - 2 - totalContentWidth - leftPadding

		paddingStyle := lipgloss.NewStyle().Width(leftPadding)
		rightPaddingStyle := lipgloss.NewStyle().Width(rightPadding)

		return barStyle.Render(left + paddingStyle.Render("") + center + rightPaddingStyle.Render("") + right)
	}

	return barStyle.Render(left + " " + center + " " + right)
}
