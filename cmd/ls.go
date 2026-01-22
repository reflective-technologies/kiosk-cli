package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/style"
	"github.com/spf13/cobra"
)

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

		// Validate filesystem
		exists := idx.ValidateFilesystem()

		// Create styler for stdout
		s := style.Stdout()

		// Calculate column widths
		maxName := len("APP")
		maxAuthor := len("AUTHOR")
		for _, key := range keys {
			author, name := splitAppKey(key)
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
			s.Apply(style.Dim, padRight("APP", maxName)),
			s.Apply(style.Dim, padRight("AUTHOR", maxAuthor)),
			s.Apply(style.Dim, "INSTALLED"),
		)
		fmt.Println(s.Apply(style.Dim, "  "+strings.Repeat("─", maxName+maxAuthor+13)))

		// Print rows
		for _, key := range keys {
			author, name := splitAppKey(key)

			entry := idx.Get(key)
			installedAt := "unknown"
			if entry != nil && !entry.InstalledAt.IsZero() {
				installedAt = entry.InstalledAt.Format("01/02/06")
			}

			status := ""
			if !exists[key] {
				status = s.Apply(style.Yellow, " (missing)")
			}

			fmt.Printf("  %s  %s  %s%s\n",
				s.Apply(style.Bold, padRight(name, maxName)),
				padRight(author, maxAuthor),
				s.Apply(style.Dim, installedAt),
				status,
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

func splitAppKey(key string) (author, name string) {
	parts := strings.SplitN(key, "/", 2)
	author = parts[0]
	name = parts[0]
	if len(parts) == 2 {
		name = parts[1]
	}
	return author, name
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
