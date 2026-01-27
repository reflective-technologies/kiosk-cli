package cmd

import (
	"fmt"
	"io"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/claude"
	"github.com/reflective-technologies/kiosk-cli/internal/config"
	"github.com/reflective-technologies/kiosk-cli/internal/sessions"
	"github.com/reflective-technologies/kiosk-cli/internal/tui"
)

type sessionExec struct {
	appArg   string
	sessions *sessions.Store
	stdin    io.Reader
	stdout   io.Writer
	stderr   io.Writer
}

func (e *sessionExec) Run() error {
	ioCfg := claude.SessionIO{
		Stdin:  e.stdin,
		Stdout: e.stdout,
		Stderr: e.stderr,
	}
	return runAppWithSession(e.appArg, e.sessions, ioCfg)
}

func (e *sessionExec) SetStdin(r io.Reader)  { e.stdin = r }
func (e *sessionExec) SetStdout(w io.Writer) { e.stdout = w }
func (e *sessionExec) SetStderr(w io.Writer) { e.stderr = w }

func runAppWithSession(appArg string, store *sessions.Store, ioCfg claude.SessionIO) error {
	if store == nil {
		return fmt.Errorf("session store unavailable")
	}

	if err := config.EnsureInitialized(); err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	idx, err := appindex.Load()
	if err != nil {
		return fmt.Errorf("failed to load app index: %w", err)
	}

	key := normalizeAppKey(appArg)
	sessionCfg := &claudeSessionConfig{
		Store: store,
		IO:    ioCfg,
	}

	if idx.Has(key) {
		return runInstalledApp(key, nil, false, sessionCfg)
	}

	return installAndRunApp(cfg, idx, appArg, key, nil, false, sessionCfg)
}

func runAppSessionCmd(appArg string, store *sessions.Store) tea.Cmd {
	return tea.Exec(&sessionExec{appArg: appArg, sessions: store}, func(err error) tea.Msg {
		if err == nil {
			return tui.StatusMsg{Message: fmt.Sprintf("Session ended: %s", appArg)}
		}
		if err == claude.ErrDetached {
			return tui.SessionSuspendedMsg{
				AppKey:  appArg,
				Message: "Session saved. Resume anytime from My Apps.",
				Timeout: 3 * time.Second,
			}
		}
		return tui.ErrorMsg{Err: err}
	})
}
