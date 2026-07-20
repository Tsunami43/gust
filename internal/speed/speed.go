// Package speed measures internet connection performance — round-trip
// latency together with download and upload throughput — using the public
// Cloudflare speed test endpoints (https://speed.cloudflare.com).
package speed

import "time"

// endpoint is the base URL of the Cloudflare speed test service. It is a
// variable rather than a constant so tests can redirect it at a local server.
var endpoint = "https://speed.cloudflare.com"

// Result describes the outcome of a single throughput measurement. Bytes and
// Elapsed cover the whole transfer; rate, when set, is the steady-state
// throughput measured after the TCP slow-start ramp has been discarded.
type Result struct {
	Bytes   int64         // number of bytes transferred
	Elapsed time.Duration // wall-clock time spent transferring them
	rate    float64       // steady-state bits/sec, or 0 when unmeasured
}

// BitsPerSecond returns the measured throughput in bits per second. It prefers
// the steady-state rate (which excludes slow-start) and otherwise falls back
// to the whole-transfer average. It returns 0 when no time has elapsed,
// avoiding a division by zero.
func (r Result) BitsPerSecond() float64 {
	if r.rate > 0 {
		return r.rate
	}
	if r.Elapsed <= 0 {
		return 0
	}
	return float64(r.Bytes) * 8 / r.Elapsed.Seconds()
}
