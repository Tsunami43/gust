package tui

// Key represents a decoded keypress.
type Key int

const (
	KeyUnknown Key = iota
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyEnter
	KeyEscape
	KeyQuit  // 'q'
	KeyCtrlC // Ctrl-C
	KeySpace
)

// decode interprets a raw byte sequence read from the terminal into a Key.
// It understands ANSI arrow-key sequences as well as vi-style j/k and a few
// control characters.
func decode(buf []byte) Key {
	if len(buf) == 0 {
		return KeyUnknown
	}
	switch buf[0] {
	case '\r', '\n':
		return KeyEnter
	case 3:
		return KeyCtrlC
	case 27: // ESC or the start of a CSI sequence
		if len(buf) >= 3 && buf[1] == '[' {
			switch buf[2] {
			case 'A':
				return KeyUp
			case 'B':
				return KeyDown
			case 'C':
				return KeyRight
			case 'D':
				return KeyLeft
			}
		}
		return KeyEscape
	case 'q', 'Q':
		return KeyQuit
	case 'k':
		return KeyUp
	case 'j':
		return KeyDown
	case ' ':
		return KeySpace
	}
	return KeyUnknown
}
