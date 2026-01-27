package cmd

import (
	"fmt"

	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/reflective-technologies/kiosk-cli/internal/clistyle"
	"github.com/spf13/cobra"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display the currently authenticated user",
	Long:  `Display information about the currently authenticated GitHub user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		user, err := auth.GetUser()
		if err != nil {
			return err
		}

		fmt.Print(clistyle.FormatWhoami(user.Name, user.Username, ""))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}
