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
func Download(ctx context.Context, client *http.Client, totalBytes int64, streams int) (Result, error) {
	if streams < 1 {
		streams = 1
	}
	perStream := totalBytes / int64(streams)
	if perStream <= 0 {
		// Not enough data to split; fall back to a single stream.
		perStream = totalBytes
		streams = 1
	}

	var transferred int64
	var wg sync.WaitGroup
	errs := make(chan error, streams)

	start := time.Now()
	for i := 0; i < streams; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			url := fmt.Sprintf("%s/__down?bytes=%d", endpoint, perStream)
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
			if _, err := io.Copy(&countingWriter{counter: &transferred}, resp.Body); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	close(errs)
	if err := <-errs; err != nil {
		return Result{}, fmt.Errorf("download: %w", err)
	}
	return Result{Bytes: atomic.LoadInt64(&transferred), Elapsed: elapsed}, nil
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
