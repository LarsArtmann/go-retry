# AGENTS.md

Concise, enduring context for every AI session working in `go-retry`.
Long-form narratives for select gotchas: `docs/engineering-notes.md` — read a
section there before acting on its topic.

## What This Is

A single-package Go **library** (not an application) providing a dependency-light
retry loop with exponential backoff and jitter. Module:
`github.com/larsartmann/go-retry`, package `retry`, Go 1.26 (see `go.mod`).

This is the **core** retry primitive — intentionally free of CQRS message types
and OpenTelemetry. The CQRS-wrapped variant (`MessageAdapter`, OTel spans,
dead-letter entries carrying `StreamID`) lives in
`github.com/larsartmann/go-cqrs-lite/middleware/v4`; do **not** add such
imports here — consumers who need only retry import this package to avoid
them. See `doc.go`.

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
Dev tools (actionlint, govulncheck) are pinned in the nested `tools/` module
via Go `tool` directives — never suffix `@version`. The classic blank-import
`tools.go` is dead: Go 1.26 rejects importing main packages, and pinning tools
in the library module would force `go.mod` off its guarded `go 1.26` directive.
Consumers download none of the tools module.

## Architecture & Data Flow

Flat single-package layout — no internal subpackages (why
`.go-structure-linter.yaml` selects the `flat` preset; root package files are
correct, `internal/` would make the library unimportable):

| File                         | Responsibility                                                                                                                                 |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `retry.go`                   | `Do` + generic `DoWithValue` (loops), backoff helpers, `Backoff`, `ComputeDelay`, sentinels `ErrExhausted`/`ErrCanceled`/`ErrDeadlineExceeded` |
| `config.go`                  | `Config` struct, `DefaultConfig()`, `FromPolicy()`, `Validate()`                                                                               |
| `options.go`                 | `Option` funcs (`WithIsRetryable`, `WithDelayFunc`, `WithOnRetry`, `WithExhausted`) + `JitterStrategy` (`JitterAdditive`, `JitterNone`)        |
| `doc.go`                     | Package doc stating the no-CQRS/no-OTel boundary                                                                                               |
| `retry_test.go`                              | External test package (`retry_test`)                                                                                                                                                                         |
| `dispositions_test.go`                       | Config-disposition guards: `.buildflow.yml` skip_steps/patch-floor, interface-typed sentinels, error-family floor within the go pin, `lychee.toml` private-namespace exclude (each proven failing on drift)  |
| `docs_test.go`               | Markdown guard tests (strike rendering, table code spans, archive verdicts, status index)                                                      |
| `workflows_test.go`          | Input-allowlist guard: every pinned `uses:` action's `with:` keys checked against allowlists verified from action.yml at each SHA              |
| `tools/`                     | Nested module pinning dev tools via Go `tool` directives                                                                                       |
| `.buildflow.yml`             | BuildFlow dispositions (deliberate policy — see Gotchas)                                                                                       |
| `.go-structure-linter.yaml`  | `flat` preset — single-package library at repo root by design                                                                                  |
| `lychee.toml`                | Link-check excludes: private LarsArtmann repos 404 unauthenticated                                                                             |
| `.golangci.yml`              | Lint config: standard defaults + ~100 extra linters; `mnd`/`exhaustruct_v5` and friends excluded from `_test.go`                               |
| `.github/workflows/ci.yml`   | Push/PR CI: vet (root + `tools/`), race tests, govulncheck, 95% coverage floor, actionlint, dprint check, golangci-lint                        |
| `.github/workflows/fuzz.yml` | Daily 03:17 UTC 30-min fuzz campaign; crash-corpus artifact on failure                                                                         |
| `testdata/fuzz/...`          | Committed fuzz corpus (mirrors the `f.Add` seeds)                                                                                              |
| `docs/engineering-notes.md`  | Long-form narratives behind the compressed gotchas below                                                                                       |
| `docs/status/`               | Point-in-time session reports; resolved ones annotated inline, moved to `docs/status/archived/` (index: `docs/status/README.md`)               |

**Control flow of `Do`**: validate config → loop `attempt` 1..`MaxAttempts` →
call `fn(ctx, attempt)` → on `nil` return → if not retryable, return → else
`awaitBackoff` (exponential + `DelayFunc` override, fire `OnRetry`, sleep in a
`select` on `timer.C` vs `ctx.Done()`; a context end classifies into
`ErrDeadlineExceeded` vs `ErrCanceled`) → on exhaustion call `OnExhausted` and
return `ErrExhausted` wrapping the last error. **`DoWithValue[T]`** wraps `Do`
with a `ResultFunc[T]`; on failure it returns the zero `T` with the same error,
never leaking a partial value from an earlier attempt.

## The error-family Dependency

The sole external dependency is `github.com/larsartmann/go-error-family`
(`errorfamily` alias). Errors carry a **family** classification and a string
**code**. This package uses:

- `errorfamily.NewInfrastructure(code, msg)` — the three sentinels (exhaustion
  and context endings are downstream/infra concern)
- `errorfamily.NewRejection(code, msg)` — `Config.Validate()` failures (invalid caller input)
- `errorfamily.NewTransient(...)` — tests' retryable error
- `errorfamily.IsRetryable(err)` — the **default** predicate when `Config.IsRetryable` is nil
- `errorfamily.WrapInfrastructure(...).WithCause(err)` — chains the last error
- `errorfamily.Classify(err)` — the family (asserted `Rejection` in the invalid-config test)

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
- **Concurrent call counting in tests uses `atomic.Int32`** (`sync/atomic`), not
  mutexes. Follow the same style.
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
- **The go directive is pinned at `go 1.26`** (guarded by
  `TestModuleGoDirectiveStaysPinned`; every doc states it). Both tools that
  fight the pin are dispositioned: `.buildflow.yml` skips `go-mod-update`
  (unconditional minor-bump; Dependabot owns gomod bumps) and
  `.go-structure-linter.yaml` suppresses its `go-version` rule (bumps to the
  installed toolchain). `tools/go.mod` keeps its dep-forced patch floor
  `go 1.26.0` via `respect_patch_floor`. A real re-pin updates every doc too.
- **No `flake.nix` despite the global AGENTS.md convention.** This repo predates
  / doesn't follow the LarsArtmann flake.nix pattern. Do not invent nix targets.
- **`//nolint:` directives are deliberate**, not leftover (`exhaustruct_v5` on
  `DefaultConfig`, `gosec` on the jitter line, `errorlint` on the identity
  assertion in `TestDo_DoesNotRetryNonRetryableError`). Removing any produces a
  real finding; sweep markers when linters are added/renamed.
- **Never cite line numbers in prose docs — ours or dependencies'.** Cite by
  function/type name only (`retry.go` (`Do`)); for dependencies cite the symbol
  and, when precision matters, the version from `go.mod`.
- **The committed fuzz corpus mirrors the `f.Add` seeds**
  (`TestFuzzCorpusMirrorsSeeds` fails naming the offender); a new constant
  expression in a seed needs its corpus file (and maybe a
  `seedConstExpressions` entry) in the same change; the daily fuzz workflow
  adds crashers to the corpus AND as distilled seeds.
- **Release notes are GitHub-only**, composed at release time from the
  CHANGELOG section; there is deliberately no `docs/releases/` directory.
- **Go files use tabs** (`.editorconfig`); YAML/JSON/Nix use 2 spaces.
- **Remote-action `with:` keys are SHA-keyed allowlisted.** actionlint cannot
  see remote-action inputs; `TestRemoteActionInputsAreAllowlisted` fails any
  re-pin until inputs are re-verified from `action.yml` at the new SHA and
  `actionInputAllowlist` is updated in the same change (last: upload-artifact
  `v7.0.2`, 2026-10-08); the parser is fail-closed on unattributable YAML.
- **`setup-go`'s version manifest lags `go.dev` by hours.** A fresh patch
  release can resolve to the previous patch while toolchain switching downloads
  the exact one; fix when it bites with `GOTOOLCHAIN: go1.26.x`.
- **After touching `.golangci.yml`, run `golangci-lint config verify`.** Plain
  `run` tolerates settings the strict CI schema rejects (`exhaustruct_v5`
  accepts no `exclude` key — it bit once).
- **Terminal-error codes/messages are single-sourced constants** at the top of
  `retry.go`; the sentinels and their `WrapInfrastructure` call sites must use
  them — never re-inline the strings.
- **The go-auto-upgrade `lo.Map` suggestion is a deliberate non-fix** (adding
  `samber/lo` breaks the dependency-light contract for one test loop). See
  `docs/engineering-notes.md` before "fixing" it.

## Testing Patterns

- External test package (`package retry_test`) — test the public API only.
  Every test calls `t.Parallel()`; table-driven subtests use `t.Run`.
- `fastConfig()` returns millisecond-scale delays — reuse it; real-second delays
  only in the two context-ending tests (cancel + deadline, `5s`).
- Assertions use `t.Fatalf` with the actual value; error identity via
  `errors.Is`, family via `errorfamily.Classify(err) == errorfamily.<Family>`.
- **Repo-level doc/code invariants get a guard test** (shape:
  `TestModuleGoDirectiveStaysPinned`). A new guard must be shown to fail on the
  drift it guards; drill with `-count=1` (data-file mutations are not in Go's
  test cache key, so a cached PASS can hide them).
- **Consumer sweeps run through go-cqrs-lite's committed `go.work`** (it lists
  this repo). Never create a temp `go.work` there — it is tracked, and an
  overwrite silently discards its committed module list.

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
  log`; never trust stale status-report claims (reports are point-in-time).
- **Annotate-as-you-land:** plan/report tables get their `done at <hash>`
  verdict in the same change that completes the work (re-check `git status`
  immediately before `git add` — the daemon races explicit commits), and a
  report cites only the final hash of the work it describes.
- **Pipeline masking:** never judge a gate by a filtered tail (`| rg ... |
  head`) — filters match test names and hide failing `ok`/`FAIL` summaries.
- **Delete-then-build:** after deleting any file/package/symbol, run
  `go build ./...` immediately, before editing dependents — LSP caches lie,
  builds don't.

## `.config/metadata.yaml`

Machine-written metadata (`tags: [lib]`, `importance`, timestamps) produced
by Lars's external repo tooling — nothing in this repo reads or writes it.
Do not edit or delete it by hand; an external writer owns the file and its
timestamps.
