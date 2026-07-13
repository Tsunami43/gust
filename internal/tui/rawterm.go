package tui

import (
	"os"
	"os/exec"
)

// rawMode switches the controlling terminal to raw (no-echo, character-at-a-
// time) mode via the `stty` utility and returns a function that restores the
// previous settings. Shelling out to stty keeps gust free of any Go module
// dependencies while still supporting interactive key input.
func rawMode() (restore func(), err error) {
	saved, err := stty("-g")
	if err != nil {
		return func() {}, err
	}
	// -echo: do not print typed keys.
	// -icanon: deliver input byte-by-byte instead of line-by-line.
	// min 1 time 0: each read blocks until at least one byte is available.
	if _, err := stty("-echo", "-icanon", "min", "1", "time", "0"); err != nil {
		return func() {}, err
	}
	return func() { _, _ = stty(saved) }, nil
}

// stty runs the stty utility against the controlling terminal and returns its
// trimmed standard output.
func stty(args ...string) (string, error) {
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return "", err
	}
	defer tty.Close()

	cmd := exec.Command("stty", args...)
	cmd.Stdin = tty
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return trimNewline(string(out)), nil
}

// trimNewline removes trailing carriage returns and newlines.
func trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

// readKey blocks for a single decoded keypress read from in.
func readKey(in *os.File) Key {
	buf := make([]byte, 8)
	n, err := in.Read(buf)
	if err != nil || n == 0 {
		return KeyUnknown
	}
	return decode(buf[:n])
}

// WaitKey briefly enters raw mode to wait for a single keypress. It is used to
// pause after showing results before returning to a menu.
func WaitKey(in *os.File) {
	restore, err := rawMode()
	if err != nil {
		return
	}
	defer restore()
	readKey(in)
}
