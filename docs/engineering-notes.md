# Engineering Notes

Long-form narratives behind the compressed gotchas in `AGENTS.md`. Each section
is the single detailed source for its topic; `AGENTS.md` carries only the
one-line summary and points here. Read a section before acting on its topic.

## The go-directive pin and the BuildFlow fight

The module pins `go 1.26` in `go.mod` (major.minor, no patch). Everything
states that floor: the README, FEATURES, CONTRIBUTING, the CHANGELOG, and the
guard test `TestModuleGoDirectiveStaysPinned` (proven failing on drift before
it landed). CI resolves the toolchain from `go.mod` itself, so the directive is
load-bearing, not cosmetic.

BuildFlow's `go-mod-update` step bumps the directive to the installed
toolchain's minor unconditionally — that policy silently re-lost the pin three
times in twenty commits (the guard test caught each one). From 2026-10-08 the
repo carried a `.buildflow.yml` skip for it; on 2026-10-27 the owner decided
the repo should need no BuildFlow config at all, so the file and its guard
tests were removed and the behavior was filed upstream instead (links in
`ROADMAP.md`). Until upstream lands a pin-respecting default, any
`buildflow update`/`--fix` run may re-write the directive: the guard trips,
restore the pin and re-run the battery. Dependabot (weekly, gomod at `/` and
`/tools`) remains the primary dependency-update surface, and intentional
bumps happen through the release flow with the docs re-pinned in the same
change. Never "fix" a red `TestModuleGoDirectiveStaysPinned` by updating the
constant alone — a real re-pin means updating every doc that states 1.26 in
the same change.

`go-mod-update` is not the only writer: the `go-structure-linter` tool's
`go-version` repair rule also bumps the directive ("go.mod specifies Go 1.26,
but Go 1.27 is available", error-severity, so it trips the findings gate too).
Caught on 2026-10-08 by watching go.mod during a full `--fix` run and
correlating the flip timestamp with the pipeline's step log — every BuildFlow
gomod step was exonerated individually first, so the writer was hiding in a
tool whose name suggests structure-only checks. `.go-structure-linter.yaml`
suppresses the rule for `go.mod` with the pin rationale. When a file flips
between runs, bisect steps AND check repair-only rules in non-obvious tools.

The `tools/` nested module declares `go 1.26.0` (patch form) because six
`golang.org/x/*` dependencies each declare a patch-form `go 1.26.0` floor, so
a major.minor-only directive cannot represent the true minimum (`go mod
tidy` re-raises it after any flattening). BuildFlow's go-line rewriters
downgrade such patch pins by default; the former `respect_patch_floor` opt-in
lived in `.buildflow.yml` and was removed with it on 2026-10-27 — the default
is filed upstream too (links in `ROADMAP.md`). If a `--fix` run flattens
`go 1.26.0` to `go 1.26`, restore it — the flip-flop is the documented BF2
class. If the x/* stack is ever bumped to versions with major.minor-only
floors, tighten the directive.

## Sentinel errors are typed `error`, not `*errorfamily.Error`

The three terminal sentinels (`ErrExhausted`, `ErrCanceled`,
`ErrDeadlineExceeded`) are declared `var X error = errorfamily.NewInfrastructure(...)`.
The interface type is the erraudit sentinel pattern: with a concrete
`*errorfamily.Error` type, erraudit flags every `errors.Is(err, ErrX)` call
site as `legacy_is` because the second argument is not recognized as a
sentinel value. Do not narrow the declarations back to the concrete type; if
erraudit findings reappear on these call sites, check the declaration type
first.

## The go-auto-upgrade lo.Map suggestion is a deliberate non-fix

`go-auto-upgrade` (stdlib2lo migrator) suggests replacing a manual
transform-and-append loop in `retry_test.go` with `lo.Map`. The suggestion is
sound Go advice in general but wrong here: this library's design contract is
dependency-lightness, `go-error-family` is the sole external dependency, and
adopting `samber/lo` for one three-line test loop would land a new require
line in the module every consumer resolves. The finding is warning-severity
and does not trip the findings gate. If it ever becomes an error, revisit the
disposition (skip the step or adopt lo deliberately), do not suppress blindly.

## go-structure-linter: the flat preset is deliberate

`.go-structure-linter.yaml` selects `presets: [flat]`. go-retry is a
single-package library whose package lives at the repo root by design;
`internal/` would make it unimportable by consumers entirely, and `pkg/`
would churn the import path for zero benefit. The `root-package-files` and
`internal-directory` findings this disables are false positives for this
layout. Do not "fix" them by moving files, and do not delete the config file.

## lychee: private-repo links are excluded, not authenticated

`lychee.toml` excludes `https://github.com/larsartmann/.*` (case-insensitive).
Markdown in this repo links private LarsArtmann repositories, which 404 under
unauthenticated link checks, and the fleet-wide authenticate-vs-exclude policy
is undecided — so this repo excludes the namespace instead of requiring
`GITHUB_TOKEN` in every environment. If the fleet picks a policy, revisit.

## Jitter is additive, not symmetric, and hard-capped

`computeDelay` adds `rand.Int64N(half)` on top of the capped exponential delay
and then caps the sum at `MaxDelay`, so the actual wait lands in
`[base, min(base * 1.5, MaxDelay)]` — never above `MaxDelay`. The cap applies
to the jittered sum: capping before jitter would let real sleeps reach 1.5x
`MaxDelay` while the docs promise a hard cap. Tests that compare two sampled
delays can be flaky for this reason; the existing exponential-growth test
verifies the formula, not sampled values. Follow that pattern.

The jitter strategy question closed in v0.7.0: `JitterAdditive` (zero value =
the historical default, byte-identical) and `JitterNone` (deterministic) are
`WithJitter` options only, never `Config` fields. `Full`, `Equal`, and
`Decorrelated` strategies are recorded ROADMAP follow-ups — no constants exist
for them, and an unknown strategy value falls back to additive by design. Do
not re-propose `Jitter`/`JitterStrategy` as a public `Config` field, and do
not add the deferred strategy constants without implementing them.

## Remote-action inputs: the allowlist is keyed by SHA

actionlint validates workflow schema, not remote-action inputs — the runner
silently ignores unknown `with:` keys. `TestRemoteActionInputsAreAllowlisted`
(`workflows_test.go`) closes that gap: every `with:` key is checked against
allowlists verified from each action's `action.yml` at the pinned SHA, and the
allowlist is keyed by `action@SHA`, so a re-pin fails the test until the
inputs are re-verified (an upstream input rename can no longer pass silently).
The parser is fail-closed on YAML shapes it cannot attribute (flow mappings,
anchors, merge keys).

Re-verification flow, proven 2026-09-13 and again 2026-10-08 (the
`actions/upload-artifact` re-pin to `cf430e030ddbb5b0abf93d22962f4752f3646cd9`,
tag `v7.0.2`): fetch the pinned commit's `action.yml` by SHA —
`gh api "repos/<owner>/<repo>/contents/action.yml?ref=<sha>" --jq .content | base64 -d`
— diff its inputs against the repo's `with:` usage, update
`actionInputAllowlist` under the new `action@SHA` key in the same change, and
post the evidence. Annotated-tag pins may be re-pinned by Dependabot to the
peeled commit (zero code change — tag object to the same commit).

## Dependabot PRs are an expected supply-chain surface

Configured weekly (`dependabot.yml`: gomod at `/` and `/tools` +
github-actions). Review flow that works: fetch each pinned commit by SHA from
upstream (content-addressed — the hash proves what will run), read its
`action.yml` inputs, diff against this repo's `with:` usage (the
`workflows_test.go` allowlist fails on unknown keys — update it in the same
change), post the evidence, merge. The gomod watcher will not touch the `go`
directive regardless. go-error-family bumps additionally get one line in the
`ROADMAP.md` bump trace when they merge. Note: BuildFlow's `go-mod-update`
step is skipped here (see the go-directive section), so Dependabot is the
primary gomod update surface.

## setup-go's version manifest lags go.dev

A fresh Go patch release can resolve `go-version-file`/`go-version` to the
previous patch, while Go's own toolchain switching downloads the exact version
regardless — a tag-CI run can then disagree with the local toolchain. Fix when
it bites: pin `GOTOOLCHAIN: go1.26.x` at job or workflow level (go-release
skill Phase 4; verified against the skill source 2026-09-16).

## nolint markers are deliberate

The `//nolint:` directives in this repo are deliberate, and the referenced
linters are enabled in `.golangci.yml`: `exhaustruct_v5` on `DefaultConfig`
(optional callbacks omitted), `gosec` on the jitter line (weak rand is
intentional and safe here; `mnd` does not fire — `2` is in its
ignored-numbers), and `errorlint` on the identity comparison in
`TestDo_DoesNotRetryNonRetryableError` (the whole point of the assertion is
`err != rejection`). Removing any marker produces a real finding; a renamed
linter marker silences nothing. Sweep markers whenever linters are added or
renamed.

## Never cite line numbers in prose docs

Own-file citations rot on the next insertion above them (this happened twice:
the T10 const block shifted every citation in FEATURES/DOMAIN_LANGUAGE within
a day). Cite by function/type name only (`retry.go` (`Do`)); function names
are unique in this package, so nothing is lost. The same applies to dependency
sources: no `classify.go:NN` refs into `go-error-family` — cite the symbol
(`Classify`, `IsRetryable`) and, when precision matters, the dependency
version from `go.mod`.

## BuildFlow health-check tools: only dprint and lychee apply here

`buildflow doctor` checks ~50 tools across every ecosystem it supports; most
show as unavailable in any single repo. For this Go library exactly two are
applicable, and both are now satisfied: `lychee` (the link-scan step —
installed 2026-10-08 via `nix profile install nixpkgs#lychee`, first real scan
green: 75 links, 0 errors, 26 excluded by the private-namespace rule) and
`dprint` (the format gate — installed the same day; note the local profile
floats with nixpkgs, currently 0.60.1, while CI pins 0.57.4 in `dprint/check`
— the recorded float-posture question in ROADMAP covers this). Everything else
unavailable (bandit, cargo-*, ruff/mypy/pytest, jest/tsc/vue/svelte, prettier,
eslint, knip/madge/c8, protolint, hadolint, shellcheck, markdownlint, gci,
vulnix, …) targets languages or gates this repo does not have; their absence
is a disposition, not a gap. Re-audit only if a BuildFlow step for one of them
starts appearing in this repo's runs.

## erraudit stays a local gate, not a CI step (2026-10-08 disposition)

The natural home for an erraudit CI gate is the pinned `tools/` module, but
the binary (`github.com/larsartmann/erraudit`, verified
`v0.5.1-0.20260922174106-1c6809adf02e`) declares `go 1.27`. Pinning it would
drag `tools/go.mod` off its deliberate patch-form `go 1.26.0` story (six x/*
deps), silently retire the `respect_patch_floor` disposition by floor-rise
instead of by need, and add a toolchain download to every lint job — a
Verschlimmbesserung against the go-pin war's clarity. The regression that
motivated the step (terminal sentinels retyped to a concrete type) is already
CI-guarded by `TestTerminalSentinelsStayInterfaceTyped`, and every full
`buildflow` run executes the erraudit step locally, where the binary is
installed. Revisit when erraudit ships a go-1.26-compatible or stable
release. Until then the working invocation is
`erraudit lint ./... --type legacy_as` (exits 0 on root and `tools/`;
`--type-aware` additionally reports six sentinel `errors.Is` advisories in
`retry_test.go` — the keep-as-is class its own message describes).

## File-by-file map (moved out of AGENTS.md 2026-10-27)

AGENTS.md is length-capped by `go-structure-linter` (`agent-config` rule,
maximum 220 lines — the file is injected into every AI session and must stay
concise), so the per-file responsibility table lives here:

| File                           | Responsibility                                                                                                                                                                                                                       |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `retry.go`                     | `Do` + generic `DoWithValue` (loops), backoff helpers, `Backoff`, `ComputeDelay`, sentinels `ErrExhausted`/`ErrCanceled`/`ErrDeadlineExceeded`                                                                                       |
| `config.go`                    | `Config` struct, `DefaultConfig()`, `FromPolicy()`, `Validate()`                                                                                                                                                                     |
| `options.go`                   | `Option` funcs (`WithIsRetryable`, `WithDelayFunc`, `WithOnRetry`, `WithExhausted`) + `JitterStrategy` (`JitterAdditive`, `JitterNone`)                                                                                              |
| `doc.go`                       | Package doc stating the no-CQRS/no-OTel boundary                                                                                                                                                                                     |
| `retry_test.go`                | External test package (`retry_test`)                                                                                                                                                                                                 |
| `dispositions_test.go`         | Config-disposition guards: interface-typed sentinels, error-family floor within the go pin, `lychee.toml` private-namespace exclude (each proven failing on drift)                                                                  |
| `docs_test.go`                 | Markdown guard tests (strike rendering, table code spans, archive verdicts, status index)                                                                                                                                            |
| `workflows_test.go`            | Input-allowlist guard: every pinned `uses:` action's `with:` keys checked against allowlists verified from action.yml at each SHA                                                                                                    |
| `tools/`                       | Nested module pinning dev tools via Go `tool` directives                                                                                                                                                                             |
| `.go-structure-linter.yaml`    | `flat` preset — single-package library at repo root by design                                                                                                                                                                        |
| `lychee.toml`                  | Link-check excludes: private LarsArtmann repos 404 unauthenticated                                                                                                                                                                   |
| `.golangci.yml`                | Lint config: standard defaults + ~100 extra linters; `mnd`/`exhaustruct_v5` and friends excluded from `_test.go`                                                                                                                     |
| `.github/workflows/ci.yml`     | Push/PR CI: vet (root + `tools/`), race tests, govulncheck, 95% coverage floor, actionlint, dprint check, golangci-lint                                                                                                              |
| `.github/workflows/fuzz.yml`   | Daily 03:17 UTC 30-min fuzz campaign; crash-corpus artifact on failure                                                                                                                                                               |
| `.github/workflows/stress.yml` | Daily 04:33 UTC 10-sweep shuffled race suite (`-count=10`); per-push CI stays `-count=1` — the 2026-10-08 select-race flakes were invisible at 1 and reproducible at 10                                                              |
| `testdata/fuzz/...`            | Committed fuzz corpus (mirrors the `f.Add` seeds)                                                                                                                                                                                    |
| `docs/engineering-notes.md`    | Long-form narratives behind the compressed AGENTS.md gotchas                                                                                                                                                                         |
| `docs/status/`                 | Point-in-time session reports; resolved ones annotated inline, moved to `docs/status/archived/` (index: `docs/status/README.md`)                                                                                                     |

## The 2026-10-15 errorfamily bump that broke the go pin (restored 2026-10-27)

An auto-committed dependency sweep landed `go-error-family` v0.12.0 and a
`go 1.27` directive on 2026-10-15 (commit `0f61e18`). v0.12.0 declares
`go 1.27`, and a dependency's floor forces the module directive up — exactly
the silent re-pin `TestModuleGoDirectiveStaysPinned` and
`TestErrorFamilyFloorStaysWithinGoPin` exist to catch. Both guards sat red on
master for 12 days (the fossilized-red-state class the ROADMAP open question
tracks; the no-pre-commit-hook decision means nothing blocked the daemon's
commit). The restore to v0.11.0 loses nothing: v0.12.0's only change is a
nil-receiver `Unwrap` guard this package never exercises, and Go requirements
are minimums — consumers wanting v0.12.0 still resolve it via MVS while this
module keeps its `go 1.26` floor for retry-only consumers. Re-adopt a newer
go-error-family only as a deliberate re-pin with the full doc sweep the
guards demand.

## Small verified facts (2026-10-27 docs-health pass)

- **The go-cqrs-lite consumer bump hash is `065b2e9c5`** (2026-10-09): the
  commit in that repo's `middleware/` module that pinned
  `github.com/larsartmann/go-retry v0.8.0`, closing the ledger gap the
  2026-10-09 02:14 report flagged as §b5/§f13.
- **The root-level `reports/` directory is the gitignored local coverage
  output** (`.gitignore` `reports/`); it holds `coverage.out` written by the
  local battery. It is a scratch dir by design, not a ghost — do not document
  it further or commit it.
- **The nightly stress workflow was the red-master alarm.** Its scheduled
  runs failed every night 2026-10-22 → 10-27 (04:4x UTC) on the two pin
  guards — faithfully reporting the fossilized red master nobody watched.
  Master went green again with the 2026-10-27 pin restore (CI runs
  40571159852 / 40572338572); the first post-restore stress run (2026-10-28
  04:33 UTC) is the expected green.
- **The daemon deleted the guarded `.buildflow.yml` twice on 2026-10-27.**
  First mid-report (commit `7d19442`, restored same hour), then again
  ~21:02 during the docs-health pass as an uncommitted working-tree
  deletion, restored from HEAD (`6791e1d`) before the daemon could commit
  it; both disposition guards re-verified green. The empirical answer to
  the 20:43 report's §f2: the guards already fail loud on the file's
  absence (`TestBuildflowDispositionsStayCommitted` fatals on the read
  error) — the missing piece is prevention, not detection.
