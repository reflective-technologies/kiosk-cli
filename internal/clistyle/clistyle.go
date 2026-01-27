package clistyle

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// Styles for CLI output
var (
	// Title for headers
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary)

	// Command name styling
	Command = lipgloss.NewStyle().
		Bold(true)

	// Flag styling
	Flag = lipgloss.NewStyle().
		Foreground(styles.Secondary)

	// Description text
	Description = lipgloss.NewStyle()

	// Muted/dim text
	Muted = lipgloss.NewStyle().
		Foreground(styles.Muted)

	// Success styling
	Success = lipgloss.NewStyle().
		Foreground(styles.Success)

	// Warning styling
	Warning = lipgloss.NewStyle().
		Foreground(styles.Warning)

	// Error styling
	Error = lipgloss.NewStyle().
		Foreground(styles.Error).
		Bold(true)

	// Section header
	Section = lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.Primary).
		MarginTop(1)

	// Box for highlighting content
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.Primary).
		Padding(0, 1)
)

// Logo renders the kiosk logo with styling
func Logo() string {
	logo := `.++++.   .=++++= .++++    .-++*****++-.    .-++*****++-.   ++++.   .=++++=
:****: .+****+.  :****. .+*****+++*****+..-****+=-=*****= .****: .=****+:.
:****:+****=.    :****..*****:     :****+.+***+:.. ..:::. .****-=****+.
:**********:     :****..****=       -****.:***********+=. .**********-
:*****++****+.   :****..****=.      =****.   .:--==+*****-.*****++*****.
:****: .=*****:. :****. -****=.. .:+****-.****+..  .+****=.****:  -*****-.
:****.   .*****+.:****. .:+***********+.  :+************- .****:   .+****+.`
	return lipgloss.NewStyle().Foreground(styles.Primary).Render(logo)
}

// FormatHelp creates a styled help output
func FormatHelp(use, short, long string, commands []CommandInfo, flags []FlagInfo) string {
	var b strings.Builder

	// Logo
	b.WriteString(Logo())
	b.WriteString("\n\n")

	// Tagline
	if long != "" {
		b.WriteString(Description.Render(long))
	} else if short != "" {
		b.WriteString(Description.Render(short))
	}
	b.WriteString("\n\n")

	// Usage
	b.WriteString(Section.Render("Usage"))
	b.WriteString("\n")
	b.WriteString("  ")
	b.WriteString(Command.Render(use))
	b.WriteString(" ")
	b.WriteString(Muted.Render("[command]"))
	b.WriteString("\n\n")

	// Commands
	if len(commands) > 0 {
		b.WriteString(Section.Render("Commands"))
		b.WriteString("\n")

		// Calculate max command length for alignment
		maxLen := 0
		for _, cmd := range commands {
			if len(cmd.Name) > maxLen {
				maxLen = len(cmd.Name)
			}
		}

		for _, cmd := range commands {
			b.WriteString("  ")
			b.WriteString(Command.Render(padRight(cmd.Name, maxLen+2)))
			b.WriteString(Muted.Render(cmd.Short))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Flags
	if len(flags) > 0 {
		b.WriteString(Section.Render("Flags"))
		b.WriteString("\n")

		// Calculate max flag length for alignment
		maxLen := 0
		for _, f := range flags {
			flagStr := formatFlag(f)
			if len(flagStr) > maxLen {
				maxLen = len(flagStr)
			}
		}

		for _, f := range flags {
			b.WriteString("  ")
			flagStr := formatFlag(f)
			b.WriteString(Flag.Render(padRight(flagStr, maxLen+2)))
			b.WriteString(Muted.Render(f.Usage))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Footer
	b.WriteString(Muted.Render("Use \"kiosk [command] --help\" for more information about a command."))
	b.WriteString("\n")

	return b.String()
}

// CommandInfo holds command information for help display
type CommandInfo struct {
	Name  string
	Short string
}

// FlagInfo holds flag information for help display
type FlagInfo struct {
	Short string
	Long  string
	Usage string
}

func formatFlag(f FlagInfo) string {
	if f.Short != "" && f.Long != "" {
		return "-" + f.Short + ", --" + f.Long
	} else if f.Long != "" {
		return "    --" + f.Long
	} else if f.Short != "" {
		return "-" + f.Short
	}
	return ""
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// FormatList renders a list of apps in a styled format (for CLI, not TUI)
func FormatList(apps []AppInfo) string {
	if len(apps) == 0 {
		return Muted.Render("No apps installed.\n\n") +
			"Run " + Command.Render("kiosk run <app>") + " to install an app.\n"
	}

	var b strings.Builder

	b.WriteString("\n")

	for _, app := range apps {
		// App name with author
		title := Command.Render(app.Name)
		if app.Author != "" {
			title += Muted.Render(" by " + app.Author)
		}
		if app.Missing {
			title += " " + Warning.Render("(missing)")
		}
		b.WriteString("  ")
		b.WriteString(title)
		b.WriteString("\n")

		// Description (if available)
		if app.Description != "" {
			desc := app.Description
			if len(desc) > 70 {
				desc = desc[:67] + "..."
			}
			b.WriteString("  ")
			b.WriteString(Muted.Render(desc))
			b.WriteString("\n")
		}

		b.WriteString("\n")
	}

	b.WriteString(Muted.Render(padRight("", 2)))
	b.WriteString(Muted.Render(strings.Repeat("─", 40)))
	b.WriteString("\n")
	b.WriteString(Muted.Render("  "))
	countStr := lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", len(apps)))
	b.WriteString(countStr)
	b.WriteString(Muted.Render(" app(s) installed"))
	b.WriteString("\n")

	return b.String()
}

// AppInfo holds app information for list display
type AppInfo struct {
	Name        string
	Author      string
	Description string
	InstalledAt string
	Missing     bool
}

// FormatWhoami renders user info in a styled format
func FormatWhoami(name, username, avatarURL string) string {
	var b strings.Builder

	b.WriteString("\n")

	b.WriteString("  ")
	if name != "" {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render(name))
		b.WriteString("\n")
		b.WriteString("  ")
		b.WriteString(Muted.Render("@" + username))
	} else {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("@" + username))
	}
	b.WriteString("\n\n")

	b.WriteString("  ")
	b.WriteString(Muted.Render("Authenticated with GitHub"))
	b.WriteString("\n")

	return b.String()
}
