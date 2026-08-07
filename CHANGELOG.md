# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

_Nothing yet._

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

[Unreleased]: https://github.com/LarsArtmann/go-retry/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/LarsArtmann/go-retry/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/LarsArtmann/go-retry/releases/tag/v0.1.0
