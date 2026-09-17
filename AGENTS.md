# AGENTS.md

Concise, enduring context for every AI session working in `go-retry`.

## What This Is

A single-package Go **library** (not an application) providing a dependency-light
retry loop with exponential backoff and jitter. Module:
`github.com/larsartmann/go-retry`, package `retry`, Go 1.26 (see `go.mod`).

This is the **core** retry primitive — intentionally free of CQRS message types
and OpenTelemetry. The CQRS-wrapped variant (`MessageAdapter`, OTel spans,
dead-letter entries carrying `StreamID`) lives in
`github.com/larsartmann/go-cqrs-lite/middleware/v4`. Do **not** add CQRS or OTel
imports here; consumers who need only retry (CLIs, batch jobs, simple services)
import this package to avoid pulling in those deps. See `doc.go`.

## Commands

No `flake.nix`, `Makefile`, or `justfile` exists in this repo — `go` and
`golangci-lint` are the only build/test tools; [dprint](https://dprint.dev)
(committed `dprint.json`) formats markdown/JSON/YAML/Dockerfile and runs via
nix. Use raw Go commands:

```bash
go test ./... -race             # tests (always with -race; backoff uses math/rand/v2)
go test ./... -race -count=10   # flake-prone jitter/backoff tests
golangci-lint run ./...         # lint (committed .golangci.yml: standard defaults + ~100 extra linters, incl. gosec/mnd/exhaustruct_v5)
go vet ./...
go test -run '^$' -fuzz '^FuzzComputeDelayNeverPanics$' -fuzztime 5m .   # fuzz campaign
go test -run '^FuzzComputeDelayNeverPanics$' .                          # seeded corpus run (no fuzzing)
go -C tools install github.com/rhysd/actionlint/cmd/actionlint golang.org/x/vuln/cmd/govulncheck  # install pinned dev tools into ~/go/bin (rerun per bump)
actionlint -verbose             # workflow schema gate (also first step of CI lint job)
govulncheck ./...               # vulnerability scan (CI uses the official action)
nix run nixpkgs#dprint -- check # markdown/JSON/YAML format gate (fmt to fix; CHANGELOG.md excluded; also gated in CI via dprint/check)
```

`go test` is the only verification gate. There is no build step beyond `go build`
(the package is consumed as a library). Development tools (actionlint,
govulncheck) are version-pinned in the nested `tools/` module via Go `tool`
directives — never suffix `@version` when running them. The classic
blank-import `tools.go` is dead: Go 1.26 rejects importing main packages, and
pinning tools in the library module would force `go.mod` off its guarded
relaxed `go 1.26` directive (current x/* tool versions declare `go 1.26.0`).
Consumers download none of the tools module.

## Architecture & Data Flow

Flat single-package layout — no internal subpackages:

| File                                         | Responsibility                                                                                                                                                                             |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `retry.go`                                   | `Do` + generic `DoWithValue` (loops), `awaitBackoff`/`nextDelay`/`contextEnded` helpers, `Backoff`, `ComputeDelay`, sentinels `ErrExhausted` / `ErrCanceled` / `ErrDeadlineExceeded        |
| `config.go`                                  | `Config` struct, `DefaultConfig()`, `FromPolicy()`, `Validate()`                                                                                                                           |
| `doc.go`                                     | Package doc stating the no-CQRS/no-OTel boundary                                                                                                                                           |
| `retry_test.go`                              | External test package (`retry_test`)                                                                                                                                                       |
| `workflows_test.go`                          | Input-allowlist guard: every pinned `uses:` action's `with:` keys checked against allowlists verified from action.yml at each SHA (catches the `namee:` typo class)                        |
| `tools/`                                     | Nested module pinning dev tools (actionlint, govulncheck) via Go `tool` directives; `tools.go` documents usage and the dprint reference                                                    |
| `.golangci.yml`                              | Lint config: standard defaults + ~100 extra linters; `mnd`/`exhaustruct_v5` and friends excluded from `_test.go`                                                                           |
| `.github/workflows/ci.yml`                   | Push/PR CI: vet, race tests, govulncheck, 95% coverage floor, actionlint (built from `tools/go.mod` pin), dprint check (SHA-pinned `dprint/check`), golangci-lint (version pinned in-repo) |
| `.github/workflows/fuzz.yml`                 | Daily 03:17 UTC 30-min fuzz campaign; crash-corpus artifact on failure                                                                                                                     |
| `testdata/fuzz/FuzzComputeDelayNeverPanics/` | Committed fuzz corpus (mirrors the `f.Add` seeds)                                                                                                                                          |
| `docs/status/`                               | Point-in-time session reports; resolved ones are annotated inline and moved to `docs/status/archived/` (index: `docs/status/README.md`)                                                    |

**Control flow of `Do`**: validate config → loop `attempt` from 1 to
`MaxAttempts` → call `fn(ctx, attempt)` → on `nil` return immediately → if not
retryable, return immediately → else `awaitBackoff` (compute delay via
`nextDelay` = exponential + `DelayFunc` override, fire `OnRetry`, sleep in a
`select` on `timer.C` vs `ctx.Done()`; a context end classifies via
`contextEnded` into `ErrDeadlineExceeded` vs `ErrCanceled`) → on exhaustion
call `OnExhausted` and return `ErrExhausted` wrapping the last error via
`.WithCause()`.

**`DoWithValue[T]`** wraps `Do` with a `ResultFunc[T]`; on any failure it
returns the zero `T` with the same error `Do` would produce, and never leaks
a partial value from an earlier attempt.

## The error-family Dependency

The sole external dependency is
`github.com/larsartmann/go-error-family` (`errorfamily` import alias). Errors
carry a **family** classification and a string **code**. This package uses:

- `errorfamily.NewInfrastructure(code, msg)` — for `ErrExhausted`,
  `ErrCanceled`, `ErrDeadlineExceeded` (retry exhaustion / context endings =
  downstream/infra concern)
- `errorfamily.NewRejection(code, msg)` — for `Config.Validate()` failures
  (caller-supplied input was invalid)
- `errorfamily.NewTransient(...)` — used in tests as a retryable error
- `errorfamily.IsRetryable(err)` — the **default** retry predicate when
  `Config.IsRetryable` is nil
- `errorfamily.WrapInfrastructure(...).WithCause(err)` — to chain the last error
- `errorfamily.Classify(err)` — returns the family (asserted as `Rejection` in
  the invalid-config test)

Error codes follow a `retry.<snake_case_event>` convention
(`retry.exhausted`, `retry.canceled`, `retry.deadline`,
`retry.invalid_max_attempts`, etc.).

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
- **Jitter is additive, not symmetric, and hard-capped.** `computeDelay`
  adds `rand.Int64N(half)` _on top of_ the capped exponential delay and then
  caps the sum at `MaxDelay`, so the actual wait is in
  `[base, min(base * 1.5, MaxDelay)]` — never above `MaxDelay`. The cap
  applies to the jittered sum; capping before jitter would let real sleeps
  reach 1.5× `MaxDelay` while the docs promise a hard cap.
  Tests that compare two sampled delays can be flaky; the existing
  exponential-growth test verifies the **formula**, not sampled values, for
  this reason. Follow that pattern.
- **`computeDelay` is panic-proof by design.** It sits on the failure path, so
  it must never crash: an unset/zero `MaxDelay` degrades to "no growth beyond
  `InitialDelay`", sub-2ns delays skip jitter, and `math.Pow` overflow saturates
  to `MaxDelay` instead of wrapping negative. A matrix property test
  (`TestComputeDelay_NeverPanicsAcrossMatrix`) guards this. Do not reintroduce
  an unguarded `rand.Int64N` call.
- **Concurrent call counting in tests uses `atomic.Int32`** (`sync/atomic`), not
  mutexes. Follow the same style.
- **Callback timing: `OnRetry` fires before the sleep**, after a failed
  attempt but only when more attempts remain. `OnExhausted` fires once after
  the final failure **and never on context end** (cancel or deadline
  termination returns without it; pinned by
  `TestDo_OnExhaustedNotCalledOnCancel` / `...OnDeadline`). Neither callback
  is called on success.
- **Context endings during backoff are distinguished.** A deadline
  exceeded returns `ErrDeadlineExceeded` (unwraps to
  `context.DeadlineExceeded`); an explicit cancel returns `ErrCanceled`
  (unwraps to `context.Canceled`). Both also chain the last attempt error
  via Go 1.20 multi-`%w`, and `Error.Is` matches by code+family — so
  `errors.Is(err, retry.ErrCanceled)` is false for deadline errors. Do not
  collapse the two branches; operators debug timeouts vs shutdowns
  differently.
- **Jitter strategy landed via `WithJitter` — do not regress it to a field.**
  The twice-deferred jitter question closed in v0.7.0: `JitterAdditive`
  (zero value = the historical default, byte-identical) and `JitterNone`
  (deterministic) are options-only, never `Config` fields. `Full`,
  `Equal`, and `Decorrelated` strategies are recorded ROADMAP follow-ups —
  no constants exist for them, and an unknown strategy value falls back to
  additive by design. Do not re-propose `Jitter`/`JitterStrategy` as a
  public `Config` field, and do not add the deferred strategy constants
  without implementing them.
- **No `flake.nix` despite the global AGENTS.md convention.** This repo predates
  / doesn't follow the LarsArtmann flake.nix pattern. Do not invent nix targets.
- **`//nolint:` directives are deliberate**, not leftover, and the referenced
  linters are enabled in `.golangci.yml`: `exhaustruct_v5` on `DefaultConfig`
  (optional callbacks omitted), `gosec` on the jitter line (weak rand is
  intentional and safe here; `mnd` does not fire — `2` is in its
  ignored-numbers), and `errorlint` on the identity comparison in
  `TestDo_DoesNotRetryNonRetryableError` (the whole point of the assertion is
  `err != rejection`). Removing any marker produces a real finding.
  Preserve them when editing.
- **Never cite line numbers in prose docs — ours or dependencies'.** Own-file
  citations rot on the next insertion above them (this happened twice: the
  T10 const block shifted every citation in FEATURES/DOMAIN_LANGUAGE within a
  day). Cite by function/type name only (`retry.go` (`Do`)); function names
  are unique in this package, so nothing is lost. The same applies to
  dependency sources: no `classify.go:NN` refs into `go-error-family` — cite
  the symbol (`Classify`, `IsRetryable`) and, when precision matters, the
  dependency version from `go.mod`.
- **Dependabot PRs are now an expected supply-chain surface — verify, then
  merge.** Configured weekly (`dependabot.yml`: gomod at `/` and `/tools` +
  github-actions); it stayed silent until 2026-09-13, then opened its first
  PR (#1, actions group), which was SHA-verified and merged the same day
  (`e67a70e`). Review flow that worked: fetch each pinned commit **by SHA**
  from upstream (content-addressed — the hash proves what will run), read its
  `action.yml` inputs, diff against this repo's `with:` usage (the
  `workflows_test.go` allowlist now fails on unknown keys — update it in the
  same change), post the evidence, merge. Annotated-tag pins may be re-pinned
  by Dependabot to the peeled commit (zero code change — `v9` tag object →
  same commit). The gomod watcher will not touch the `go` directive
  regardless; go-error-family bumps additionally get one line in the
  `ROADMAP.md` bump trace when they merge.
- **After touching `.golangci.yml`, run `golangci-lint config verify`.** Plain
  `golangci-lint run` tolerates settings that strict schema validation (what
  the CI action executes first) rejects. This bit once: `exhaustruct_v5`
  accepts no `exclude` settings key (unlike v4), local `run` stayed green,
  and the CI lint job went red on push.
- **Terminal-error codes/messages are single-sourced constants** at the top of
  `retry.go` (`codeExhausted`/`msgExhausted`, `codeCanceled`/`msgCanceled`,
  `codeDeadline`/`msgDeadline`). The sentinels and their
  `WrapInfrastructure` call sites must use them — never re-inline the strings.
- **The committed fuzz corpus mirrors the `f.Add` seeds, enforced by a test.**
  `TestFuzzCorpusMirrorsSeeds` fails with the offending entry named when a
  seed lacks its `testdata/fuzz/FuzzComputeDelayNeverPanics/` file or a
  corpus entry matches no seed; a seed introducing a new constant expression
  needs it added to `seedConstExpressions` in the same change. A daily
  scheduled workflow (`.github/workflows/fuzz.yml`) fuzzes for 30 minutes;
  new crashers land in the corpus AND as distilled seeds, together.
- **Release notes are GitHub-only.** Bodies are composed at release time from
  the CHANGELOG section (the `go-release` skill flow); there is deliberately
  no `docs/releases/` directory.
- **Go files use tabs** (`.editorconfig`); YAML/JSON/Nix use 2 spaces.
- **actionlint validates structure, not remote-action inputs.** Cron formats,
  expressions, and workflow schema errors die in seconds (the gate is the first
  lint-job step), but a typo'd _input key_ on a `uses:` action passes actionlint
  silently — the runner ignores unknown inputs. Since 2026-09-17 the gap is
  guarded by `TestRemoteActionInputsAreAllowlisted` (`workflows_test.go`):
  every `with:` key is checked against allowlists verified from each action's
  `action.yml` at the pinned SHA; the allowlist is keyed by `action@SHA`, so
  a re-pin fails the test until the inputs are re-verified (an upstream
  input rename can no longer pass silently). Pinning a new action or
  re-pinning an existing one means re-verifying inputs at the new SHA and
  updating `actionInputAllowlist` under the new `action@SHA` key in the same
  change; the parser is fail-closed on
  YAML shapes it cannot attribute (flow mappings, anchors, merge keys).
- **`setup-go`'s version manifest lags `go.dev` by hours.** A fresh Go patch
  release can resolve `go-version-file`/`go-version` to the _previous_ patch,
  while Go's own toolchain switching downloads the exact version regardless —
  a tag-CI run can then disagree with the local toolchain. Fix when it bites:
  pin `GOTOOLCHAIN: go1.26.x` at job or workflow level (go-release skill Phase
  4; verified against the skill source 2026-09-16).

## Testing Patterns

- External test package (`package retry_test`) — test the public API only.
- Every test calls `t.Parallel()`.
- Table-driven subtests use `t.Run(tt.name, ...)` (see `TestDo_InvalidConfig...`).
- `fastConfig()` helper returns a `Config` with millisecond-scale delays so the
  suite stays fast. Reuse it; don't introduce real-second delays except the
  two context-ending tests (cancel + deadline) which deliberately use `5s`
  delays so the context end fires during the wait.
- Assertions use `t.Fatalf` with a descriptive message including the actual value.
- Validate error identity with `errors.Is`, and family with
  `errorfamily.Classify(err) == errorfamily.<Family>`.
- **Repo-level doc/code invariants get a guard test.** `go.mod`'s `go`
  directive is pinned by `TestModuleGoDirectiveStaysPinned` (proven failing on
  drift before it landed), so external tooling cannot silently re-pin it away
  from the `go 1.26` every living doc states. Add the same shape when a
  documented claim has no other enforcement.
- **Consumer sweeps run through go-cqrs-lite's committed `go.work`** — it
  already lists `/home/lars/projects/go-retry` as a `use` target, so running
  `commandlifecycle`/`integration`/`example/taskmanager` suites there resolves
  local go-retry master with no overlay files. Never create a temp `go.work`
  in that repo: the file is tracked, and an overwrite silently discards its
  committed module list (recovered once, 2026-09-17).

## Session Ritual (self-checks before claiming done)

- **Gate order:** `gofmt -l .` → `go vet ./...` → `go test ./... -race
  -count=10` → `golangci-lint run ./...` → `golangci-lint config verify`
  (mandatory after any `.golangci.yml` touch — plain `run` tolerates schema
  violations the CI action rejects) → coverage if tests changed →
  `govulncheck ./...` (release cuts scan root + `tools/`) →
  `nix run nixpkgs#dprint -- check` (markdown/JSON/YAML drift) →
  `./scripts/check-compare-links.sh` (after any CHANGELOG link edit).
- **Coverage canonical format:** `go test -cover ./...` — read the
  `coverage: 100.0% of statements` line; the CI floor is 95% (decided
  2026-09-16: keep 95 — rationale in `ROADMAP.md`). Never quote coverage from `go tool cover`
  output without the `go test -cover` line as source.
- **Test-failure proof:** a new guard test must be shown to FAIL on the drift
  it guards (temporarily break the fixture, observe the named failure,
  restore). A test that was never seen failing is unverified.
- **Hash verification:** every commit hash cited as evidence is verified with
  `git show`/`git log` before writing it into a report; status-report claims
  are re-verified against fresh CLI runs, never trusted (reports are
  point-in-time).
- **Annotate-as-you-land:** plan/report tables get their `done at <hash>`
  verdict in the same change that completes the work (re-check `git status`
  immediately before `git add` — the daemon races explicit commits), and a
  report cites only the final hash of the work it describes, never an
  intermediate one.
- **Pipeline masking:** never judge a gate by a filtered tail (`| rg ... |
  head`); read the raw `ok`/`FAIL` summary lines — filters match test names
  and hide failing summaries.
- **Delete-then-build:** after deleting any file/package/symbol, run
  `go build ./...` immediately, before editing dependents — LSP caches lie,
  builds don't.
- **Marker coverage:** when linters are added/renamed, sweep the `//nolint:`
  markers; a renamed linter marker silences nothing and re-produces findings.

## `.config/metadata.yaml`

Machine-written metadata (`tags: [lib]`, `importance`, timestamps) produced
by Lars's external repo tooling — nothing in this repo reads or writes it,
and its `updated_at` changes without repo activity. Do not edit or delete it
by hand; an external writer owns the file and its timestamps.
