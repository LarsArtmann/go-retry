# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Documentation guard tests.** Four gates now read the repo's markdown:
  `TestMarkdownStrikethroughSpansRender` (the three strikethrough bug classes
  from the 2026-09-17 catastrophe: space-preceded closers, lone `~` inside a
  span, unclosed spans), `TestMarkdownTableCellsCloseCodeSpans` (unclosed
  backticks per table cell, `\|`-aware), `TestArchivedReportItemsCarryVerdicts`
  (every numbered item in archived `§b`/`§c`/`§f`/`§g` sections and every
  Verdict/Status-column table row carries a verdict), and
  `TestStatusIndexCoversArchive` (archive↔index consistency, non-empty State
  cells). Each proven failing on injected drift. `docs_test.go`.
- **`scripts/check-docs.sh`.** One command for the doc battery: guard tests
  (`-count=1` — a data-file mutation is not part of Go's test cache key, so a
  cached PASS can hide it), dprint check, compare-links. Wired into the AGENTS
  Session Ritual.
- **`tools/` module vetting.** `go -C tools vet ./...` added to the CI lint job
  and the Session Ritual — the nested module was invisible to root `./...` and
  stood outside every quality gate.

### Fixed

- **Rendering repairs caught by the guard prototype.** AGENTS.md's architecture
  table carried an unclosed code span (`ErrDeadlineExceeded` rendered with a
  literal backtick); the status index's 13:19 row ended in an unclosed
  backtick; two archived strike spans carried lone tildes (`~14:55`,
  `~10 minutes` — class-B survivors of the same-day repair pass); the 05:53
  report's `§d.3` cell had mangled backticks around an escaped pipe.
- **Two archive annotations.** The 08:47 report's `§f.8` (T37) is now struck
  done (v0.7.1 pkg.go.dev page verified) and `§f.15` (T42) carries its
  `→ routed to` pointer — the two items the marker gate found genuinely
  unannotated.

## [0.7.1] - 2026-09-17

A documentation-and-tooling patch: library code, public API, and dependencies
are byte-identical to v0.7.0. Cut so pkg.go.dev finally renders the example
set and the new guarantees ship on a tagged page.

### Added

- **Post-v0.7.0 godoc examples.** `ExampleBackoff` (the `Rejection` path for
  `attempt < 1`) and `ExampleComputeDelay` (the hard-cap determinism case) are
  output-pinned; they landed after the v0.7.0 tag, so v0.7.0's pkg.go.dev page
  does not carry them and they surface with the next cut. `retry_test.go`.
- **Module go-directive guard.** `TestModuleGoDirectiveStaysPinned` fails when
  `go.mod`'s `go` directive drifts from the deliberately relaxed `go 1.26`
  that every living doc states, so external tooling can no longer re-pin it
  silently. `retry_test.go`.
- **Pinned development-tools module.** A nested `tools/` module pins
  `actionlint` (v1.7.12) and `govulncheck` (v1.8.0) with Go `tool`
  directives — the classic blank-import `tools.go` is rejected by the Go
  1.26 toolchain, and pinning inside the library module would force the
  guarded `go 1.26` directive to a patch pin. The CI lint job builds
  actionlint from this pin instead of `go install ...@version`; dprint (a
  Rust binary) is referenced in `tools/tools.go` and pinned where it runs in
  CI. Dependabot's gomod watcher now covers `/tools`. `tools/go.mod`,
  `tools/tools.go`.
- **Remote-action input-allowlist guard.**
  `TestRemoteActionInputsAreAllowlisted` extracts every pinned `uses:` step
  and its `with:` keys from `.github/workflows` and fails when a key is not
  a real input of that action (allowlists verified from each action.yml at
  the exact pinned SHA) — the 2026-09-13 `namee:` typo class that
  actionlint's schema check cannot see. The parser is fail-closed (flow
  mappings, anchors, and merge keys abort the test), also asserts SHA
  pinning, and flags stale allowlist rows; the allowlist is keyed by
  `action@SHA`, so re-pinning an action fails the test until its inputs are
  re-verified from the new action.yml, closing the upstream-rename blind
  spot. `workflows_test.go`.
- **`ExampleDo_withOptions`.** The options tail gets its output-pinned godoc
  counterpart to the README's verified snippet: `WithOnRetry` plus
  `WithJitter(JitterNone)` with deterministic exact delays. `retry_test.go`.

### Changed

- **dprint check in CI.** The lint job gains a `dprint/check` step
  (action SHA-pinned, `dprint-version: 0.57.4`, attestation-verified
  download), so markdown/JSON/YAML drift from non-session writers fails on
  master instead of only in the local ritual. `.github/workflows/ci.yml`.
- **`go.mod` directive restored to `go 1.26`.** An external writer had
  re-pinned it to `go 1.27.1`, which contradicted the recorded decision, broke
  the documented local gate ritual under `GOTOOLCHAIN=local`, and drifted every
  doc that states the Go version. Nothing here needs 1.27: the sole dependency
  declares `go 1.26` and the code uses no post-1.26 feature. `go.mod`.

### Fixed

- Nothing yet.

## [0.7.0] - 2026-09-16

### Added

- **Per-call options (`Option` tail).** `Do` and `DoWithValue` grow a
  variadic `opts ...Option` tail — a purely additive, source-compatible
  change. Options apply left-to-right after the caller's `Config` is copied
  and before `Validate`; a later option wins over the field and earlier
  options, nil options are ignored, and the passed `Config` is never
  modified. `Config`'s public 8-field shape is frozen and pinned by tests.
  `options.go`, `retry.go`.
- **Callback mirror options.** `WithIsRetryable`, `WithDelayFunc`,
  `WithOnRetry`, and `WithExhausted` mirror Config's four callbacks for one
  call; the fields keep working forever. `options.go`.
- **Jitter strategy (`WithJitter`).** The twice-deferred jitter question
  lands as a designed capability: `JitterAdditive` (the zero-value default,
  byte-identical to the historical behavior) and `JitterNone` (pure capped
  exponential — deterministic delays for tests and previews). Unknown
  strategy values fall back to additive without panicking.
  `Backoff`/`ComputeDelay` always preview the additive default.
  `options.go`, `retry.go` (`computeDelay`).
- **Deterministic RNG (`WithRandomSource`).** Inject a `math/rand/v2`
  `Source` (e.g. seeded `rand.NewPCG`) to reproduce exact jittered delay
  sequences; nil keeps the goroutine-safe global generator.
  `options.go`, `retry.go` (`jitterValue`).

### Changed

- **`TestBackoff_IncreasesExponentially` asserts real delays.** With
  `JitterNone` making delays deterministic, the test verifies the exact
  computed sequence through the public API instead of re-deriving the
  formula inside the test. `retry_test.go`.

## [0.6.1] - 2026-09-16

### Added

- **Deadline-budget recipe.** The README's "Deadline budgets" section
  documents the count-based retry contract and the pattern of setting
  `MaxDelay` below a remaining time budget, so an over-deadline run fails
  fast with `ErrDeadlineExceeded` instead of sleeping past the cutoff.
  Documentation only; no API change. `README.md`.
- **Corpus↔seeds mirror is enforced by a test.**
  `TestFuzzCorpusMirrorsSeeds` parses the `f.Add` seeds from the test source
  and the committed corpus files, normalizes both (including the constant
  expressions `time.Millisecond`, `math.MaxInt64`, …), and fails naming the
  offending entry when either side drifts. The hand-discipline documented in
  `AGENTS.md` is now machine-checked. `retry_test.go`.

### Changed

- **Fuzz workflow hardening.** The crash-corpus artifact is now named with
  the commit SHA (`fuzz-crash-corpus-<sha>`) so artifacts from different
  failing runs cannot shadow each other, and crash minimization is bounded
  with `-fuzzminimizetime 5m` so it cannot consume the job's 45-minute
  timeout on a hit. `.github/workflows/fuzz.yml`.
- **CI tests run with `-shuffle=on`.** Go shuffles test order with a logged
  seed on every CI run, catching order dependencies early; a trial of 5
  shuffled runs plus 4 fixed-seed runs found none. `.github/workflows/ci.yml`.
- **Workflow schema gate.** The CI `lint` job now runs
  [actionlint](https://github.com/rhysd/actionlint) (pinned `v1.7.12`,
  installed from the Go module proxy) over every workflow before
  `golangci-lint`, so invalid workflow YAML fails in seconds with a named
  error instead of surfacing as mysterious job failures. Locally:
  `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.12`.
  `.github/workflows/ci.yml`, `CONTRIBUTING.md`.

## [0.6.0] - 2026-09-13

### Added

- **Scheduled fuzz campaign in CI.** A daily `Fuzz` workflow runs
  `FuzzComputeDelayNeverPanics` for 30 minutes (03:17 UTC, plus manual
  `workflow_dispatch`) and uploads any crash corpus it discovers as an
  artifact on failure. `.github/workflows/fuzz.yml`.
- **Committed fuzz corpus.** The seven `f.Add` seeds are mirrored in
  `testdata/fuzz/FuzzComputeDelayNeverPanics/`, so the corpus survives local
  cache loss and future campaign discoveries accumulate in-repo; seeded runs
  (`go test -run '^FuzzComputeDelayNeverPanics$'`) exercise it without
  fuzzing. `testdata/fuzz/`.
- **`govulncheck` in CI.** The `test` job now scans the module with the
  official `golang/govulncheck-action` (SHA-pinned, Go pinned to `go.mod`).
  Local run against the current code: no known vulnerabilities.
  `.github/workflows/ci.yml`.
- **Non-retryable errors are pinned by identity.**
  `TestDo_DoesNotRetryNonRetryableError` now asserts `err != rejection`, not
  just `errors.Is`, so `Do` can never silently start re-wrapping a typed
  non-retryable error. `retry_test.go`.
- **Decision record: release-notes bodies are GitHub-only.** Composed at
  release time from the matching `CHANGELOG.md` section (the `go-release`
  flow); there is deliberately no `docs/releases/` mirror — a third copy
  would drift. `ROADMAP.md`, `AGENTS.md`.
- **Nesting fail-closed guarantee is now pinned by a test.** An outer `Do`
  makes exactly one attempt when an inner loop returns `ErrExhausted`
  (`Infrastructure` is not retryable by default) — the guarantee was
  documented in the README and godoc but previously enforced by nothing.
  `TestDo_NestedRetriesAreFailClosed`. `retry_test.go`.
- **`OnExhausted` is pinned to never fire on context end.** Cancellation and
  deadline termination return without the exhaustion callback.
  `TestDo_OnExhaustedNotCalledOnCancel` /
  `TestDo_OnExhaustedNotCalledOnDeadline`. `retry_test.go`.
- **`ExampleDoWithValue` godoc example.** The value-returning API — v0.5.0's
  headline — now has a runnable, output-pinned example rendering on
  `pkg.go.dev`; the README snippet is its verified form. `retry_test.go`.
- **Nested amplification is pinned in the override direction.** A deliberate
  `IsRetryable` override re-enables nested retries exactly as documented:
  `outer(3) × inner(3) = 9` attempts
  (`TestDo_NestedRetriesAmplifyWhenOverridden`), completing the fail-closed
  pin from the default-predicate side. `retry_test.go`.

### Changed

- **Lint config migrated `exhaustruct` → `exhaustruct_v5`** (the old linter is
  deprecated since golangci-lint v2.13.0 and printed a warning on every run);
  the `DefaultConfig` `//nolint:` marker moved with it, and CI's pinned
  golangci-lint was bumped v2.12.2 → v2.13.2 to match the version the repo
  develops against. The v4 stdlib `exclude` settings block was dropped — v5's
  schema rejects that key (strict `config verify`, which CI runs, caught it
  after a plain local `run` did not); the `_test.go` exclusion rule is
  unaffected. `.golangci.yml`, `config.go`, `.github/workflows/ci.yml`.
- **CI jobs carry `timeout-minutes: 10` and a concurrency group** (superseded
  pushes to the same ref cancel in-flight runs), mirroring the fuzz workflow.
  `.github/workflows/ci.yml`.
- **Terminal-error codes and messages are single-sourced.** `contextEnded`
  and the exhaustion wrapper now derive their code/message from the same
  constants as the `ErrCanceled`, `ErrDeadlineExceeded`, and `ErrExhausted`
  sentinels, so a rename cannot drift the two sites apart. User-visible
  effect: exhaustion errors now carry the sentinel's message ("all retry
  attempts failed") where the wrapper previously said "all attempts failed".
  `retry.go`.
- **CI runs on `actions/checkout` v6** (SHA-pinned), off the deprecated Node
  20 runtime that v4 used; `setup-go` was already v6.
  `.github/workflows/ci.yml`.

## [0.5.0] - 2026-09-06

### Added

- **`DoWithValue[T]`** — generic companion to `Do` for retries that produce a
  value: returns the successful attempt's result with a nil error, and the
  zero value with the same error `Do` would produce on any failure
  (non-retryable, exhaustion, context end). `retry.go`, `retry_test.go`.
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

[Unreleased]: https://github.com/LarsArtmann/go-retry/compare/v0.7.1...HEAD
[0.7.1]: https://github.com/LarsArtmann/go-retry/compare/v0.7.0...v0.7.1
[0.7.0]: https://github.com/LarsArtmann/go-retry/compare/v0.6.1...v0.7.0
[0.6.1]: https://github.com/LarsArtmann/go-retry/compare/v0.6.0...v0.6.1
[0.6.0]: https://github.com/LarsArtmann/go-retry/compare/v0.5.0...v0.6.0
[0.5.0]: https://github.com/LarsArtmann/go-retry/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/LarsArtmann/go-retry/compare/v0.3.1...v0.4.0
[0.3.1]: https://github.com/LarsArtmann/go-retry/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/LarsArtmann/go-retry/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/LarsArtmann/go-retry/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/LarsArtmann/go-retry/releases/tag/v0.1.0
