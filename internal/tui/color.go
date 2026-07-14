// Package tui provides a minimal, dependency-free interactive terminal
// interface (menus and key handling) built on ANSI escape codes and the
// system `stty` utility for raw input. Colours follow the shared violet theme.
package tui

import "github.com/Tsunami43/gust/internal/theme"

// Cursor visibility escape sequences.
const (
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
)

func bold(s string) string   { return theme.Bold + s + theme.Reset }
func dim(s string) string    { return theme.Dim + s + theme.Reset }
func accent(s string) string { return theme.FG(theme.Accent) + s + theme.Reset }
func alt(s string) string    { return theme.FG(theme.Alt) + s + theme.Reset }
