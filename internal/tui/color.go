// Package tui provides a minimal, dependency-free interactive terminal
// interface (menus and key handling) built on ANSI escape codes and the
// system `stty` utility for raw input.
package tui

// ANSI escape sequences used by the interactive interface.
const (
	reset      = "\033[0m"
	boldSeq    = "\033[1m"
	dimSeq     = "\033[2m"
	cyanSeq    = "\033[36m"
	yellowSeq  = "\033[33m"
	greenSeq   = "\033[32m"
	hideCursor = "\033[?25l"
	showCursor = "\033[?25h"
)

func bold(s string) string   { return boldSeq + s + reset }
func dim(s string) string    { return dimSeq + s + reset }
func cyan(s string) string   { return cyanSeq + s + reset }
func yellow(s string) string { return yellowSeq + s + reset }
func green(s string) string  { return greenSeq + s + reset }
