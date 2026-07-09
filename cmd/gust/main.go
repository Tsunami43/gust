// Command gust is a small, dependency-free CLI that reports your public IP
// address and measures internet connection speed (latency, download and
// upload) using the public Cloudflare speed test endpoints.
//
// On an interactive terminal it renders an animated, colourful interface;
// when piped or asked for JSON it emits clean, machine-readable output.
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
const version = "1.1.0"

// config holds the parsed command-line options for a single run.
type config struct {
	jsonOut    bool
	ipOnly     bool
	noDownload bool
	noUpload   bool
	noColor    bool
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

	// Colour and animation are used only on an interactive terminal, unless
	// explicitly disabled or when machine-readable output is requested.
	colorOK := !cfg.noColor && !cfg.jsonOut && os.Getenv("NO_COLOR") == ""
	r := ui.NewRenderer(stdout, colorOK)
	fancy := r.Fancy()

	var report ui.Report

	if fancy {
		r.Header(version)
	}

	// Resolve public network information (with provider fallback) first.
	var meta netinfo.Meta
	if err := r.RunSpinner("resolving network", func() error {
		var e error
		meta, e = netinfo.Lookup(ctx, client)
		return e
	}); err != nil {
		return err
	}
	report.Network = &meta
	r.NetworkCard(meta)

	if !cfg.ipOnly {
		var lat speed.Latency
		if err := r.RunSpinner("measuring latency", func() error {
			var e error
			lat, e = speed.MeasureLatency(ctx, client, cfg.pings)
			return e
		}); err != nil {
			return err
		}
		report.Latency = &lat
		r.LatencyLine(lat)

		if !cfg.noDownload {
			dl, err := r.RunBar("Download", totalBytes, func(p *int64) (speed.Result, error) {
				return speed.Download(ctx, client, totalBytes, cfg.streams, p)
			})
			if err != nil {
				return err
			}
			report.Download = &dl
		}

		if !cfg.noUpload {
			ul, err := r.RunBar("Upload", totalBytes, func(p *int64) (speed.Result, error) {
				return speed.Upload(ctx, client, totalBytes, cfg.streams, p)
			})
			if err != nil {
				return err
			}
			report.Upload = &ul
		}
	}

	r.Footer()

	// Fancy runs have already drawn their results live; only the plain and
	// JSON modes need to print a report at the end.
	switch {
	case cfg.jsonOut:
		return report.WriteJSON(stdout)
	case !fancy:
		report.WriteText(stdout)
	}
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
	fs.BoolVar(&cfg.noColor, "no-color", false, "disable coloured and animated output")
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
