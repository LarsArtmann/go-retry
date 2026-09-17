# Status Report — go-retry — Docs-Health Pass 3: Annotation, Archive & Living-Doc Reseed

**Written:** 2026-09-16 16:55 CEST
**Scope:** the 2026-09-16 docs-health session — full AUDIT over all `2026-0*`
historical files, inline annotation + archive of the 2026-09-14 03:30 report,
and rebuild of all living docs (TODO_LIST, CHANGELOG-audit, AGENTS, README,
ROADMAP, FEATURES, CONTRIBUTING, status index).
**Baseline at session start:** master `7040b44` (03:30 report unarchived,
TODO_LIST all-resolved with "no open work"), tree clean.
**End state:** master `9eb87ee` (clean), TODO_LIST reseeded with T22–T33,
03:30 report fully annotated + archived (12 indexed reports), all gates green
— verified twice, see §d.2.
**Self-checks run before writing:** all cited hashes verified via
`git show --stat` (`b4efa02`, `7aeaae7`, `37f54a2`, `9eb87ee`); fuzz run IDs
read fresh from `gh run list`; pkg.go.dev fetched live; benchmark re-run;
unexpected `go.mod`/`go.sum` diff in `9eb87ee` read and judged (not reverted);
gates re-run green against the post-`9eb87ee` tree.
**Format note:** `.md` per explicit user instruction (skill default is a
styled HTML dashboard — override honored, not propagated).

---

## a) FULLY DONE

| #  | Item                                                         | What                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           | Evidence                                                       | Verification                                                                           |
| -- | ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| 1  | **All 14 `2026-0*` files read and annotation-audited**       | 1 unarchived status report (2026-09-14 03:30), 2 planning files, 11 archived files: every one opened or sweep-verified. The 11 archived files conform to the repo's 2026-09-13 annotation convention (the Pareto plan carries verdicts on 25/25 rows; the HTML report is Status-column annotated; unmarked numbered items live only in self-evidencing ledger/retro sections)                                                                                                                                                                                                                  | `docs/status/archived/`, `docs/planning/archived/`             | `grep -c '~~'` per file + targeted greps for `done at`/`done —`                        |
| 2  | **03:30 report fully annotated inline (110 markers)**        | §f.1–50: every item struck through with a verdict (`done at <hash>` / verified-with-evidence / `→ routed to TODO_LIST T#` / `→ routed to ROADMAP`); §g.1–3 routed to ROADMAP → Open questions; changed-state §b.2/§b.4/§b.5/§c.6 corrected in place; §b.1/§b.3/§c.1–5 left unmarked deliberately (still-accurate narrative per repo convention)                                                                                                                                                                                                                                                | `docs/status/archived/2026-09-14_03-30_…md`; marker grep = 110 | Post-annotation completeness sweep: zero unmarked actionable items                     |
| 3  | **03:30 report archived + indexed**                          | `git mv` to `docs/status/archived/` (no plain `mv`); `docs/status/README.md` index row added at top with State set; archive now 12 reports + 2 planning files                                                                                                                                                                                                                                                                                                                                                                                                                                  | `37f54a2`; `docs/status/README.md`                             | Index re-read; completeness gate (no marker-less archive) passes                       |
| 4  | **TODO_LIST rebuilt: resolved items deleted, harvest lands** | T16–T21 removed (their story lives in CHANGELOG + archives — no trophy case); reseeded from the 03:30 report §f as T22–T33 with per-item evidence and source refs: T22 options migration (P1), T23 consumer sweep, T24 fuzz crash drill, T25 next-cut decision (P2), T26–T33 polish + watchlist (P3). 1×P1 / 4×P2 / 8×P3, every item bounded and evidence-cited                                                                                                                                                                                                                                | `7aeaae7`; `TODO_LIST.md`                                      | Cross-checked each surviving item against code/CHANGELOG first                         |
| 5  | **AGENTS: two new load-bearing gotchas (18 → 20, at cap)**   | (1) actionlint validates structure, not remote-action _inputs_ (the `namee:` class — broken-but-green master window); (2) setup-go's version manifest lags `go.dev` — `GOTOOLCHAIN` pin recipe. The setup-go claim was verified against the go-release skill source BEFORE writing (verify-external-claims)                                                                                                                                                                                                                                                                                    | `b4efa02`; `AGENTS.md` → Gotchas                               | Gotcha count re-counted: exactly 20 (documented cap)                                   |
| 6  | **FEATURES: CI row + testing guarantees current**            | CI row now includes `-shuffle=on`, the actionlint gate, and runner evidence refreshed to run 3479724601 (2026-09-14); new guarantee row for `TestFuzzCorpusMirrorsSeeds`; fuzz row notes schedule-trigger runs verified green 2026-09-14 → 2026-09-16                                                                                                                                                                                                                                                                                                                                          | `b4efa02`; `FEATURES.md`                                       | Fresh `gh run list` + bench re-run (row 9)                                             |
| 7  | **README: deadline-budget recipe + dev-command parity**      | New "Deadline budgets" section (the M19 recipe: set `MaxDelay` below remaining budget, count-based semantics untouched); Development block gains the seeded-corpus run + actionlint one-liner; CI paragraph mentions `-shuffle=on` and the schema gate                                                                                                                                                                                                                                                                                                                                         | `b4efa02`; `README.md`                                         | Rendered content re-read post-edit                                                     |
| 8  | **ROADMAP: skeleton, owner questions, CI ideas**             | 20-line `Option` type skeleton written into the options-design record (mutator shape, unexported fields, apply-then-validate; §f.25); THREE owner questions added (§g.2 third-local-tool, §g.3 consumer-bump permission, §f.42 metadata.yaml); CI-ideas bundle (input-allowlist test, concurrency-group shape, release-workflow decision, verified setup-go note); archive-retention count 11 → 12                                                                                                                                                                                             | `7aeaae7`; `ROADMAP.md`                                        | Stitch-verified after edit (caught one duplication — see §d.1)                         |
| 9  | **Verification sweep — everything re-proven, not assumed**   | Gates green: `gofmt -l` clean, `go vet` clean, `go test ./... -race` pass, coverage **100.0%**, golangci-lint **0 issues**; scheduled fuzz green ×3 (runs 34802993739 / 34925264051 / 35052069172, 2026-09-14→16); pkg.go.dev v0.6.0 renders live with all five examples; benchmark 20.1–21.1 ns/op, **0 allocs** (FEATURES' "~20–35 ns/op" holds); DOMAIN_LANGUAGE codes 8/8 match source; 7/7 CHANGELOG compare links resolve; internal md links resolve; mermaid fences in the archived plan intact; §f.31/§f.32 resolved by evidence (`modernize` + `gofumpt`/`golines` are gate-enforced) | gate outputs; `gh run list`; live fetch; bench output          | Each claim executed this session, none copied from prior reports                       |
| 10 | **CONTRIBUTING: formatting truth-pass**                      | New Formatting section: Go formatting is lint-enforced (`gci`, `goimports`, `gofumpt`, `golines` — max-len **120**, not the folk-claimed 128), dprint owns markdown/JSON/YAML/Dockerfile (CHANGELOG excluded), tabs/2-space per `.editorconfig`                                                                                                                                                                                                                                                                                                                                                | `b4efa02`; `CONTRIBUTING.md`; `.golangci.yml` (`formatters`)   | Read from `.golangci.yml` before writing — premise of §f.33 was wrong and is corrected |
| 11 | **Session lifecycle closed**                                 | Four daemon commits landed and verified: `b4efa02` (AGENTS/CONTRIBUTING/FEATURES/README), `7aeaae7` (ROADMAP/TODO_LIST/report annotations), `37f54a2` (archive mv + index), `9eb87ee` (external go.mod bump + formatter pass — see §d.2); tree clean at write time                                                                                                                                                                                                                                                                                                                             | `git show --stat` ×4; `git status`                             | Every hash cited here was read from `git log`, never from memory                       |

## b) PARTIALLY DONE

1. ~~**Canonical markdown formatting is unverified by tool.** `dprint` is not on
   PATH (nix-only per the 03:30 report) and I stopped at one `which` — yet
   `9eb87ee` contains emphasis rewrites (`*input key*` → `_input key_`) in my
   new AGENTS text, proving SOME formatter automation runs somewhere. Which
   tool, invoked how, is unknown; this session's markdown was hand-checked
   only.~~**— done** (M2, `f65ee77`): dprint identified, documented, gated.
2. ~~**ROADMAP compat matrix + DOMAIN_LANGUAGE version citation are already
   behind.** Both cite go-error-family `v0.10.0`; master now consumes
   `v0.10.1` (`9eb87ee`). The matrix's own rule accepts within-v0.x patch
   bumps "ad hoc", so no minor bump is required — but the row/note update is
   pending. My "DOMAIN_LANGUAGE verified" claim from mid-session was true for
   ~10 minutes.~~**— done** (M1, `047f075`): matrix + version cite updated after the v0.10.1 bump.
3. ~~**pkg.go.dev "Latest" banner oddity.** The live versioned fetch rendered
   v0.6.0 completely (5/5 examples) but displayed "This package is not in the
   latest version of its module" — likely a versioned-URL quirk (the 03:30
   report verified the badge via its own self-checks on 09-14). Unresolved;
   canonical-URL re-check queued (§f.19).~~**— done** (M3.8/M12.7): canonical-URL check recorded in the release ritual.
4. ~~**Annotation hash attribution is coarse.** Verdicts for ROADMAP/TODO_LIST
   work cite "this pass" + TODO IDs rather than per-edit hashes — the daemon
   batches multi-file commits, making per-edit attribution impractical (same
   accepted limitation as the 14:48 pass, noted there as §b.6).~~**— accepted limitation**: verdicts cite the pass or a batched commit; per-edit attribution stays impractical (standing repo convention).
5. ~~**§b.1/§b.3/§c.1–5 of the archived report remain unmarked.** Deliberate
   (still-accurate narrative), but their routing pointers (T22/T23/T24) live
   only in TODO_LIST, not inline where a skimming reader would see them.~~**— done** (this pass): the archived report's remaining §b/§c items are resolved inline.
6. ~~**GitHub-side rendering unverified.** New tables (TODO_LIST harvest, status
   index row, ROADMAP additions) and the fenced Go skeleton were verified as
   text only, not as rendered GitHub markdown.~~**— done** (M12.5): GitHub-render check ran in the doc-hygiene bundle.

## c) NOT STARTED

1. ~~**T22 — Options-pattern migration** (design + skeleton exist; zero code).~~**— done** (M4–M8, v0.7.0).
2. ~~**T23 — Consumer sweep** to v0.6.0+ (owner-gated; live pkg.go.dev still
   shows "Imported by: 0", confirming the premise).~~**— done** (M9; `go-cqrs-lite` 680f2d4d0).
3. ~~**T24 — Fuzz crash drill** (failure path still never executed).~~**— done** (M10; run 35142442193).
4. ~~**T25 — Next-cut decision** (`[Unreleased]` holds 4 entries, no queue).~~**— done** (M3/M8: v0.6.1 then v0.7.0).
5. ~~**T26–T33** — the P3 tail (ci.yml workflow_dispatch, Backoff/ComputeDelay
   godoc examples, compare-link guard, link-rot sweep, local govulncheck,
   release-notes skeleton, coverage-floor decision, verification watchlist).~~**— done** (M10–M12, M15–M16; see `CHANGELOG.md` [0.6.1]/[0.7.0]).
6. ~~**CHANGELOG `[Unreleased]` entry for the README deadline-budgets section**
   — CONTRIBUTING's policy says README changes that alter documented guidance
   get an entry; found unfinished at report time (§d.3).~~**— done** (M1.1, `047f075`).
7. ~~**ROADMAP matrix / DOMAIN_LANGUAGE v0.10.1 note** (§b.2).~~**— done** (M1, `047f075`).
8. ~~**Next docs-health pass** to annotate THIS report once §f resolves.~~**— done** (this pass).
9. ~~**Docs website** — post-v1.0 preconditions unchanged.~~**— Open**: docs website stays post-v1.0 (`ROADMAP.md`).

## d) TOTALLY FUCKED UP

1. **I repeated the edit-hygiene mistake the previous session flagged.** My
   ROADMAP multiedit used an anchor that ended mid-line ("…keeps the manual")
   and my replacement re-included the continuation, duplicating `path cheap.`
   — the exact "multiedit left broken intermediate states" class from the
   03:30 report's §d.4. My post-edit stitch check caught it within one step,
   but the correct behavior was a full-line anchor, not a catch-and-fix.
2. **My verification raced external automation and lost.** I verified the
   gates and doc claims against v0.10.0 / `go 1.26.7`; minutes later
   `9eb87ee` — authored by the daemon but NOT by me — bumped go-error-family
   to v0.10.1, relaxed `go 1.26.7` → `go 1.26`, and reformat-rewrote my new
   AGENTS text. My "verified fresh" claims decayed before the pass even
   ended. What went right: I read the unexpected diff and judged it instead
   of reverting, and re-ran the full gates green against the new tree before
   claiming anything. What went wrong: I nearly wrote the report on stale
   evidence — end-of-pass re-verification must happen AFTER the daemon
   settles, not before.
3. **I forgot CONTRIBUTING's own CHANGELOG policy while editing CONTRIBUTING
   and README.** The deadline-budgets section alters documented guidance →
   per the policy it earns a `[Unreleased]` entry. I wrote the policy's
   subject matter into the repo and skipped its consequence. Found during
   this report's self-review, not during the pass.
4. **One `which dprint` and I gave up.** The 03:30 session ran dprint 0.57.4
   "via nix"; I never tried `nix run` or located the profile binary, so the
   canonical markdown gate stayed unexecuted — and then a mystery formatter
   edit landed anyway, proving the gap is real and lived-in.
5. **I hand-annotated 50 items against the report's own instruction.** §e.6/§f.26
   of the source report said to use the docs-health `annotate-*.py` scripts.
   My deviation (heterogeneous routed verdicts) is sanctioned by the repo's
   hand-marker convention, but I skipped the dry-run cross-check that would
   have validated marker placement mechanically.

## e) WHAT WE SHOULD IMPROVE

1. **Treat "clean" as a snapshot, not a state:** run `git status` immediately
   before every final claim and before writing any report — the daemon and
   external bump tooling commit between commands.
2. **Full-line anchors, always.** Truncating a multiedit anchor mid-line is
   how §d.1 happened; the anchor budget of "a few extra lines" is cheaper
   than every stitch-check save.
3. **Re-run the cheapest gate at the very END of a pass,** after the daemon
   settles — mid-pass green does not certify end-of-pass state (§d.2).
4. **Locate and document the canonical markdown formatter first,** then edit
   markdown, then run it — not hand-format now, discover the tool later.
5. **Use `annotate-*.py --dry-run` as a cross-check even for hand-written
   annotations** — calibrated placement validation for free.
6. **Stop hardcoding dependency patch versions in prose docs** (DOMAIN_LANGUAGE's
   `v0.10.0`): cite the symbol and "see `go.mod`" — same lesson as the
   line-number ban; patch pins rot the same way.
7. **Map §f-item → T-number in the report footer** so the next ANNOTATE pass
   is a mechanical grep, not archaeology.
8. **One extra call on surprising evidence:** the pkg.go.dev banner oddity
   deserved a canonical-URL re-fetch before being logged unresolved (§b.3).

## f) Up to 50 things to get done next

> Brainstorm sorted by impact, not a commitment list. Tags: `[VERIFY]` ·
> `[RELEASE]` · `[CI]` · `[CODE]` · `[DOC]` · `[ROADMAP]` · `[OWNER]` ·
> `[PROCESS]`.

1. ~~`[DOC]` Add the `[Unreleased]` CHANGELOG entry for the README
   deadline-budgets section (CONTRIBUTING policy; §d.3).~~**— done** (M1.1, `047f075`).
2. ~~`[VERIFY]` Identify the markdown formatter that rewrote emphasis in
   `9eb87ee`; document its canonical invocation in CONTRIBUTING; run it over
   this session's files (§b.1, §d.4).~~**— done** (M2, `f65ee77`).
3. ~~`[VERIFY]` Watch the CI run for `9eb87ee` (pending at write time; only
   `37f54a2` had a green push run when checked).~~**— done** (CI green through `000a479`).
4. ~~`[DOC]` Update ROADMAP's compat matrix for go-error-family v0.10.1 (patch
   bump, "accepted ad hoc" per the matrix's own rule — add the note/row).~~**— done** (M1.2, `047f075`).
5. ~~`[DOC]` Fix DOMAIN_LANGUAGE's hardcoded `v0.10.0` → symbol + "see go.mod"
   (§e.6).~~**— done** (M1.3, `047f075`).
6. ~~`[VERIFY]` Decide whether `go 1.26.7` → `go 1.26` was intentional (§g.3);
   if float-patch is the policy, note it; if not, restore the pin.~~**— done** (resolved 2026-09-17: reverted to `go 1.26`; guard test `e2eb9f1`).
7. ~~`[CODE]` T22 — options migration per the ROADMAP skeleton (WithIsRetryable,
   WithDelayFunc, WithOnRetry, WithExhausted, WithJitter, WithRandomSource).~~**— done** (M4–M8, v0.7.0).
8. ~~`[CODE]` T23 — consumer sweep: bump `go-cqrs-lite/middleware/v4` (owner-
   gated via ROADMAP → Open questions; kills the "Imported by: 0" weakness).~~**— done** (M9).
9. ~~`[VERIFY]` T24 — fuzz crash drill (SHA-named artifact + minimization
   budget, first execution of the failure path).~~**— done** (M10).
10. ~~`[RELEASE]` T25 — decide v0.6.1 vs v0.7.0 ([Unreleased] = 4 entries).~~**— done** (M3/M8).
11. ~~`[CI]` T26 — `workflow_dispatch` on ci.yml.~~**— done** (M10.1, `75a4816`).
12. ~~`[DOC]` T27 — godoc examples for `Backoff` and `ComputeDelay`.~~**— done** (M11, `4c48d13`).
13. ~~`[DOC]` T28 — CHANGELOG compare-link guard (script or test).~~**— done** (M12.1, `d11703b`).
14. ~~`[DOC]` T29 — link-rot sweep over README/CONTRIBUTING external links.~~**— done** (M12.3, `d11703b`).
15. ~~`[RELEASE]` T30 — local `govulncheck` in the pre-release ritual.~~**— done** (M3.5).
16. ~~`[DOC]` T31 — capture the v0.6.0 release-notes skeleton as template.~~**— done** (M3.9, `6469515`).
17. ~~`[CODE]` T32 — coverage-floor 95→99 decision.~~**— done** (M15.1, `f7f1bfb`).
18. ~~`[VERIFY]` T33 — watchlist (Dependabot rebase behavior; concurrency-group
    cancel; `go mod tidy` no-diff at next tag).~~**— done** (M3.10/M10.8/M10.9; the Dependabot tail stays open → TODO_LIST T42).
19. ~~`[VERIFY]` Re-check pkg.go.dev's "Latest" badge via the canonical URL
    (§b.3).~~**— done** (M3.8).
20. ~~`[VERIFY]` GitHub-render check of this session's new tables/fences (§b.6).~~**— done** (M12.5).
21. ~~`[PROCESS]` Next docs-health pass: annotate THIS report (§f-item → T-number
    mapping is in TODO_LIST source refs).~~**— done** (this pass).
22. ~~`[DOC]` AGENTS gotcha cap reached (20/20) — prune pass due before the next
    addition.~~**— done** (M13.3, `f7f1bfb`; reviewed again this pass).
23. ~~`[OWNER]` Daemon push policy — third consecutive report; now it also
    carries external dependency bumps unattended (§g.1).~~**— done** (answered (owner 2026-09-16)).
24. ~~`[OWNER]` Markdown-formatter canonicalization (§g.2).~~**— done** (answered — dprint).
25. ~~`[OWNER]` Consumer-bump proactive vs on-request (§g.3, gates T23).~~**— done** (answered — owner-authorized).
26. ~~`[OWNER]` Archive retention (12 files, keep-forever vs pruning).~~**— done** (answered — keep forever).
27. ~~`[OWNER]` 0.x full-release confirmation (v0.6.0 = third data point).~~**— done** (answered — full releases for 0.x).
28. ~~`[OWNER]` `.config/metadata.yaml` keep-or-remove.~~**— done** (answered — keep).
29. ~~`[OWNER]` Third-local-tool / `tools.go` blessing (actionlint gate is
    CI-only today).~~**— done** (answered — blessed → TODO_LIST T35).
30. `[CI]` Remote-action input-allowlist test (the `namee:` class actionlint
    cannot see; ROADMAP CI ideas). → TODO_LIST T36
31. `[CI]` Concurrency-group shape across tag ref and master (ROADMAP CI
    ideas). Open (→ ROADMAP CI ideas)
32. `[CI]` Release-workflow vs manual `go-release` skill decision (ROADMAP CI
    ideas). Open (→ ROADMAP CI ideas)
33. ~~`[DOC]` Consider a FEATURES row for the markdown formatter once identified
    (tool inventory honesty).~~**— done** (FEATURES dprint row).
34. ~~`[DOC]` AGENTS "error-family Dependency" section: confirm surface wording
    survives v0.10.1 (no API change observed — gates green — but say so once).~~**— done** (M1.4).
35. ~~`[PROCESS]` Inline routing pointers for the archived report's remaining
    unmarked §b/§c items IF a future pass judges them load-bearing (§b.5).~~**— done** (this pass).
36. ~~`[ROADMAP]` Revisit "Imported by: 0" after T23 — strengthens the v1.0
    freeze claim.~~**— done** (M9).
37. ~~`[CODE]` With `WithRandomSource` (T22): simplify
    `TestBackoff_IncreasesExponentially` to sampled assertions.~~**— done** (M7).
38. ~~`[DOC]` README: consider linking the ROADMAP skeleton from the options
    discussion once T22 starts (keep one source of truth).~~**— done** (M8).
39. ~~`[VERIFY]` Confirm `exhaustruct_v5` markers still map after any future
    linter renames (standing marker-coverage ritual; no trigger this session).~~**— done** (M16.1, `f7f1bfb`).
40. `[PROCESS]` Split report-time annotation batches per section AND verify
    daemon commits land between batches where feasible (bisectability; §f.49
    of the archived report, partially followed this session). Open — process suggestion
41. ~~`[DOC]` Add "who bumps go-error-family and when" to the ROADMAP matrix
    note — the v0.10.1 bump arrived with no in-repo trace of the decider.~~**— done** (M1.2).
42. ~~`[CI]` Evaluate whether the fuzz workflow should also run on the weekly
    cadence for go-error-family bumps (supply-chain-triggered fuzzing; idea,
    unpriced).~~**— done** (M15.3 — deferred with pricing).
43. ~~`[DOC]` CONTRIBUTING: note that `_emphasis_` (not `*emphasis*`) is the
    formatter-enforced style, once §f.2 identifies the tool.~~**— done** (M2.3, `f65ee77`).
44. ~~`[VERIFY]` Re-run the seeded fuzz corpus run after the v0.10.1 bump
    (`go test -run '^FuzzComputeDelayNeverPanics$' .`) — cheap freshness
    proof for the new dependency.~~**— done** (M1.6).
45. ~~`[DOC]` FEATURES bench row: numbers re-verified this session (20.1–21.1
    ns/op, 0 allocs) — no action needed; re-check at next hardware/toolchain
    change.~~**— done** (M16.4, `f7f1bfb`).
46. ~~`[PROCESS]` When harvesting, capture the §f→T mapping as a table in the
    TODO_LIST header comment (mechanical reverse-navigation for ANNOTATE).~~**— done** (superseded — TODO_LIST rebuilt with a Source column).
47. ~~`[ROADMAP]` If T22 lands, delete the "Options-based configuration" raw
    idea and graduate its decisions into FEATURES rows.~~**— done** (M8.3).
48. ~~`[VERIFY]` Spot-check that `9eb87ee`'s formatter pass missed nothing in
    TODO_LIST (it touched ROADMAP-adjacent files; TODO_LIST landed in
    `7aeaae7`, before the formatter ran — possible drift).~~**— done** (M1.5).
49. ~~`[DOC]` Add the coverage-command output note ("(statements) 100.0%") to
    the Session Ritual's gate order as the canonical coverage check format.~~**— done** (M12.6).
50. ~~`[PROCESS]` Before the next status report: poll `git log` for daemon
    commits made while writing, and cite final hashes, not mid-session ones
    (this report cites four; a fifth may land before it is read).~~**— done** (standing practice).

## g) Questions I can NOT figure out myself

1. ~~**Daemon push policy (third ask, now with supply-chain teeth):** today the
   daemon carried my doc blobs AND an external go-error-family v0.10.1 bump +
   `go`-directive relaxation to `master` unattended, with no in-repo trace of
   who decided the bump. Is unattended auto-push to `master` — including
   dependency changes nobody reviewed in-repo — the intended posture, or
   should the daemon stop pushing (or push to a side branch)?~~**— done**: answered 2026-09-16 — the daemon keeps committing and pushing; agent commits amend over daemon races only while unpushed.
2. ~~**Which markdown formatter is canonical?** Something rewrote `*x*` → `_x_`
   across my AGENTS text in `9eb87ee`, but `dprint` is not on PATH and no
   config in-repo names the tool that does emphasis normalization. Which tool
   owns markdown formatting here, how is it invoked, and should CONTRIBUTING
   pin that invocation so agents stop guessing (and hand-formatting wrong)?~~**— done**: answered — dprint is the canonical markdown formatter (M2, `f65ee77`); `_emphasis_` documented.
3. ~~**Was `go 1.26.7` → `go 1.26` intentional?** The patch pin was removed in
   the same commit as the v0.10.1 bump. If floating patch toolchains is the
   policy, the AGENTS setup-go gotcha should say the repo deliberately floats;
   if it was collateral, the pin should be restored. Which is it?~~**— done**: resolved 2026-09-17 — the directive was reverted to `go 1.26` and is now guarded by `TestModuleGoDirectiveStaysPinned`.

---

_Point-in-time snapshot — will go stale. Route section (f) items via
docs-health HARVEST into `TODO_LIST.md` / `ROADMAP.md` rather than reading
this file as a backlog (T22–T33 already live in TODO_LIST; owner questions
already live in ROADMAP). All hashes, run IDs, and gate outputs in this
report were verified against fresh CLI output at write time._
