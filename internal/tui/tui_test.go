package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecode(t *testing.T) {
	cases := []struct {
		name string
		in   []byte
		want Key
	}{
		{"enter-cr", []byte{'\r'}, KeyEnter},
		{"enter-lf", []byte{'\n'}, KeyEnter},
		{"ctrl-c", []byte{3}, KeyCtrlC},
		{"up-arrow", []byte{27, '[', 'A'}, KeyUp},
		{"down-arrow", []byte{27, '[', 'B'}, KeyDown},
		{"right-arrow", []byte{27, '[', 'C'}, KeyRight},
		{"left-arrow", []byte{27, '[', 'D'}, KeyLeft},
		{"escape", []byte{27}, KeyEscape},
		{"quit", []byte{'q'}, KeyQuit},
		{"vi-k", []byte{'k'}, KeyUp},
		{"vi-j", []byte{'j'}, KeyDown},
		{"space", []byte{' '}, KeySpace},
		{"unknown", []byte{'x'}, KeyUnknown},
		{"empty", nil, KeyUnknown},
	}
	for _, c := range cases {
		if got := decode(c.in); got != c.want {
			t.Errorf("decode(%v) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestMenuRender(t *testing.T) {
	path := filepath.Join(t.TempDir(), "frame")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	items := []Item{{Label: "Alpha", Desc: "first"}, {Label: "Beta"}}
	m := NewMenu("pick one", items, nil, f)
	m.render(1, true) // second item selected

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := string(data)
	for _, want := range []string{"pick one", "Alpha", "Beta", "❯"} {
		if !strings.Contains(out, want) {
			t.Errorf("rendered frame missing %q\n%s", want, out)
		}
	}
}
