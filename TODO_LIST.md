# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish.

---

## P1

### T16 — First fuzz workflow run on a real runner (ci.yml half DONE)

**DONE (2026-09-13 18:40 CEST):** the `ci.yml` batch — `exhaustruct_v5`
migration, golangci-lint v2.12.2 → v2.13.2, per-job `timeout-minutes: 10`,
concurrency group, and the config-verify fix — is **verified green on real
runners** for `cf3cd40` (run 34766471885), `b356f63` (run 34766604503), and
`9d6700a` (run 34766960403), each 3/3 jobs (test ✅ coverage ✅ lint ✅).
The previous batch (checkout v6, govulncheck action, coverage floor) was
already green for `691744b` (run 34755167105).

**Remaining:** `fuzz.yml` has **zero runs so far**: trigger a
`workflow_dispatch` run, confirm it goes green, loads the committed corpus
(expect 14 entries in the log: 7 seeds + 7 corpus files), and only exercises
the crash-artifact step logic on failure.
Evidence: `.github/workflows/ci.yml`, `.github/workflows/fuzz.yml`.
Source: `docs/status/archived/2026-09-13_12-57_t9-t15-execution-status.md` §f.1–3.

### T21 — Cut the next release (v0.5.1 or v0.6.0) — owner-gated

`[Unreleased]` holds user-visible changes: the exhaustion error message now
reads "all retry attempts failed" (was "all attempts failed"), plus the new
example/pins and CI hardening. Decide patch vs minor (the message wording is
observable in error strings → minor-leaning). When cutting: move **all**
`[Unreleased]` entries together (including the two pre-session pinning-test
entries), compose GitHub-only release notes from the CHANGELOG section (the
T14 decision), and re-check `pkg.go.dev` rendering after the tag.
Evidence: `CHANGELOG.md` `[Unreleased]`. Source: 12-57 report §f.19–20, §f.49.

## P2

### T17 — Dependabot policy — RESOLVED (2026-09-13)

Update 2026-09-13 ~14:51 CEST: Dependabot **opened its first-ever PR**
([#1](https://github.com/LarsArtmann/go-retry/pull/1), "bump the actions group
with 3 updates") — after a repo-history of zero PRs (`gh pr list --state all`
was empty that morning). So version updates DO run; the earlier silence stays
unexplained. The update policy is decided and documented (manual SHA bumps,
`AGENTS.md` gotcha). Remaining: review PR #1 (its CI ran red because the
branch carried the invalid `exhaustruct_v5` settings block from master —
fixed on master; Dependabot auto-rebases), then merge or close it and decide
Dependabot-owned vs manual going forward. Evidence: `.github/dependabot.yml`.
Source: 12-57 report §f.4–5; this report's §d.

**DONE (2026-09-13 18:45 CEST):** PR #1 reviewed with SHA-level evidence
(each new pin fetched by SHA from upstream, `action.yml` inputs diffed
against this repo's usage; the golangci bump proved to be a zero-diff
tag-object→commit re-pin), evidence review posted on the PR, merged squash
as `e67a70e`. Policy settled: Dependabot stays enabled for github-actions +
gomod; every PR gets the verify-by-SHA review before merge.
Evidence: AGENTS.md Dependabot gotcha (updated).

### T20 — Corpus↔seeds sync test — RESOLVED (2026-09-13)

~~A test asserting every `f.Add` seed in `FuzzComputeDelayNeverPanics` has a
matching `go test fuzz v1` file in `testdata/fuzz/FuzzComputeDelayNeverPanics/`.
The mirror is currently maintained by hand with a documented sync rule
(`AGENTS.md` gotcha); the test removes the human-discipline requirement.
Evidence: `retry_test.go` (`FuzzComputeDelayNeverPanics`), `testdata/fuzz/`.
Source: 12-57 report §f.11.~~

**DONE (2026-09-13 ~19:05 CEST):** `TestFuzzCorpusMirrorsSeeds` lands —
value-level 1:1 check (seeds normalized via a `seedConstExpressions` table,
corpus files parsed in `go test fuzz v1` form, `slices.Sorted` +
`BinarySearch` multiset compare), verified to fail naming the offender in
both drift directions. Gate green: `-race -count=10`, lint 0 issues,
`config verify`, coverage 100.0%. AGENTS gotcha updated to "enforced by a
test"; CHANGELOG `[Unreleased]` entry added.

### T18 — Workflow schema gate before push

Add `actionlint` (or an equivalent workflow schema check) as a pre-push gate
so workflow YAML errors die locally instead of on the runner. Note: this repo
deliberately keeps its tool surface at `go` + `golangci-lint`, so the gate
must either justify a third tool or reuse an existing one. Evidence:
`.github/workflows/*.yml`. Source: 12-57 report §f.10.

## P3

### T19 — Fuzz job polish

Name the crash artifact with the commit SHA
(`fuzz-crash-corpus-${{ github.sha }}`) for multi-crash archaeology, and
consider `-fuzzminimizetime` tuning so crash minimization cannot eat the
45-minute job timeout on a hit. Evidence: `.github/workflows/fuzz.yml:28-34`.
Source: 12-57 report §f.40–41.
