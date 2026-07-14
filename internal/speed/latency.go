package speed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Latency summarises a series of round-trip time samples.
type Latency struct {
	Min     time.Duration
	Avg     time.Duration
	Max     time.Duration
	Jitter  time.Duration   // mean absolute deviation between consecutive samples
	Samples int             // number of probes taken
	Series  []time.Duration // individual round-trip times, in order
}

// MeasureLatency issues `samples` tiny requests and reports round-trip
// statistics. A zero-byte download is used so that the measured time is
// dominated by the network round-trip rather than payload transfer.
func MeasureLatency(ctx context.Context, client *http.Client, samples int) (Latency, error) {
	if samples < 1 {
		samples = 1
	}
	url := fmt.Sprintf("%s/__down?bytes=0", endpoint)

	// Warm up the connection first. The initial request pays for DNS, the TCP
	// handshake and the TLS handshake; charging that to the first sample would
	// inflate the average and jitter by hundreds of milliseconds. Its timing is
	// discarded so every recorded sample reuses the established connection.
	if _, err := probe(ctx, client, url); err != nil {
		return Latency{}, err
	}

	rtts := make([]time.Duration, 0, samples)
	for i := 0; i < samples; i++ {
		d, err := probe(ctx, client, url)
		if err != nil {
			return Latency{}, err
		}
		rtts = append(rtts, d)
	}
	return summarise(rtts), nil
}

// probe issues a single zero-byte request and returns its round-trip time.
func probe(ctx context.Context, client *http.Client, url string) (time.Duration, error) {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("build latency request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("latency probe: %w", err)
	}
	// Drain and close so the connection can be reused for the next probe.
	_, _ = io.Copy(io.Discard, resp.Body)
	resp.Body.Close()

	return time.Since(start), nil
}

// summarise computes min/avg/max/jitter from raw round-trip samples.
func summarise(rtts []time.Duration) Latency {
	l := Latency{Samples: len(rtts), Min: rtts[0], Max: rtts[0], Series: rtts}

	var sum time.Duration
	for _, d := range rtts {
		if d < l.Min {
			l.Min = d
		}
		if d > l.Max {
			l.Max = d
		}
		sum += d
	}
	l.Avg = sum / time.Duration(len(rtts))

	// Jitter is the mean absolute difference between consecutive samples.
	if len(rtts) > 1 {
		var jitterSum time.Duration
		for i := 1; i < len(rtts); i++ {
			diff := rtts[i] - rtts[i-1]
			if diff < 0 {
				diff = -diff
			}
			jitterSum += diff
		}
		l.Jitter = jitterSum / time.Duration(len(rtts)-1)
	}
	return l
}
