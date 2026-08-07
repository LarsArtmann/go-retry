# AGENTS.md

Concise, enduring context for every AI session working in `go-retry`.

## What This Is

A single-package Go **library** (not an application) providing a dependency-light
retry loop with exponential backoff and jitter. Module:
`github.com/larsartmann/go-retry`, package `retry`, Go 1.26.5.

This is the **core** retry primitive — intentionally free of CQRS message types
and OpenTelemetry. The CQRS-wrapped variant (`MessageAdapter`, OTel spans,
dead-letter entries carrying `StreamID`) lives in
`github.com/larsartmann/go-cqrs-lite/middleware/v4`. Do **not** add CQRS or OTel
imports here; consumers who need only retry (CLIs, batch jobs, simple services)
import this package to avoid pulling in those deps. See `doc.go`.

## Commands

No `flake.nix`, `Makefile`, or `justfile` exists in this repo — `go` and
`golangci-lint` are the only tools. Use raw Go commands:

```bash
go test ./... -race             # tests (always with -race; backoff uses math/rand/v2)
go test ./... -race -count=10   # flake-prone jitter/backoff tests
golangci-lint run ./...         # lint (committed .golangci.yml enables gosec/mnd/exhaustruct + defaults)
go vet ./...
```

`go test` is the only verification gate. There is no build step beyond `go build`
(the package is consumed as a library).

## Architecture & Data Flow

Flat single-package layout — no internal subpackages:

| File            | Responsibility                                                                         |
| --------------- | -------------------------------------------------------------------------------------- |
| `retry.go`      | `Do` (loop), `Backoff`, `ComputeDelay`, sentinel errors `ErrExhausted` / `ErrCanceled` |
| `config.go`     | `Config` struct, `DefaultConfig()`, `Validate()`                                       |
| `doc.go`        | Package doc stating the no-CQRS/no-OTel boundary                                       |
| `retry_test.go` | External test package (`retry_test`)                                                   |

**Control flow of `Do`**: validate config → loop `attempt` from 1 to
`MaxAttempts` → call `fn(ctx, attempt)` → on `nil` return immediately → if not
retryable, return immediately → else compute delay, fire `OnRetry`, sleep in a
`select` on `timer.C` vs `ctx.Done()` → on exhaustion call `OnExhausted` and
return `ErrExhausted` wrapping the last error via `.WithCause()`.

## The error-family Dependency

The sole external dependency is
`github.com/larsartmann/go-error-family` (`errorfamily` import alias). Errors
carry a **family** classification and a string **code**. This package uses:

- `errorfamily.NewInfrastructure(code, msg)` — for `ErrExhausted`, `ErrCanceled`
  (retry exhaustion/cancellation = downstream/infra concern)
- `errorfamily.NewRejection(code, msg)` — for `Config.Validate()` failures
  (caller-supplied input was invalid)
- `errorfamily.NewTransient(...)` — used in tests as a retryable error
- `errorfamily.IsRetryable(err)` — the **default** retry predicate when
  `Config.IsRetryable` is nil
- `errorfamily.WrapInfrastructure(...).WithCause(err)` — to chain the last error
- `errorfamily.Classify(err)` — returns the family (asserted as `Rejection` in
  the invalid-config test)

Error codes follow a `retry.<snake_case_event>` convention
(`retry.exhausted`, `retry.canceled`, `retry.invalid_max_attempts`, etc.).

## Gotchas & Non-Obvious Conventions

- **`MaxAttempts` counts the first call, not retries on top.** `MaxAttempts: 3`
  means 1 initial call + 2 retries = 3 total invocations. Must be `>= 1`.
- **`IsRetryable` is nullable.** When `nil`, `Do` substitutes
  `errorfamily.IsRetryable` — do not assume a bare nil check means "retry
  nothing". `DefaultConfig()` pre-populates it with the same function.
- **`Backoff`/`ComputeDelay` return `(time.Duration, error)`.** An
  `attempt < 1` yields a `Rejection` (`retry.invalid_attempt`). The internal
  `Do` loop calls the unexported `computeDelay` (no error tax on a
  loop-controlled value).
- **Jitter is additive, not symmetric.** `computeDelay` adds `rand.Int64N(half)`
  _on top of_ the computed delay, so the actual wait is in `[base, base * 1.5]`,
  not centered on `base`. Tests that compare two sampled delays can be flaky;
  the existing exponential-growth test verifies the **formula**, not sampled
  values, for this reason. Follow that pattern.
- **`computeDelay` is panic-proof by design.** It sits on the failure path, so
  it must never crash: an unset/zero `MaxDelay` degrades to "no growth beyond
  `InitialDelay`", sub-2ns delays skip jitter, and `math.Pow` overflow saturates
  to `MaxDelay` instead of wrapping negative. A matrix property test
  (`TestComputeDelay_NeverPanicsAcrossMatrix`) guards this. Do not reintroduce
  an unguarded `rand.Int64N` call.
- **Concurrent call counting in tests uses `atomic.Int32`** (`sync/atomic`), not
  mutexes. Follow the same style.
- **`OnRetry` fires before the sleep**, after a failed attempt but only when more
  attempts remain. `OnExhausted` fires once after the final failure. Neither is
  called on success.
- **Cancellation during backoff returns `ErrCanceled`** wrapping the _last fn
  error_ as cause — not `context.Canceled` directly. Use `errors.Is(err,
retry.ErrCanceled)` to detect it.
- **No `flake.nix` despite the global AGENTS.md convention.** This repo predates
  / doesn't follow the LarsArtmann flake.nix pattern. Do not invent nix targets.
- **`//nolint:` directives are deliberate**, not leftover, and the referenced
  linters are now enabled in `.golangci.yml`: `exhaustruct` on `DefaultConfig`
  (optional callbacks omitted), `mnd,gosec` on the jitter divisor (weak rand is
  intentional and safe here). Removing either marker produces a real finding.
  Preserve them when editing.
- **Go files use tabs** (`.editorconfig`); YAML/JSON/Nix use 2 spaces.

## Testing Patterns

- External test package (`package retry_test`) — test the public API only.
- Every test calls `t.Parallel()`.
- Table-driven subtests use `t.Run(tt.name, ...)` (see `TestDo_InvalidConfig...`).
- `fastConfig()` helper returns a `Config` with millisecond-scale delays so the
  suite stays fast. Reuse it; don't introduce real-second delays except the
  dedicated cancellation test which deliberately uses `5s` + a goroutine cancel.
- Assertions use `t.Fatalf` with a descriptive message including the actual value.
- Validate error identity with `errors.Is`, and family with
  `errorfamily.Classify(err) == errorfamily.<Family>`.
