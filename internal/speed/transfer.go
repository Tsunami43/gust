package speed

import (
	"context"
	"sync/atomic"
	"time"
)

// splitSizes divides total into n parts as evenly as possible, handing the
// remainder to the earliest parts so the sizes always sum back to total
// exactly. This avoids the byte loss of a plain total/n division, so a test
// transfers the full requested amount.
func splitSizes(total int64, n int) []int64 {
	if n < 1 {
		n = 1
	}
	if total < 0 {
		total = 0
	}
	base := total / int64(n)
	rem := total % int64(n)
	sizes := make([]int64, n)
	for i := range sizes {
		sizes[i] = base
		if int64(i) < rem {
			sizes[i]++
		}
	}
	return sizes
}

// checkpoint records how many bytes had moved at a given instant.
type checkpoint struct {
	bytes int64
	at    time.Time
}

// slowStartTrimmer watches a shared progress counter during a parallel
// transfer and records a checkpoint once warmupBytes have moved. Measuring
// throughput from that checkpoint to the end excludes the TCP slow-start ramp,
// which would otherwise deflate the figure on fast links (the connection is
// still accelerating when a byte-bounded test finishes).
type slowStartTrimmer struct {
	start time.Time
	stop  chan struct{}
	ck    chan checkpoint
}

// newSlowStartTrimmer starts watching progress. Passing warmupBytes <= 0
// disables trimming, and the whole transfer is measured instead.
func newSlowStartTrimmer(progress *int64, warmupBytes int64, start time.Time) *slowStartTrimmer {
	t := &slowStartTrimmer{start: start, stop: make(chan struct{}), ck: make(chan checkpoint, 1)}
	if warmupBytes <= 0 {
		return t
	}
	go func() {
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-t.stop:
				return
			case <-ticker.C:
				if p := atomic.LoadInt64(progress); p >= warmupBytes {
					t.ck <- checkpoint{bytes: p, at: time.Now()}
					return
				}
			}
		}
	}()
	return t
}

// result finalises the measurement for a transfer that reached `final` bytes
// and ended at `end`. Bytes and Elapsed always describe the whole transfer;
// the steady-state rate is filled in when a post-warmup window was captured.
func (t *slowStartTrimmer) result(final int64, end time.Time) Result {
	close(t.stop)
	res := Result{Bytes: final, Elapsed: end.Sub(t.start)}
	select {
	case ck := <-t.ck:
		if d := end.Sub(ck.at); d > 0 {
			res.rate = float64(final-ck.bytes) * 8 / d.Seconds()
		}
	default:
	}
	return res
}

// transferError inspects the per-stream errors of a parallel transfer. A single
// dropped connection should not void an otherwise good measurement, so it
// returns an error only when the context was cancelled or every launched
// stream failed. errs must already be closed.
func transferError(ctx context.Context, errs <-chan error, launched int) error {
	var failed int
	var first error
	for e := range errs {
		failed++
		if first == nil {
			first = e
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if launched > 0 && failed == launched {
		return first
	}
	return nil
}
