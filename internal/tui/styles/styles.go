package styles

import (
	"github.com/charmbracelet/lipgloss"
)

// IsDarkMode indicates whether the terminal has a dark background
var IsDarkMode = lipgloss.HasDarkBackground()

// Colors - these are set based on terminal background
var (
	Primary    lipgloss.Color
	Secondary  lipgloss.Color
	Success    lipgloss.Color
	Warning    lipgloss.Color
	Error      lipgloss.Color
	Muted      lipgloss.Color
	Background lipgloss.Color
	Foreground lipgloss.Color
)

func init() {
	if IsDarkMode {
		// Dark mode colors
		Primary = lipgloss.Color("#3B5BF7")    // Blue
		Secondary = lipgloss.Color("#35C1FF")  // Light blue
		Success = lipgloss.Color("#10B981")    // Green
		Warning = lipgloss.Color("#F59E0B")    // Amber
		Error = lipgloss.Color("#EF4444")      // Red
		Muted = lipgloss.Color("#9CA3AF")      // Gray (lighter for dark bg)
		Background = lipgloss.Color("#1F2937") // Dark gray
		Foreground = lipgloss.Color("#F9FAFB") // Light gray
	} else {
		// Light mode colors
		Primary = lipgloss.Color("#2563EB")    // Darker blue for light bg
		Secondary = lipgloss.Color("#0284C7")  // Darker light blue
		Success = lipgloss.Color("#059669")    // Darker green
		Warning = lipgloss.Color("#D97706")    // Darker amber
		Error = lipgloss.Color("#DC2626")      // Darker red
		Muted = lipgloss.Color("#6B7280")      // Gray
		Background = lipgloss.Color("#F3F4F6") // Light gray
		Foreground = lipgloss.Color("#111827") // Dark gray (almost black)
	}

	// Re-initialize styles with the correct colors
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

	// Code/monospace text - use appropriate background for mode
	codeBg := lipgloss.Color("#374151") // dark mode
	if !IsDarkMode {
		codeBg = lipgloss.Color("#E5E7EB") // light mode
	}
	Code = lipgloss.NewStyle().
		Foreground(Secondary).
		Background(codeBg).
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
