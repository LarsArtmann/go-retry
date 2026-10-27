# AGENTS.md

Concise, enduring context for every AI session working in `go-retry`. Long-form
narratives: `docs/engineering-notes.md` — read a section there before acting on
its topic. This file sits at its line budget: one gotcha in, one gotcha out;
new context lands in `docs/engineering-notes.md` with a one-line pointer here.

## What This Is

A single-package Go **library** (not an app): a dependency-light retry loop
with exponential backoff and jitter. Module `github.com/larsartmann/go-retry`,
package `retry`, Go 1.26. It is the **core** primitive — deliberately free of
CQRS message types and OpenTelemetry; the CQRS-wrapped variant (`MessageAdapter`,
OTel spans, `StreamID` dead-letter entries) lives in
`github.com/larsartmann/go-cqrs-lite/middleware/v4`. Do **not** add such
imports here — retry-only consumers import this package to avoid them (`doc.go`).

## Commands

No `flake.nix`, `Makefile`, or `justfile` — `go` and `golangci-lint` are the
only build/test tools; [dprint](https://dprint.dev) (committed `dprint.json`)
formats markdown/JSON/YAML/Dockerfile via nix. Raw Go commands:

```bash
go test ./... -race             # tests (always with -race; backoff uses math/rand/v2)
go test ./... -race -count=10   # flake-prone jitter/backoff tests
golangci-lint run ./...         # lint (committed .golangci.yml: standard defaults + ~100 extra linters, incl. gosec/mnd/exhaustruct_v5)
go vet ./... && go -C tools vet ./...   # vet; the nested tools module stands outside root ./...
./scripts/check-docs.sh         # doc battery: guard tests + dprint + compare-links
go test -run '^$' -fuzz '^FuzzComputeDelayNeverPanics$' -fuzztime 5m .   # fuzz campaign
go test -run '^FuzzComputeDelayNeverPanics$' .                          # seeded corpus run (no fuzzing)
go -C tools install github.com/rhysd/actionlint/cmd/actionlint golang.org/x/vuln/cmd/govulncheck  # install pinned dev tools into ~/go/bin (rerun per bump)
actionlint -verbose             # workflow schema gate (also first step of CI lint job)
govulncheck ./...               # vulnerability scan (CI uses the official action)
nix run nixpkgs#dprint -- check # markdown/JSON/YAML format gate (fmt to fix; CHANGELOG.md excluded; also gated in CI via dprint/check)
```

`go test` is the only verification gate; the package is consumed as a library.
Dev tools (actionlint, govulncheck) are pinned in the nested `tools/` module via
Go `tool` directives (never suffix `@version`); the classic blank-import
`tools.go` is dead — Go 1.26 rejects importing main packages, and consumers
download none of the tools module.

## Architecture & Data Flow

Flat single-package layout — no internal subpackages (`internal/` would make
the library unimportable; why `.go-structure-linter.yaml` selects the `flat`
preset); nothing lives in subdirectories except `tools/` (nested dev-tools
module), `docs/`, `testdata/fuzz/`, and `scripts/`. The file-by-file
responsibility map lives in `docs/engineering-notes.md` ("File-by-file map").

**Control flow of `Do`**: validate config → loop `attempt` 1..`MaxAttempts` →
call `fn(ctx, attempt)` → on `nil` return → if not retryable, return → else
`awaitBackoff` (exponential + `DelayFunc` override, fire `OnRetry`, sleep in a
`select` on `timer.C` vs `ctx.Done()`; a context end classifies into
`ErrDeadlineExceeded` vs `ErrCanceled`) → on exhaustion call `OnExhausted` and
return `ErrExhausted` wrapping the last error (every terminal promotes a coded
last attempt's code outermost — gotcha below). **`DoWithValue[T]`** wraps `Do`
with a `ResultFunc[T]`; on failure it returns the zero `T` with the same error,
never leaking a partial value from an earlier attempt.

## The error-family Dependency

The sole external dependency is `github.com/larsartmann/go-error-family`
(`errorfamily` alias). Errors carry a **family** classification and a string
**code**. This package uses:

- `errorfamily.NewInfrastructure(code, msg)` — the three sentinels (exhaustion
  and context endings are downstream/infra concern); `NewRejection(...)` —
  `Config.Validate()` failures; `NewTransient(...)` — tests' retryable error
- `errorfamily.IsRetryable(err)` — the **default** predicate when `Config.IsRetryable` is nil
- `errorfamily.WrapInfrastructure(...).WithCause(err)` — chains the last error
- `errorfamily.Classify(err)` — the family (asserted `Rejection` in the
  invalid-config test); `errorfamily.Code(err)` — the **outermost** coded layer
  (terminals promote the last attempt's code, so it answers with the
  operational code, not the `retry.*` bookkeeping code)

Error codes follow a `retry.<snake_case_event>` convention (`retry.exhausted`,
`retry.canceled`, `retry.deadline`, `retry.invalid_max_attempts`, etc.).

## Gotchas & Non-Obvious Conventions

- **`MaxAttempts` counts the first call, not retries on top.** `MaxAttempts: 3`
  = 1 initial call + 2 retries = 3 total invocations. Must be `>= 1`.
- **`IsRetryable` is nullable.** When `nil`, `Do` substitutes
  `errorfamily.IsRetryable` — `DefaultConfig()` pre-populates the same
  function; a bare nil check never means "retry nothing".
- **`Backoff`/`ComputeDelay` return `(time.Duration, error)`.** An
  `attempt < 1` yields a `Rejection` (`retry.invalid_attempt`); the internal
  `Do` loop calls the unexported `computeDelay` (no error tax there).
- **Sentinels are declared `var X error = errorfamily.NewInfrastructure(...)`.**
  The interface type is the erraudit sentinel pattern; narrowing back to
  `*errorfamily.Error` re-flags every `errors.Is` call site as `legacy_is`.
- **Jitter is additive and hard-capped** — the cap applies to the jittered sum
  (`[base, min(base * 1.5, MaxDelay)]`); tests verify the **formula**, not
  sampled values. Strategies are `WithJitter` options only (`JitterAdditive`
  zero value = historical default, `JitterNone` deterministic; `Full`/`Equal`/
  `Decorrelated` are ROADMAP follow-ups, unknown values fall back to additive).
  Never a `Config` field. Details: `docs/engineering-notes.md`.
- **`computeDelay` is panic-proof by design.** It sits on the failure path: an
  unset/zero `MaxDelay` degrades to "no growth beyond `InitialDelay`", sub-2ns
  delays skip jitter, `math.Pow` overflow saturates to `MaxDelay`. Guarded by
  `TestComputeDelay_NeverPanicsAcrossMatrix`; never reintroduce an unguarded
  `rand.Int64N` call.
- **Callback timing: `OnRetry` fires before the sleep**, after a failed attempt
  but only when more attempts remain. `OnExhausted` fires once after the final
  failure and never on context end (pinned by `TestDo_OnExhaustedNotCalledOnCancel`
  / `...OnDeadline`); neither fires on success.
- **Context endings during backoff are distinguished.** Deadline exceeded
  returns `ErrDeadlineExceeded` (unwraps to `context.DeadlineExceeded`); cancel
  returns `ErrCanceled` (unwraps to `context.Canceled`). Both chain the last
  attempt error via multi-`%w`; `Error.Is` matches by code+family, so
  `errors.Is(err, retry.ErrCanceled)` is false for deadline errors. Do not
  collapse the branches — operators debug timeouts vs shutdowns differently.
- **Terminals promote the last attempt's code outermost** (issue #3). A coded
  final attempt error adds an outermost `WrapInfrastructure(terminal,
  innerCode, msg)` layer: `Code` answers the inner code, `Classify` stays
  `Infrastructure` (fail-closed nesting pinned by
  `TestDo_NestedExhaustionWithPromotedCodeStaysTerminal`), `errors.Is`/`AsType`
  surfaces unchanged; uncoded last errors keep the plain terminal. The rendered
  string repeats the terminal message — accepted cost of every layer staying a
  native `errorfamily.Error`.
- **The go directive is pinned at `go 1.26`** (guarded by
  `TestModuleGoDirectiveStaysPinned`; every doc states it); there is
  deliberately **no `.buildflow.yml`** (owner decision 2026-10-27). BuildFlow's
  `go-mod-update` bumps the directive to the installed minor and its rewriters
  strip patch-form floors — both filed upstream (links in `ROADMAP.md`); until
  then `buildflow update`/`--fix` may re-lose the pin: guards trip, restore per
  `docs/engineering-notes.md`. `.go-structure-linter.yaml` suppresses its
  `go-version` rule; `tools/go.mod` keeps its dep-forced `go 1.26.0` floor;
  dep side: `TestErrorFamilyFloorStaysWithinGoPin` (v0.12.0 saga: notes).
- **No `flake.nix` despite the global AGENTS.md convention.** Do not invent nix
  targets. Same stance for go-auto-upgrade's `lo.Map` suggestion: adding
  `samber/lo` breaks the dependency-light contract for one test loop — see
  `docs/engineering-notes.md` before "fixing" it.
- **Never cite line numbers in prose docs — ours or dependencies'.** Cite by
  function/type name only (`retry.go` (`Do`)); for dependencies cite the symbol
  and, when precision matters, the version from `go.mod`.
- **The committed fuzz corpus mirrors the `f.Add` seeds**
  (`TestFuzzCorpusMirrorsSeeds` fails naming the offender); a seed's new
  constant expression needs its corpus file in the same change; the daily fuzz
  workflow adds crashers to the corpus AND as distilled seeds.
- **Release notes are GitHub-only**, composed at release time from the
  CHANGELOG section; there is deliberately no `docs/releases/` directory.
  **Go files use tabs** (`.editorconfig`); YAML/JSON/Nix use 2 spaces.
- **Remote-action `with:` keys are SHA-keyed allowlisted.** actionlint cannot
  see remote-action inputs; `TestRemoteActionInputsAreAllowlisted` fails any
  re-pin until inputs are re-verified from `action.yml` at the new SHA and
  `actionInputAllowlist` is updated in the same change (last: upload-artifact
  `v7.0.2`, 2026-10-08); the parser is fail-closed on unattributable YAML.
- **`setup-go`'s version manifest lags `go.dev` by hours.** A fresh patch
  release can resolve to the previous patch; fix with `GOTOOLCHAIN: go1.26.x`
  when it bites. With `go-version-file: go.mod` and the minor-only directive
  it cannot produce a wrong-toolchain build (narrative:
  `docs/engineering-notes.md`).
- **No pre-commit hook by decision (2026-10-08).** The auto-commit daemon
  commits every few minutes; a gate-running hook would run the full battery
  on each daemon tick. Fossilized red states are a fleet-level problem —
  see the ROADMAP open question on a daemon-side guard.
- **Lint-gate quirks:** after touching `.golangci.yml`, run
  `golangci-lint config verify` (plain `run` tolerates settings the strict CI
  schema rejects — `exhaustruct_v5` accepts no `exclude` key, it bit once);
  the in-source `//nolint:` markers (per-marker rationale in
  `CONTRIBUTING.md`) and the hash-precision `// art-dupl:accept` directives in
  `retry_test.go` (reviewed clones; they suppress standalone — verified
  2026-10-27 with no BuildFlow config, findings gate green) are deliberate —
  removing any produces a real finding.
- **Terminal-error codes/messages are single-sourced constants** at the top of
  `retry.go`; the sentinels and their `WrapInfrastructure` call sites must use
  them — never re-inline the strings.
- **`.config/metadata.yaml` is machine-written** (`tags: [lib]`, `importance`,
  timestamps) by Lars's external repo tooling; nothing in this repo reads or
  writes it — never edit or delete it by hand.

## Testing Patterns

- External test package (`package retry_test`) — test the public API only.
  Every test calls `t.Parallel()`; table-driven subtests use `t.Run`.
- Concurrent call counting in tests uses `atomic.Int32` (`sync/atomic`), not
  mutexes. Follow the same style.
- `fastConfig()` returns millisecond-scale delays — reuse it; real-second delays
  only in the two context-ending tests (cancel + deadline, `5s`).
- Assertions use `t.Fatalf` with the actual value; error identity via
  `errors.Is`, family via `errorfamily.Classify(err) == errorfamily.<Family>`.
- **Repo-level doc/code invariants get a guard test** (shape:
  `TestModuleGoDirectiveStaysPinned`). A new guard must be shown to fail on the
  drift it guards; drill with `-count=1` (data-file mutations are not in Go's
  test cache key, so a cached PASS can hide them).
- **Consumer sweeps bump the pin, then run through go-cqrs-lite's committed
  `go.work`.** The retry consumer there is the separate `middleware` module
  (its own `go.mod`, pinning go-retry); the workspace `use` list does **not**
  include this repo (verified 2026-10-08), so a sweep means `go get
  github.com/larsartmann/go-retry@<version>` in `middleware/`, then its suite
  plus the ritual suites (`commandlifecycle`, `integration`,
  `example/taskmanager`) with `-race`. Never create a temp `go.work` there — it
  is tracked, and an overwrite silently discards its committed module list.

## Session Ritual (self-checks before claiming done)

- **Gate order:** `gofmt -l .` → `go vet ./...` → `go -C tools vet ./...` →
  `go test ./... -race -count=10` → `golangci-lint run ./...` →
  `golangci-lint config verify` (mandatory after any `.golangci.yml` touch) →
  coverage if tests changed → `govulncheck ./...` (release cuts scan root +
  `tools/`) → `./scripts/check-docs.sh` (after any doc edit).
- **Coverage canonical format:** `go test -cover ./...` — read the
  `coverage: 100.0% of statements` line; the CI floor is 95%. Never quote
  coverage from `go tool cover` output without the `go test -cover` line.
- **Hash verification:** verify every cited commit hash with `git show`/`git
  log`; never trust stale status-report claims or handoff summaries — open
  sessions with fresh `git status --branch` before acting on repo state.
- **Annotate-as-you-land:** plan/report tables get their `done at <hash>`
  verdict in the same change that completes the work (re-check `git status`
  immediately before `git add` — the daemon races explicit commits), and a
  report cites only the final hash of the work it describes.
- **Pipeline masking:** never judge a gate by a filtered tail (`| rg ... |
  head`) — filters match test names and hide failing `ok`/`FAIL` summaries.
  **Delete-then-build:** after deleting any file/package/symbol, run
  `go build ./...` immediately, before editing dependents — LSP caches lie,
  builds don't.
