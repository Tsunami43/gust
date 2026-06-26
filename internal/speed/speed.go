// Package speed measures internet connection performance — round-trip
// latency together with download and upload throughput — using the public
// Cloudflare speed test endpoints (https://speed.cloudflare.com).
package speed

import "time"

// endpoint is the base URL of the Cloudflare speed test service.
const endpoint = "https://speed.cloudflare.com"

// Result describes the outcome of a single throughput measurement.
type Result struct {
	Bytes   int64         // number of bytes transferred
	Elapsed time.Duration // wall-clock time spent transferring them
}

// BitsPerSecond returns the measured throughput in bits per second.
// It returns 0 when no time has elapsed, avoiding a division by zero.
func (r Result) BitsPerSecond() float64 {
	if r.Elapsed <= 0 {
		return 0
	}
	return float64(r.Bytes) * 8 / r.Elapsed.Seconds()
}
