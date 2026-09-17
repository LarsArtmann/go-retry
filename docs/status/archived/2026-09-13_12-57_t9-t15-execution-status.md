# Status Report — go-retry — TODO T9–T15 Execution & Brutal Self-Review

**When:** 2026-09-13 12:57 CEST
**Scope of this session:** Execute `TODO_LIST.md` items T9–T15 (the 2026-09-13
harvest), end to end: code, CI, corpus, decision records, docs, verification.
**End state:** TODO_LIST.md empty. All T-items shipped. Nothing broken found by
the final gate (`go build`, `go vet`, `go test -race`, `go test -race
-count=10`, `golangci-lint run` → 0 issues). ~~CI changes are **not yet verified
on real GitHub runners** (no push happened).~~ Superseded same day: the tree was
pushed and CI ran **green on real runners** for the tip `691744b` — run
34755167105, all three jobs, 2026-09-13. `fuzz.yml` still has zero runs
(`TODO_LIST.md` T16).

---

## a) FULLY DONE

| Item                                                          | What                                                                                                                                                                                                                                                                                                                                               | Evidence                                   | Verification                                                                                           |
| ------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------ | ------------------------------------------------------------------------------------------------------ |
| **T9** — non-retryable error pinned by identity               | `TestDo_DoesNotRetryNonRetryableError` now asserts `err != rejection`, not just `errors.Is`, so `Do` can never silently start re-wrapping a typed non-retryable error. Added a deliberate `//nolint:errorlint` with reason (repo convention: explained, enabled linter).                                                                           | `retry_test.go:156-158`                    | Test passes with `-race`; lint 0 issues                                                                |
| **T10** — terminal-error codes/messages single-sourced        | New const block (`codeExhausted`/`msgExhausted`, `codeCanceled`/`msgCanceled`, `codeDeadline`/`msgDeadline`) at `retry.go:14-24`; used by all three sentinels AND their `WrapInfrastructure` call sites (`retry.go:118-119`, `retry.go:205-211`).                                                                                                  | `retry.go`                                 | Build/vet/tests/lint pass                                                                              |
| **T10+ (found beyond the TODO)** — pre-existing message drift | The exhaustion wrapper said `"all attempts failed"` while the `ErrExhausted` sentinel says `"all retry attempts failed"` — the exact drift class T10 targets, at a third site the TODO didn't cite. Unified on the sentinel wording; updated the `ExampleDo_delayFunc` `// Output:` pin accordingly.                                               | `retry.go:16`, `retry_test.go:1109`        | Example passes                                                                                         |
| **T13** — committed fuzz corpus                               | 7 `go test fuzz v1` files in `testdata/fuzz/FuzzComputeDelayNeverPanics/`, one per `f.Add` seed, descriptively named (`exponential-basic`, `maxint64-saturation`, `near-max-fractional-multiplier`, `negative-multiplier`, `negative-initial`, `negative-attempt`, `uncapped-maxdelay`).                                                           | `testdata/fuzz/`                           | `go test -run '^FuzzComputeDelayNeverPanics$' -v` → all 14 entries (7 seeds + 7 corpus) parse and pass |
| **T14** — release-notes decision                              | **GitHub-only.** Bodies composed at release time from the CHANGELOG section (matches the `go-release` skill Phase 2.4/7 flow). No `docs/releases/` mirror — CHANGELOG stays the single in-repo copy. Recorded in ROADMAP (decided, kept for record), AGENTS.md gotcha, CHANGELOG.                                                                  | `ROADMAP.md` → Open questions, `AGENTS.md` | n/a (decision)                                                                                         |
| **Docs maintenance**                                          | CHANGELOG `[Unreleased]` gained 5 Added + 2 Changed entries; TODO_LIST emptied per its own convention (done items live in CHANGELOG only); FEATURES CI/fuzz rows updated; AGENTS.md gotchas extended (single-sourced constants, corpus↔seeds sync rule, fuzz workflow, GitHub-only notes, nolint inventory now includes the new errorlint marker). | 5 doc files, auto-committed as `3ec60b0`   | Docs-health ownership model respected (no fact in two places)                                          |
| **Drift fix found during self-review**                        | `docs/DOMAIN_LANGUAGE.md` code table cited `retry.go:28/36/47/230`; my const block shifted the file — refs corrected to `39/47/58/241`.                                                                                                                                                                                                            | `docs/DOMAIN_LANGUAGE.md:84-91`            | Verified against live `retry.go`                                                                       |

**Verification gate (whole session):** `go build ./...` ✅ · `go vet ./...` ✅ ·
`go test ./... -race` ✅ · `go test ./... -race -count=10` ✅ (final summary
line inspected, not just exit code) · `golangci-lint run ./...` → 0 issues ✅ ·
both workflow YAMLs parse ✅.

## b) PARTIALLY DONE

1. **T11 — checkout v4 → v6.** Applied to all three jobs
   (`ci.yml:14/27/48`), pinned `d23441a4…` = `v6` tag = v6.1.0, verified via
   `git ls-remote` + the tag's `action.yml` (`using: node24` — deprecated-Node
   rationale confirmed). **Partial because:** never executed on a real
   Actions runner (no push). ~~YAML-parse-level confidence only.~~ →
   Runner-verified 2026-09-13: run 34755167105 green for `691744b`.
2. **T15 — govulncheck in CI.** Step added to the `test` job
   (`ci.yml:24-28`): `golang/govulncheck-action@032d455… # v1` (v1.1.0,
   verified), with `repo-checkout: false` (the action checks out on its own —
   read from its `action.yml`, this input is easy to miss) and
   `go-version-file: go.mod`. **Locally verified:** `go run
   golang.org/x/vuln/cmd/govulncheck@latest ./...` → "No vulnerabilities
   found." **Partial because:** the action's behavior on the runner (its
   internal checkout/setup-go dance, `check-latest` interaction with
   `go-version-file`) is unverified until the first real run. ~~the action's
   behavior on the runner … unverified until the first real run.~~ →
   Runner-verified 2026-09-13: the `test` job (incl. the govulncheck step)
   passed in run 34755167105.
3. **T12 — scheduled fuzz workflow.** `.github/workflows/fuzz.yml` written:
   daily 03:17 UTC (off-peak minute), `workflow_dispatch` escape hatch,
   concurrency-cancel, 30-min campaign, crash-corpus artifact upload on
   failure (`upload-artifact@043fb46… # v7` — verified latest major).
   Command verified locally: 10 s campaign, 2.9M execs, 0 failures.
   **Partial because:** ~~the workflow itself has never run; scheduled
   workflows also only trigger after push to the default branch.~~ → Still
   true that it has zero runs (first scheduled fire pending) — routed to
   `TODO_LIST.md` (T16).
4. **Lint-config description.** Partially corrected: AGENTS.md nolint
   inventory updated (3 markers), but the "enables gosec/mnd/exhaustruct +
   defaults" phrasing survives at `AGENTS.md:26` and `FEATURES.md:143-144` —
   the config actually enables ≈70 linters (I read the list when fighting
   errorlint). Understatement, not a lie; ~~unfixed.~~ → Fixed in the
   same-day docs-health pass (AGENTS.md Commands line, FEATURES lint row,
   CONTRIBUTING lint policy — all now state the ~100-linter truth).
5. **TODO_LIST harvest loop.** TODO_LIST is empty; this report's section (f)
   is the new backlog source. The docs-health HARVEST pass (route (f) items
   into TODO_LIST/ROADMAP) is **pending, awaiting your instruction** — you
   said "wait for instructions". → Executed in the same-day docs-health pass:
   §f routed into `TODO_LIST.md` (T16–T21) and `ROADMAP.md`.

## c) NOT STARTED

- ~~**Real-runner verification of all CI changes** (blocked on push decision).~~
  → ci.yml-side done (run 34755167105 green for `691744b`); the newest batch
  (exhaustruct_v5, golangci v2.13.2, timeouts, concurrency) + first fuzz run
  routed to `TODO_LIST.md` (T16).
- ~~**`exhaustruct` → `exhaustruct_v5` migration** — every `golangci-lint run`
  this session printed the deprecation warning ("Replaced by
  exhaustruct_v5"). Pre-existing; out of session scope; still open.~~ → Done
  in the same-day docs-health pass (`.golangci.yml`, `config.go` marker, CI
  golangci pinned v2.13.2; lint 0 issues, warning gone).
- **Dependabot behavior investigation** — `.github/dependabot.yml` has a
  weekly `github-actions` entry; I did not know it existed until the
  self-review pass. Why checkout sat stale on v4 despite a weekly watcher is
  unexplained (closed PRs? SHA+comment pin not matched?). → Investigated
  2026-09-13: **zero PRs have ever been opened** (`gh pr list --state all` is
  empty), and GitHub docs confirm SHA+comment pins ARE supported — so version
  updates are simply not running (settings-side, cause unknown). Residual
  routed to `TODO_LIST.md` (T17); manual-SHA-bump policy documented in
  `AGENTS.md`.
- ~~**ANNOTATE the harvest source** — `docs/status/archived/2026-08-22_…v0.4.0…`
  §f items T9–T15 are now all done; the archived report carries no `done at`
  markers yet (docs-health ANNOTATE mode).~~ → Done in the same-day
  docs-health pass (inline markers).
- **Release cut** — CHANGELOG `[Unreleased]` has real, user-visible content
  (the exhaustion-message change is observable in error strings). No version
  bump/tag was requested or performed. → Routed to `TODO_LIST.md`
  (T21, owner-gated).
- **ROADMAP open question** — "0.x: full release or prerelease?" still needs
  a human decision (pre-dates this session; blocks the `go-release` skill
  default). → Still open by design (`ROADMAP.md` → Open questions; owner
  decision).

## d) TOTALLY FUCKED UP

Honest verdict: **nothing is destroyed, broken, or reverted.** No data loss,
no broken builds, no fabricated claims (every action SHA/version/flag in this
session was verified against `git ls-remote` / `action.yml` / a local run
before being encoded — per the verify-external-claims gate). But three real
self-inflicted near-misses, owned here:

1. **My drift sweep had a hole.** After T10 I grepped `--type go` for the
   message strings — which **cannot** see README.md, docs/, or examples
   outside Go files. The post-hoc full-repo sweep found nothing drifted in
   README (good luck, not good process), but the review pass _did_ find
   `DOMAIN_LANGUAGE.md` carrying stale `retry.go` line refs my const block
   created — a drift **I introduced and missed in-session**. Lesson recorded:
   drift sweeps must cover all file types, and any line-number citation is
   invalidated by inserting lines above it.
2. **I never checked `.github/dependabot.yml`** while planning CI changes. I
   changed action pins blind to the repo's own update automation. Consequence
   is probably nil (my bump agrees with what Dependabot would want), but the
   blind spot is real — and it hides an unresolved puzzle (why was checkout
   stale?).
3. **I noticed the lint-config understatement** (FEATURES.md/AGENTS.md
   describing ~70 enabled linters as "defaults + 3") **during the session** —
   when reading `.golangci.yml` — and left it because it was pre-existing.
   Borderline call: the global rule says trivial doc fixes happen on sight.

## e) WHAT WE SHOULD IMPROVE

**Process (mine):**

- Drift sweeps: always repo-wide, all file types; grep for _identifiers and
  line numbers_, not just strings.
- Before touching `.github/`, inventory the whole `.github/` tree
  (workflows + dependabot + templates) first.
- Read a tool's `action.yml` before using it — the `repo-checkout: false`
  and `go-version-file` inputs were only discovered that way, and both matter.
- Verify by _running_ what CI will run, locally, before writing the workflow
  (done for fuzz + govulncheck — this worked well; keep it).

**Codebase (small, observed):**

- `AGENTS.md:26` / `FEATURES.md:143` lint-config description → say what the
  config actually enables.
- Corpus mirrors `f.Add` seeds by hand — a deliberate duplication with a
  documented sync rule, but a test (or generator) would remove the human
  discipline requirement.
- The deprecation-warning noise (`exhaustruct`) trains us to ignore lint
  stderr — kill it before it masks something real.

**Docs model (working well — keep):** single-ownership docs (ROADMAP for
decisions, AGENTS for operational rules, CHANGELOG for history, TODO_LIST for
open work) held up; the T14 decision landed in four files with zero overlap of
_content_, only of _reference_.

## f) Up to 50 things to get done next

> Brainstorm sorted by impact, not a commitment list. Tags: `[VERIFY]` do
> before/with next push · `[DOC]` docs hygiene · `[CI]` CI/tooling · `[CODE]`
> library code · `[ROADMAP]` graduate via docs-health HARVEST · `[PROCESS]`
> ways of working.

1. ~~`[VERIFY]` Push, watch all three `ci.yml` jobs + first scheduled `fuzz.yml`
   run green (checkout v6, govulncheck action, upload-artifact v7 are all
   runner-unverified).~~ → ci.yml half done (run 34755167105 green for
   `691744b`); fuzz.yml + newest ci.yml changes → `TODO_LIST.md` (T16).
2. ~~`[VERIFY]` Confirm `golang/govulncheck-action` respects `go-version-file`
   together with its `check-latest: true` default on the runner.~~ → Folded
   into `TODO_LIST.md` (T16); the step itself already passed on the runner
   (run 34755167105).
3. ~~`[VERIFY]` Confirm the fuzz workflow's baseline loads the committed corpus
   (runner log should show 14 corpus entries, not 37-from-cache).~~ → Folded
   into `TODO_LIST.md` (T16).
4. ~~`[CI]` Investigate why Dependabot's weekly `github-actions` watcher left
   checkout on v4 (closed PRs? SHA+`# v4` comment pins unsupported?).~~ → done
   (verified 2026-09-13: zero PRs ever, `gh pr list --state all` empty; GitHub
   docs confirm SHA+comment pins are supported → updates not running at all,
   settings-side). Residual → `TODO_LIST.md` (T17).
5. ~~`[CI]` Decide the action-update policy: Dependabot-owned vs manual SHA
   bumps; document the winner in AGENTS.md.~~ → done (2026-09-13 pass):
   manual SHA bumps, documented in `AGENTS.md` gotcha.
6. ~~`[CI]` Migrate `exhaustruct` → `exhaustruct_v5` in `.golangci.yml` (silence
   the deprecation warning every run prints).~~ → done at `23192cd`
   (`.golangci.yml` enable/settings/exclusions, `config.go` marker, CI
   golangci pinned v2.13.2; lint 0 issues, warning gone).
7. ~~`[DOC]` Fix `FEATURES.md:143` + `AGENTS.md:26`: describe `.golangci.yml`
   accurately (≈70 enabled linters, not "defaults + gosec/mnd/exhaustruct").~~
   → done (2026-09-13 pass; AGENTS.md, FEATURES.md, CONTRIBUTING.md).
8. ~~`[DOC]` ANNOTATE the archived 2026-08-22 report §f: T9–T15 `done at` with
   this session's hashes (docs-health ANNOTATE, inline, not appendix).~~ →
   done (2026-09-13 pass, inline markers).
9. ~~`[CI]` Add `timeout-minutes` to the three `ci.yml` jobs (fuzz job has one;
   the others run unbounded).~~ → done at `23192cd` (10 min per job).
10. ~~`[CI]` Add `actionlint` (or `gh workflow` schema check) as a pre-push gate
    so workflow YAML errors die locally, not on the runner.~~ → routed to
    `TODO_LIST.md` (T18).
11. ~~`[ROADMAP]` CODE→TEST: a sync test asserting every `f.Add` seed has a
    matching corpus file (kills the manual-mirror discipline).~~ → graduated
    to `TODO_LIST.md` (T20).
12. ~~`[ROADMAP]` CODE: generate corpus files from seeds (go:generate or test
    helper) instead of hand-mirroring.~~ → routed to `ROADMAP.md` raw ideas
    (T20's sync test removes the immediate pain).
13. ~~`[ROADMAP]` CI: schedule a periodic `-race` fuzz short-run (throughput vs
    concurrency-bug tradeoff, even monthly).~~ → routed to `ROADMAP.md`
    (CI hardening ideas).
14. ~~`[ROADMAP]` CI: decide govulncheck supply-chain posture — the action does
    `go install …@latest` internally; pin if that offends you.~~ → routed to
    `ROADMAP.md` (CI hardening ideas).
15. ~~`[DOC]` Check README for CI badges — do they reference workflow names that
    still exist (`CI` badge vs new `Fuzz` workflow — badge or drop)?~~ → done
    (verified 2026-09-13: README carries no badges at all; adding badges stays
    Won't-implement per the 2026-09-13 12:20 report f.50 decision).
16. ~~`[DOC]` SECURITY.md: verify supported-versions text is still accurate
    post-0.5.0 (not checked this session).~~ → done (verified 2026-09-13:
    "latest release on `master`" remains accurate).
17. ~~`[DOC]` CONTRIBUTING.md: mention the fuzz workflow, corpus convention, and
    `go test -run '^Fuzz…$'` seeded-run command (not checked this session).~~ →
    done (2026-09-13 pass: corpus + workflow + seeded-run all in CONTRIBUTING).
18. ~~`[DOC]` DOMAIN_LANGUAGE.md: consider adding the terminal _messages_ next
    to codes, now that messages are part of the single-sourced contract.~~ →
    done (2026-09-13 pass: Message column added to the code table).
19. ~~`[CODE]` Cut v0.5.1 or v0.6.0: `[Unreleased]` holds a user-visible change
    (exhaustion error message wording) — decide patch vs minor per the
    error-family convention.~~ → routed to `TODO_LIST.md` (T21, owner-gated).
20. ~~`[ROADMAP]` At release time: compose the GitHub-only release notes per the
    T14 decision (first real exercise of the new convention).~~ → folded into
    `TODO_LIST.md` (T21).
21. ~~`[ROADMAP]` Resolve the 0.x full-release-vs-prerelease open question (see
    questions below) so the `go-release` default stops fighting practice.~~ →
    stays `ROADMAP.md` → Open questions (owner decision, still open).
22. ~~`[PROCESS]` Add to global memory/AGENTS lessons: "drift sweeps cover all
    file types; line-number citations die on insertion."~~ → done (2026-09-13
    pass: root-caused instead — living docs no longer cite line numbers at
    all; AGENTS.md gotcha added).
23. ~~`[PROCESS]` Post-REVIEW rule: after any edit that inserts/moves lines,
    re-grep all `*.md` for `retry.go:\d+` style citations.~~ → done
    (superseded: the citations no longer exist to rot).
24. ~~`[CI]` Consider concurrency group for `ci.yml` (cancel superseded pushes)
    — fuzz.yml has one, ci.yml doesn't.~~ → done at `28fe7e4`
    (`ci-${{ github.ref }}`, cancel-in-progress).
25. ~~`[ROADMAP]` Re-run `BenchmarkComputeDelay` after the const extraction;
    update the FEATURES ns/op figure if drifted (FEATURES claims ≈32 ns/op).~~
    → done (2026-09-13: 21.3 / 34.9 ns/op across two runs — machine-noise
    band, 0 allocs; FEATURES now states the range, killing the recurring
    re-check).
26. ~~`[ROADMAP]` `go test -shuffle=on` in CI or the 10× flake pass — cheap
    order-dependency detector for the suite.~~ → routed to `ROADMAP.md`
    (CI hardening ideas).
27. ~~`[ROADMAP]` FEATURES WORTH_CONSIDERING → decide: deterministic RNG option
    (pluggable `rand` source); currently testable-formula workaround exists.~~
    → done (already tracked: `FEATURES.md` WORTH_CONSIDERING).
28. ~~`[ROADMAP]` FEATURES WORTH_CONSIDERING → decide: deadline-aware attempt
    budgeting (stop retrying when remaining context budget < next delay).~~ →
    done (already tracked: `FEATURES.md` WORTH_CONSIDERING).
29. ~~`[ROADMAP]` ROADMAP raw idea: options-based configuration migration
    (`WithOnRetry`, `WithJitter`) — large; needs a concrete migration story.~~
    → done (already tracked: `ROADMAP.md` raw ideas).
30. ~~`[ROADMAP]` ROADMAP raw idea: version-compatibility matrix with
    go-error-family majors.~~ → done (already tracked: `ROADMAP.md`).
31. ~~`[ROADMAP]` ROADMAP raw idea: public docs site (Astro/Starlight) — gated
    on v1.0 API stability, keep parked.~~ → done (already tracked:
    `ROADMAP.md`).
32. ~~`[ROADMAP]` ROADMAP v1.0 bar: run the public API surface audit (exported
    symbols + Config fields list) as a recorded checklist.~~ → done (already
    tracked: `ROADMAP.md` v1.0 section).
33. ~~`[ROADMAP]` ROADMAP v1.0 bar: decide `AttemptFunc(ctx, attempt)` vs
    passing the previous error — deliberate decision, not accident.~~ → done
    (already tracked: `ROADMAP.md` v1.0 section).
34. ~~`[ROADMAP]` FEATURES WORTH_CONSIDERING: configurable jitter strategy —
    stays deferred per the twice-made decision; only revisit via options
    pattern (do NOT re-litigate standalone).~~ → done (decision stands:
    `AGENTS.md` gotcha + `FEATURES.md`).
35. ~~`[ROADMAP]` ROADMAP raw idea: `OnSuccess(attempts)` hook — stays deferred
    until a concrete consumer exists.~~ → done (decision stands:
    `ROADMAP.md`).
36. ~~`[ROADMAP]` ROADMAP raw idea: composition primitives (breaker/bulkhead)
    — lean stays "pure loop, document the pattern".~~ → done (already
    tracked: `ROADMAP.md`).
37. ~~`[CODE]` Consider a `Config.Validate` split-brain audit: validate
    messages ↔ DOMAIN_LANGUAGE table codes are in sync (codes unchanged this
    session, but nothing enforces it).~~ → done (verified 2026-09-13: all 8
    codes match between `Validate`/`ComputeDelay` and the DOMAIN_LANGUAGE
    table, which now also carries the messages).
38. ~~`[CI]` Dependabot gomod group: confirm the weekly group PRs don't bump
    `go-error-family` across majors silently (gomod grouping is
    minor-and-patch only — majors arrive as separate PRs; fine, just watch).~~
    → folded into `TODO_LIST.md` (T17) — moot while Dependabot demonstrably
    never fires.
39. ~~`[DOC]` AGENTS.md architecture table: add rows for
    `.github/workflows/{ci,fuzz}.yml` + `testdata/fuzz/` so the file map is
    complete.~~ → done (2026-09-13 pass, plus `.golangci.yml` and
    `docs/status/` rows).
40. ~~`[CI]` Fuzz job: consider `-fuzzminimizetime` tuning so crash minimization
    doesn't eat the 45-minute timeout on a hit.~~ → routed to `TODO_LIST.md`
    (T19).
41. ~~`[CI]` Fuzz job: name the crash artifact with the run date/SHA
    (`fuzz-crash-corpus-${{ github.sha }}`) for multi-crash archaeology.~~ →
    routed to `TODO_LIST.md` (T19).
42. ~~`[ROADMAP]` Auto-PR workflow idea: crasher found by scheduled fuzz →
    auto-PR adding it to `testdata/fuzz/` (closes the discovery→corpus loop).~~
    → routed to `ROADMAP.md` (CI hardening ideas).
43. ~~`[CODE]` `DoWithValue`: zero-value return on failure is pinned where? If
    only implicitly, add a test pinning "never leaks a partial value from an
    earlier attempt" (AGENTS.md claims it; I didn't verify the pin this
    session).~~ → done (verified 2026-09-13: the pin already exists —
    `TestDoWithValue_DoesNotLeakPartialValueOnLaterFailure`, `retry_test.go`).
44. ~~`[DOC]` README: verify the "Exhaustion and nesting" section still matches
    post-T10 error strings (sweep found no message quote in README — verified;
    this item is to _keep_ it that way in the release-notes template).~~ →
    done (re-verified 2026-09-13: README quotes no error message; the keep-it-
    that-way note lives in `TODO_LIST.md` T21's release-notes step).
45. ~~`[CI]` Set `permissions: contents: read` top-level in fuzz.yml — done;
    apply the same top-level pattern review to any future workflow.~~ → done
    (confirmed in place: `fuzz.yml` top-level `permissions: contents: read`).
46. ~~`[PROCESS]` Keep the "verify CI commands locally before encoding" habit
    (fuzz CLI + govulncheck both ran clean first — this saved the workflow
    from a flag typo).~~ → done (standing practice; exercised again in the
    same-day docs-health pass).
47. ~~`[PROCESS]` When a TODO cites `file:line` evidence, re-verify the line
    before the edit, not just the file (T9/T10 evidence lines had drifted by
    the time I read them — they were still correct this time, by luck of
    ordering).~~ → done (superseded by the function-name-citation convention,
    `AGENTS.md` gotcha).
48. ~~`[ROADMAP]` Go toolchain pin (1.26.7): define the bump policy — dependabot
    gomod won't touch the `go` directive; manual, per minor.~~ → done
    (2026-09-13 pass: policy recorded in the `AGENTS.md` manual-bumps gotcha).
49. ~~`[DOC]` CHANGELOG: when cutting the release, move the two pre-session
    `[Unreleased]` pins (nested fail-closed, OnExhausted-on-context-end)
    along with this session's entries — don't split the release.~~ → folded
    into `TODO_LIST.md` (T21).
50. ~~`[PROCESS]` Session retro: skills-first flow (how-to-golang →
    verify-external-claims → go-release → docs-health) caught three
    would-be-mistakes (govulncheck-as-policy, /tmp-notes precedent,
    delete-done-from-TODO_LIST). Keep the lazy per-phase loading.~~ → done
    (standing practice, kept).

## g) Questions I cannot figure out myself

1. ~~**Push now, or bundle CI verification with the next release?** All
   workflow changes are YAML- and locally-verified but never ran on real
   runners. Pushing to `master` triggers CI immediately (and schedules fuzz
   from tomorrow). Your repo, your push policy — I won't push without an
   explicit ask.~~ → Overtaken by events same day: the tree was pushed (auto
   daemon) and CI ran green for `691744b` (run 34755167105); remaining fuzz
   verification is `TODO_LIST.md` (T16).
2. ~~**Who owns action bumps — you or Dependabot?** `dependabot.yml` watches
   `github-actions` weekly, yet checkout sat stale on v4. Did you close its
   PRs, or do SHA+`# vX` pins silently not match? I can't distinguish "you
   rejected it" from "it never fired" without archaeology you may just want
   to answer.~~ → Answered 2026-09-13: neither — **zero PRs have ever been
   opened** (`gh pr list --state all` empty); updates never fire at all.
   Policy decided: manual SHA bumps (`AGENTS.md`); residual → `TODO_LIST.md`
   (T17).
3. **The 0.x release question (pre-existing ROADMAP open question):** full
   releases (demonstrated practice: v0.4.0, v0.5.0) or prereleases
   (`go-release` skill default)? This decision unblocks the release flow
   default and only you can make it. → Still open by design
   (`ROADMAP.md` → Open questions).

---

_Point-in-time snapshot — will go stale. Route section (f) items via
docs-health HARVEST into `TODO_LIST.md` / `ROADMAP.md` rather than reading
this file as a backlog._
