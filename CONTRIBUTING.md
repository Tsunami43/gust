# Contributing to gust

Thanks for your interest in improving **gust**!

## Getting started

```sh
git clone https://github.com/Tsunami43/gust.git
cd gust
make build
make test
```

gust targets **Go 1.22+** and deliberately uses **only the standard library** —
please do not add third-party module dependencies.

## Before opening a pull request

- `make vet` — no reported issues
- `make test` — all tests pass
- `gofmt` / `goimports` clean (`make lint` if you have `golangci-lint`)
- Keep commits focused and write clear, imperative commit messages
  (e.g. `feat(ui): add sparkline to latency line`).

## Project layout

```
cmd/gust/           CLI entry point, flags, menu wiring
internal/config/    persistent user defaults
internal/netinfo/   public IP / ASN / location lookup
internal/speed/     latency, download and upload measurements
internal/tui/       interactive menu and key handling
internal/ui/        text, JSON and animated rendering
```

## Reporting bugs

Open an issue with the command you ran, the output, your OS and `gust -version`.
