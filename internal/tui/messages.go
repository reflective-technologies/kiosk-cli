package tui

import (
	"github.com/reflective-technologies/kiosk-cli/internal/api"
	"github.com/reflective-technologies/kiosk-cli/internal/appindex"
	"github.com/reflective-technologies/kiosk-cli/internal/auth"
)

// Navigation messages

// NavigateMsg tells the app to navigate to a different view
type NavigateMsg struct {
	View ViewType
}

// GoBackMsg tells the app to go back to the previous view
type GoBackMsg struct{}

// App-related messages

// AppsLoadedMsg is sent when apps have been loaded from the index
type AppsLoadedMsg struct {
	Index *appindex.Index
	Err   error
}

// AppSelectedMsg is sent when a user selects an app
type AppSelectedMsg struct {
	Key   string
	Entry *appindex.AppEntry
}

// AppInstalledMsg is sent when an app has been installed
type AppInstalledMsg struct {
	Key   string
	Entry *appindex.AppEntry
	Err   error
}

// AppRemovedMsg is sent when an app has been removed
type AppRemovedMsg struct {
	Key string
	Err error
}

// Auth-related messages

// LoginStartedMsg is sent when login flow begins
type LoginStartedMsg struct {
	DeviceCode      string
	UserCode        string
	VerificationURI string
	Interval        int // Polling interval in seconds (per RFC 8628)
}

// LoginCompleteMsg is sent when login completes
type LoginCompleteMsg struct {
	User *auth.UserInfo
	Err  error
}

// LogoutCompleteMsg is sent when logout completes
type LogoutCompleteMsg struct {
	Err error
}

// AuthStatusMsg is sent with current auth status
type AuthStatusMsg struct {
	LoggedIn bool
	User     *auth.UserInfo
}

// Operation messages

// CloneProgressMsg is sent during git clone operation
type CloneProgressMsg struct {
	Percent int
	Message string
}

// CloneCompleteMsg is sent when git clone finishes
type CloneCompleteMsg struct {
	Path string
	Err  error
}

// AuditStartedMsg is sent when audit begins
type AuditStartedMsg struct{}

// AuditCompleteMsg is sent when audit finishes
type AuditCompleteMsg struct {
	Result string
	Err    error
}

// Browse apps messages

// BrowseAppsLoadedMsg is sent when apps have been loaded from the API
type BrowseAppsLoadedMsg struct {
	Apps       []api.App
	NextCursor *string // cursor for next page, nil if no more pages
	Err        error
}

// BrowseAppsPageLoadedMsg is sent when an additional page of apps has been loaded
type BrowseAppsPageLoadedMsg struct {
	Apps       []api.App
	NextCursor *string // cursor for next page, nil if no more pages
	Err        error
}

// BrowseAppSelectedMsg is sent when a user selects an app to install
type BrowseAppSelectedMsg struct {
	App api.App
}

// App detail messages

// ShowAppDetailMsg is sent to show app detail view
type ShowAppDetailMsg struct {
	App         *api.App
	IsInstalled bool
	AppKey      string // for installed apps (e.g., "owner/repo")
}

// RunAppMsg is sent when user wants to run an app
type RunAppMsg struct {
	AppKey string
	GitURL string
}

// DeleteAppMsg is sent when user wants to delete an app
type DeleteAppMsg struct {
	AppKey string
}

// ExecAppMsg signals the TUI to quit and execute an app
type ExecAppMsg struct {
	AppKey string
}

// ExecPostInstallOptionMsg signals the TUI to quit and execute a post-install option
type ExecPostInstallOptionMsg struct {
	AppPath string
	Command string // One of: "claude", "edit", "dev", "build", "test"
	Prompt  string // Prompt to pass to Claude (if applicable)
}

// Generic messages

// ErrorMsg represents an error that occurred
type ErrorMsg struct {
	Err error
}

// SuccessMsg represents a successful operation
type SuccessMsg struct {
	Message string
}

// StatusMsg is a transient status message
type StatusMsg struct {
	Message string
}

// ClearStatusMsg clears the current status message
type ClearStatusMsg struct{}
