package speed

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync/atomic"
	"testing"
)

// withEndpoint points the package at srv for the duration of a test.
func withEndpoint(t *testing.T, url string) {
	t.Helper()
	prev := endpoint
	endpoint = url
	t.Cleanup(func() { endpoint = prev })
}

// speedServer is a tiny stand-in for the Cloudflare speed endpoints: /__down
// returns the requested number of zero bytes, /__up counts what it receives.
func speedServer(received *int64) *httptest.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/__down", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.URL.Query().Get("bytes"))
		if n > 0 {
			_, _ = w.Write(make([]byte, n))
		}
	})
	mux.HandleFunc("/__up", func(w http.ResponseWriter, r *http.Request) {
		n, _ := io.Copy(io.Discard, r.Body)
		if received != nil {
			atomic.AddInt64(received, n)
		}
		w.WriteHeader(http.StatusOK)
	})
	return httptest.NewServer(mux)
}

func TestSplitSizes(t *testing.T) {
	cases := []struct {
		total int64
		n     int
		want  []int64
	}{
		{100, 4, []int64{25, 25, 25, 25}},
		{10, 4, []int64{3, 3, 2, 2}},
		{7, 1, []int64{7}},
		{0, 3, []int64{0, 0, 0}},
		{5, 0, []int64{5}}, // n normalised to 1
	}
	for _, c := range cases {
		got := splitSizes(c.total, c.n)
		var sum int64
		for _, v := range got {
			sum += v
		}
		if sum != c.total {
			t.Errorf("splitSizes(%d,%d) sums to %d, want %d", c.total, c.n, sum, c.total)
		}
		if len(got) != len(c.want) {
			t.Fatalf("splitSizes(%d,%d) = %v, want %v", c.total, c.n, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitSizes(%d,%d) = %v, want %v", c.total, c.n, got, c.want)
				break
			}
		}
	}
}

func TestDownloadCountsFullTransfer(t *testing.T) {
	srv := speedServer(nil)
	defer srv.Close()
	withEndpoint(t, srv.URL)

	const total = 64 * 1024
	var progress int64
	res, err := Download(context.Background(), srv.Client(), total, 4, &progress)
	if err != nil {
		t.Fatal(err)
	}
	if res.Bytes != total {
		t.Errorf("downloaded %d bytes, want %d", res.Bytes, total)
	}
	if atomic.LoadInt64(&progress) != total {
		t.Errorf("progress = %d, want %d", progress, total)
	}
	if res.BitsPerSecond() <= 0 {
		t.Error("BitsPerSecond should be positive")
	}
}

func TestUploadCountsFullTransfer(t *testing.T) {
	var received int64
	srv := speedServer(&received)
	defer srv.Close()
	withEndpoint(t, srv.URL)

	const total = 48 * 1024
	var progress int64
	res, err := Upload(context.Background(), srv.Client(), total, 3, &progress)
	if err != nil {
		t.Fatal(err)
	}
	if res.Bytes != total {
		t.Errorf("uploaded %d bytes, want %d", res.Bytes, total)
	}
	if got := atomic.LoadInt64(&received); got != total {
		t.Errorf("server received %d bytes, want %d", got, total)
	}
}

func TestDownloadUnreachable(t *testing.T) {
	srv := speedServer(nil)
	srv.Close() // closed immediately: every stream must fail
	withEndpoint(t, srv.URL)

	_, err := Download(context.Background(), srv.Client(), 4096, 4, nil)
	if err == nil {
		t.Fatal("expected an error when the endpoint is unreachable")
	}
}

func TestDownloadCancelled(t *testing.T) {
	srv := speedServer(nil)
	defer srv.Close()
	withEndpoint(t, srv.URL)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Download(ctx, srv.Client(), 4096, 4, nil)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestMeasureLatency(t *testing.T) {
	srv := speedServer(nil)
	defer srv.Close()
	withEndpoint(t, srv.URL)

	lat, err := MeasureLatency(context.Background(), srv.Client(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if lat.Samples != 5 || len(lat.Series) != 5 {
		t.Errorf("samples/series = %d/%d, want 5/5", lat.Samples, len(lat.Series))
	}
	if lat.Avg < 0 || lat.Min < 0 || lat.Max < 0 {
		t.Errorf("negative latency: %+v", lat)
	}
}

func TestTransferError(t *testing.T) {
	// Every stream failed -> surface the first error.
	ch := make(chan error, 2)
	ch <- errors.New("a")
	ch <- errors.New("b")
	close(ch)
	if err := transferError(context.Background(), ch, 2); err == nil {
		t.Error("all streams failed: want an error")
	}

	// Only some streams failed -> tolerate, the measurement is still useful.
	ch = make(chan error, 2)
	ch <- errors.New("a")
	close(ch)
	if err := transferError(context.Background(), ch, 2); err != nil {
		t.Errorf("partial failure should be tolerated, got %v", err)
	}

	// Context cancelled wins even without stream errors.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ch = make(chan error, 1)
	close(ch)
	if err := transferError(ctx, ch, 2); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
