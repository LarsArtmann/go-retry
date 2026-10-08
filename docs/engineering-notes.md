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
times in twenty commits (the guard test caught each one). Resolution,
2026-10-08: `.buildflow.yml` skips `go-mod-update` in this repo; Dependabot
(weekly, gomod at `/` and `/tools`) is the primary dependency-update surface,
and intentional bumps happen through the release flow with the docs re-pinned
in the same change. Do not remove the `skip_steps` entry without re-deciding
this policy, and never "fix" a red `TestModuleGoDirectiveStaysPinned` by
updating the constant alone — a real re-pin means updating every doc that
states 1.26 in the same change.

The `tools/` nested module declares `go 1.26.0` (patch form) because six
`golang.org/x/*` dependencies each declare a patch-form `go 1.26.0` floor, so
a major.minor-only directive cannot represent the true minimum. That is why
`.buildflow.yml` sets
`tool_options.go-version-auto-configure.respect_patch_floor: true`: patch-form
directives stay alone while they do not exceed the installed toolchain. If the
x/* stack is ever bumped to versions with major.minor-only floors, tighten the
directive and drop the option together.

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
