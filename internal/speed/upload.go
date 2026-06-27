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
func Upload(ctx context.Context, client *http.Client, totalBytes int64, streams int) (Result, error) {
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
	url := endpoint + "/__up"

	start := time.Now()
	for i := 0; i < streams; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			body := &countingReader{remaining: perStream, counter: &transferred}
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
			if err != nil {
				errs <- err
				return
			}
			req.ContentLength = perStream
			req.Header.Set("Content-Type", "application/octet-stream")

			resp, err := client.Do(req)
			if err != nil {
				errs <- err
				return
			}
			// Drain and close so the connection can be reused.
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	close(errs)
	if err := <-errs; err != nil {
		return Result{}, fmt.Errorf("upload: %w", err)
	}
	return Result{Bytes: atomic.LoadInt64(&transferred), Elapsed: elapsed}, nil
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
