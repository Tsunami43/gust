# Changelog

All notable changes to this project are documented in this file.

## [1.1.0] - 2026-07-10

### Added
- Rich, animated terminal interface: a banner, a rounded network panel, a
  latency spinner and live download/upload progress bars with colour.
- `-no-color` flag to force plain output; colour is also disabled
  automatically when the output is not a terminal, when `NO_COLOR` is set or
  when `-json` is used.

## [1.0.1] - 2026-07-04

### Fixed
- Send a browser `User-Agent` so Cloudflare no longer rejects requests with
  `403 Forbidden`.

### Added
- Fall back to `ipwho.is` and `ip-api.com` for IP/location lookup when the
  Cloudflare edge is unavailable.

## [1.0.0] - 2026-07-01

Initial release.

### Added
- Public IP, ASN and approximate location lookup via the Cloudflare edge.
- Latency measurement reporting min / avg / max and jitter.
- Download and upload throughput tests over configurable parallel connections.
- Human-readable text output and machine-readable JSON output (`-json`).
- Flags to tune payload size, stream count, sample count and timeout.
