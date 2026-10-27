# Status 2026-10-27 22:10 — docs-health AUDIT: six reports annotated and archived, backlog rebuilt, BuildFlow config gone

Session scope: "View ALL `2026-0*` files, execute the docs-health skill properly" —
which became a full AUDIT (VERIFY + HARVEST + ANNOTATE + ARCHIVE) over every
non-archived status report, the unarchived planning file, and all six living
docs. The two non-archived `2026-0*` reports (09-17 20:32, 09-18 05:41) were
the named targets; the four newer ones (10-08, 10-09 ×2, 10-27 20:43) were
pulled in by the live-index rot and the never-harvested 20:43 §f. A
concurrent session (owner-driven BuildFlow-config removal) landed mid-pass;
its state was reconciled, never reverted.

Baseline at start: HEAD `6791e1d`, tree clean, master synced with origin,
CI green (runs 40571159852 / 40572338572), latest tag v0.8.0.
End state at write time: HEAD `ff196eb`, this report plus the audit's
renames/edits uncommitted (daemon will sweep); `docs/status/` holds only
`archived/` (23 files) + `README.md`; `docs/planning/archived/` holds four
plans; `./scripts/check-docs.sh` all green; full battery green (below).

## a) FULLY DONE

1. **All six open status reports annotated inline and archived.** Every
   gated §b/§c/§f/§g item carries a verdict (strike + `done at`/`resolved`,
   `→ routed to TODO_LIST/ROADMAP`, `**Won't implement**`), every table row
   carries a column or cell verdict, using the docs-health `annotate.go
   repair` tool (dry-run first; 46/54/22/30/27 specs per file). The two
   2026-0* reports the session was ordered to view are among them.
2. **The 2026-10-08 planning file annotated (all 27 M-rows, per-outcome) and
   archived** — `docs/planning/archived/` now holds four plans, all fully
   annotated; the status-index planning note updated (three → four files).
3. **The status index rebuilt with a bulk-archive manifest** — six new rows
   (newest-first) with classification and deciding reason per file; the
   live-index-rot Critical is closed and now gate-enforced by
   `TestStatusIndexCoversArchive` (23 archived files ↔ 23 rows).
4. **HARVEST executed with verify-before-routing: TODO_LIST rebuilt.**
   T42/T51–T54/T60 kept (premises re-verified open); 18 new rows T77–T94
   routed from the six reports' §f tails — headline: T77 (P1: cut v0.9.0 +
   consumer sweep), T78 (P2: red-master alarm), T79–T81 (docs-gate gaps),
   T85 (file the two BuildFlow issues), T88–T94 (small bounded polish).
5. **ROADMAP extended:** three new owner questions (go-pin endgame,
   SECURITY.md response-time numbers, per-push flake net), the lychee
   scheduled-CI idea, and — superseded by the daemon's own better
   daemon-guard update that landed mid-pass, which was kept.
6. **Living docs synced to verified reality:** README (nightly stress
   workflow in the CI paragraph, the doubled-terminal-message render note,
   the unpromoted-`OnExhausted` table row), FEATURES (fuzz verified through
   2026-10-27, "v0.8.1" → next cut v0.9.0, stress workflow row), CHANGELOG
   (pin-restore entry moved Changed → Fixed per the 20:43 report §f14).
7. **AGENTS.md rebalanced under its budget with the one-in-one-out rule
   written into the header** — exactly 220 lines at the pass's midpoint,
   20 gotcha rows (cap honored): the `atomic.Int32` note moved to Testing
   Patterns (it is a testing pattern), the setup-go gotcha condensed with
   its narrative kept in `docs/engineering-notes.md`, the session-start
   `git status --branch` ritual folded into the hash-verification bullet.
8. **docs/status/README.md gained the two adopted archive-edit rules**
   (09-18 report §f17/§f20): format-only repairs, never content; new verdict
   phrasings are fixed by consciously extending the recognizer, never by
   rewording the archive.
9. **Two report items proven done on the spot instead of routed:** the
   done-at citation sampling (three 05:53-report citations — `047f075`,
   `f65ee77`, `a0a3968` — all resolve via `git cat-file`) and the v0.7.1
   release-notes render check (`gh release view`: notes live, both code
   fences present).
10. **Three live facts established and recorded in
    `docs/engineering-notes.md` → "Small verified facts":** the go-cqrs-lite
    consumer-bump hash (`065b2e9c5`, closing the 02:14 report's §b5/§f13
    ledger gap); `reports/` is the gitignored local coverage output
    (resolving 20:43 §f15 and the plan's M14); the nightly stress workflow
    was the red-master alarm — its 10-22 → 10-27 failures were the fossilized
    red master (runs 39724195287…40451635911), master went green with the
    pin restore, first post-fix run due 2026-10-28 04:33 UTC.
11. **The `.buildflow.yml` incident chain handled to the owner's landing
    state:** restored the file from HEAD when the working tree lost it at
    ≈21:02 (second same-day deletion; guards re-verified green), then — when
    the concurrent session committed the owner decision (no repo-level
    BuildFlow config; guards removed with the file; art-dupl honors the
    directives standalone) — completed the `git rm` and re-synced this
    session's own now-false texts (engineering-notes fact bullet, TODO_LIST
    T94 premise, the 20:43 report's §f2 verdict → superseded).
12. **Full gate battery green at close:** `gofmt -l` clean; `go vet` root +
    `tools/` clean; `go test ./... -race -count=10` ok; `golangci-lint run`
    0 issues; `golangci-lint config verify` clean; coverage **100.0% of
    statements** (canonical `go test -cover` line); `./scripts/check-docs.sh`
    all green (archive verdict gate, index↔archive, strike rendering, dprint
    after one `fmt` re-pad of 9 files, compare-links).

## b) PARTIALLY DONE

1. **Concurrency reconciliation.** A second session landed owner-driven
   changes mid-pass (the `.buildflow.yml` removal, guard deletions,
   AGENTS/FEATURES/CHANGELOG/CONTRIBUTING rewrites). This session synced its
   own outputs to the landed state but did not touch their in-flight work;
   their CHANGELOG entry references upstream links "in ROADMAP.md" that are
   not there yet, and `docs/drafts/` holds their verify-before-filing
   BuildFlow issue drafts — both are theirs to finish.
2. **Claim-level verification of the 17 pre-existing archived reports was
   deliberately skipped** per the docs-health canon (closed history):
   verified mechanically (every archived `.md` carries `~~` verdict markers;
   the HTML one exempt by design) and structurally (the archive gates pass),
   but no factual re-sweep of their contents.
3. **The AGENTS budget holds by one edit.** 219 lines / 20 gotcha rows at
   close: the one-in-one-out rule exists, but the next addition without a
   fold still trips the doctor cap.
4. **The go-pin endgame question is routed, not answered** (ROADMAP owner
   question): hold `go 1.26` behind the guards + restore ritual, or
   deliberately re-pin to `go 1.27` with the full doc sweep. The BuildFlow
   config removal changes the cost side of this (see §g3).
5. **The nightly stress verification is pending by construction:** the
   first post-restore run fires 2026-10-28 04:33 UTC; predicted green, not
   yet observed.

## c) NOT STARTED

1. **Filing the two BuildFlow upstream issues (T85)** — the concurrent
   session's drafts sit in `docs/drafts/`; nothing filed by this session.
2. **T77: the v0.9.0 cut + go-cqrs-lite middleware consumer sweep** — issue
   #3's code promotion is still unreleased; jitter examples still unrendered
   on pkg.go.dev; go-paperless still waiting.
3. **T78: the red-master alarm** — 12 days of unread red (and this session's
   discovery that the nightly stress failures were read by nobody) stand as
   evidence; no watchdog exists.
4. **T79–T94** — the remaining harvested backlog (docs-gate scanner
   self-tests, planning-dir marker gating, compare-links CI step,
   structure-linter guard, benchmark re-run, suppression removal,
   go-paperless docs, examples, errorfamilytest adoption, micro-benchmarks,
   harness fold, consumer-docs sweep, dead-directive experiment,
   PreCanceled comment, drift loudness).
5. **The fourth docs-health point** — the drift-rate trend this session's
   inline report feeds; not scheduled.

## d) TOTALLY FUCKED UP

1. **An annotate.go spec off-by-one misplaced two verdicts into the wrong
   rows.** My §b/§c line numbers for the 10-08 report were computed one row
   early (line 37 was the table header), so b5's verdict landed on b2 and
   c5's on c2. The tool's action-line guard refused the header spec ("45 of
   46") — the only reason it surfaced — but I applied the batch before
   reading the dry-run diff, then hand-corrected the two rows and marked the
   right ones. Verifying placement in the dry-run, not the applied file, is
   the missing step.
2. **38 prose verdicts were appended mid-item** (repair writes to the spec's
   physical line; my spec lines were the items' first lines), rendering
   verdicts mid-sentence. Caught by post-batch spot-check, fixed with a
   placement fixer that moves each verdict to the item's last line. Same
   root cause as (1): the dry-run was checked for count, not shape.
3. **The ≈21:02 `.buildflow.yml` restore fought the owner's decision for
   its last forty minutes.** Restoring was defensible on the visible state
   (the deletion matched the 20:43 report's documented malfunction and the
   guards went green after restore), but the second same-day deletion was
   already evidence of drift toward removal, not a repeated glitch — and the
   concurrent session had to land the removal around my restore.
4. **Two wasted multiedit rounds on ROADMAP** (mod-time refusals after daemon
   touches) before re-reading — and the re-read then revealed the daemon had
   committed a better daemon-guard update than the one I was about to apply.
   Skipping my version was right; discovering it by refusal instead of by
   checking was luck, not process.
5. **dprint formatting was left to battery time:** nine files sat unformatted
   between my table edits and the gate. Nothing bad swept them into history
   this time, but the window was the 09-18 report's red-master class.
6. **The ritual "re-check `git status` immediately before writes" was
   skipped for most batches** — the two refusals above are the visible cost;
   the invisible cost is not knowing how often the tree moved under me.

## e) WHAT WE SHOULD IMPROVE

1. **Dry-run inspection must diff shape, not count.** "Would annotate N
   lines" hides off-by-ones and mid-item placement; the pre-write check
   should be the rendered diff (or emit-keys round-trip), every time.
2. **A second occurrence of the same anomaly is signal drift, not repeat
   malfunction.** The `.buildflow.yml` deletion looked identical twice;
   the second one was the owner's decision germinating. When an anomaly
   recurs, look for the intent before restoring a third time.
3. **Detect concurrent sessions early and reconcile before batch edits.**
   Unknown commits landing under a running pass (caught here ≈90 minutes in,
   via `git status` naming files I never touched) should pause the pass, not
   race it.
4. **`dprint fmt` belongs immediately after markdown-table edits**, not at
   battery time — same lesson as the 09-18 report's lint-first rule, one
   tool over.
5. **Workflow-health belongs in the doc-freshness check.** This pass found
   the 12-day red master only because verifying FEATURES' fuzz claim
   required `gh run list`. A "recent workflow runs green?" step (or the T78
   watchdog) closes the gap the nightly stress failures exposed.
6. **Hand over annotation conventions to the tool's guards fully:** the
   off-by-one survived precisely because line-number specs are hand-built;
   emit-keys-based spec generation (the tool's own command) would have made
   the header collision impossible.

## f) Up to 50 things we should get done next

Ranked; the first block is the open TODO_LIST verbatim (T42, T51–T54, T60,
T77–T94 — 24 rows, all evidence-cited in `TODO_LIST.md`).

1. **T77 (P1)** — cut v0.9.0 (govulncheck root + `tools/`, CHANGELOG cut,
   annotated tag, GitHub-only notes) and run the consumer sweep: bump
   go-cqrs-lite `middleware/` to v0.9.0, its suite + `commandlifecycle` +
   `integration` + `example/taskmanager` with `-race`, ROADMAP bump-trace row.
2. **T78 (P2)** — CI red-master alarm: required checks, merge queue, or a
   watchdog workflow that pings on non-green master.
3. **T85 (P2)** — file the two BuildFlow issues (verify-before-filing
   first): the daemon-style file-deletion miscount and the dry-mode run that
   wrote `.golangci.yml`; the concurrent session's `docs/drafts/` already
   covers the go-mod-update pin-blind half.
4. **T79 (P2)** — table-driven self-tests for the docs-gate scanner edges.
5. **T80 (P2)** — extend marker gating to `docs/planning/archived/*.md`
   (the four plans now stand outside `TestArchivedReportItemsCarryVerdicts`).
6. **T81 (P2)** — `check-compare-links.sh` as a CI step (lint job, after
   actionlint).
7. **T82 (P3)** — guard test for `.go-structure-linter.yaml` (the
   go-version suppression has no enforcement of its own).
8. **T83 (P3)** — re-run `BenchmarkComputeDelay` on the current toolchain;
   update FEATURES numbers if they moved.
9. **T84 (P3)** — remove the `.go-structure-linter.yaml` go-version
   suppression once the deployed structure-linter is ≥ v0.11.3.
10. **T86 (P3)** — go-paperless: retire the `errors.AsType` workaround in
    its consumer docs (the promotion is the fix it waited for).
11. **T87 (P3)** — `ExampleDoWithValue` error-surface example (promoted code
    on failure).
12. **T88 (P3)** — adopt `errorfamilytest.AssertCode`/`AssertFamily` in the
    promotion tests.
13. **T89 (P3)** — benchmark the extra terminal-error allocation.
14. **T90 (P3)** — stress harness: fold the DoWithValue check into the same
    loop (halves harness cost).
15. **T91 (P3)** — go-cqrs-lite docs sweep for the stale go.work claim.
16. **T92 (P3)** — the dead-directive experiment, recorded in
    engineering-notes.
17. **T93 (P3)** — one-line comment on `TestDo_PreCanceledContextReturnsErrCanceled`
    (why hand-rolled values, not `longBackoffConfig()`).
18. **T94 (P3)** — make clone-hash drift loud (directive-line guard or
    strict art-dupl invocation); re-decide warnings-don't-gate in the same
    change.
19. **T42 (P3)** — Dependabot watchlist: next actions-group PR, post-merge
    auto-rebase, the `/tools` bump flow end-to-end.
20. **T51 (P3)** — safe probe fixtures for guard drills.
21. **T52 (P3)** — index/backlog conventions (planning State column,
    Verdict/Status format, planning README).
22. **T53 (P3)** — sweep archived plans for dangling `done at` hashes.
23. **T54 (P3)** — go-directive guard's docs-match half (living docs' stated
    Go version vs `go.mod`).
24. **T60 (P3)** — rebuild/reinstall the BuildFlow binary; doctor.
25. Confirm the nightly stress run green on 2026-10-28 04:33 UTC (first
    post-fix execution) and record it in FEATURES' stress row.
26. Finish the concurrent session's tails: the upstream links in ROADMAP
    that the CHANGELOG entry promises, and the two issue drafts in
    `docs/drafts/` (fold into T85 when filing).
27. Decide the go-pin endgame (ROADMAP owner question) — now re-priced by
    the BuildFlow config removal (see §g3).
28. Decide the SECURITY.md response-time numbers (ROADMAP owner question);
    edit or keep the committed sentence.
29. Decide the per-push flake net (`-count=2`/`3` vs nightly-only; ROADMAP
    owner question).
30. Fourth docs-health point (trend): rerun this audit shape and compute
    the drift rate against today's inline scores (Accuracy 7.0, Fitness
    7.75 as-found).
31. Re-verify AGENTS stays at ≤220 lines / ≤20 gotcha rows on the next
    gotcha (one-in-one-out, now written into the header).
32. Consider a `docs/status/` convention for concurrent-session collisions
    (two sessions annotating the same report within an hour is now a real
    case; the 20:43 report's §b2 verdict was rewritten twice today).
33. When T80 lands, backfill the four planning files into the marker gate
    and re-run the battery (`-count=1` — data files are not in Go's test
    cache key).
34. Sample three more archived `done at` citations at the next archive pass
    (standing drill; this pass sampled the 05:53 report).
35. Check go-cqrs-lite's own docs for the stale go.work claim before its
    next release (pairs with T91).
36. Keep CHANGELOG `[Unreleased]` fed as work lands (currently carries the
    code promotion, the pin restore, and the BuildFlow-config removal).
37. After v0.9.0: confirm the corpus-mirror guard, the jitter examples, and
    the promotion examples render on the pkg.go.dev page (standing
    post-release ritual).
38. Decide the fleet lychee policy (ROADMAP owner question; this repo's
    exclusion is committed and guarded).
39. Decide the daemon pre-commit guard (ROADMAP owner question; three-plus
    incident classes of evidence).
40. Decide the tools-pin topology (ROADMAP owner question; gates the
    tools-lint CI half).
41. Decide the dprint local pinning posture (ROADMAP owner question; the
    float-vs-CI divergence is live).
42. Decide the daemon commit-message policy (ROADMAP owner question;
    `done at` citations still point at heuristic batches).
43. Watch the next Dependabot actions-group PR: verify upload-artifact
    stays at the commit-verified v7.0.2 SHA; re-key the allowlist in the
    same change.
44. After the next crush-config deploy, re-run the skill fan-out guard
    (46 symlinks clean as of 2026-10-09's check).
45. Consider a scheduled CI job for the lychee link scan (ROADMAP idea).
46. Sweep the remaining 8 `context.WithCancel` test sites for dupl findings
    (golangci dupl is currently silent at 150 tokens — confirmed this pass).
47. Consider whether `reports/` (gitignored coverage output) should gain a
    one-line mention in CONTRIBUTING's coverage section (it is now
    documented in engineering-notes only).
48. Re-check the ROADMAP "Doc-freshness scheduled CI job" raw idea once T79
    and T80 land (its premise — gates worth scheduling — grows teeth).
49. Fold the stress harness DoWithValue loop before the next harness edit
    (T90) so the v0.9.0 cut's tests do not pay double cost.
50. Revisit the doubled-message render cost after consumer feedback lands
    on v0.9.0 (the README now warns; the optional shorter promoted-message
    constant stays deferred pending owner taste).

## g) QUESTIONS I CAN **NOT** FIGURE OUT MYSELF

1. **Concurrent-session protocol:** the BuildFlow-config removal session ran
   alongside this pass for roughly its second half (its commits landed
   under my edits; my ROADMAP multiedit nearly overwrote its daemon-guard
   update). Is that session yours, still running, and do you want doc-health
   passes to stand down or checkpoint-and-wait when foreign commits land
   mid-pass — or is optimistic merging like today's acceptable?
2. **The BuildFlow issues (T85) and the drafts:** your session left
   verify-before-filing drafts in `docs/drafts/` (go-mod-update pin-blind,
   plus a second body). File them upstream as-is, review first, or hold?
   Related: with repo-level BuildFlow config gone, are art-dupl's six
   warnings (the three accepted clone groups) accepted noise until upstream
   flips the default, or do you want the repo to pin/strict them (T94)?
3. **Go-pin endgame, re-priced:** with `.buildflow.yml` gone, nothing stops
   `buildflow update`/`--fix` from re-losing the `go 1.26` pin until the
   upstream fix lands — the guards trip and a human/agent restores. Is the
   restore-ritual posture the intended steady state, or does the removal
   tip the balance toward the deliberate `go 1.27` re-pin + full doc sweep
   (the ROADMAP question added today)?

---

_Point-in-time snapshot written 2026-10-27 22:10 CET. Sections (a)–(g) follow
the standing report format; (f) is the HARVEST input for `TODO_LIST.md` /
`ROADMAP.md` (routed this session); this file gets inline annotations as work
lands — never rewritten. Gates at write time: `check-docs.sh` green; full
battery green (race `-count=10`, vet ×2, lint 0 issues, coverage 100.0%,
compare-links, dprint). Concurrent session active: its in-flight tails are
§b1, §c1, §f26, not this session's._
