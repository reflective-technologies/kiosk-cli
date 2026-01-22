package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ANSI color codes
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	italic = "\033[3m"
	under  = "\033[4m"

	// Standard colors
	black   = "\033[30m"
	red     = "\033[31m"
	green   = "\033[32m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	magenta = "\033[35m"
	cyan    = "\033[36m"
	white   = "\033[37m"

	// Bright colors
	brightBlack   = "\033[90m"
	brightRed     = "\033[91m"
	brightGreen   = "\033[92m"
	brightYellow  = "\033[93m"
	brightBlue    = "\033[94m"
	brightMagenta = "\033[95m"
	brightCyan    = "\033[96m"
	brightWhite   = "\033[97m"

	// Background colors

	bgWhite = "\033[47m"
)

// useColor determines if we should use ANSI colors
func styleUseColor() bool {
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// style applies ANSI codes if colors are enabled
func style(codes string, text string) string {
	if !styleUseColor() {
		return text
	}
	return codes + text + reset
}

var styleTestCmd = &cobra.Command{
	Use:    "style-test",
	Short:  "Test CLI styling options",
	Hidden: true, // Hide from help output
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println()
		fmt.Println(style(bold, "=== CLI Styling Test ==="))
		fmt.Println()

		// Text styles
		fmt.Println(style(bold, "Text Styles:"))
		fmt.Printf("  %s  %s  %s  %s\n",
			style(bold, "bold"),
			style(dim, "dim"),
			style(italic, "italic"),
			style(under, "underline"),
		)
		fmt.Println()

		// Standard colors
		fmt.Println(style(bold, "Standard Colors:"))
		fmt.Printf("  %s %s %s %s %s %s %s %s\n",
			style(black+bgWhite, " black "),
			style(red, "red"),
			style(green, "green"),
			style(yellow, "yellow"),
			style(blue, "blue"),
			style(magenta, "magenta"),
			style(cyan, "cyan"),
			style(white, "white"),
		)
		fmt.Println()

		// Bright colors
		fmt.Println(style(bold, "Bright Colors:"))
		fmt.Printf("  %s %s %s %s %s %s %s %s\n",
			style(brightBlack, "black"),
			style(brightRed, "red"),
			style(brightGreen, "green"),
			style(brightYellow, "yellow"),
			style(brightBlue, "blue"),
			style(brightMagenta, "magenta"),
			style(brightCyan, "cyan"),
			style(brightWhite, "white"),
		)
		fmt.Println()

		// Combinations
		fmt.Println(style(bold, "Combinations:"))
		fmt.Printf("  %s  %s  %s\n",
			style(bold+green, "bold green"),
			style(dim+cyan, "dim cyan"),
			style(bold+under+yellow, "bold underline yellow"),
		)
		fmt.Println()

		// Sample outputs
		fmt.Println(style(bold, "Sample CLI Output Styles:"))
		fmt.Println()

		// Success message
		fmt.Printf("  %s App installed successfully!\n", style(green, "✓"))

		// Warning message
		fmt.Printf("  %s Config file not found, using defaults\n", style(yellow, "⚠"))

		// Error message
		fmt.Printf("  %s Failed to connect to server\n", style(red, "✗"))

		// Info message
		fmt.Printf("  %s Checking for updates...\n", style(cyan, "ℹ"))

		fmt.Println()

		// List style
		fmt.Println(style(bold, "Installed apps:"))
		fmt.Println()
		apps := []struct {
			name    string
			version string
			status  string
		}{
			{"my-app", "v1.2.3", "active"},
			{"another-app", "v0.9.0", "active"},
			{"old-app", "v0.1.0", "missing"},
		}
		for _, app := range apps {
			status := style(green, "●")
			if app.status == "missing" {
				status = style(red, "●")
			}
			fmt.Printf("  %s %s %s\n",
				status,
				style(bold, app.name),
				style(dim, app.version),
			)
		}
		fmt.Println()

		// Table style
		fmt.Println(style(bold, "Table Style:"))
		fmt.Println()
		fmt.Printf("  %s  %s  %s\n",
			style(dim, "NAME"),
			style(dim, "VERSION"),
			style(dim, "STATUS"),
		)
		fmt.Println(style(dim, "  "+strings.Repeat("─", 35)))
		for _, app := range apps {
			statusText := style(green, "active")
			if app.status == "missing" {
				statusText = style(red, "missing")
			}
			fmt.Printf("  %-12s %-10s %s\n", app.name, app.version, statusText)
		}
		fmt.Println()

		// Progress/spinner style
		fmt.Println(style(bold, "Progress Indicators:"))
		fmt.Printf("  %s Downloading...\n", style(cyan, "⠋"))
		fmt.Printf("  %s\n", style(dim, "  [████████████░░░░░░░░] 60%%"))
		fmt.Println()

		// Box style
		fmt.Println(style(bold, "Box Style:"))
		fmt.Println(style(dim, "  ┌────────────────────────────┐"))
		fmt.Println(style(dim, "  │") + " " + style(bold+cyan, "Welcome to Kiosk") + "          " + style(dim, "│"))
		fmt.Println(style(dim, "  │") + " The app store for Claude   " + style(dim, "│"))
		fmt.Println(style(dim, "  └────────────────────────────┘"))
		fmt.Println()

		// Key-value pairs
		fmt.Println(style(bold, "Key-Value Style:"))
		fmt.Printf("  %s %s\n", style(dim, "User:"), "john_doe")
		fmt.Printf("  %s %s\n", style(dim, "Email:"), "john@example.com")
		fmt.Printf("  %s %s\n", style(dim, "Plan:"), style(green, "Pro"))
		fmt.Println()
	},
}

func init() {
	rootCmd.AddCommand(styleTestCmd)
}
