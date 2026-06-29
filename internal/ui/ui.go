// Package ui renders network information and speed results for humans
// (aligned plain text) and for machines (indented JSON).
package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Tsunami43/gust/internal/netinfo"
	"github.com/Tsunami43/gust/internal/speed"
)

// Report aggregates everything a single gust run has measured. Any field may
// be nil when the corresponding step was skipped or failed.
type Report struct {
	Network  *netinfo.Meta
	Latency  *speed.Latency
	Download *speed.Result
	Upload   *speed.Result
}

// WriteText renders the report as aligned, human-readable text.
func (r Report) WriteText(w io.Writer) {
	if r.Network != nil {
		n := r.Network
		fmt.Fprintln(w, "Network")
		fmt.Fprintf(w, "  IP        %s\n", orNA(n.IP))
		fmt.Fprintf(w, "  Location  %s\n", location(n))
		fmt.Fprintf(w, "  Provider  %s\n", provider(n))
		if n.Colo != "" || n.Protocol != "" {
			fmt.Fprintf(w, "  Edge      %s\n", edge(n))
		}
		if r.Latency != nil || r.Download != nil || r.Upload != nil {
			fmt.Fprintln(w)
		}
	}

	if l := r.Latency; l != nil {
		fmt.Fprintf(w, "Latency   %s  (min %s / max %s / jitter %s)\n",
			ms(l.Avg), ms(l.Min), ms(l.Max), ms(l.Jitter))
	}
	if d := r.Download; d != nil {
		fmt.Fprintf(w, "Download  %s  (%s in %s)\n",
			HumanBits(d.BitsPerSecond()), HumanBytes(d.Bytes), shortDur(d.Elapsed))
	}
	if u := r.Upload; u != nil {
		fmt.Fprintf(w, "Upload    %s  (%s in %s)\n",
			HumanBits(u.BitsPerSecond()), HumanBytes(u.Bytes), shortDur(u.Elapsed))
	}
}

// WriteJSON renders the report as indented JSON on a single top-level object.
func (r Report) WriteJSON(w io.Writer) error {
	view := make(map[string]any)
	if r.Network != nil {
		view["network"] = r.Network
	}
	if l := r.Latency; l != nil {
		view["latency"] = map[string]any{
			"min_ms":    round2(l.Min),
			"avg_ms":    round2(l.Avg),
			"max_ms":    round2(l.Max),
			"jitter_ms": round2(l.Jitter),
			"samples":   l.Samples,
		}
	}
	if d := r.Download; d != nil {
		view["download"] = throughputJSON(*d)
	}
	if u := r.Upload; u != nil {
		view["upload"] = throughputJSON(*u)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(view)
}

// throughputJSON is the machine-readable view of a throughput measurement.
func throughputJSON(res speed.Result) map[string]any {
	return map[string]any{
		"bytes":        res.Bytes,
		"seconds":      round2seconds(res.Elapsed),
		"mbps":         round2f(res.BitsPerSecond() / 1e6),
		"bits_per_sec": int64(res.BitsPerSecond()),
	}
}

// --- text helpers -----------------------------------------------------------

func location(n *netinfo.Meta) string {
	parts := make([]string, 0, 3)
	for _, p := range []string{n.City, n.Region, n.Country} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return "n/a"
	}
	return strings.Join(parts, ", ")
}

func provider(n *netinfo.Meta) string {
	switch {
	case n.ASN != 0 && n.Org != "":
		return fmt.Sprintf("AS%d %s", n.ASN, n.Org)
	case n.Org != "":
		return n.Org
	case n.ASN != 0:
		return fmt.Sprintf("AS%d", n.ASN)
	default:
		return "n/a"
	}
}

func edge(n *netinfo.Meta) string {
	switch {
	case n.Colo != "" && n.Protocol != "":
		return fmt.Sprintf("%s via %s", n.Colo, n.Protocol)
	case n.Colo != "":
		return n.Colo
	default:
		return n.Protocol
	}
}

func orNA(s string) string {
	if s == "" {
		return "n/a"
	}
	return s
}

// HumanBits formats a bits-per-second value using decimal (SI) units.
func HumanBits(bps float64) string {
	const unit = 1000.0
	units := []string{"bps", "Kbps", "Mbps", "Gbps", "Tbps"}
	v, i := bps, 0
	for v >= unit && i < len(units)-1 {
		v /= unit
		i++
	}
	return fmt.Sprintf("%.2f %s", v, units[i])
}

// HumanBytes formats a byte count using binary (IEC) units.
func HumanBytes(b int64) string {
	const unit = 1024.0
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	v, i := float64(b), 0
	for v >= unit && i < len(units)-1 {
		v /= unit
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d %s", b, units[i])
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}

// ms renders a duration in milliseconds with two decimals (e.g. "12.34 ms").
func ms(d time.Duration) string {
	return fmt.Sprintf("%.2f ms", float64(d.Microseconds())/1000)
}

// shortDur renders a duration compactly (e.g. "2.1s").
func shortDur(d time.Duration) string {
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// --- json numeric helpers ---------------------------------------------------

func round2(d time.Duration) float64        { return round2f(float64(d.Microseconds()) / 1000) }
func round2seconds(d time.Duration) float64 { return round2f(d.Seconds()) }

func round2f(f float64) float64 {
	return float64(int64(f*100+0.5)) / 100
}
