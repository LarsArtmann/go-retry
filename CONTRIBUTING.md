# Contributing

Thanks for your interest in contributing to `go-retry`!

## Prerequisites

- **Go 1.26** or later (`go.mod` pins `1.26.5`)
- **[golangci-lint](https://golangci-lint.run/)** v2.x (config is committed at
  [`.golangci.yml`](.golangci.yml))

There is no `justfile`, `Makefile`, or `flake.nix` — `go` and `golangci-lint`
are the only tools required.

## Development commands

```bash
go test ./... -race        # tests (always with -race; backoff uses math/rand/v2)
golangci-lint run ./...    # lint (uses the committed .golangci.yml)
go vet ./...               # vet
```

### Coverage

`reports/` is gitignored (`.gitignore`). Regenerate coverage locally:

```bash
go test ./... -race -coverprofile=reports/coverage.out \
  && go tool cover -func=reports/coverage.out
```

Current statement coverage is **100%**. New code should keep it there.

## Lint policy

The committed [`.golangci.yml`](.golangci.yml) enables the default linters plus
`gosec`, `mnd`, and `exhaustruct`. The following in-source `//nolint:` markers
are **deliberate** — do not "fix" them by removing the marker or restructuring
the code:

- `config.go` (`DefaultConfig`) — `//nolint:exhaustruct`: `OnRetry` and
  `OnExhausted` are intentionally omitted (they are optional callbacks).
- `retry.go` (`ComputeDelay`) — `//nolint:mnd,gosec`: the `delay / 2` jitter
  divisor uses a weak RNG on purpose; this is jitter, not security-sensitive
  randomness.

`mnd` and `exhaustruct` are excluded from `*_test.go` (see `.golangci.yml`),
where partial struct literals and bare scalars are legitimate.

## Testing conventions

- External test package (`package retry_test`) — exercise the public API only.
- Every test calls `t.Parallel()`.
- Counters use `sync/atomic` (`atomic.Int32`), not mutexes.
- Keep delays millisecond-scale (see the `fastConfig()` helper) so the suite
  stays fast; the cancellation test is the only deliberate exception.

See [`AGENTS.md`](AGENTS.md) for the deeper architectural context (the
`error-family` dependency, the no-CQRS/no-OTel boundary, control flow of `Do`).

## How to contribute

1. Fork the repository and create a feature branch.
2. Make your change with tests; keep coverage at 100%.
3. Ensure `go test ./... -race`, `golangci-lint run ./...`, and `go vet ./...`
   all pass.
4. Keep the [no-CQRS / no-OTel boundary](doc.go) — features needing CQRS message
   types or OpenTelemetry belong in `go-cqrs-lite/middleware/v4`, not here.
5. Submit a pull request.

## Reporting issues

Please use [GitHub Issues](https://github.com/larsartmann/go-retry/issues) to
report bugs or request features.
