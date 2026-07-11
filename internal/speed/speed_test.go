package speed

import (
	"io"
	"testing"
	"time"
)

func TestResultBitsPerSecond(t *testing.T) {
	// 1,000,000 bytes in one second == 8 Mbps.
	r := Result{Bytes: 1_000_000, Elapsed: time.Second}
	if got := r.BitsPerSecond(); got != 8_000_000 {
		t.Errorf("BitsPerSecond = %v, want 8000000", got)
	}
	// Zero elapsed must not divide by zero.
	if got := (Result{Bytes: 10}).BitsPerSecond(); got != 0 {
		t.Errorf("BitsPerSecond(zero elapsed) = %v, want 0", got)
	}
}

func TestSummarise(t *testing.T) {
	rtts := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		30 * time.Millisecond,
	}
	l := summarise(rtts)
	if l.Min != 10*time.Millisecond || l.Max != 30*time.Millisecond {
		t.Errorf("min/max = %v/%v", l.Min, l.Max)
	}
	if l.Avg != 20*time.Millisecond {
		t.Errorf("avg = %v, want 20ms", l.Avg)
	}
	if l.Samples != 3 || len(l.Series) != 3 {
		t.Errorf("samples/series = %d/%d", l.Samples, len(l.Series))
	}
	// Jitter: |20-10| + |30-20| averaged == 10ms.
	if l.Jitter != 10*time.Millisecond {
		t.Errorf("jitter = %v, want 10ms", l.Jitter)
	}
}

func TestCountingReader(t *testing.T) {
	var counter int64
	r := &countingReader{remaining: 2500, counter: &counter}
	n, err := io.Copy(io.Discard, r)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2500 || counter != 2500 {
		t.Errorf("read %d bytes (counter %d), want 2500", n, counter)
	}
}
