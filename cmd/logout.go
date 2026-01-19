package cmd

import (
	"fmt"

	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out from GitHub",
	Long:  `Remove stored GitHub credentials from this machine.`,
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}

func runLogout(cmd *cobra.Command, args []string) error {
	if !auth.IsLoggedIn() {
		fmt.Println("You are not logged in.")
		return nil
	}

	if err := auth.DeleteCredentials(); err != nil {
		return fmt.Errorf("failed to logout: %w", err)
	}

	fmt.Println("Successfully logged out.")
	return nil
}
