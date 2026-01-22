package style

import (
	"os"

	"golang.org/x/term"
)

// ANSI color codes
const (
	Reset  = "\033[0m"
	Bold   = "\033[1m"
	Dim    = "\033[2m"
	Red    = "\033[31m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
)

// UseColor determines if ANSI colors should be used for the given file descriptor.
func UseColor(fd uintptr) bool {
	if !term.IsTerminal(int(fd)) {
		return false
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	if os.Getenv("TERM") == "dumb" {
		return false
	}
	return true
}

// Apply wraps text in ANSI codes if colors are enabled for the given file descriptor.
func Apply(fd uintptr, codes, text string) string {
	if !UseColor(fd) {
		return text
	}
	return codes + text + Reset
}

// Stdout returns a Styler configured for stdout.
func Stdout() *Styler {
	return &Styler{fd: os.Stdout.Fd()}
}

// Stderr returns a Styler configured for stderr.
func Stderr() *Styler {
	return &Styler{fd: os.Stderr.Fd()}
}

// Styler applies styles to text for a specific output stream.
type Styler struct {
	fd uintptr
}

// Apply wraps text in ANSI codes if colors are enabled.
func (s *Styler) Apply(codes, text string) string {
	return Apply(s.fd, codes, text)
}

// UseColor returns whether colors are enabled for this styler.
func (s *Styler) UseColor() bool {
	return UseColor(s.fd)
}
