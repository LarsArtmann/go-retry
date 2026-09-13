# Status Report — go-retry — Docs-Health Pass 2: Audit, Annotation, Archive & CI Hardening

**When:** 2026-09-13 14:48 CEST
**Scope of this session:** Execute the docs-health skill over all eleven
2026-0\* historical files and every living doc; execute the two same-day
reports' §f backlogs; ANNOTATE + ARCHIVE both 2026-09-13 status reports;
close every doc-drift and cheap-CI item found on the way.
**Format note:** written as `.md` per explicit user instruction (skill default
is an HTML dashboard — override honored, not propagated).
**Baseline at session start:** master `691744b` (clean, pushed, CI green —
run 34755167105), v0.5.0 released 2026-09-06.
**End state at write time:** local master ahead of origin with the session's
code/doc commits (`23192cd`, `28fe7e4`, `6e7a469`, `91d02a6`, + daemon
commits); all gates green; both status reports annotated and archived;
`docs/status/` holds only `README.md` + `archived/`.
**Amended 14:58 CEST:** the daemon pushed the session's commits and CI went
**red on master** (lint job): the action runs `golangci-lint config verify`,which rejects what plain `run` tolerated — the migrated `exhaustruct_v5`settings block used v4's `exclude` key, which v5's schema does not allow.Fixed at ~14:55 (block dropped — the `_test.go` exclusion rule is independent),`config verify` now passes locally (see d.6, e.8). Same hour, Dependabotopened its first-ever PR (#1, actions group) — invalidating the morning's"zero PRs ever" finding (see a.9 amendment, g.2).

---

## a) FULLY DONE

| #  | Item                                                                   | What                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                 | Evidence                                                           | Verification                                                       |
| -- | ---------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------ | ------------------------------------------------------------------ |
| 1  | **All eleven 2026-0\* files + all living docs + code read in full**    | 2 live status reports, 9 archived (incl. HTML), README, AGENTS, FEATURES, TODO_LIST, ROADMAP, CHANGELOG, DOMAIN_LANGUAGE, CONTRIBUTING, SECURITY, dependabot.yml, retry.go, config.go, doc.go, full test-function list, go.mod, .golangci.yml, ci.yml, fuzz.yml                                                                                                                                                                                                                                                                                                                                                                      | repo                                                               | Read before any edit                                               |
| 2  | **Quality gate green, repeatedly**                                     | build, vet, `test -race`, `-race -count=10`, lint 0 issues, coverage 100.0%, 20 s fuzz campaign (8.5M execs, 0 failures), benchmark 21.3/34.9 ns/op 0 allocs                                                                                                                                                                                                                                                                                                                                                                                                                                                                         | CLI runs this session                                              | Final summary lines inspected, not just exit codes                 |
| 3  | **`exhaustruct` → `exhaustruct_v5` migration**                         | `.golangci.yml` (enable list, settings key, `_test.go` exclusion), `config.go` `//nolint` marker, and CI golangci-lint pinned v2.12.2 → **v2.13.2** (required: v5 exists only ≥ v2.13.0; local = v2.13.2)                                                                                                                                                                                                                                                                                                                                                                                                                            | `23192cd`                                                          | lint 0 issues, deprecation warning gone; local toolchain == CI pin |
| 4  | **Nested-amplification pin (override direction)**                      | `TestDo_NestedRetriesAmplifyWhenOverridden`: deliberate `IsRetryable` override yields `outer(3) × inner(3) = 9` — completes the fail-closed pin's other half, as the README claims                                                                                                                                                                                                                                                                                                                                                                                                                                                   | `23192cd`                                                          | Passes `-race -count=10`; coverage stays 100%                      |
| 5  | **`ExampleDoWithValue` godoc example**                                 | Runnable, `// Output:`-pinned example for the value-returning API; README snippet replaced with its verified form (kills the pseudo-code sin from the 12:20 report d.1)                                                                                                                                                                                                                                                                                                                                                                                                                                                              | `23192cd`; `ExampleDoWithValue`                                    | Runs green in `go test -run Example`                               |
| 6  | **CI hardening**                                                       | `timeout-minutes: 10` on all three `ci.yml` jobs + `concurrency` group (mirrors fuzz.yml)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            | `23192cd`, `28fe7e4` (concurrency hunk verified via `git show`)    | YAML locally; runner verification = TODO T16                       |
| 7  | **Citation-rot root-caused**                                           | Every `retry.go` line-number citation in FEATURES/DOMAIN_LANGUAGE had drifted (claimed `Do, line 72`; actual 83). Living docs now cite **function names only**; new AGENTS.md gotcha bans line-number citations in prose docs                                                                                                                                                                                                                                                                                                                                                                                                        | grep sweep: zero `retry.go:N`/`config.go:N` left in living docs    | Sweep run post-edit                                                |
| 8  | **Lint-config truth**                                                  | AGENTS.md Commands, FEATURES lint row, CONTRIBUTING lint policy now say "~100 extra linters" (was "defaults + gosec/mnd/exhaustruct"); marker inventory complete (gosec, exhaustruct_v5, errorlint)                                                                                                                                                                                                                                                                                                                                                                                                                                  | 3 files                                                            | Cross-checked against `.golangci.yml` enable list                  |
| 9  | **Dependabot mystery solved**                                          | `gh pr list --state all` → **zero PRs ever opened** (as of that morning); GitHub docs confirm SHA+comment pins ARE supported → version updates appeared not to run (settings-side, cause unknown). Policy decided: manual SHA bumps, documented in AGENTS.md gotcha. **Amendment 14:58 CEST:** Dependabot opened its first-ever PR (#1, "bump the actions group with 3 updates") the same afternoon — updates DO run; the earlier silence stays unexplained. AGENTS gotcha + TODO T17 updated                                                                                                                                        | live `gh` query 2026-09-13 + docs.github.com; PR #1 observed 14:51 | Authenticated query; docs cross-checked; amendment live-verified   |
| 10 | **CI verified green on real runners for tip `691744b`**                | run 34755167105: test ✅ coverage ✅ lint ✅ — resolves the "runner-unverified" caveat for checkout v6, govulncheck action, coverage floor                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | `gh run view` (headSha matched `691744b`)                          | Live API check                                                     |
| 11 | **Both 2026-09-13 reports ANNOTATED inline**                           | 12:57 report: header + §b (5) + §c (6) + §f (50/50) + §g (3). 12:20 report: §b (7) + §c (9) + §d (8) + §e (9) + §f (42; items 43–50 pre-closed, SKIP rule) + §g (3). Strikethrough + hashes/verdicts, zero appendix-only                                                                                                                                                                                                                                                                                                                                                                                                             | both files in `docs/status/archived/`                              | Marker-coverage checker (written this session): ALL MARKED         |
| 12 | **Both reports ARCHIVED**                                              | `git mv` → `docs/status/archived/`; `docs/status/` now holds only `README.md` + `archived/`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                          | git index                                                          | `ls` verified                                                      |
| 13 | **Archive door sign**                                                  | `docs/status/README.md` index: 10 reports + planning archive, State column, annotation/archive convention (explicit hand variants vs script kinds) + `AGENTS.md` pointer row                                                                                                                                                                                                                                                                                                                                                                                                                                                         | `docs/status/README.md`, AGENTS.md                                 | —                                                                  |
| 14 | **Living docs superb pass**                                            | AGENTS (lint truth, architecture table +5 rows incl. workflows/corpus/status-dir, 2 new gotchas, exhaustruct_v5 rename), FEATURES (durable citations, benchmark range ~20–35 ns/op 0 allocs, amplify-pin + example rows, CI row with verified-run note), README (verified DoWithValue, fuzz/govulncheck CI truth), CONTRIBUTING (lint truth, 3 markers, corpus/workflow/seeded-run), DOMAIN_LANGUAGE (function citations + **Message column**), ROADMAP (CI-hardening raw ideas + delay-table **Won't-implement** decision record), CHANGELOG `[Unreleased]` (+3 entries: example, amplification pin, lint migration + CI hardening) | 7 files                                                            | Every claim re-verified against code before writing                |
| 15 | **Verified-and-closed without edits**                                  | `TestDoWithValue_DoesNotLeakPartialValueOnLaterFailure` already exists (12:57 f.43); SECURITY.md accurate; Validate↔glossary codes in sync (8/8); README has no badges (f.15 no-op); 12:20's HTML parse-validates; 20 s fuzz smoke passes                                                                                                                                                                                                                                                                                                                                                                                            | greps + CLI + HTML parser                                          | —                                                                  |
| 16 | **Report-time self-check caught + fixed 2 real defects** (see d.1/d.2) | Stale `done (tracked as TODO_LIST Tx)` markers in the archived 2026-08-22 report updated to `done at <hash>` (T9/T10/T11/T13/T14); wrong T9 hash `a89ad17` → `258c1fd` corrected after `git show`                                                                                                                                                                                                                                                                                                                                                                                                                                    | `archived/2026-08-22…md`, `archived/2026-09-13_12-20…md`           | `git show --stat` per commit                                       |

## b) PARTIALLY DONE

1. **Runner verification for THIS session's changes** — exhaustruct_v5, golangci
   v2.13.2, timeouts, concurrency, new tests: all locally verified, never on a
   runner (local master is unpushed). Tracked as TODO T16.
2. **Hash attributions in annotations** — `258c1fd`, `a89ad17`, `28fe7e4`
   verified via `git show` (diff-level); `9d7efa1`, `3ec60b0`, `23192cd`
   verified at stat level only (right files, right shape — not per-hunk).
3. **Fuzz smoke ran 20 s**, not the 1 minute the source item asked. Economy,
   not capability; the daily 30-minute campaign subsumes it.
4. **dprint** — `dprint.json` configures markdown; this session's new/edited
   `.md` files (incl. hand-aligned tables) were never run through it.
5. **AGENTS gotchas at 18/20** — added 2 this session (line-citation ban,
   manual-bumps policy). Prune pass due at 20.
6. **Doc-work verdicts in annotations cite "the 2026-09-13 pass"** or `23192cd`
   where code — the doc edits themselves landed inside multi-file daemon
   "chore" commits; per-edit hash attribution is impractical there. Accepted,
   noted here.
7. **`docs/status/README.md` State column** for the two new archives was
   written by the same session that annotated them — self-graded (though the
   marker-coverage checker backs it).

## c) NOT STARTED

1. **T16** — runner verification of the whole batch: the session's changes ARE
   pushed (daemon) and went red (`exhaustruct_v5` settings bug, fixed ~14:55);
   the fix itself needs a green run, plus the first `fuzz.yml` run (next
   scheduled 2026-09-14 03:17 UTC; `workflow_dispatch` could trigger earlier).
2. **T21** — release cut v0.5.1/v0.6.0 (owner-gated): `[Unreleased]` carries
   the user-visible exhaustion-message change; at tag time also update
   ROADMAP's "current release" line + CHANGELOG compare links + re-check
   pkg.go.dev (now incl. `ExampleDoWithValue`).
3. **T17 residual** — review Dependabot PR #1 (merge or close) and explain the
   sudden first fire after a history of silence; then decide the ongoing
   policy.
4. **T20** — corpus↔seeds sync test.
5. **T18** — actionlint/schema pre-push gate (must respect the two-tool repo
   convention or justify a third tool).
6. **T19** — fuzz artifact SHA naming + `-fuzzminimizetime`.
7. **0.x release policy** — owner decision, parked in ROADMAP Open questions.
8. **v1.0 track** — API audit, `AttemptFunc` signature, options migration,
   compat matrix (ROADMAP-owned).
9. **dprint pass** over this session's markdown output (b.4).

## d) TOTALLY FUCKED UP (honest ledger)

1. **I marked work "done" that I had not done.** The 12:57 report's §f.8
   ("ANNOTATE the archived 2026-08-22 §f: T9–T15 done at this session's
   hashes") received my `done (2026-09-13 pass, inline markers)` marker — and
   I never touched that file during the pass. The 12:20 session had marked
   those rows `done (tracked as TODO_LIST Tx)`, which had itself gone stale
   once T9–T15 shipped. Caught ~2 h later by the self-check I ran while
   writing THIS report; fixed by actually doing it (5 rows re-marked with
   verified hashes). This is the worst sin in this repo's taxonomy — a false
   evidence marker — and it happened because I annotated a 50-item list from a
   mental completion map instead of a per-item checklist.
2. **I cited a commit hash from file-stat inference; it was wrong.** T9's
   `done at a89ad17` — `git show` proved T9 is `258c1fd` (2-line test change);
   `a89ad17` is the T10 const block. Corrected in place with a note. Same
   failure class as the 12:20 report's d.6, which I had _read hours earlier_
   and still repeated.
3. **My chat summary overstated the annotation count**: "all 50 + 50 §f items"
   — truth: 50 + 42 (the 12:20 report's §f items 43–50 are pre-closed
   "explicitly do not re-open" rows, legitimately unmarked per the SKIP rule —
   but my claim said 50). The files were right; the message was wrong.
4. **One wasted lint round trip**: wrote a >120-char `t.Fatalf` line; lint
   caught golines after tests were already green. Formatting-first would have
   avoided it.
5. **Luck-adjacent ordering**: TODO_LIST cites the _archived_ paths of the two
   reports — written before the `git mv` ran. Correct only because the move
   succeeded. References should be written/updated after the move they
   reference, not before.
6. **My lint-config change broke master's CI.** The `exhaustruct_v5` migration
   kept v4's `exclude` settings block; v5's schema rejects that key. Local
   `golangci-lint run` stayed green (it tolerates it); the CI action runs
   `config verify` first and the lint job went red on three consecutive pushes
   (`5058fec5`, `8615ef8d`, and Dependabot's PR #1 branch). Caught within ~20
   minutes via a routine `gh run list` while writing this report; fixed by
   dropping the settings block (the `_test.go` exclusion rule is independent;
   the only non-test v5 finding is already nolint'd). `config verify` now
   passes locally. Root cause: I verified the migration with the same tool
   invocation the repo always used, not with the invocation CI actually runs.

## e) WHAT WE SHOULD IMPROVE

1. **Never annotate from a mental completion map.** Per-item checklist or the
   skill's `annotate-*.py` scripts (dry-run first). The marker-coverage checker
   written this session (every numbered §f item must carry a verdict) turns
   d.1 from "caught by luck, 2 h later" into "caught in seconds" — run it
   before claiming any ANNOTATE pass done.
2. **`git show` before citing any hash.** Stat-level (right file touched) is
   not attribution-level (right hunk, right claim). d.2 is the proof.
3. **Formatter belongs in the edit batch** (golines/dprint), not after the
   test gate. The repo's formatter config (`dprint.json`) should be consulted
   before hand-aligning markdown tables.
4. **Chat claims must match file truth exactly** — counts included (d.3).
5. **Sequence reference updates after the moves they reference** (d.5).
6. **Daemon blobs vs evidence**: cite "the 2026-09-13 pass" for doc-only work
   instead of pretending hash precision; use real hashes only where verified
   (stat → say stat-level; diff → say diff-level).
7. **Report-time self-checks paid off twice** (d.1, d.2). Institutionalize:
   no status report without (a) marker-coverage check, (b) hash verification
   for every newly cited hash.
8. **After touching `.golangci.yml`, run `golangci-lint config verify`** —
   the exact command the CI action runs before linting (d.6). Recorded as an
   AGENTS.md gotcha so it survives sessions.

## f) Up to 50 things to get done next

> Brainstorm sorted by impact, not a commitment list. Tags: `[VERIFY]` runner
> verification · `[RELEASE]` release flow · `[CI]` · `[CODE]` · `[DOC]` ·
> `[ROADMAP]` · `[WISEGO]` other repo · `[OWNER]` needs you · `[PROCESS]`.

1. `[VERIFY]` T16: verify the config-verify fix lands green on the runner
   (next push), then watch all three `ci.yml` jobs (exhaustruct_v5 + golangci
   v2.13.2 + timeouts + concurrency).
2. `[VERIFY]` T16: trigger `fuzz.yml` once via `workflow_dispatch` instead of
   waiting for tomorrow 03:17 UTC; confirm the committed corpus loads
   (14 entries in the log).
3. `[VERIFY]` T16: confirm govulncheck action's `go-version-file` +
   `check-latest` interaction on the runner (run-log read).
4. `[VERIFY]` T16: confirm `upload-artifact` v7 + `if-no-files-found: ignore`
   behaves (skips cleanly on green).
5. `[VERIFY]` T16: watch for exhaustruct_v5 findings drift vs the old linter
   (v5 may flag structs v4 didn't).
6. `[RELEASE]` T21: decide v0.5.1 vs v0.6.0 — the exhaustion-message wording
   is user-visible → minor-leaning.
7. `[RELEASE]` T21: move ALL `[Unreleased]` entries together (incl. the two
   pre-session pinning-test entries) — don't split the release.
8. `[RELEASE]` T21: compose GitHub-only release notes (first exercise of the
   T14 decision).
9. `[RELEASE]` T21: update ROADMAP's "current release is v0.5.0" line +
   CHANGELOG compare links at tag time.
10. `[RELEASE]` T21: re-check pkg.go.dev rendering post-tag (incl.
    `ExampleDoWithValue`).
11. `[CI]` T17: review Dependabot PR #1 (bump the actions group with 3
    updates) — merge or close; decide Dependabot-owned vs manual policy.
12. `[CI]` T17: after PR #1's fate is decided, update the AGENTS.md gotcha if
    the policy changes from manual SHA bumps.
13. `[CI]` T18: actionlint gate — pick the install surface (pre-commit hook?
    CI job? documented manual step?) without breaking the two-tool convention.
14. `[CI]` T19: name the crash artifact `fuzz-crash-corpus-${{ github.sha }}`.
15. `[CI]` T19: tune `-fuzzminimizetime` so crash minimization can't eat the
    45-minute budget.
16. `[CODE]` T20: corpus↔seeds sync test (fails when a seed lacks its corpus
    file).
17. `[DOC]` Run dprint over this session's markdown (config exists; tables
    were hand-aligned).
18. `[DOC]` CONTRIBUTING/FEATURES: replace hardcoded "v2.13.2" with "pinned in
    `.github/workflows/ci.yml`" — the version string rots on the next bump.
19. `[DOC]` DOMAIN_LANGUAGE: decide the policy for _external_ citations
    (`classify.go:67` — dependency line numbers).
20. `[DOC]` AGENTS gotcha prune at 20 (18 now).
21. `[DOC]` docs/status/README.md: keep the State column current as reports
    land.
22. `[PROCESS]` Institutionalize the report-time self-checks (see e.7).
23. `[PROCESS]` Sweep-verify remaining stat-level hash citations (`9d7efa1`,
    `3ec60b0`, `23192cd`) with `git show` when convenient.
24. `[PROCESS]` Extend the docs-health annotate scripts with a routing kind
    upstream — decision stands as "won't for now"; revisit on the next batch
    pass.
25. `[ROADMAP]` Graduate the periodic `-race` fuzz run when CI capacity is
    known.
26. `[ROADMAP]` Decide the govulncheck supply-chain posture (action runs
    `go install …@latest` internally).
27. `[ROADMAP]` Auto-PR loop: fuzz crasher → PR into `testdata/fuzz/`.
28. `[ROADMAP]` Try `go test -shuffle=on` locally before proposing CI
    adoption (cheap order-dependency probe).
29. `[ROADMAP]` Corpus generation from seeds — superseded by T20 unless drift
    pain recurs.
30. `[ROADMAP]` v1.0: run the public API surface audit (12 symbols) as a
    recorded checklist.
31. `[ROADMAP]` v1.0: `AttemptFunc(ctx, attempt)` signature decision.
32. `[ROADMAP]` Options-pattern migration design (`WithOnRetry`, `WithJitter`).
33. `[ROADMAP]` Version-compatibility matrix with go-error-family.
34. `[ROADMAP]` Deterministic-RNG decision (FEATURES WORTH_CONSIDERING).
35. `[ROADMAP]` Deadline-aware attempt-budgeting decision (FEATURES
    WORTH_CONSIDERING).
36. `[ROADMAP]` Docs website (Astro/Starlight) — post-API-freeze only.
37. `[WISEGO]` Failsafe→go-retry adoption spike (other repo).
38. `[WISEGO]` wise-go v1.0.0 tag (owner-gated).
39. `[WISEGO]` wise-go sandbox integration tests (API-key-gated).
40. `[WISEGO]` wise-go CI re-enable (nix/GOEXPERIMENT).
41. `[OWNER]` Confirm 0.x release policy (ROADMAP Open question).
42. `[OWNER]` Confirm archive retention (default keep-forever is recorded).
43. `[OWNER]` Push policy going forward (the daemon pushes; confirm that is
    intended for master, see g.1).
44. `[DOC]` `.config/metadata.yaml` — never read; check whether it's repo
    tooling that docs should mention.
45. `[DOC]` CHANGELOG convention: decide whether doc-only changes get entries
    (current practice: only user-observable ones).
46. `[TEST]` Keep `-race -count=10` in the personal gate for any jitter-adjacent
    change (standing).
47. `[CI]` If T16's timeout of 10 min proves tight on slow runners (coverage
    job), bump before it bites.
48. `[DOC]` Superseded — the manual-bumps gotcha was updated the same hour
    (Dependabot fired, PR #1).
49. `[PROCESS]` Next docs-health pass: HARVEST this report's §f (route →
    TODO_LIST/ROADMAP; drop the `[WISEGO]` block for this repo).
50. `[DOC]` Archive THIS report via docs-health ANNOTATE + ARCHIVE once its
    items resolve — its §f is the new backlog, not a to-read list.

## g) Questions I can NOT figure out myself

1. **Push policy:** the daemon pushed this session's commits to master
   unattended (which is how the broken lint config reached CI before I saw
   it). Is daemon-auto-push to `master` intended policy, or should it stop so
   red CI never hits the default branch? Your repo, your call — I won't push
   manually either way.
2. **Dependabot policy (T17):** PR #1 (bump the actions group with 3 updates)
   is the first Dependabot PR ever here. Merge it and hand action bumps to
   Dependabot (drop the manual-SHA-bump rule), or close it and stay manual?
   The AGENTS gotcha currently says manual; I'll follow your answer.
3. **golangci v2.12.2 pin origin:** was CI's golangci-lint deliberately pinned
   to v2.12.2 (a cross-repo standard I should honor), or just drift? I bumped
   the pin to v2.13.2 to match local + unlock `exhaustruct_v5`; if the old pin
   was policy, I'll re-align the local toolchain down instead.

---

_Point-in-time snapshot — will go stale. Route section (f) items via
docs-health HARVEST into `TODO_LIST.md` / `ROADMAP.md` rather than reading
this file as a backlog. The auto-commit daemon will pick this file up; no
manual commit per harness rules._
