# Changelog

All notable changes to this project are documented in this file.

## [1.3.1] - 2026-07-14

### Fixed
- Warm up the connection before latency sampling. The first probe used to
  include DNS/TCP/TLS setup, which inflated the reported average and jitter by
  hundreds of milliseconds; samples now reflect the true round-trip time.

## [1.3.0] - 2026-07-14

### Changed
- Reworked the interface around a calm violet colour theme (24-bit colour):
  a gradient GUST word-mark, `❯` pointer, muted rule dividers and accent bars.
- Menu, cards, latency line, summary and watch view restyled to match.

### Added
- Shared `theme` package with the palette and icon set.

## [1.2.0] - 2026-07-14

### Added
- Interactive, arrow-key **menu** shown when gust is run on a terminal with no
  action flags: full test, download/upload/ping only, IP info and settings.
- **Settings** screen and persistent defaults in `~/.config/gust/config.json`
  (also writable via `-save-config`).
- **Watch mode** (`-watch <interval>`): repeat the download test and show a
  rolling speed sparkline.
- Latency **sparkline** of the individual round-trip samples.
- Connection **quality grade** (A+…E) in the results summary.
- `-no-menu` flag to force one-shot behaviour.
- Test suite and CI workflow.

## [1.1.0] - 2026-07-14

### Added
- Rich, animated terminal interface: a banner, a rounded network panel, a
  latency spinner and live download/upload progress bars with colour.
- `-no-color` flag to force plain output; colour is also disabled
  automatically when the output is not a terminal, when `NO_COLOR` is set or
  when `-json` is used.
- Live progress reporting for the download and upload measurements.

## [1.0.1] - 2026-07-14

### Fixed
- Send a browser `User-Agent` so Cloudflare no longer rejects requests with
  `403 Forbidden`.

### Added
- Fall back to `ipwho.is` and `ip-api.com` for IP/location lookup when the
  Cloudflare edge is unavailable, so `-ip-only` keeps working even where
  Cloudflare is blocked.

## [1.0.0] - 2026-07-14

Initial release.

### Added
- Public IP, ASN and approximate location lookup via the Cloudflare edge.
- Latency measurement reporting min / avg / max and jitter.
- Download and upload throughput tests over configurable parallel connections.
- Human-readable text output and machine-readable JSON output (`-json`).
- Flags to tune payload size, stream count, sample count and timeout.
