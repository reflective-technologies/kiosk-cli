package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
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
	browseView := views.NewBrowseModel()
	publishView := views.NewPublishModel()
	helpView := views.NewHelpModel()
	loginView := views.NewLoginModel()
	auditView := views.NewAuditModel()

	// Set views on the main model (pass as pointers)
	m.SetHomeView(&homeView)
	m.SetAppListView(&appListView)
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
	_, err := p.Run()
	if err != nil {
		return fmt.Errorf("error running TUI: %w", err)
	}

	return nil
}

// RunTUIPostInstall runs the TUI in post-install mode for a specific app
func RunTUIPostInstall(appName, appKey, appPath string) error {
	// Create the main TUI model
	m := tui.New()

	// Create post-install view
	postInstallView := views.NewPostInstallModel(appName, appKey, appPath)

	// Set the view
	m.SetPostInstallView(&postInstallView)

	// Navigate directly to post-install
	// We need to send a message to navigate
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
	started bool
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

// Optional: Make the TUI the default when no arguments are provided
func init() {
	// Check if we should launch TUI by default
	// This is optional and can be enabled if desired
	if len(os.Args) == 1 {
		// Uncomment to make TUI the default:
		// os.Args = append(os.Args, "tui")
	}
}
