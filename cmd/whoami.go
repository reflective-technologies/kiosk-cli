package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/reflective-technologies/kiosk-cli/internal/auth"
	"github.com/spf13/cobra"
)

type GitHubUser struct {
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Display the currently authenticated user",
	Long:  `Display information about the currently authenticated GitHub user.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, err := auth.GetToken()
		if err != nil {
			return err
		}

		user, err := getGitHubUser(token)
		if err != nil {
			return fmt.Errorf("failed to fetch user info: %w", err)
		}

		if user.Name != "" {
			fmt.Printf("%s (%s)\n", user.Name, user.Login)
		} else {
			fmt.Println(user.Login)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(whoamiCmd)
}

func getGitHubUser(token string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}
