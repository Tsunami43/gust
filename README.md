```
 ██████╗ ██╗   ██╗███████╗████████╗
██╔════╝ ██║   ██║██╔════╝╚══██╔══╝
██║  ███╗██║   ██║███████╗   ██║
██║   ██║██║   ██║╚════██║   ██║
╚██████╔╝╚██████╔╝███████║   ██║
 ╚═════╝  ╚═════╝ ╚══════╝   ╚═╝
```

**Fast, dependency-free internet speed test & public-IP checker for your terminal.**

`gust` shows your public IP, provider and location, then measures latency,
download and upload speed. On an interactive terminal it opens a **menu-driven
interface** with live animated bars, sparklines and a quality grade; piped or
with flags it behaves as a classic one-shot command with plain-text or JSON
output.

Written in pure Go: **no third-party dependencies**, just the standard library.

---

## Features

- 🌍 Public IP, ASN, provider and location (with automatic provider fallback)
- ⏱️ Latency, jitter and a per-sample sparkline
- ⬇️⬆️ Download & upload throughput over parallel connections
- 🎛️ Interactive menu — pick a test, tweak settings, all with arrow keys
- 📈 `watch` mode with a rolling speed sparkline
- 🏅 Quality grade (A+…E) summarising the result
- 💾 Persistent defaults in `~/.config/gust/config.json`
- 🧾 Clean plain-text and JSON output for scripts

---

## Install

Requires **Go 1.22+**.

```sh
go install github.com/Tsunami43/gust/cmd/gust@latest
```

Or build from source:

```sh
git clone https://github.com/Tsunami43/gust.git
cd gust
make build      # or: go build -o gust ./cmd/gust
```

---

## Usage

Just run it on a terminal to open the interactive menu:

```sh
gust
```

```
   ██████╗ ██╗   ██╗███████╗████████╗
  ██╔════╝ ██║   ██║██╔════╝╚══██╔══╝
  ██║  ███╗██║   ██║███████╗   ██║
  ██║   ██║██║   ██║╚════██║   ██║
  ╚██████╔╝╚██████╔╝███████║   ██║
   ╚═════╝  ╚═════╝ ╚══════╝   ╚═╝
  internet speed test · v1.3.0

  main menu
  ❯ Full test       latency + download + upload
    Download only   measure download speed
    Upload only     measure upload speed
    Ping            latency and jitter
    Public IP       show network info
    Settings        size, streams, pings
    Quit            exit gust

  ↑↓ move   ↵ select   q back
```

The interface uses a calm violet colour theme with a gradient word-mark, a
`❯` pointer and sparklines — colours are 24-bit and shown on any modern
terminal.

A finished full test looks like this:

```
  ╭─ Network ───────────────────────────────────╮
  │ IP         203.0.113.7                       │
  │ Location   Berlin, Germany                   │
  │ Provider   AS3320 Deutsche Telekom AG        │
  │ Server     TXL · HTTP/2.0 · cloudflare       │
  ╰──────────────────────────────────────────────╯

  ✓ Latency   12.34 ms   · jitter 0.70 ms   ▂▃▁▂▄▁
  ✓ Download  ██████████████████████  94.21 Mbps
  ✓ Upload    ██████████████████████  18.44 Mbps

   A  great   ↓ 94.21 Mbps   ↑ 18.44 Mbps   · 12.34 ms
```

### One-shot & scripting

```sh
gust -no-menu          # run a full test without the menu
gust -ip-only          # only show public IP / network info
gust -json             # machine-readable output (for jq & scripts)
gust -no-upload        # skip the upload test
gust -size 100 -streams 8   # heavier test over 8 connections
```

### Watch mode

Repeat the download test on an interval and watch a rolling sparkline:

```sh
gust -watch 5s
```

```
  16:04:11  ↓    91.20 Mbps  ▅▆▄
  16:04:16  ↓    88.70 Mbps  ▄▅▃▂
  16:04:21  ↓    94.80 Mbps  ▆▇▅▃█
```

### Saving defaults

Change settings interactively in **Settings**, or save flags as defaults:

```sh
gust -size 50 -streams 8 -pings 10 -save-config
```

Defaults are stored in `~/.config/gust/config.json` and reused on every run.

---

## Options

| Flag           | Default | Description                                      |
| -------------- | ------- | ------------------------------------------------ |
| `-ip-only`     | `false` | Only show public IP / network info.              |
| `-json`        | `false` | Output results as JSON.                          |
| `-watch`       | `0`     | Repeat the download test on this interval (e.g. `5s`). |
| `-no-download` | `false` | Skip the download test.                          |
| `-no-upload`   | `false` | Skip the upload test.                            |
| `-no-menu`     | `false` | Never open the interactive menu.                 |
| `-no-color`    | `false` | Disable coloured / animated output.              |
| `-size`        | `25`    | Data to transfer per test, in MiB.               |
| `-streams`     | `4`     | Parallel connections per test.                   |
| `-pings`       | `6`     | Number of latency samples.                       |
| `-timeout`     | `60s`   | Overall timeout for a single measurement.        |
| `-save-config` | `false` | Save size/streams/pings as defaults and exit.    |
| `-version`     | `false` | Print version and exit.                          |

Press `Ctrl-C` to cancel at any time.

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). In short: Go 1.22+, standard library
only, `make test` and `make vet` before a PR.

## License

[MIT](LICENSE)
