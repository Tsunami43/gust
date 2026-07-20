// Command gust is a small, dependency-free CLI that reports your public IP
// address and measures internet connection speed (latency, download and
// upload) using the public Cloudflare speed test endpoints.
//
// Run on an interactive terminal it opens a menu-driven interface; with flags
// or when piped it behaves as a classic one-shot command with plain-text or
// JSON output.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Tsunami43/gust/internal/config"
	"github.com/Tsunami43/gust/internal/netinfo"
	"github.com/Tsunami43/gust/internal/speed"
	"github.com/Tsunami43/gust/internal/ui"
)

// version is reported by the -version flag.
const version = "1.3.1"

// options holds the parsed command-line options for a single run.
type options struct {
	jsonOut    bool
	ipOnly     bool
	noDownload bool
	noUpload   bool
	noColor    bool
	noMenu     bool
	saveConfig bool
	sizeMB     int
	streams    int
	pings      int
	watch      time.Duration
	timeout    time.Duration
}

// plan selects which measurements a single execution performs.
type plan struct {
	latency  bool
	download bool
	upload   bool
}

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gust: "+err.Error())
		os.Exit(1)
	}
}

// run wires configuration, flags and I/O together and dispatches to the menu,
// watch or one-shot mode. It is separated from main so it can be tested.
func run(args []string, stdin, stdout, stderr *os.File) error {
	base, _ := config.Load() // fall back to defaults on error
	opt, showVersion, err := parseFlags(args, stderr, base)
	if err != nil {
		return err
	}
	if showVersion {
		fmt.Fprintf(stdout, "gust %s\n", version)
		return nil
	}
	if opt.saveConfig {
		return config.Save(config.Config{
			SizeMB: opt.sizeMB, Streams: opt.streams, Pings: opt.pings, NoColor: opt.noColor,
		})
	}

	// The root context is cancelled only by Ctrl-C: it spans the whole session,
	// which in menu mode outlives any single measurement. The -timeout budget is
	// applied per measurement instead, in execute and runWatch.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	colorOK := !opt.noColor && !opt.jsonOut && os.Getenv("NO_COLOR") == ""
	a := &app{
		opt:    opt,
		client: newClient(opt.streams),
		in:     stdin,
		out:    stdout,
		r:      ui.NewRenderer(stdout, colorOK),
	}

	switch {
	case opt.watch > 0:
		return a.runWatch(ctx)
	case a.wantsMenu(stdin, stdout):
		return a.runMenu(ctx)
	default:
		return a.runOnce(ctx)
	}
}

// app carries the shared state for one invocation.
type app struct {
	opt    options
	client *http.Client
	in     *os.File
	out    *os.File
	r      *ui.Renderer
}

// total returns the per-test transfer size in bytes.
func (a *app) total() int64 { return int64(a.opt.sizeMB) << 20 }

// wantsMenu reports whether the interactive menu should be shown: a bare,
// colourful, fully-interactive terminal session with no action flags.
func (a *app) wantsMenu(stdin, stdout *os.File) bool {
	if a.opt.noMenu || a.opt.jsonOut || !a.r.Fancy() || !ui.IsTerminal(stdin) {
		return false
	}
	return !a.opt.ipOnly && !a.opt.noDownload && !a.opt.noUpload
}

// runOnce performs a single classic measurement based on the flags.
func (a *app) runOnce(ctx context.Context) error {
	a.r.Header(version)
	p := plan{
		latency:  !a.opt.ipOnly,
		download: !a.opt.ipOnly && !a.opt.noDownload,
		upload:   !a.opt.ipOnly && !a.opt.noUpload,
	}
	rep, err := a.execute(ctx, p)
	if err != nil {
		return err
	}
	a.r.Footer()

	switch {
	case a.opt.jsonOut:
		return rep.WriteJSON(a.out)
	case !a.r.Fancy():
		rep.WriteText(a.out)
	}
	return nil
}

// execute resolves network info and runs the measurements selected by p,
// rendering each result as it completes.
func (a *app) execute(ctx context.Context, p plan) (ui.Report, error) {
	var rep ui.Report

	ctx, cancel := context.WithTimeout(ctx, a.opt.timeout)
	defer cancel()

	var meta netinfo.Meta
	if err := a.r.RunSpinner("resolving network", func() error {
		var e error
		meta, e = netinfo.Lookup(ctx, a.client)
		return e
	}); err != nil {
		return rep, err
	}
	rep.Network = &meta
	a.r.NetworkCard(meta)

	if p.latency {
		var lat speed.Latency
		if err := a.r.RunSpinner("measuring latency", func() error {
			var e error
			lat, e = speed.MeasureLatency(ctx, a.client, a.opt.pings)
			return e
		}); err != nil {
			return rep, err
		}
		rep.Latency = &lat
		a.r.LatencyLine(lat)
	}
	if p.download {
		dl, err := a.r.RunBar("Download", a.total(), func(pr *int64) (speed.Result, error) {
			return speed.Download(ctx, a.client, a.total(), a.opt.streams, pr)
		})
		if err != nil {
			return rep, err
		}
		rep.Download = &dl
	}
	if p.upload {
		ul, err := a.r.RunBar("Upload", a.total(), func(pr *int64) (speed.Result, error) {
			return speed.Upload(ctx, a.client, a.total(), a.opt.streams, pr)
		})
		if err != nil {
			return rep, err
		}
		rep.Upload = &ul
	}

	a.r.SummaryCard(rep)
	return rep, nil
}

// parseFlags interprets the command-line arguments into options, seeding the
// defaults from the persisted configuration.
func parseFlags(args []string, stderr *os.File, base config.Config) (opt options, showVersion bool, err error) {
	fs := flag.NewFlagSet("gust", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		fmt.Fprintf(stderr, "gust %s — internet speed & public IP checker\n\n", version)
		fmt.Fprintln(stderr, "Usage: gust [options]")
		fmt.Fprintln(stderr, "\nOptions:")
		fs.PrintDefaults()
	}

	fs.BoolVar(&opt.jsonOut, "json", false, "output results as JSON")
	fs.BoolVar(&opt.ipOnly, "ip-only", false, "only show public IP / network info and exit")
	fs.BoolVar(&opt.noDownload, "no-download", false, "skip the download test")
	fs.BoolVar(&opt.noUpload, "no-upload", false, "skip the upload test")
	fs.BoolVar(&opt.noColor, "no-color", base.NoColor, "disable coloured and animated output")
	fs.BoolVar(&opt.noMenu, "no-menu", false, "never open the interactive menu")
	fs.BoolVar(&opt.saveConfig, "save-config", false, "save the given size/streams/pings as defaults and exit")
	fs.IntVar(&opt.sizeMB, "size", base.SizeMB, "amount of data to transfer per test, in MiB")
	fs.IntVar(&opt.streams, "streams", base.Streams, "number of parallel connections per test")
	fs.IntVar(&opt.pings, "pings", base.Pings, "number of latency samples")
	fs.DurationVar(&opt.watch, "watch", 0, "repeat the download test on this interval (e.g. 5s); 0 disables")
	fs.DurationVar(&opt.timeout, "timeout", 60*time.Second, "overall timeout for a single measurement")
	fs.BoolVar(&showVersion, "version", false, "print version and exit")

	if err = fs.Parse(args); err != nil {
		return options{}, false, err
	}
	return opt, showVersion, nil
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

// userAgentTransport sets browserUserAgent on every request that lacks one.
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
