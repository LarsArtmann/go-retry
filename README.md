# go-retry

> A dependency-light Go retry loop with exponential backoff and jitter, built on
> [`go-error-family`](https://github.com/larsartmann/go-error-family).

## Why

Most retry libraries either pull in a heavy framework or reinvent error
classification. `go-retry` is the small core: a retry loop, exponential backoff
with additive jitter, and a pluggable retryable predicate — nothing more. It
intentionally has **no CQRS message types and no OpenTelemetry SDK dependency**
(`doc.go`), so CLIs, batch jobs, and simple services can retry an operation
without importing an entire platform. For the CQRS-wrapped variant (message
adapter, OTel spans, dead-letter entries) use
[`go-cqrs-lite/middleware/v4`](https://github.com/larsartmann/go-cqrs-lite).

## Installation

Requires **Go 1.26** or later.

```bash
go get github.com/larsartmann/go-retry@latest
```

## Quick start

```go
package main

import (
	"context"
	"fmt"

	errorfamily "github.com/larsartmann/go-error-family"
	retry "github.com/larsartmann/go-retry"
)

func main() {
	var attempt int
	err := retry.Do(context.Background(), retry.DefaultConfig(),
		func(ctx context.Context, n int) error {
			attempt = n
			if n < 3 {
				// a Transient error is retryable by default
				return errorfamily.NewTransient("db.timeout", "database unavailable")
			}
			return nil
		},
	)
	if err != nil {
		fmt.Println("failed:", err)
		return
	}
	fmt.Printf("succeeded on attempt %d\n", attempt)
}
```

```text
succeeded on attempt 3
```

## Configuration

`retry.DefaultConfig()` returns sensible defaults; override only what you need.

| Field          | Default                   | Description                                                                  |
| -------------- | ------------------------- | ---------------------------------------------------------------------------- |
| `MaxAttempts`  | `3`                       | Total attempts **including the first call** (not retries on top).            |
| `InitialDelay` | `100ms`                   | Delay before the second attempt.                                             |
| `MaxDelay`     | `5s`                      | Cap on the backoff delay.                                                    |
| `Multiplier`   | `2.0`                     | Exponential backoff factor. Delay for attempt _n_ is `Initial * Mult^(n-1)`. |
| `IsRetryable`  | `errorfamily.IsRetryable` | Decides whether an error triggers a retry. Set `nil` for the default.        |
| `OnRetry`      | `nil`                     | Called after each failed attempt, before sleeping.                           |
| `OnExhausted`  | `nil`                     | Called once after all attempts have failed.                                  |

Delay for attempt _n_ is `InitialDelay * Multiplier^(n-1)`, capped at
`MaxDelay`, plus random jitter of up to 50% of the capped delay. Use the
exported `Backoff(config, attempt)` or `ComputeDelay(...)` to preview the planned
delay without running the loop.

### Custom retryable predicate and observability hooks

```go
cfg := retry.DefaultConfig()
cfg.MaxAttempts = 5
cfg.IsRetryable = func(err error) bool {
	// only retry your own sentinel, not every Transient error
	return errors.Is(err, errServiceOverloaded)
}
cfg.OnRetry = func(attempt int, delay time.Duration, err error) {
	log.Printf("attempt %d failed (%v); retrying in %s", attempt, err, delay)
}
cfg.OnExhausted = func(attempts int, err error) {
	log.Printf("gave up after %d attempts: %v", attempts, err)
}
```

## Errors

`Do` never returns a bare `context.Canceled`. On exhaustion or cancellation
during backoff it returns an `error-family` `Infrastructure` error wrapping a
stable sentinel, with the last operation error chained as its cause:

| Situation                       | Sentinel             | Code              |
| ------------------------------- | -------------------- | ----------------- |
| All attempts failed             | `retry.ErrExhausted` | `retry.exhausted` |
| Context canceled during backoff | `retry.ErrCanceled`  | `retry.canceled`  |

Detect them with `errors.Is`:

```go
if errors.Is(err, retry.ErrExhausted) { /* gave up */ }
```

Invalid `Config` returns a `Rejection`-family error (`retry.invalid_*` codes)
before the first attempt runs. See [`FEATURES.md`](FEATURES.md) for the full
inventory.

## Development

```bash
go test ./... -race        # tests (always with -race)
golangci-lint run ./...    # lint
go vet ./...               # vet
go test ./... -race -coverprofile=reports/coverage.out \
  && go tool cover -func=reports/coverage.out   # coverage (currently 100%)
```

There is no `justfile`, `Makefile`, or `flake.nix` — `go` and `golangci-lint`
are the only tools. See [`CONTRIBUTING.md`](CONTRIBUTING.md).

## License

Proprietary — all rights reserved. See [`LICENSE`](LICENSE).
