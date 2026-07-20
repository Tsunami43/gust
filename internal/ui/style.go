package ui

import (
	"os"
	"strings"

	"github.com/Tsunami43/gust/internal/theme"
)

// palette holds the ANSI escape codes used for styling. When colour is
// disabled every field is an empty string, so styled output degrades to plain
// text automatically. Colours follow the shared violet theme.
type palette struct {
	reset, bold, dim   string
	red, green, yellow string
	accent, rule, text string
}

// newPalette returns a populated palette when enabled, or an all-empty palette
// (producing plain text) when not.
func newPalette(enabled bool) palette {
	if !enabled {
		return palette{}
	}
	return palette{
		reset:  theme.Reset,
		bold:   theme.Bold,
		dim:    theme.Dim,
		red:    theme.FG(theme.Bad),
		green:  theme.FG(theme.Good),
		yellow: theme.FG(theme.Warn),
		accent: theme.FG(theme.Accent),
		rule:   theme.FG(theme.Rule),
		text:   theme.FG(theme.Text),
	}
}

// IsTerminal reports whether f is attached to a character device (a TTY).
// It relies only on the standard library, keeping gust dependency-free.
func IsTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// visibleLen returns the number of printable columns in s, ignoring ANSI
// escape sequences so that padding calculations stay correct for coloured
// strings.
func visibleLen(s string) int {
	n, inEscape := 0, false
	for _, r := range s {
		switch {
		case inEscape:
			if r == 'm' {
				inEscape = false
			}
		case r == '\033':
			inEscape = true
		default:
			n++
		}
	}
	return n
}

// renderBar draws a fixed-width progress bar filled to the given fraction.
func renderBar(fraction float64, width int, p palette) string {
	if fraction < 0 {
		fraction = 0
	}
	if fraction > 1 {
		fraction = 1
	}
	filled := int(fraction*float64(width) + 0.5)
	return p.accent + strings.Repeat("█", filled) +
		p.rule + strings.Repeat("░", width-filled) + p.reset
}
