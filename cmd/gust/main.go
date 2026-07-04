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
	"os/signal"
	"time"

	"github.com/Tsunami43/gust/internal/netinfo"
	"github.com/Tsunami43/gust/internal/speed"
	"github.com/Tsunami43/gust/internal/ui"
)

// version is reported by the -version flag.
const version = "1.0.1"

// config holds the parsed command-line options for a single run.
type config struct {
	jsonOut    bool
	ipOnly     bool
	noDownload bool
	noUpload   bool
	sizeMB     int
	streams    int
	pings      int
	timeout    time.Duration
}

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gust: "+err.Error())
		os.Exit(1)
	}
}

// run parses flags, performs the requested measurements and writes the report.
// It is separated from main so behaviour can be exercised in tests.
func run(args []string, stdout, stderr *os.File) error {
	cfg, showVersion, err := parseFlags(args, stderr)
	if err != nil {
		return err
	}
	if showVersion {
		fmt.Fprintf(stdout, "gust %s\n", version)
		return nil
	}

	// Cancel the whole run on the overall timeout or on Ctrl-C.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.timeout)
	defer cancel()
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	client := newClient(cfg.streams)
	totalBytes := int64(cfg.sizeMB) << 20 // MiB -> bytes

	var report ui.Report

	// Always resolve public network information first.
	step(stderr, cfg.jsonOut, "Looking up network information")
	meta, err := netinfo.Lookup(ctx, client)
	if err != nil {
		return err
	}
	report.Network = &meta

	if !cfg.ipOnly {
		step(stderr, cfg.jsonOut, "Measuring latency")
		lat, err := speed.MeasureLatency(ctx, client, cfg.pings)
		if err != nil {
			return err
		}
		report.Latency = &lat

		if !cfg.noDownload {
			step(stderr, cfg.jsonOut, "Testing download speed")
			dl, err := speed.Download(ctx, client, totalBytes, cfg.streams)
			if err != nil {
				return err
			}
			report.Download = &dl
		}

		if !cfg.noUpload {
			step(stderr, cfg.jsonOut, "Testing upload speed")
			ul, err := speed.Upload(ctx, client, totalBytes, cfg.streams)
			if err != nil {
				return err
			}
			report.Upload = &ul
		}
	}

	if cfg.jsonOut {
		return report.WriteJSON(stdout)
	}
	report.WriteText(stdout)
	return nil
}

// parseFlags interprets the command-line arguments into a config value.
func parseFlags(args []string, stderr *os.File) (cfg config, showVersion bool, err error) {
	fs := flag.NewFlagSet("gust", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "gust %s — internet speed & public IP checker\n\n", version)
		fmt.Fprintln(stderr, "Usage: gust [options]")
		fmt.Fprintln(stderr, "\nOptions:")
		fs.PrintDefaults()
	}

	fs.BoolVar(&cfg.jsonOut, "json", false, "output results as JSON")
	fs.BoolVar(&cfg.ipOnly, "ip-only", false, "only show public IP / network info and exit")
	fs.BoolVar(&cfg.noDownload, "no-download", false, "skip the download test")
	fs.BoolVar(&cfg.noUpload, "no-upload", false, "skip the upload test")
	fs.IntVar(&cfg.sizeMB, "size", 25, "amount of data to transfer per test, in MiB")
	fs.IntVar(&cfg.streams, "streams", 4, "number of parallel connections per test")
	fs.IntVar(&cfg.pings, "pings", 6, "number of latency samples")
	fs.DurationVar(&cfg.timeout, "timeout", 60*time.Second, "overall timeout for the whole run")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")

	if err = fs.Parse(args); err != nil {
		return config{}, false, err
	}
	return cfg, showVersion, nil
}

// newClient builds an HTTP client tuned for parallel speed tests. Per-request
// deadlines come from the context, so no client-level timeout is set.
func newClient(streams int) *http.Client {
	if streams < 1 {
		streams = 1
	}
	return &http.Client{
		Transport: &userAgentTransport{
			agent: browserUserAgent,
			base: &http.Transport{
				MaxIdleConns:        streams * 2,
				MaxIdleConnsPerHost: streams * 2,
				ForceAttemptHTTP2:   true,
			},
		},
	}
}

// browserUserAgent is a realistic browser identity. Cloudflare's bot
// protection replies 403 to the default Go user agent on the speed test
// endpoints, so we present ourselves as a common browser instead.
const browserUserAgent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"

// userAgentTransport sets a stable User-Agent on every request. Some edge
// services reject the default Go user agent, so we identify ourselves.
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

// step prints a progress line to stderr unless JSON output is requested, in
// which case stdout must stay a clean, machine-readable stream.
func step(stderr *os.File, jsonOut bool, msg string) {
	if jsonOut {
		return
	}
	fmt.Fprintf(stderr, "→ %s...\n", msg)
}
