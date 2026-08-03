# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

> **Note on links:** this repo currently has no git remote, so the usual
> `[Unreleased]`/version compare links at the bottom of a Keep-a-Changelog file
> are intentionally omitted. They will be added when a remote is published
> (tracked in `TODO_LIST.md` → T7). There is nothing to link to today.

## [Unreleased]

### Added

- `AGENTS.md` — non-obvious project context for AI sessions working in this
  repo (commands, the `error-family` dependency map, jitter/cancellation
  gotchas, testing patterns). Commit `4f7a57c`.
- `FEATURES.md` — honest feature inventory by status, every entry cited to
  code; statement coverage is 100%.
- `TODO_LIST.md` — short-term, actionable open work (broken README, missing
  lint config, no examples/benchmarks, no CI).
- `ROADMAP.md` — long-term direction and raw ideas (v1.0 bar, options-based
  config, non-goals).

### Changed

- Corrected the `[0.1.0]` release date (was `2026-01-01`, actually
  `2026-08-03` — see the signed annotated tag `v0.1.0`) and expanded its entry
  to describe what shipped instead of the placeholder "Initial release".

## [0.1.0] - 2026-08-03

The initial release. Tagged at commit `eae60c5`
("feat(retry): implement core retry package with configurable backoff
strategies") as a signed annotated tag.

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
  table-driven validation tests, 100% statement coverage.
  `retry_test.go`.
- **Repository scaffolding** — `doc.go` (documents the no-CQRS/no-OTel
  boundary), `.editorconfig`, `.gitattributes`, `.gitignore`, `LICENSE`
  (proprietary), `README.md`, `CONTRIBUTING.md`, `go.mod`
  (`github.com/larsartmann/go-retry`, Go 1.26.5, depends on
  `github.com/larsartmann/go-error-family v0.10.0`).
