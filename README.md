# gust

`gust` is a small, dependency-free CLI written in Go that shows your public IP
and measures internet connection speed (latency, download and upload) using the
public Cloudflare speed test endpoints.

## Install

Requires Go 1.22+.

```sh
go install github.com/Tsunami43/gust/cmd/gust@latest
```

## Usage

```sh
gust            # full test
gust -ip-only   # only show public IP / network info
gust -json      # machine-readable output
```

## Options

| Flag           | Default | Description                          |
| -------------- | ------- | ------------------------------------ |
| `-ip-only`     | `false` | Only show public IP / network info.  |
| `-json`        | `false` | Output results as JSON.              |
| `-no-download` | `false` | Skip the download test.              |
| `-no-upload`   | `false` | Skip the upload test.                |
| `-size`        | `25`    | Data to transfer per test, in MiB.   |
| `-streams`     | `4`     | Parallel connections per test.       |
| `-pings`       | `6`     | Number of latency samples.           |
| `-timeout`     | `60s`   | Overall timeout.                     |
| `-version`     | `false` | Print version and exit.              |

## License

MIT
