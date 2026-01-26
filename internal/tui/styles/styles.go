package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors
var (
	Primary    = lipgloss.Color("#3B5BF7") // Blue
	Secondary  = lipgloss.Color("#35C1FF") // Light blue
	Success    = lipgloss.Color("#10B981") // Green
	Warning    = lipgloss.Color("#F59E0B") // Amber
	Error      = lipgloss.Color("#EF4444") // Red
	Muted      = lipgloss.Color("#6B7280") // Gray
	Background = lipgloss.Color("#1F2937") // Dark gray
	Foreground = lipgloss.Color("#F9FAFB") // Light gray
)

// Common styles
var (
	// Title styling
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(Primary).
		MarginBottom(1)

	// Subtitle styling
	Subtitle = lipgloss.NewStyle().
			Foreground(Muted).
			Italic(true)

	// Highlighted text
	Highlight = lipgloss.NewStyle().
			Foreground(Secondary).
			Bold(true)

	// Success message
	SuccessStyle = lipgloss.NewStyle().
			Foreground(Success)

	// Warning message
	WarningStyle = lipgloss.NewStyle().
			Foreground(Warning)

	// Error message
	ErrorStyle = lipgloss.NewStyle().
			Foreground(Error).
			Bold(true)

	// Muted/dim text
	MutedStyle = lipgloss.NewStyle().
			Foreground(Muted)

	// Bold text
	Bold = lipgloss.NewStyle().
		Bold(true)

	// Help text at bottom
	HelpStyle = lipgloss.NewStyle().
			Foreground(Muted).
			MarginTop(1)

	// Box styling for panels
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Primary).
		Padding(1, 2)

	// Active/focused box
	ActiveBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Secondary).
			Padding(1, 2)

	// Status bar at bottom
	StatusBar = lipgloss.NewStyle().
			Background(Background).
			Foreground(Foreground).
			Padding(0, 1)

	// Code/monospace text
	Code = lipgloss.NewStyle().
		Foreground(Secondary).
		Background(lipgloss.Color("#374151")).
		Padding(0, 1)

	// App name in list
	AppName = lipgloss.NewStyle().
		Bold(true).
		Foreground(Foreground)

	// App description
	AppDescription = lipgloss.NewStyle().
			Foreground(Muted)

	// Selected list item
	SelectedItem = lipgloss.NewStyle().
			Foreground(Primary).
			Bold(true)

	// Unselected list item
	NormalItem = lipgloss.NewStyle().
			Foreground(Foreground)
)

// Logo is the ASCII art kiosk text logo
const Logo = `.++++.   .=++++= .++++    .-++*****++-.    .-++*****++-.   ++++.   .=++++=          
:****: .+****+.  :****. .+*****+++*****+..-****+=-=*****= .****: .=****+:.          
:****:+****=.    :****..*****:     :****+.+***+:.. ..:::. .****-=****+.             
:**********:     :****..****=       -****.:***********+=. .**********-              
:*****++****+.   :****..****=.      =****.   .:--==+*****-.*****++*****.            
:****: .=*****:. :****. -****=.. .:+****-.****+..  .+****=.****:  -*****-.          
:****.   .*****+.:****. .:+***********+.  :+************- .****:   .+****+.`

// LogoStyled returns the logo with styling applied
func LogoStyled() string {
	return lipgloss.NewStyle().
		Foreground(Primary).
		Render(Logo)
}
