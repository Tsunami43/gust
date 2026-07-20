// Package theme defines gust's shared colour palette and icon set — a calm
// violet terminal aesthetic. Colours are emitted as 24-bit ("truecolor") ANSI
// escape sequences.
package theme

import "fmt"

// Palette colours (hex).
const (
	Accent = "#a78bfa" // primary violet
	Text   = "#e9e4f5" // near-white lavender
	Alt    = "#b9a7e6" // muted lavender
	Good   = "#86d6a2" // soft green
	Warn   = "#f0c560" // amber
	Bad    = "#ee7d92" // soft red
	Bright = "#d8b4fe" // light purple
	Rule   = "#6b6577" // muted divider grey
	Base   = "#7c5cd6" // logo mid tone
	Shade  = "#4c3a8a" // logo shadow
	Hi     = "#ffffff" // logo highlight
)

// Icons.
const (
	Pointer = "❯"
	Bar     = "▌"
	Done    = "✓"
	Fail    = "✗"
	Dot     = "·"
	Down    = "↓"
	Up      = "↑"
)

// Style codes.
const (
	Reset = "\033[0m"
	Bold  = "\033[1m"
	Dim   = "\033[2m"
)

// FG returns the escape sequence that sets the foreground to the given hex.
func FG(hex string) string {
	r, g, b := rgb(hex)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// LerpFG interpolates between two hex colours (t in [0,1]) and returns the
// escape sequence for the resulting foreground.
func LerpFG(a, b string, t float64) string {
	ar, ag, ab := rgb(a)
	br, bg, bb := rgb(b)
	mix := func(x, y int) int { return int(float64(x) + (float64(y)-float64(x))*t + 0.5) }
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", mix(ar, br), mix(ag, bg), mix(ab, bb))
}

func rgb(hex string) (r, g, b int) {
	if len(hex) == 7 && hex[0] == '#' {
		var n int64
		fmt.Sscanf(hex[1:], "%x", &n)
		return int((n >> 16) & 255), int((n >> 8) & 255), int(n & 255)
	}
	return 255, 255, 255
}
