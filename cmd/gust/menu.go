package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Tsunami43/gust/internal/config"
	"github.com/Tsunami43/gust/internal/netinfo"
	"github.com/Tsunami43/gust/internal/speed"
	"github.com/Tsunami43/gust/internal/tui"
)

// clearScreen erases the terminal and homes the cursor.
const clearScreen = "\033[2J\033[H"

// mainMenuItems are the entries of the top-level interactive menu.
var mainMenuItems = []tui.Item{
	{Label: "Full test", Desc: "latency + download + upload"},
	{Label: "Download only", Desc: "measure download speed"},
	{Label: "Upload only", Desc: "measure upload speed"},
	{Label: "Ping", Desc: "latency and jitter"},
	{Label: "Public IP", Desc: "show network info"},
	{Label: "Settings", Desc: "size, streams, pings"},
	{Label: "Quit", Desc: "exit gust"},
}

// runMenu drives the interactive, menu-based interface.
func (a *app) runMenu(ctx context.Context) error {
	menu := tui.NewMenu("main menu", mainMenuItems, a.in, a.out)
	for {
		fmt.Fprint(a.out, clearScreen)
		a.r.Logo("internet speed test · v" + version)
		idx, ok := menu.Run()
		if !ok {
			return nil
		}

		fmt.Fprint(a.out, clearScreen)
		var err error
		switch idx {
		case 0:
			_, err = a.execute(ctx, plan{latency: true, download: true, upload: true})
		case 1:
			_, err = a.execute(ctx, plan{download: true})
		case 2:
			_, err = a.execute(ctx, plan{upload: true})
		case 3:
			_, err = a.execute(ctx, plan{latency: true})
		case 4:
			_, err = a.execute(ctx, plan{})
		case 5:
			a.runSettings()
			continue
		case 6:
			return nil
		}
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			fmt.Fprintln(a.out, "  error: "+err.Error())
		}
		fmt.Fprintln(a.out, "\n  press any key to return to the menu…")
		tui.WaitKey(a.in)
	}
}

// runSettings lets the user cycle the transfer size, stream count and ping
// count and optionally persist them as defaults.
func (a *app) runSettings() {
	for {
		items := []tui.Item{
			{Label: fmt.Sprintf("Size: %d MiB", a.opt.sizeMB), Desc: "data transferred per test"},
			{Label: fmt.Sprintf("Streams: %d", a.opt.streams), Desc: "parallel connections"},
			{Label: fmt.Sprintf("Pings: %d", a.opt.pings), Desc: "latency samples"},
			{Label: "Save as defaults", Desc: "write ~/.config/gust/config.json"},
			{Label: "Back", Desc: "return to the main menu"},
		}
		fmt.Fprint(a.out, clearScreen)
		idx, ok := tui.NewMenu("settings", items, a.in, a.out).Run()
		if !ok {
			return
		}
		switch idx {
		case 0:
			a.opt.sizeMB = cycleInt(a.opt.sizeMB, []int{5, 10, 25, 50, 100})
		case 1:
			a.opt.streams = cycleInt(a.opt.streams, []int{1, 2, 4, 8, 16})
		case 2:
			a.opt.pings = cycleInt(a.opt.pings, []int{3, 6, 10, 20})
		case 3:
			_ = config.Save(config.Config{
				SizeMB: a.opt.sizeMB, Streams: a.opt.streams, Pings: a.opt.pings, NoColor: a.opt.noColor,
			})
			return
		case 4:
			return
		}
	}
}

// runWatch repeatedly measures download speed on the configured interval,
// printing a rolling sparkline until the context is cancelled (Ctrl-C).
func (a *app) runWatch(ctx context.Context) error {
	const historyLen = 40
	var history []float64

	for {
		if ctx.Err() != nil {
			return nil
		}
		var meta netinfo.Meta
		if len(history) == 0 {
			// Resolve and show network info once, on the first iteration.
			if m, err := netinfo.Lookup(ctx, a.client); err == nil {
				meta = m
				a.r.NetworkCard(meta)
			}
		}

		dl, err := speed.Download(ctx, a.client, a.total(), a.opt.streams, nil)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		mbps := dl.BitsPerSecond() / 1e6
		history = append(history, mbps)
		if len(history) > historyLen {
			history = history[len(history)-historyLen:]
		}
		a.r.WatchLine(time.Now().Format("15:04:05"), mbps, history)

		select {
		case <-time.After(a.opt.watch):
		case <-ctx.Done():
			return nil
		}
	}
}

// cycleInt returns the value after cur in opts, wrapping around. If cur is not
// present, the first option is returned.
func cycleInt(cur int, opts []int) int {
	for i, v := range opts {
		if v == cur {
			return opts[(i+1)%len(opts)]
		}
	}
	return opts[0]
}
