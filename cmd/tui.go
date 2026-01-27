package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/views"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:     "tui",
	Aliases: []string{"ui", "interactive"},
	Short:   "Launch the interactive TUI",
	Long: `Launch the interactive terminal user interface for managing kiosk apps.

The TUI provides a visual interface for:
  - Viewing and managing installed apps
  - Running security audits
  - Authentication with GitHub
  - Post-installation workflows`,
	RunE: runTUI,
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}

func runTUI(cmd *cobra.Command, args []string) error {
	// Create the main TUI model
	m := tui.New()

	// Create view models (as pointers so SetSize works correctly)
	homeView := views.NewHomeModel()
	appListView := views.NewAppListModel()
	appDetailView := views.NewAppDetailModel()
	browseView := views.NewBrowseModel()
	publishView := views.NewPublishModel()
	helpView := views.NewHelpModel()
	loginView := views.NewLoginModel()
	auditView := views.NewAuditModel()

	// Set views on the main model (pass as pointers)
	m.SetHomeView(&homeView)
	m.SetAppListView(&appListView)
	m.SetAppDetailView(&appDetailView)
	m.SetBrowseView(&browseView)
	m.SetPublishView(&publishView)
	m.SetHelpView(&helpView)
	m.SetLoginView(&loginView)
	m.SetAuditView(&auditView)

	// Create the program with alternate screen buffer
	p := tea.NewProgram(
		&m,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	// Run the TUI
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	// Check if we need to execute an app after TUI exits
	if model, ok := finalModel.(*tui.Model); ok && model.ExecApp != "" {
		// Execute the app using kiosk run
		return executeApp(model.ExecApp)
	}

	return nil
}

// executeApp runs an app after TUI exits using the same logic as `kiosk run`
func executeApp(appKey string) error {
	// Ensure working directory is initialized
	if err := config.EnsureInitialized(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// Load config and index
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	idx, err := appindex.Load()
	if err != nil {
		return fmt.Errorf("failed to load app index: %w", err)
	}

	// Normalize key to org/repo format for index lookup
	key := normalizeAppKey(appKey)

	// Check if app is installed
	if idx.Has(key) {
		return runInstalledApp(key, nil, false)
	}

	// App not installed - fetch from API and install
	return installAndRunApp(cfg, idx, appKey, key, nil, false)
}

// RunTUIPostInstall runs the TUI in post-install mode for a specific app
func RunTUIPostInstall(appName, appKey, appPath string) error {
	// Create the main TUI model
	m := tui.New()

	// The post-install view is created in postInstallModel.Init() to avoid
	// creating it twice (Init sets the view and navigates to it)
	p := tea.NewProgram(
		&postInstallModel{
			model:   &m,
			appName: appName,
			appKey:  appKey,
			appPath: appPath,
		},
		tea.WithAltScreen(),
	)

	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running post-install TUI: %w", err)
	}

	return nil
}

// postInstallModel wraps the TUI model to start in post-install mode
type postInstallModel struct {
	model   *tui.Model
	appName string
	appKey  string
	appPath string
}

func (m *postInstallModel) Init() tea.Cmd {
	postInstallView := views.NewPostInstallModel(m.appName, m.appKey, m.appPath)
	m.model.SetPostInstallView(&postInstallView)

	return tea.Batch(
		m.model.Init(),
		func() tea.Msg {
			return tui.NavigateMsg{View: tui.ViewPostInstall}
		},
	)
}

func (m *postInstallModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle quit from post-install
	if _, ok := msg.(tui.GoBackMsg); ok {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	_, cmd = m.model.Update(msg)
	return m, cmd
}

func (m *postInstallModel) View() string {
	return m.model.View()
}
