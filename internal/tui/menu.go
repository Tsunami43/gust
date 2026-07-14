package tui

import (
	"fmt"
	"os"

	"github.com/Tsunami43/gust/internal/theme"
)

// Item is a selectable menu entry.
type Item struct {
	Label string
	Desc  string
}

// Menu is an interactive, arrow-navigable selection list.
type Menu struct {
	Title string
	Items []Item
	in    *os.File
	out   *os.File
}

// NewMenu builds a Menu that reads from in and draws to out.
func NewMenu(title string, items []Item, in, out *os.File) *Menu {
	return &Menu{Title: title, Items: items, in: in, out: out}
}

// Run displays the menu and returns the selected index, or ok=false if the
// user cancelled with Esc, q or Ctrl-C.
func (m *Menu) Run() (index int, ok bool) {
	if len(m.Items) == 0 {
		return 0, false
	}
	restore, err := rawMode()
	if err != nil {
		return 0, false
	}
	defer restore()

	fmt.Fprint(m.out, hideCursor)
	defer fmt.Fprint(m.out, showCursor)

	sel, first := 0, true
	for {
		m.render(sel, first)
		first = false

		switch readKey(m.in) {
		case KeyUp:
			sel = (sel - 1 + len(m.Items)) % len(m.Items)
		case KeyDown:
			sel = (sel + 1) % len(m.Items)
		case KeyEnter, KeySpace:
			return sel, true
		case KeyQuit, KeyEscape, KeyCtrlC:
			return 0, false
		}
	}
}

// lineCount is the number of terminal lines a full render occupies.
func (m *Menu) lineCount() int { return len(m.Items) + 2 } // title + items + hint

// render draws the menu, moving the cursor back up over the previous frame on
// every redraw after the first.
func (m *Menu) render(sel int, first bool) {
	if !first {
		fmt.Fprintf(m.out, "\033[%dA", m.lineCount())
	}
	title := theme.FG(theme.Accent) + theme.Bold + m.Title + theme.Reset
	fmt.Fprintf(m.out, "\r\033[K  %s\n", title)
	for i, it := range m.Items {
		marker, label := "  ", dim(it.Label)
		if i == sel {
			marker = theme.FG(theme.Accent) + theme.Pointer + theme.Reset + " "
			label = theme.FG(theme.Accent) + theme.Bold + it.Label + theme.Reset
		}
		line := "  " + marker + label
		if it.Desc != "" {
			line += "  " + dim(it.Desc)
		}
		fmt.Fprintf(m.out, "\r\033[K%s\n", line)
	}
	hint := alt("↑↓") + dim(" move") + "   " + alt("↵") + dim(" select") + "   " + alt("q") + dim(" back")
	fmt.Fprintf(m.out, "\r\033[K  %s\n", hint)
}
