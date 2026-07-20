package speed

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// Download measures download throughput by fetching `totalBytes` of random
// data from the speed endpoint, split evenly across `streams` parallel
// connections. The returned Result covers the aggregate transfer.
//
// If progress is non-nil it is updated atomically with the running byte count
// as data arrives, so a caller can render a live progress bar.
func Download(ctx context.Context, client *http.Client, totalBytes int64, streams int, progress *int64) (Result, error) {
	if streams < 1 {
		streams = 1
	}
	if progress == nil {
		progress = new(int64)
	}
	sizes := splitSizes(totalBytes, streams)

	var wg sync.WaitGroup
	errs := make(chan error, len(sizes))
	launched := 0

	start := time.Now()
	trimmer := newSlowStartTrimmer(progress, totalBytes/10, start)

	for _, sz := range sizes {
		if sz <= 0 {
			continue
		}
		launched++
		wg.Add(1)
		go func(sz int64) {
			defer wg.Done()

			url := fmt.Sprintf("%s/__down?bytes=%d", endpoint, sz)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
			if err != nil {
				errs <- err
				return
			}
			resp, err := client.Do(req)
			if err != nil {
				errs <- err
				return
			}
			defer resp.Body.Close()

			// Discard the payload while counting the bytes as they arrive.
			if _, err := io.Copy(&countingWriter{counter: progress}, resp.Body); err != nil {
				errs <- err
			}
		}(sz)
	}
	wg.Wait()
	res := trimmer.result(atomic.LoadInt64(progress), time.Now())

	close(errs)
	if err := transferError(ctx, errs, launched); err != nil {
		return Result{}, fmt.Errorf("download: %w", err)
	}
	return res, nil
}

// countingWriter discards everything written to it while atomically counting
// the number of bytes. Multiple download streams share one counter safely.
type countingWriter struct {
	counter *int64
}

func (w *countingWriter) Write(p []byte) (int, error) {
	atomic.AddInt64(w.counter, int64(len(p)))
	return len(p), nil
}
