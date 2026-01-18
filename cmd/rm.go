package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

var rmForce bool

var rmCmd = &cobra.Command{
	Use:   "rm <org/repo>",
	Short: "Remove an installed app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]

		// Load index
		idx, err := appindex.Load()
		if err != nil {
			return fmt.Errorf("failed to load app index: %w", err)
		}

		// Check if app is in index
		if !idx.Has(key) {
			return fmt.Errorf("app %q is not installed", key)
		}

		// Confirm unless --force
		if !rmForce {
			fmt.Printf("Remove %q? This will delete the local copy. [y/N] ", key)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		// Remove directory if it exists
		appPath := config.AppPath(key, "")
		// Parse org/repo properly
		parts := strings.SplitN(key, "/", 2)
		if len(parts) == 2 {
			appPath = config.AppPath(parts[0], parts[1])
		}

		if _, err := os.Stat(appPath); err == nil {
			if err := os.RemoveAll(appPath); err != nil {
				return fmt.Errorf("failed to remove directory: %w", err)
			}
		}

		// Remove from index
		idx.Remove(key)
		if err := appindex.Save(idx); err != nil {
			return fmt.Errorf("failed to save app index: %w", err)
		}

		fmt.Printf("Removed %s\n", key)
		return nil
	},
}

func init() {
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(rmCmd)
}
