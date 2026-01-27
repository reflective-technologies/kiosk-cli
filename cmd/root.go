package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/reflective-technologies/kiosk-cli/internal/clistyle"
	"github.com/reflective-technologies/kiosk-cli/internal/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var rootCmd = &cobra.Command{
	Use:           "kiosk",
	Short:         "Kiosk CLI",
	Long:          "The app store for Claude Code apps.",
	Version:       Version,
	SilenceErrors: true, // We handle error display ourselves
	SilenceUsage:  true, // Don't show usage on errors
	// Run TUI when no subcommand is provided
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI(cmd, args)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		errors.PrintError(err)
		os.Exit(1)
	}
}

func init() {
	// Enable verbose error logging in dev mode
	errors.DevMode = Version == "dev"

	// Custom help function
	rootCmd.SetHelpFunc(styledHelp)
}

// styledHelp renders a styled help output
func styledHelp(cmd *cobra.Command, args []string) {
	// Collect commands (excluding hidden ones)
	var commands []clistyle.CommandInfo
	for _, c := range cmd.Commands() {
		if !c.Hidden && c.Name() != "help" && c.Name() != "completion" {
			commands = append(commands, clistyle.CommandInfo{
				Name:  c.Name(),
				Short: c.Short,
			})
		}
	}

	// Sort commands alphabetically
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})

	// Collect flags
	var flags []clistyle.FlagInfo
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" || f.Name == "version" {
			return
		}
		flags = append(flags, clistyle.FlagInfo{
			Short: f.Shorthand,
			Long:  f.Name,
			Usage: f.Usage,
		})
	})

	// Add standard flags
	flags = append(flags,
		clistyle.FlagInfo{Short: "h", Long: "help", Usage: "help for kiosk"},
		clistyle.FlagInfo{Short: "v", Long: "version", Usage: "version for kiosk"},
	)

	// Render styled help
	help := clistyle.FormatHelp(cmd.Use, cmd.Short, cmd.Long, commands, flags)
	fmt.Print(help)
}
