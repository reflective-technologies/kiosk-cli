package cmd

import (
	"fmt"
	"sort"

	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
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
		apps := idx.List()
		sort.Strings(apps)

		// Validate filesystem
		exists := idx.ValidateFilesystem()

		fmt.Println("Installed apps:")
		fmt.Println()

		for _, key := range apps {
			status := ""
			if !exists[key] {
				status = " (missing)"
			}
			fmt.Printf("  %s%s\n", key, status)
		}

		fmt.Println()
		fmt.Printf("%d app(s) installed\n", idx.Count())

		return nil
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
