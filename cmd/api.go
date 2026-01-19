package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/spf13/cobra"
)

var inputFile string

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Interact directly with the Kiosk API",
	Long:  `Direct interaction with Kiosk API endpoints. Useful for scripting and agents.`,
}

var apiListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all published apps",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIUrl)
		apps, err := client.ListApps()
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(apps)
	},
}

var apiGetCmd = &cobra.Command{
	Use:   "get [appId]",
	Short: "Get app details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIUrl)
		app, err := client.GetApp(args[0])
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(app)
	},
}

var apiCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Publish a new app",
	RunE: func(cmd *cobra.Command, args []string) error {
		var req api.CreateAppRequest
		if err := readJSONInput(inputFile, &req); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIUrl)
		app, err := client.CreateApp(req)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(app)
	},
}

var apiUpdateCmd = &cobra.Command{
	Use:   "update [appId]",
	Short: "Update an existing app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var req api.UpdateAppRequest
		if err := readJSONInput(inputFile, &req); err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIUrl)
		app, err := client.UpdateApp(args[0], req)
		if err != nil {
			return err
		}

		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(app)
	},
}

var apiDeleteCmd = &cobra.Command{
	Use:   "delete [appId]",
	Short: "Delete an app",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIUrl)
		if err := client.DeleteApp(args[0]); err != nil {
			return err
		}

		fmt.Printf("App %q deleted successfully\n", args[0])
		return nil
	},
}

var apiRefreshCmd = &cobra.Command{
	Use:   "refresh [appId]",
	Short: "Refresh app's Kiosk.md from repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return err
		}

		client := api.NewClient(cfg.APIUrl)
		if err := client.RefreshApp(args[0]); err != nil {
			return err
		}

		fmt.Printf("App %q refreshed successfully\n", args[0])
		return nil
	},
}

func readJSONInput(path string, v interface{}) error {
	var r io.Reader

	if path == "" || path == "-" {
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			return fmt.Errorf("input file required (use -f) or pipe data to stdin")
		}
		r = os.Stdin
	} else {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func init() {
	rootCmd.AddCommand(apiCmd)
	apiCmd.AddCommand(apiListCmd)
	apiCmd.AddCommand(apiGetCmd)
	apiCmd.AddCommand(apiCreateCmd)
	apiCmd.AddCommand(apiUpdateCmd)
	apiCmd.AddCommand(apiDeleteCmd)
	apiCmd.AddCommand(apiRefreshCmd)

	apiCreateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Path to JSON file (use - for stdin)")
	apiUpdateCmd.Flags().StringVarP(&inputFile, "file", "f", "", "Path to JSON file (use - for stdin)")
}