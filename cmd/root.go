package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

const logo = `
   ____ ___   ________  ________  ________  ____ ___
  ╱    ╱   ╲ ╱        ╲╱        ╲╱        ╲╱    ╱   ╲
 ╱         ╱_╱       ╱╱         ╱        _╱         ╱
╱╱       _╱╱         ╱         ╱-        ╱        _╱
╲╲___╱___╱ ╲________╱╲________╱╲________╱╲____╱___╱
`

var rootCmd = &cobra.Command{
	Use:     "kiosk",
	Short:   "Kiosk CLI",
	Long:    logo + `The app store for Claude Code apps.`,
	Version: Version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kiosk.yaml)")
}
