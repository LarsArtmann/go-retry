# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **CI coverage floor** — a dedicated `coverage` job fails below 95% statement
  coverage (local coverage is 100%; the floor leaves room for a legitimately
  hard-to-test edge). `.github/workflows/ci.yml`.
- **CI `go vet` step** — vet now runs in CI before the race-detector tests;
  previously it was a local-only check. `.github/workflows/ci.yml`.
- **Decision records** — `OnSuccess(attempts)` hook deferred (no consumer;
  derivable today; widens the API to migrate); jitter-deferral decision
  reaffirmed after the v0.4.0 cap fix removed the correctness pressure.
  `ROADMAP.md`, `FEATURES.md`.

### Changed

- **`BenchmarkComputeDelay` uses `b.Loop()`** — the Go 1.24+ idiom; timer
  reset and allocation reporting are now framework-owned. `retry_test.go`.
- **Fuzz seed corpus widened** — the original three `f.Add` seeds grew to
  seven, distilling the four input classes the 5-minute campaign discovered
  (see Added). `retry_test.go` (`FuzzComputeDelayNeverPanics`).

### Fixed

- Nothing yet.

## [0.4.0] - 2026-08-22

### Fixed

- **Backoff delay could exceed `MaxDelay` by up to 50%.** `computeDelay`
  applied the cap *before* adding jitter, so real sleeps reached
  1.5× `MaxDelay` (measured: ~300 ms against a declared 200 ms cap over a
  20 000-sample probe) while `Backoff`/`ComputeDelay` documented a hard
  cap. The jittered sum is now capped: `min(exponential + jitter,
  MaxDelay)`. Migration: if you sized timeouts or SLAs around the old
  (buggy) upper bound, re-check them — worst-case delays are now up to a
  third shorter. Pinned by `TestComputeDelay_NeverExceedsMaxDelay`
  (20 000 samples). `retry.go` (`computeDelay`).
- **Deadline-exceeded was mislabeled as cancellation.** A context whose
  deadline expired during a backoff delay returned `ErrCanceled`,
  indistinguishable from an explicit shutdown cancel — yet operators debug
  timeouts and shutdowns differently. The backoff wait now branches on
  `ctx.Err()`: an expired deadline returns the new `ErrDeadlineExceeded`
  sentinel (unwraps to `context.DeadlineExceeded`); an explicit cancel
  keeps `ErrCanceled`. Migration: deadline errors no longer match
  `ErrCanceled` — code branching on "canceled during backoff" as shutdown
  should check `ErrDeadlineExceeded` first. `retry.go` (`contextEnded`,
  `awaitBackoff`).

### Added

- **`ErrDeadlineExceeded`** — `Infrastructure` sentinel (`retry.deadline`)
  returned when the context deadline ends the loop during a backoff delay.
  Errors matching it also unwrap to `context.DeadlineExceeded`, and the
  last attempt error stays in the chain. `retry.go`.
- **Terminal errors chain the context error.** Cancel and deadline errors
  now wrap both the context error and the last attempt error (Go 1.20
  multi-`%w`), so `errors.Is(err, context.Canceled)` /
  `errors.Is(err, context.DeadlineExceeded)` hold without losing the
  attempt cause. Previously the `ErrCanceled` doc claimed it wrapped
  `context.Canceled` while the code chained only the attempt error — the
  chain now tells the truth. `retry.go` (`contextEnded`).
- **Godoc examples** — `ExampleDo_delayFunc` (honoring a server-provided
  Retry-After via `DelayFunc`) and `ExampleFromPolicy` (error-family
  `RetryPolicy` → `Config`), both deterministic with `// Output:` blocks.
  `retry_test.go`.
- **README "Exhaustion and nesting" section** — documents that exhaustion
  errors unwrap to the last attempt's error (`errors.Is` reaches it) and
  that nested retry loops are fail-closed: an outer loop does not amplify
  an inner loop's exhaustion because `Infrastructure` is not retryable by
  default. `README.md`.

### Changed

- **`Do` restructured into `awaitBackoff` + `nextDelay` helpers** —
  cyclomatic complexity dropped from 13 (over the `cyclop` max of 12, the
  repo's only lint warning) to well below it. Behavior is unchanged; the
  suite passes with `-race -count=10`. `retry.go`.
- **Configurable jitter remains deferred** (decision 2026-08-08). The cap
  fix makes the current additive strategy contract-safe, so the deferral
  stands unchanged. See `ROADMAP.md` (v1.0 section) for the rationale.

## [0.3.1] - 2026-08-08

### Fixed

- **`DelayFunc` returning 0 now falls back to default exponential backoff**
  instead of meaning "no delay." Previously a `DelayFunc` that returned `0`
  silently zeroed the wait; now `0` means "use the computed exponential backoff
  with jitter," and only a positive return overrides it. This lets callers
  override only when a server-provided delay (e.g. HTTP `Retry-After`) is
  present and fall through to the normal backoff otherwise. `retry.go` (`Do`),
  `config.go` (`DelayFunc` doc comment).

### Changed

- `DelayFunc` doc comment rewritten to clarify the zero-return semantics: a
  return `> 0` overrides; `0` means "use the default." `config.go`.
- Test renamed: `TestDo_DelayFuncZeroMeansNoWait` →
  `TestDo_DelayFuncZeroFallsBackToExponential` (asserts positive backoff, not
  near-instant completion). `retry_test.go`.

## [0.3.0] - 2026-08-07

### Added

- **`Config.DelayFunc`** — optional callback that overrides the exponential
  backoff delay for a single attempt. Receives the current attempt number and
  the error from the failed attempt, so callers can honor server-provided
  delays (e.g. HTTP `Retry-After` headers) or implement custom backoff
  strategies. `config.go` (`DelayFunc` field), `retry.go` (`Do`).
- **`FromPolicy(errorfamily.RetryPolicy) Config`** — converts an `error-family`
  retry policy into this package's `Config`, mapping `MinDelay` to
  `InitialDelay`. Retains the default multiplier, retry predicate, and unset
  hooks. `config.go` (`FromPolicy`).
- **Concurrent retry isolation test** — 100-goroutine test proving `Do`
  invocations share no mutable state. `retry_test.go`
  (`TestDo_ConcurrentInvocationsShareNoMutableState`).
- **Fuzz target for `ComputeDelay`** — `FuzzComputeDelayNeverPanics` with seeds
  for ordinary, zero-cap, overflow, and near-`MaxInt64` inputs.
  `retry_test.go`.
- **`FromPolicy` tests** — field mapping, default-preservation, and
  non-retryable-family validation. `retry_test.go`.
- **`DelayFunc` tests** — override behavior, error-receiving, zero-return
  fallback, and `OnRetry` integration. `retry_test.go`.
- **API cross-links** — `Backoff` and `ComputeDelay` doc comments now reference
  each other. `retry.go`.

## [0.2.0] - 2026-08-07

### Changed

- **Breaking:** `Backoff` and `ComputeDelay` now return
  `(time.Duration, error)`. Passing `attempt < 1` returns a `Rejection`-family
  error (`retry.invalid_attempt`) instead of computing a meaningless value via
  a negative exponent. The internal `Do` loop is unaffected — it calls an
  unexported `computeDelay` directly, so validation is enforced only at the
  external boundary where untrusted values arrive. `retry.go`.
- `Config.Validate()` now rejects `MaxDelay <= 0` with a `Rejection`-family
  error (`retry.invalid_max_delay`). Previously an unset `MaxDelay` was the
  most common trigger for the B1 panic below. `config.go`.

### Fixed

- **Three `rand.Int64N` panics on the retry failure path.** A retry library
  must never panic when a downstream call fails — a panic here converts a
  recoverable blip into a process crash. `computeDelay` is now hardened so no
  input combination can panic or return a negative duration:
  - **B1 — omitted/zero `MaxDelay`:** `min(delay, 0) == 0` made
    `Int64N(0)` panic. Now an unset cap degrades to "no growth beyond
    `InitialDelay`" instead of crashing, and `Validate()` rejects it up front.
  - **B2 — sub-2ns delays:** `int64(delay)/2 == 0` made `Int64N(0)` panic.
    Delays too small to halve now return as-is.
  - **B3 — `math.Pow` overflow:** at high attempts (e.g. plain `DefaultConfig()`
    at attempt 38) the `float64 → time.Duration` conversion wrapped to
    `INT64_MIN`; the comparison is now done in float space and saturates to
    `MaxDelay` instead of wrapping.
    `retry.go`. All three were reproduced against `v0.1.0` source before fixing.

### Added

- `FromPolicy(errorfamily.RetryPolicy)` converts advisory error-family retry
  defaults into this package's `Config`, mapping `MinDelay` to `InitialDelay`.
- `TestComputeDelay_NeverPanicsOnExtremeInputs` — regression test for each of
  B1/B2/B3 plus overflow edges; asserts no panic, non-negative delay, and the
  documented `MaxDelay + 50%` bound.
- `TestComputeDelay_NeverPanicsAcrossMatrix` — property test sweeping
  `initial × maxDelay × multiplier × attempt` to prove `computeDelay` cannot
  panic for any reachable input combination. Statement coverage could not catch
  B1/B2/B3 because the panicking lines were already exercised with benign
  inputs; this covers the input domain instead.
- `TestValidate_RejectsInvalidMaxDelay` and a `zero max delay` row in the
  table-driven invalid-config test.

## [0.1.0] - 2026-08-03

Initial public release. Signed annotated tag `v0.1.0`.

### Added

- **Core retry loop** — `Do(ctx, config, fn)` executes an `AttemptFunc` up to
  `Config.MaxAttempts` times, returning immediately on the first success.
  `retry.go`.
- **Exponential backoff with additive jitter** — `Backoff(config, attempt)`
  and the dependency-free `ComputeDelay(initial, max, mult, attempt)` compute
  `initial * mult^(n-1)`, capped at `MaxDelay`, plus random jitter up to 50% of
  the capped delay. Exported so callers can preview/log the planned delay.
  `retry.go`.
- **`Config` with defaults** — `MaxAttempts` (3), `InitialDelay` (100ms),
  `MaxDelay` (5s), `Multiplier` (2.0), plus `IsRetryable`, `OnRetry`, and
  `OnExhausted` hooks. `config.go`.
- **`Validate()`** — rejects `MaxAttempts < 1`, `InitialDelay <= 0`, and
  `Multiplier <= 1` with `Rejection`-family errors. `config.go`.
- **`error-family` integration** — `IsRetryable` defaults to
  `errorfamily.IsRetryable`; `ErrExhausted` and `ErrCanceled` are
  `Infrastructure`-family sentinels carrying stable codes (`retry.exhausted`,
  `retry.canceled`); the last `fn` error is chained via `WithCause`.
  `retry.go`, `config.go`.
- **Context cancellation during backoff** — canceling the context during a
  backoff delay returns an error wrapping `ErrCanceled`. `retry.go`.
- **Test suite** — external `retry_test` package, `t.Parallel()` on every test,
  table-driven validation tests, 100% statement coverage. `retry_test.go`.
- **Behavioral-guarantee tests** — assert that `OnRetry` does not fire after the
  final failed attempt, that a pre-canceled context yields `ErrCanceled`, and
  that `OnExhausted` receives the exact last error by identity.
  `retry_test.go`.
- godoc **`ExampleDo`** and **`ExampleDo_customIsRetryable`** — runnable,
  deterministic examples that render on `pkg.go.dev`. `retry_test.go`.
- **`BenchmarkComputeDelay`** — surfaces the backoff path's cost (~18 ns/op,
  0 allocations; the jitter path is allocation-free). `retry_test.go`.
- `.golangci.yml` — pins the golangci-lint **v2** config: default linters plus
  `gosec`, `mnd`, and `exhaustruct` (the linters the in-source `//nolint:`
  markers already reference), with `mnd`/`exhaustruct` excluded from `_test.go`.
- `docs/DOMAIN_LANGUAGE.md` — ubiquitous vocabulary for the package: retry
  terms (attempt, `MaxAttempts`, backoff, jitter, exhaustion, cancellation),
  the three `error-family` families used here (Transient / Rejection /
  Infrastructure), `IsRetryable` / `Classify` / `WithCause`, and the
  `retry.<event>` code table.
- `AGENTS.md` — non-obvious project context for AI sessions (commands, the
  `error-family` dependency map, jitter/cancellation gotchas, testing patterns).
- `FEATURES.md` — honest feature inventory by status, every entry cited to
  code; statement coverage is 100%.
- `TODO_LIST.md` — short-term, actionable open work.
- `ROADMAP.md` — long-term direction and raw ideas (v1.0 bar, options-based
  config, non-goals).
- `CONTRIBUTING.md` — prerequisites (Go 1.26, golangci-lint v2), real dev
  commands, coverage workflow, lint policy, and testing conventions.
- **Repository scaffolding** — `doc.go` (documents the no-CQRS/no-OTel
  boundary), `.editorconfig`, `.gitattributes`, `.gitignore`, `LICENSE` (MIT),
  `README.md` (comprehensive package description with runnable quick start,
  configuration table, and error model), `go.mod`
  (`github.com/larsartmann/go-retry`, Go 1.26.5, depends on
  `github.com/larsartmann/go-error-family v0.10.0`).
- **Keep-a-Changelog compare links** — `[Unreleased]` and `[0.1.0]` footer
  links resolve against the public GitHub remote.

[Unreleased]: https://github.com/LarsArtmann/go-retry/compare/v0.4.0...HEAD
[0.4.0]: https://github.com/LarsArtmann/go-retry/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/LarsArtmann/go-retry/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/LarsArtmann/go-retry/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/LarsArtmann/go-retry/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/LarsArtmann/go-retry/releases/tag/v0.1.0
