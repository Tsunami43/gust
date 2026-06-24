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
)

// version is reported by the -version flag.
const version = "1.0.0"

func main() {
	timeout := flag.Duration("timeout", 30*time.Second, "overall timeout for the whole run")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("gust %s\n", version)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	meta, err := netinfo.Lookup(ctx, newClient())
	if err != nil {
		fmt.Fprintln(os.Stderr, "gust: "+err.Error())
		os.Exit(1)
	}

	fmt.Println("Network")
	fmt.Printf("  IP        %s\n", meta.IP)
	fmt.Printf("  Location  %s, %s\n", meta.City, meta.Country)
	fmt.Printf("  Provider  AS%d %s\n", meta.ASN, meta.Org)
}

// newClient builds an HTTP client that identifies itself with a stable
// User-Agent, which some edge services require.
func newClient() *http.Client {
	return &http.Client{Transport: &userAgentTransport{
		agent: "gust/" + version,
		base:  http.DefaultTransport,
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
