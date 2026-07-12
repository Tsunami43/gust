package ui

import "testing"

func TestHumanBits(t *testing.T) {
	cases := []struct {
		bps  float64
		want string
	}{
		{0, "0.00 bps"},
		{512, "512.00 bps"},
		{1000, "1.00 Kbps"},
		{94_210_000, "94.21 Mbps"},
		{2_500_000_000, "2.50 Gbps"},
	}
	for _, c := range cases {
		if got := HumanBits(c.bps); got != c.want {
			t.Errorf("HumanBits(%v) = %q, want %q", c.bps, got, c.want)
		}
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		b    int64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KiB"},
		{26_214_400, "25.0 MiB"},
	}
	for _, c := range cases {
		if got := HumanBytes(c.b); got != c.want {
			t.Errorf("HumanBytes(%d) = %q, want %q", c.b, got, c.want)
		}
	}
}

func TestVisibleLen(t *testing.T) {
	// "abc" coloured with ANSI codes still measures three visible columns.
	coloured := "\033[1m\033[36mabc\033[0m"
	if got := visibleLen(coloured); got != 3 {
		t.Errorf("visibleLen = %d, want 3", got)
	}
	if got := visibleLen("héllo"); got != 5 {
		t.Errorf("visibleLen(unicode) = %d, want 5", got)
	}
}

func TestSparkline(t *testing.T) {
	if got := Sparkline(nil); got != "" {
		t.Errorf("Sparkline(nil) = %q, want empty", got)
	}
	// A rising ramp should start at the lowest block and end at the highest.
	line := []rune(Sparkline([]float64{1, 2, 3, 4, 5, 6, 7, 8}))
	if line[0] != '▁' || line[len(line)-1] != '█' {
		t.Errorf("Sparkline ramp = %q", string(line))
	}
	// Flat input must not divide by zero and should render the lowest block.
	if got := Sparkline([]float64{5, 5, 5}); got != "▁▁▁" {
		t.Errorf("Sparkline(flat) = %q, want ▁▁▁", got)
	}
}

func TestGradeResult(t *testing.T) {
	cases := []struct {
		down, lat float64
		want      string
	}{
		{300, 10, "A+"},
		{120, 20, "A"},
		{60, 40, "B"},
		{30, 40, "C"},
		{10, 90, "D"},
		{1, 200, "E"},
	}
	for _, c := range cases {
		if got := GradeResult(c.down, c.lat).Letter; got != c.want {
			t.Errorf("GradeResult(%v,%v) = %q, want %q", c.down, c.lat, got, c.want)
		}
	}
}
