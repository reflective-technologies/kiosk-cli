package tui

// ViewType represents different views/screens in the TUI
type ViewType int

const (
	ViewHome ViewType = iota
	ViewAppList
	ViewAppDetail
	ViewBrowse
	ViewLibrary
	ViewHelp
	ViewLogin
	ViewAudit
	ViewPostInstall
	ViewSettings
)

// String returns the string representation of the view type
func (v ViewType) String() string {
	switch v {
	case ViewHome:
		return "Home"
	case ViewAppList:
		return "My Apps"
	case ViewAppDetail:
		return "App Detail"
	case ViewBrowse:
		return "Browse Apps"
	case ViewLibrary:
		return "Library"
	case ViewHelp:
		return "Help"
	case ViewLogin:
		return "Login"
	case ViewAudit:
		return "Audit"
	case ViewPostInstall:
		return "Post Install"
	case ViewSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}
