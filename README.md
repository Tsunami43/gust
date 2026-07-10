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
download and upload speed — with an animated, colourful interface that
gracefully degrades to plain text or JSON when piped.

Written in pure Go: **no third-party dependencies**, just the standard library.

---

## Install

Requires **Go 1.22+**.

```sh
go install github.com/Tsunami43/gust/cmd/gust@latest
```

## Usage

```sh
gust
```

```
  ⚡ gust  v1.1.0  ·  internet speed test

  ╭─ Network ───────────────────────────────────╮
  │ IP         203.0.113.7                       │
  │ Location   Berlin, Germany                   │
  │ Provider   AS3320 Deutsche Telekom AG        │
  │ Server     TXL · HTTP/2.0 · cloudflare       │
  ╰──────────────────────────────────────────────╯

  ✔ Latency   12.34 ms   · jitter 0.70 ms
  ✔ Download  ██████████████████████  94.21 Mbps
  ✔ Upload    ██████████████████████  18.44 Mbps
```

```sh
gust -ip-only   # only show public IP / network info
gust -json      # machine-readable output
gust -no-color  # disable colour / animation
```

## Options

| Flag           | Default | Description                          |
| -------------- | ------- | ------------------------------------ |
| `-ip-only`     | `false` | Only show public IP / network info.  |
| `-json`        | `false` | Output results as JSON.              |
| `-no-download` | `false` | Skip the download test.              |
| `-no-upload`   | `false` | Skip the upload test.                |
| `-no-color`    | `false` | Disable coloured / animated output.  |
| `-size`        | `25`    | Data to transfer per test, in MiB.   |
| `-streams`     | `4`     | Parallel connections per test.       |
| `-pings`       | `6`     | Number of latency samples.           |
| `-timeout`     | `60s`   | Overall timeout.                     |
| `-version`     | `false` | Print version and exit.              |

## License

[MIT](LICENSE)
