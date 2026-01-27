package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/reflective-technologies/kiosk-cli/internal/tui/styles"
)

// AppItemDelegate is a custom delegate for app list items that supports
// multi-line descriptions with word wrapping.
type AppItemDelegate struct {
	Styles             AppItemStyles
	shortHeight        int // height for items with short descriptions (1 line)
	tallHeight         int // height for items with long descriptions (2 lines)
	spacing            int
	maxDescLines       int // max number of lines for description
	shortDescThreshold int // character threshold for short descriptions
	showSelected       bool
}

// AppItemStyles defines the styles for app items
type AppItemStyles struct {
	NormalTitle   lipgloss.Style
	NormalDesc    lipgloss.Style
	SelectedTitle lipgloss.Style
	SelectedDesc  lipgloss.Style
}

// NewAppItemDelegate creates a new app item delegate with multi-line support
func NewAppItemDelegate() AppItemDelegate {
	return AppItemDelegate{
		Styles: AppItemStyles{
			NormalTitle: lipgloss.NewStyle().
				Foreground(styles.Foreground).
				Padding(0, 0, 0, 2),
			NormalDesc: lipgloss.NewStyle().
				Foreground(styles.Muted).
				Padding(0, 0, 0, 2),
			SelectedTitle: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(styles.Primary).
				Foreground(styles.Foreground).
				Padding(0, 0, 0, 1),
			SelectedDesc: lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(styles.Primary).
				Foreground(styles.Muted).
				Padding(0, 0, 0, 1),
		},
		shortHeight:        2, // Title + 1 desc line
		tallHeight:         3, // Title + 2 desc lines
		spacing:            0,
		maxDescLines:       2,
		shortDescThreshold: 80, // descriptions <= 80 chars get 1 line
		showSelected:       true,
	}
}

// Height returns the height of each item (uses tall height to accommodate all items)
func (d AppItemDelegate) Height() int {
	return d.tallHeight
}

// Spacing returns the spacing between items
func (d AppItemDelegate) Spacing() int {
	return d.spacing
}

// Update handles item updates
func (d AppItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

// Render renders an item
func (d AppItemDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	var (
		title, desc string
		titleStyle  lipgloss.Style
		descStyle   lipgloss.Style
	)

	if i, ok := item.(list.DefaultItem); ok {
		title = i.Title()
		desc = i.Description()
	} else {
		return
	}

	isSelected := index == m.Index()

	if isSelected && d.showSelected {
		titleStyle = d.Styles.SelectedTitle
		descStyle = d.Styles.SelectedDesc
	} else {
		titleStyle = d.Styles.NormalTitle
		descStyle = d.Styles.NormalDesc
	}

	// Calculate available width for text
	// Account for padding (2 chars on left for normal, 1 char + border for selected)
	width := m.Width() - 4
	if width < 20 {
		width = 20
	}

	// Render title (single line, truncated if needed)
	renderedTitle := truncate(title, width)
	fmt.Fprint(w, titleStyle.Render(renderedTitle))
	fmt.Fprint(w, "\n")

	// Determine how many description lines to use based on length
	descLines := 1
	if len(desc) > d.shortDescThreshold {
		descLines = d.maxDescLines
	}

	// Wrap and render description
	if desc != "" {
		wrapped := wordwrap.String(desc, width)
		lines := strings.Split(wrapped, "\n")

		// Only render up to descLines
		linesToRender := len(lines)
		if linesToRender > descLines {
			linesToRender = descLines
		}

		for i := 0; i < linesToRender; i++ {
			line := lines[i]
			// If this is the last allowed line and there's more text, add ellipsis
			if i == descLines-1 && len(lines) > descLines {
				line = truncate(line, width-3) + "..."
			}
			fmt.Fprint(w, descStyle.Render(line))
			fmt.Fprint(w, "\n")
		}
	}
}

// truncate truncates a string to the given width
func truncate(s string, width int) string {
	if len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-3] + "..."
}
