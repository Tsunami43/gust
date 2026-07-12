package ui

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Tsunami43/gust/internal/netinfo"
	"github.com/Tsunami43/gust/internal/speed"
)

// barWidth is the number of cells used to draw throughput progress bars.
const barWidth = 22

// spinnerFrames animates in-progress work using Braille dots.
var spinnerFrames = []rune("⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏")

// tickInterval controls the animation frame rate.
const tickInterval = 90 * time.Millisecond

// Renderer draws gust's rich, animated terminal interface. When fancy output
// is disabled (no TTY, --no-color, NO_COLOR or JSON mode) every method still
// runs the supplied work but produces no decoration, so the caller can fall
// back to plain text or JSON.
type Renderer struct {
	out   *os.File
	fancy bool
	pal   palette
}

// NewRenderer builds a Renderer. Fancy output is enabled only when colour is
// requested and the destination is an interactive terminal.
func NewRenderer(out *os.File, color bool) *Renderer {
	fancy := color && isTerminal(out)
	return &Renderer{out: out, fancy: fancy, pal: newPalette(fancy)}
}

// Fancy reports whether the renderer draws the rich interface.
func (r *Renderer) Fancy() bool { return r.fancy }

// draw overwrites the current terminal line with s (no trailing newline).
func (r *Renderer) draw(s string) { fmt.Fprint(r.out, "\r\033[K"+s) }

// line writes s followed by a newline.
func (r *Renderer) line(s string) { fmt.Fprintln(r.out, s) }

// nl ends the current line.
func (r *Renderer) nl() { fmt.Fprint(r.out, "\n") }

// clear erases the current line and returns the cursor to its start.
func (r *Renderer) clear() { fmt.Fprint(r.out, "\r\033[K") }

// Header prints the program banner.
func (r *Renderer) Header(version string) {
	if !r.fancy {
		return
	}
	p := r.pal
	r.line("")
	r.line(fmt.Sprintf("  %s⚡%s %sgust%s  %sv%s%s  %s·%s  %sinternet speed test%s",
		p.yellow, p.reset, p.bold, p.reset, p.dim, version, p.reset,
		p.dim, p.reset, p.dim, p.reset))
	r.line("")
}

// Footer adds trailing breathing room after the results.
func (r *Renderer) Footer() {
	if !r.fancy {
		return
	}
	r.line("")
}

// NetworkCard renders public network information inside a rounded box.
func (r *Renderer) NetworkCard(m netinfo.Meta) {
	if !r.fancy {
		return
	}
	rows := [][2]string{
		{"IP", orNA(m.IP)},
		{"Location", location(&m)},
		{"Provider", provider(&m)},
	}
	if server := serverLine(m); server != "" {
		rows = append(rows, [2]string{"Server", server})
	}
	r.card("Network", rows)
	r.line("")
}

// LatencyLine prints the finished latency measurement, including a sparkline
// of the individual round-trip samples.
func (r *Renderer) LatencyLine(l speed.Latency) {
	if !r.fancy {
		return
	}
	p := r.pal
	col := latencyColor(l.Avg, p)
	spark := ""
	if s := latencySpark(l); s != "" {
		spark = fmt.Sprintf("   %s%s%s", p.dim, s, p.reset)
	}
	r.line(fmt.Sprintf("  %s✔%s %-9s %s%s%s   %s· jitter %s%s%s",
		p.green, p.reset, "Latency", col, ms(l.Avg), p.reset,
		p.dim, ms(l.Jitter), p.reset, spark))
}

// SummaryCard renders a compact quality summary once a full test has finished.
func (r *Renderer) SummaryCard(rep Report) {
	if !r.fancy || rep.Download == nil {
		return
	}
	p := r.pal

	downMbps := rep.Download.BitsPerSecond() / 1e6
	latMs := 0.0
	if rep.Latency != nil {
		latMs = float64(rep.Latency.Avg.Microseconds()) / 1000
	}
	g := GradeResult(downMbps, latMs)
	gc := gradeColor(g.Letter, p)

	up := "—"
	if rep.Upload != nil {
		up = HumanBits(rep.Upload.BitsPerSecond())
	}
	lat := "—"
	if rep.Latency != nil {
		lat = ms(rep.Latency.Avg)
	}

	r.line("")
	r.line(fmt.Sprintf("  %s %s %s  %s%s%s   %s↓%s %s   %s↑%s %s   %s%s%s",
		gc+p.bold, g.Letter, p.reset,
		gc, g.Label, p.reset,
		p.green, p.reset, HumanBits(rep.Download.BitsPerSecond()),
		p.cyan, p.reset, up,
		p.dim, lat, p.reset))
}

// WatchLine prints a single sample line for watch mode: a timestamp label, the
// latest download speed and a rolling sparkline of recent speeds.
func (r *Renderer) WatchLine(label string, mbps float64, history []float64) {
	p := r.pal
	spark := Sparkline(history)
	r.line(fmt.Sprintf("  %s%s%s  %s↓%s %s%8.2f Mbps%s  %s%s%s",
		p.dim, label, p.reset,
		p.green, p.reset, p.bold, mbps, p.reset,
		p.cyan, spark, p.reset))
}

// RunSpinner runs fn while animating a spinner labelled with text. When fancy
// output is disabled it simply runs fn. The line is cleared on completion so
// the caller can print a final result.
func (r *Renderer) RunSpinner(text string, fn func() error) error {
	if !r.fancy {
		return fn()
	}
	p := r.pal
	ch := make(chan error, 1)
	go func() { ch <- fn() }()

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for frame := 0; ; frame++ {
		select {
		case err := <-ch:
			r.clear()
			return err
		case <-ticker.C:
			sp := string(spinnerFrames[frame%len(spinnerFrames)])
			r.draw(fmt.Sprintf("  %s%s%s %s%s…%s", p.cyan, sp, p.reset, p.dim, text, p.reset))
		}
	}
}

// RunBar runs fn while animating a live progress bar for a throughput test.
// fn receives a counter it must update atomically as bytes are transferred.
// When fancy output is disabled it simply runs fn with a nil counter.
func (r *Renderer) RunBar(label string, target int64, fn func(progress *int64) (speed.Result, error)) (speed.Result, error) {
	if !r.fancy {
		return fn(nil)
	}
	p := r.pal
	var progress int64

	type outcome struct {
		res speed.Result
		err error
	}
	ch := make(chan outcome, 1)
	start := time.Now()
	go func() {
		res, err := fn(&progress)
		ch <- outcome{res, err}
	}()

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for frame := 0; ; frame++ {
		select {
		case o := <-ch:
			if o.err != nil {
				r.clear()
				return o.res, o.err
			}
			r.draw(fmt.Sprintf("  %s✔%s %-9s %s  %s%s%s",
				p.green, p.reset, label, renderBar(1, barWidth, p),
				p.bold, HumanBits(o.res.BitsPerSecond()), p.reset))
			r.nl()
			return o.res, nil
		case <-ticker.C:
			done := atomic.LoadInt64(&progress)
			frac := 0.0
			if target > 0 {
				frac = float64(done) / float64(target)
			}
			var bps float64
			if secs := time.Since(start).Seconds(); secs > 0 {
				bps = float64(done) * 8 / secs
			}
			sp := string(spinnerFrames[frame%len(spinnerFrames)])
			r.draw(fmt.Sprintf("  %s%s%s %-9s %s  %s%3d%%%s  %s%s%s",
				p.cyan, sp, p.reset, label, renderBar(frac, barWidth, p),
				p.dim, int(frac*100), p.reset, p.dim, HumanBits(bps), p.reset))
		}
	}
}

// card draws a rounded box titled `title` with the given label/value rows.
func (r *Renderer) card(title string, rows [][2]string) {
	p := r.pal

	labelW := 0
	for _, row := range rows {
		if l := visibleLen(row[0]); l > labelW {
			labelW = l
		}
	}

	// Determine the inner width from the widest content row and the title.
	plains := make([]string, len(rows))
	inner := visibleLen(title) + 2
	for i, row := range rows {
		plains[i] = fmt.Sprintf("%-*s   %s", labelW, row[0], row[1])
		if l := visibleLen(plains[i]); l > inner {
			inner = l
		}
	}

	// Top border with embedded title: "╭─ Network ─────╮".
	dashes := inner - visibleLen(title) - 1
	if dashes < 1 {
		dashes = 1
	}
	r.line(fmt.Sprintf("  %s╭─ %s%s%s %s%s╮%s",
		p.dim, p.reset+p.bold, title, p.reset+p.dim, strings.Repeat("─", dashes), p.dim, p.reset))

	// Content rows.
	for i, row := range rows {
		pad := strings.Repeat(" ", inner-visibleLen(plains[i]))
		content := fmt.Sprintf("%s%-*s%s   %s%s%s",
			p.dim, labelW, row[0], p.reset, p.bold, row[1], p.reset)
		r.line(fmt.Sprintf("  %s│%s %s%s %s│%s", p.dim, p.reset, content, pad, p.dim, p.reset))
	}

	// Bottom border.
	r.line(fmt.Sprintf("  %s╰%s╯%s", p.dim, strings.Repeat("─", inner+2), p.reset))
}

// serverLine joins the edge PoP, HTTP protocol and provider name of the meta
// source into a single descriptive string.
func serverLine(m netinfo.Meta) string {
	var parts []string
	for _, s := range []string{m.Colo, m.Protocol, m.Source} {
		if s != "" {
			parts = append(parts, s)
		}
	}
	return strings.Join(parts, " · ")
}

// latencyColor grades a round-trip time green/yellow/red.
func latencyColor(d time.Duration, p palette) string {
	switch {
	case d < 30*time.Millisecond:
		return p.green
	case d < 100*time.Millisecond:
		return p.yellow
	default:
		return p.red
	}
}

// latencySpark renders the individual latency samples as a sparkline.
func latencySpark(l speed.Latency) string {
	if len(l.Series) < 2 {
		return ""
	}
	vals := make([]float64, len(l.Series))
	for i, d := range l.Series {
		vals[i] = float64(d.Microseconds()) / 1000
	}
	return Sparkline(vals)
}

// gradeColor maps a quality grade letter to a palette colour.
func gradeColor(letter string, p palette) string {
	switch letter {
	case "A+", "A":
		return p.green
	case "B":
		return p.cyan
	case "C":
		return p.yellow
	default:
		return p.red
	}
}
