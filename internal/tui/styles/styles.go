package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// Colors - using colors that work well on both light and dark backgrounds
// We avoid very light colors (like white) and very dark colors (like black)
// Instead using mid-range colors that have good contrast on both
var (
	Primary    = lipgloss.Color("#3B82F6") // Blue - visible on both
	Secondary  = lipgloss.Color("#0EA5E9") // Sky blue
	Success    = lipgloss.Color("#22C55E") // Green
	Warning    = lipgloss.Color("#F59E0B") // Amber
	Error      = lipgloss.Color("#EF4444") // Red
	Muted      = lipgloss.Color("#71717A") // Zinc gray - works on both
	Background = lipgloss.Color("")        // Use terminal default
	Foreground = lipgloss.Color("")        // Use terminal default (empty = inherit)
)

func init() {
	initStyles()
}

// Common styles - initialized in initStyles()
var (
	Title          lipgloss.Style
	Subtitle       lipgloss.Style
	Highlight      lipgloss.Style
	SuccessStyle   lipgloss.Style
	WarningStyle   lipgloss.Style
	ErrorStyle     lipgloss.Style
	MutedStyle     lipgloss.Style
	Bold           lipgloss.Style
	HelpStyle      lipgloss.Style
	Box            lipgloss.Style
	ActiveBox      lipgloss.Style
	StatusBar      lipgloss.Style
	Code           lipgloss.Style
	AppName        lipgloss.Style
	AppDescription lipgloss.Style
	SelectedItem   lipgloss.Style
	NormalItem     lipgloss.Style
)

func initStyles() {
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
}

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
