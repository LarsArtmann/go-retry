# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish.

---

## P1

### T16 — Verify CI changes on real runners — RESOLVED (2026-09-13)

~~**DONE (2026-09-13 18:40 CEST):** the `ci.yml` batch — `exhaustruct_v5`
migration, golangci-lint v2.12.2 → v2.13.2, per-job `timeout-minutes: 10`,
concurrency group, and the config-verify fix — is **verified green on real
runners** for `cf3cd40` (run 34766471885), `b356f63` (run 34766604503), and
`9d6700a` (run 34766960403), each 3/3 jobs (test ✅ coverage ✅ lint ✅).
The previous batch (checkout v6, govulncheck action, coverage floor) was
already green for `691744b` (run 34755167105).~~

~~**Remaining:** `fuzz.yml` has **zero runs so far**: trigger a
`workflow_dispatch` run, confirm it goes green, loads the committed corpus
(expect 14 entries in the log: 7 seeds + 7 corpus files), and only exercises
the crash-artifact step logic on failure.~~

**CLOSED (2026-09-13 ~19:55 CEST):** `fuzz.yml` dispatched and **green**
(run 34771258905, 30-min campaign, ~38k execs/s, 0 failures). Corpus load
confirmed: the run starts from exactly the 7 seed values — the 14 on-disk
files (7 `f.Add` seeds + 7 corpus mirrors) dedupe to those values, which is
precisely what `TestFuzzCorpusMirrorsSeeds` guarantees; the fuzz cache then
grew to 50+ interesting inputs. The crash-artifact step **skipped cleanly**
on green (`skipped`, not failed). `-fuzzminimizetime 5m` accepted by the
runner. Every CI surface in this repo has now executed on a real runner.
Source: `docs/status/archived/2026-09-13_12-57_t9-t15-execution-status.md` §f.1–3.

### T21 — Cut the next release — RESOLVED (2026-09-13): v0.6.0 shipped

~~`[Unreleased]` holds user-visible changes: the exhaustion error message now
reads "all retry attempts failed" (was "all attempts failed"), plus the new
example/pins and CI hardening. Decide patch vs minor (the message wording is
observable in error strings → minor-leaning). When cutting: move **all**
`[Unreleased]` entries together (including the two pre-session pinning-test
entries), compose GitHub-only release notes from the CHANGELOG section (the
T14 decision), and re-check `pkg.go.dev` rendering after the tag.
Evidence: `CHANGELOG.md` `[Unreleased]`. Source: 12-57 report §f.19–20, §f.49.~~

**DONE:** v0.6.0 cut 2026-09-13 — minor (user-visible message wording),
all 13 entries promoted together, signed annotated tag on `6105848`, GitHub
Release composed from the CHANGELOG section, proxy + clean-room
`go get` verified. `pkg.go.dev` render check rode the documented lag window.
Evidence: tag `v0.6.0`, release
https://github.com/LarsArtmann/go-retry/releases/tag/v0.6.0.

---

**No open work.** Every P1–P3 item is resolved. New findings go through the
docs-health HARVEST route (`TODO_LIST.md` ← status reports / session
discoveries); long-term bets live in `ROADMAP.md`.

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

### T18 — Workflow schema gate — RESOLVED (2026-09-13)

~~Add `actionlint` (or an equivalent workflow schema check) as a pre-push gate
so workflow YAML errors die locally instead of on the runner. Note: this repo
deliberately keeps its tool surface at `go` + `golangci-lint`, so the gate
must either justify a third tool or reuse an existing one. Evidence:
`.github/workflows/*.yml`. Source: 12-57 report §f.10.~~

**DONE (2026-09-13 ~19:15 CEST):** surface decided — **CI gate, not a third
local tool**. The `lint` job now installs actionlint `v1.7.12` from the Go
module proxy (fits the Go-only tool surface; no download script, version
pinned) and runs it over both workflows before `golangci-lint`. Local
surface is the same one-liner via `go run` (documented in CONTRIBUTING +
AGENTS commands); a persistent local binary/pre-push hook was rejected —
`go install` of new tools is blocked in the AI environment and the repo
keeps its two-tool convention. Local `go run` validation + a deliberate
break test on a throwaway branch prove the gate bites.

## P3

### T19 — Fuzz job polish — RESOLVED (2026-09-13)

~~Name the crash artifact with the commit SHA
(`fuzz-crash-corpus-${{ github.sha }}`) for multi-crash archaeology, and
consider `-fuzzminimizetime` tuning so crash minimization cannot eat the
45-minute job timeout on a hit. Evidence: `.github/workflows/fuzz.yml:28-34`.
Source: 12-57 report §f.40–41.~~

**DONE (2026-09-13 ~19:10 CEST):** artifact renamed to
`fuzz-crash-corpus-${{ github.sha }}`; `-fuzzminimizetime 5m` added
(30 m fuzz + 5 m minimization = 35 m, inside the 45-minute timeout).
CHANGELOG `[Unreleased]` entry added.
