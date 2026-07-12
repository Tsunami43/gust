package config

import (
	"os"
	"testing"
)

func TestLoadDefaultWhenMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg != Default() {
		t.Errorf("Load() = %+v, want defaults %+v", cfg, Default())
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	want := Config{SizeMB: 50, Streams: 8, Pings: 10, NoColor: true}
	if err := Save(want); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("round-trip = %+v, want %+v", got, want)
	}

	// The file must actually exist at the resolved path.
	path, _ := Path()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config file not written: %v", err)
	}
}
