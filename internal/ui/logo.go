package ui

import (
	"fmt"

	"github.com/Tsunami43/gust/internal/theme"
)

// logoLines is the GUST word-mark drawn in box-glyphs.
var logoLines = []string{
	" ██████╗ ██╗   ██╗███████╗████████╗",
	"██╔════╝ ██║   ██║██╔════╝╚══██╔══╝",
	"██║  ███╗██║   ██║███████╗   ██║",
	"██║   ██║██║   ██║╚════██║   ██║",
	"╚██████╔╝╚██████╔╝███████║   ██║",
	" ╚═════╝  ╚═════╝ ╚══════╝   ╚═╝",
}

// sheen returns the foreground colour for a normalised position t within the
// logo, ramping highlight → bright → accent → base → shadow.
func sheen(t float64) string {
	switch {
	case t < 0.15:
		return theme.LerpFG(theme.Hi, theme.Bright, t/0.15)
	case t < 0.40:
		return theme.LerpFG(theme.Bright, theme.Accent, (t-0.15)/0.25)
	case t < 0.70:
		return theme.LerpFG(theme.Accent, theme.Base, (t-0.40)/0.30)
	default:
		return theme.LerpFG(theme.Base, theme.Shade, (t-0.70)/0.30)
	}
}

// Logo prints the gust word-mark with a soft violet sheen, followed by a
// tagline. It is a no-op when fancy output is disabled.
func (r *Renderer) Logo(tagline string) {
	if !r.fancy {
		return
	}
	rows := len(logoLines)
	r.line("")
	for row, ln := range logoLines {
		chars := []rune(ln)
		last := len(chars) - 1
		if last < 1 {
			last = 1
		}
		line := "  "
		for i, ch := range chars {
			if ch == ' ' {
				line += " "
				continue
			}
			tX := float64(i) / float64(last)
			tY := float64(row) / float64(rows-1)
			line += sheen((tX+tY)/2) + string(ch)
		}
		line += theme.Reset
		r.line(line)
	}
	if tagline != "" {
		r.line(fmt.Sprintf("  %s%s%s", theme.Dim, tagline, theme.Reset))
	}
	r.line("")
}
