package ui

import (
	"os"
	"strings"
)

// palette holds the ANSI escape codes used for styling. When colour is
// disabled every field is an empty string, so styled output degrades to plain
// text automatically.
type palette struct {
	reset, bold, dim                  string
	red, green, yellow, cyan, magenta string
}

// newPalette returns a populated palette when enabled, or an all-empty palette
// (producing plain text) when not.
func newPalette(enabled bool) palette {
	if !enabled {
		return palette{}
	}
	return palette{
		reset:   "\033[0m",
		bold:    "\033[1m",
		dim:     "\033[2m",
		red:     "\033[31m",
		green:   "\033[32m",
		yellow:  "\033[33m",
		cyan:    "\033[36m",
		magenta: "\033[35m",
	}
}

// isTerminal reports whether f is attached to a character device (a TTY).
// It relies only on the standard library, keeping gust dependency-free.
func isTerminal(f *os.File) bool {
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
	return p.cyan + strings.Repeat("█", filled) +
		p.dim + strings.Repeat("░", width-filled) + p.reset
}
