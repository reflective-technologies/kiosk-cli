package cmd

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed apps (alias for 'ls')",
	Long:  "List installed apps. This is an alias for the 'ls' command.",
	RunE:  lsCmd.RunE, // Use the same function as the ls command
}

func init() {
	rootCmd.AddCommand(listCmd)
}
