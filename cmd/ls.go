package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// ANSI codes for ls command
const (
	lsReset = "\033[0m"
	lsDim   = "\033[2m"
	lsBold  = "\033[1m"
)

func lsUseColor() bool {
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

func lsStyle(codes string, text string) string {
	if !lsUseColor() {
		return text
	}
	return codes + text + lsReset
}

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List installed apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		idx, err := appindex.Load()
		if err != nil {
			return fmt.Errorf("failed to load app index: %w", err)
		}

		if idx.Count() == 0 {
			fmt.Println("No apps installed.")
			return nil
		}

		// Get and sort app keys
		keys := idx.List()
		sort.Strings(keys)

		// Calculate column widths
		maxName := len("APP")
		maxAuthor := len("AUTHOR")
		for _, key := range keys {
			parts := strings.SplitN(key, "/", 2)
			author := parts[0]
			name := parts[0]
			if len(parts) == 2 {
				name = parts[1]
			}
			if len(name) > maxName {
				maxName = len(name)
			}
			if len(author) > maxAuthor {
				maxAuthor = len(author)
			}
		}

		// Print header
		fmt.Println()
		fmt.Printf("  %s  %s  %s\n",
			lsStyle(lsDim, padRight("APP", maxName)),
			lsStyle(lsDim, padRight("AUTHOR", maxAuthor)),
			lsStyle(lsDim, "INSTALLED"),
		)
		fmt.Println(lsStyle(lsDim, "  "+strings.Repeat("─", maxName+maxAuthor+14)))

		// Print rows
		for _, key := range keys {
			parts := strings.SplitN(key, "/", 2)
			author := parts[0]
			name := parts[0]
			if len(parts) == 2 {
				name = parts[1]
			}

			entry := idx.Get(key)
			installedAt := entry.InstalledAt.Format("01/02/06")

			fmt.Printf("  %s  %s  %s\n",
				lsStyle(lsBold, padRight(name, maxName)),
				padRight(author, maxAuthor),
				lsStyle(lsDim, installedAt),
			)
		}

		fmt.Println()
		fmt.Printf("%d app(s) installed\n", idx.Count())

		return nil
	},
}

func padRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
