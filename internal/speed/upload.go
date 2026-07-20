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

// Upload measures upload throughput by POSTing `totalBytes` of generated data
// to the speed endpoint, split evenly across `streams` parallel connections.
// The returned Result covers the aggregate transfer.
//
// If progress is non-nil it is updated atomically with the running byte count
// as data is sent, so a caller can render a live progress bar.
func Upload(ctx context.Context, client *http.Client, totalBytes int64, streams int, progress *int64) (Result, error) {
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
	url := endpoint + "/__up"

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

			body := &countingReader{remaining: sz, counter: progress}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
			if err != nil {
				errs <- err
				return
			}
			req.ContentLength = sz
			req.Header.Set("Content-Type", "application/octet-stream")

			resp, err := client.Do(req)
			if err != nil {
				errs <- err
				return
			}
			// Drain and close so the connection can be reused.
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}(sz)
	}
	wg.Wait()
	res := trimmer.result(atomic.LoadInt64(progress), time.Now())

	close(errs)
	if err := transferError(ctx, errs, launched); err != nil {
		return Result{}, fmt.Errorf("upload: %w", err)
	}
	return res, nil
}

// countingReader yields up to `remaining` zero bytes while atomically counting
// how many bytes have been read (i.e. handed to the transport for sending).
// Multiple upload streams share one counter safely.
type countingReader struct {
	remaining int64
	counter   *int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	n := int64(len(p))
	if n > r.remaining {
		n = r.remaining
	}
	// The transport reuses buffers, so zero the slice to avoid leaking
	// arbitrary process memory into the request body.
	clear(p[:n])
	r.remaining -= n
	atomic.AddInt64(r.counter, n)
	return int(n), nil
}
