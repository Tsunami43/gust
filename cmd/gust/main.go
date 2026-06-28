// Command gust is a small, dependency-free CLI that reports your public IP
// address and measures internet connection speed (latency, download and
// upload) using the public Cloudflare speed test endpoints.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Tsunami43/gust/internal/netinfo"
	"github.com/Tsunami43/gust/internal/speed"
)

// version is reported by the -version flag.
const version = "1.0.0"

func main() {
	ipOnly := flag.Bool("ip-only", false, "only show public IP / network info and exit")
	sizeMB := flag.Int("size", 25, "amount of data to transfer per test, in MiB")
	streams := flag.Int("streams", 4, "number of parallel connections per test")
	pings := flag.Int("pings", 6, "number of latency samples")
	timeout := flag.Duration("timeout", 60*time.Second, "overall timeout for the whole run")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gust %s\n", version)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	client := newClient(*streams)
	if err := runTests(ctx, client, *ipOnly, int64(*sizeMB)<<20, *streams, *pings); err != nil {
		fmt.Fprintln(os.Stderr, "gust: "+err.Error())
		os.Exit(1)
	}
}

// runTests resolves network info and, unless ipOnly is set, measures latency,
// download and upload throughput, printing each result as it completes.
func runTests(ctx context.Context, client *http.Client, ipOnly bool, totalBytes int64, streams, pings int) error {
	meta, err := netinfo.Lookup(ctx, client)
	if err != nil {
		return err
	}
	fmt.Println("Network")
	fmt.Printf("  IP        %s\n", meta.IP)
	fmt.Printf("  Location  %s, %s\n", meta.City, meta.Country)
	fmt.Printf("  Provider  AS%d %s\n\n", meta.ASN, meta.Org)
	if ipOnly {
		return nil
	}

	lat, err := speed.MeasureLatency(ctx, client, pings)
	if err != nil {
		return err
	}
	fmt.Printf("Latency   %.2f ms\n", float64(lat.Avg.Microseconds())/1000)

	dl, err := speed.Download(ctx, client, totalBytes, streams)
	if err != nil {
		return err
	}
	fmt.Printf("Download  %.2f Mbps\n", dl.BitsPerSecond()/1e6)

	ul, err := speed.Upload(ctx, client, totalBytes, streams)
	if err != nil {
		return err
	}
	fmt.Printf("Upload    %.2f Mbps\n", ul.BitsPerSecond()/1e6)
	return nil
}

// newClient builds an HTTP client tuned for parallel speed tests and sets a
// stable User-Agent, which some edge services require.
func newClient(streams int) *http.Client {
	if streams < 1 {
		streams = 1
	}
	return &http.Client{Transport: &userAgentTransport{
		agent: "gust/" + version,
		base: &http.Transport{
			MaxIdleConns:        streams * 2,
			MaxIdleConnsPerHost: streams * 2,
			ForceAttemptHTTP2:   true,
		},
	}}
}

// userAgentTransport sets a stable User-Agent on every outgoing request.
type userAgentTransport struct {
	agent string
	base  http.RoundTripper
}

func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", t.agent)
	}
	return t.base.RoundTrip(req)
}
