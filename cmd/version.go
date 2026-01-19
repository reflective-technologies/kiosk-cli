package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time via ldflags
	Version = "dev"

	// GitHubClientID is the OAuth App client ID for GitHub authentication.
	// Can be overridden at build time via ldflags.
	GitHubClientID = "Iv23lilG55tK1ZOWxag2"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("kiosk version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
