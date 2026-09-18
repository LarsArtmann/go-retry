# Status Report — go-retry — T45–T47 Executed: Docs Gates Shipped, Rendering Bugs Found and Fixed

Written 2026-09-18 05:41 CEST; the work it reports ran 2026-09-17 ≈20:53–21:28
CEST, continuing the 20:32 report (v0.7.1 shipped, annotation catastrophe
repaired). Scope: the owner's instruction to execute the open P1/P2 backlog —
T45 (marker-completeness gate), T46 (`scripts/check-docs.sh`), T47 (`tools/`
module gating) — plus honest ledgers. This session's premise-check on arrival
found the prior summary wrong about push state; that correction is §d.1.

## a) FULLY DONE

1. **Session-start state verification caught a stale claim.** The 20:32
   summary said "all pushed"; fresh `git status` showed master **ahead 16**
   of origin with `origin/master` frozen at the release commit `082842a`
   (18:08). All 16 unpushed daemon commits were inspected (`git log --stat`)
   and confirmed to be legitimate sweeps of the documented session work —
   nothing foreign, nothing broken. Re-verified against fresh CLI runs, not
   the summary (the exact failure mode this repo's rituals warn about).
2. **Gate design prototyped against the real corpus before any Go was
   written.** Two throwaway Python scanners (`/tmp/docs-guard-proto.py`,
   `/tmp/markers-proto.py`) ran over all 30 tracked `.md` files and the 16
   archived reports (723 gated `§b/§c/§f/§g` numbered items) to tune the
   detection rules against reality first: what verdict forms actually exist
   in the archive, which of the flagged items were real bugs vs. unrecognized
   verdict phrasings. This turned gate design from guessing into measurement.
3. **Seven real documentation bugs found and fixed** (all rendering-class,
   the same catastrophe family the 20:32 session repaired — these were the
   survivors):
   - AGENTS.md architecture table: unclosed code span — 21 backtick runs,
     odd count; `ErrDeadlineExceeded` rendered with a literal dangling
     backtick on GitHub.
   - `docs/status/README.md` index, 13:19 row: cell ended in an unclosed
     backtick (`` `TODO_LIST.md`/`ROADMAP.md `` — never closed).
   - 14:48 report `§c.1`: lone `~` inside a strike span (`~14:55`) — the
     class-B bug that silently defeats span pairing; replaced with `≈14:55`.
   - 16:55 report `§b.2`: same class (`~10 minutes`) — replaced with `≈10`.
   - 05:53 report `§d.3`: garbled cell from the original annotation-script
     collision — three orphaned single backticks around an escaped pipe;
     repaired to a readable, render-clean sentence.
   - 08:47 report `§f.8` (T37): genuinely unannotated; struck **done** —
     premise verified first (the v0.7.1 pkg.go.dev page rendered all three
     examples, verified during the release session).
   - 08:47 report `§f.15` (T42): genuinely unannotated; given its
     `→ routed to TODO_LIST (T42, still open)` pointer.
4. **T45 — the marker/rendering guard shipped as a Go guard test**
   (`docs_test.go`, external `retry_test` package per repo convention; the
   script role T45 asked for is fulfilled by T46's orchestrator below):
   - `TestMarkdownStrikethroughSpansRender` — detects all three bug classes
     statically: space-preceded closers, lone `~` inside a span, unclosed
     spans at block end. Fence-aware, code-span-stripping, table
     cells-scanned independently (spans cannot cross cell boundaries).
   - `TestMarkdownTableCellsCloseCodeSpans` — per-cell backtick-run balance
     with `\|` escape handling.
   - `TestArchivedReportItemsCarryVerdicts` — every numbered item in
     archived `§b/§c/§f/§g` sections must carry a verdict; the recognized
     gesture set is fail-closed (strike present, `done at`, `routed to`,
     arrows after sentence end, `— DONE`, explicit `Open —`, `won't
     implement / won't for now`, `resolved/superseded/deferred/decided`,
     `no-op`, `verified`, `covered by`, `nothing to do`, `correct per
     convention`); unknown future phrasings FAIL the gate and force a
     conscious extension, exactly like `actionInputAllowlist`. Also enforces
     the column-wise convention: non-empty cells in Verdict/Status columns
     of gated-section tables (the 05:53 report's tables were the live
     validation target).
   - `TestStatusIndexCoversArchive` — archive↔index consistency both
     directions, non-empty State cells, link targets exist.
   - **Prove-fail drills all passed**: injected `~~text. ~~ done` (strike
     class), an unannotated `99.` item (marker class), and an unindexed
     `zzz_drill.md` (index class) each failed the right test naming the
     offender; restore returned green. `git log -S 'Drill text'` proved the
     mutations never entered history despite two daemon commits landing
     mid-drill.
   - The one archived `.html` report is out of scope by design (documented
     in the test and in `docs/status/README.md`): zero `<del>` spans, HTML
     Status column, parse-validated instead. This closes T45's HTML design
     call.
5. **T46 — `scripts/check-docs.sh` shipped and wired in.** One command for
   the doc battery: the four guard tests (with `-count=1` — see §d.7), the
   dprint format check, and `check-compare-links.sh`. Added to the AGENTS
   Commands block and Session Ritual gate order (replacing the two separate
   dprint/compare-links lines) and to CONTRIBUTING (commands block + a "Doc
   battery" bullet replacing the narrower compare-link bullet).
6. **T47 — the `tools/` module entered the quality gates.** `go -C tools
   vet ./...` added as a CI lint-job step (after actionlint) and to the
   Session Ritual. Root `./...` never saw the nested module; now both
   surfaces vet. Scope decision recorded: golangci-lint-for-tools in CI is
   deliberately deferred to the tools-pin topology answer (ROADMAP Open
   question) — it runs clean locally (`0 issues`), and wiring it via the
   action would mean a new input to SHA-verify against the allowlist;
   topology-first avoids throwaway work. `actionlint` re-verified the
   edited workflow (0 errors); the input-allowlist test stayed green (no
   new `uses:`).
7. **Living docs synced to the new reality.** CHANGELOG `[Unreleased]`:
   three Added entries (guard tests, check-docs.sh, tools vetting) and two
   Fixed entries (rendering repairs, archive annotations) — the next cut
   inherits an honest ledger. `docs/status/README.md` gained a paragraph
   naming the enforcing tests and the HTML-scope call. TODO_LIST rebuilt:
   T45/T46 removed (done, recorded in CHANGELOG), T47 narrowed to an
   observe-tail row (first green CI run carrying the vet step), T48–T54 and
   T42 untouched. AGENTS gained the tools-vet command, the check-docs.sh
   ritual line, the ci.yml row update, and a test-cache warning folded into
   the existing test-failure-proof bullet (no new gotcha row — the 20-row
   cap holds, T49 pressure unchanged).
8. **Full gate battery green at close, re-verified after the last edit:**
   `gofmt -l` clean; `go vet ./...` and `go -C tools vet ./...` clean;
   `go test ./... -race -count=10` ok; `golangci-lint run ./...` **0 issues**;
   `golangci-lint config verify` clean; `go test -cover ./...` reads
   **coverage: 100.0% of statements** (canonical format); `govulncheck
   ./...` no vulnerabilities; `actionlint -verbose` 0 errors; dprint check
   clean after `fmt` re-padded 4+2 table files; compare-links "all tag
   pairs resolve"; docs battery green.
9. **Lint debt paid in the same change, not deferred.** The new test file
   arrived with ≈55 findings across 8 linters (varnamelen, nlreturn,
   wsl_v5, intrange, prealloc, unlambda, gofumpt, golines). All fixed —
   renames (`j`→`runEnd`, `d`→`entry`, `s`→`compact`), preallocation, range
   loops, unlambda, and the whitespace classes scripted from lint output
   with an apply→recheck loop (3 iterations to zero). Not shipped red.
10. **gopls restarted at session end** after lying with phantom syntax
    errors for the entire session (see §d.8); post-restart tests and lint
    re-ran green.

## b) PARTIALLY DONE

1. **T47's remote half is unverified by construction.** The vet step is in
   `ci.yml`, actionlint-passed and locally green, but **16 commits sit
   unpushed** — the daemon has committed continuously (latest `e83a509`,
   21:28) yet nothing has been pushed since `082842a` (18:08 the previous
   evening). No CI run can carry the step until push happens, and pushing
   is not mine to decide. The TODO row was narrowed to exactly this
   observation tail (P3) instead of being closed done.
2. **T45's "graduate the paragraph render-diff into it" tail is open.** The
   GFM render-diff (`/tmp/verify-strikes.py`, GitHub API round-trip per
   file) remains a manual deep-verify tool outside the repo. The static
   gate covers the documented bug classes but has known, accepted blind
   spots: span pairing across list-item boundaries is not item-scoped
   (false negatives possible), and prose code spans crossing lines are not
   modeled. The drill-proven static gate plus the archived render-diff is
   the shipped compromise; graduating the diff into `scripts/` (network-
   and-auth-gated, so never CI) is routed as §f.3.
3. **The gate's own parsing edges have no self-tests.** Correctness is
   currently evidenced by the drills plus the live corpus (723 items, 30
   `.md` files). Edge behavior — `\|` splitting, double-backtick runs,
   fence detection, the lazy-continuation item segmenter — is exercised
   only incidentally. Table-driven unit tests for the scanner helpers are
   routed as §f.4.

## c) NOT STARTED

1. **The push.** Explicitly owner-gated; not done (see §b.1).
2. **T48–T54 (all P3):** canonical coverage recipe across
   README/CONTRIBUTING/FEATURES; AGENTS gotcha prune below 20; SECURITY.md
   posture audit against the grown supply-chain surface; safe probe
   fixtures for guard drills; index/backlog conventions; dangling-hash
   sweep in archived plans; living-docs go-version guard extension. All
   untouched this session, all still evidence-backed in TODO_LIST.
3. **T42 (Dependabot watch):** still open; no Dependabot PR appeared during
   the window.
4. **ROADMAP Open questions** (daemon commit-message/push policy,
   tools-pin topology, dprint pinning posture): owner decisions, untouched
   — though §d.1 sharpens the daemon-policy question with fresh evidence.
5. **CI wiring for compare-links:** `check-compare-links.sh` still runs
   only at release time and locally; it is not a CI step (routed §f.6).

## d) TOTALLY FUCKED UP (honest ledger)

1. **Trusted a stale session summary over the working tree — the cardinal
   sin this repo keeps re-learning.** I opened the session prepared to
   "mark todos complete" per a summary claiming everything was pushed and
   the tree clean. Fresh `git status` showed 16 unpushed commits and froze
   that plan. Nothing was damaged (I verified the commits before acting),
   but the reflex to trust the handoff artifact cost the first minutes and
   produced this report's sharpest §g question.
2. **A multiedit misfire deleted a TODO row I had just read.** Batch-
   editing TODO_LIST, my second edit anchored on the `| T48 | Canonical
   coverage recipe` prefix and replaced it with the new T47 row — erasing
   T48's task text and leaving a duplicate T47. The first edit in the same
   batch contained a garbage anchor (`hmm`) I had fumbled into both sides.
   Both were caught within one command (post-edit grep + diff), and the
   table was repaired by line surgery with T48 restored verbatim; dprint
   re-padded. But this is the exact "compose anchors from memory" failure
   the editing rules prohibit, done twice in one batch.
3. **I briefly fabricated archive evidence.** Repairing the garbled `§d.3`
   cell, my first rewrite filled its empty Evidence cell with invented
   content ("live diff, script stderr") that the original never carried.
   Self-caught in the immediately following command and reverted to the
   empty cell it always was — but in an archive whose whole contract is
   fidelity, the reflex to "complete" a row is precisely wrong. Format-only
   repairs; never content.
4. **The Go test cache nearly taught me a false lesson.** The first drill-A
   run printed nothing and I almost moved on: the mutation lived in a data
   file, which is not part of the test cache key, so `go test` returned a
   cached PASS. Caught by refusing to accept silence, re-run with
   `-count=1` failed correctly, and the lesson is now structural —
   `check-docs.sh` hardcodes `-count=1` with a comment, and the AGENTS
   test-failure-proof bullet warns about it. Before that, though, the
   drill "passed" dishonestly once.
5. **A sloppy edit corrupted the file it was fixing.** Mid-lint-fix, an
   edit with a trailing-newline anchor mismatch merged two lines
   (`j++\t\t}`) inside the function I was renaming. Self-caught by viewing
   the region, rewrote cleanly. Same root cause as §d.2: anchors composed
   from memory instead of freshly viewed text.
6. **Burned round trips on bash-as-reading.** Multiple edit attempts were
   refused because I had "read" files via `sed`/`grep` instead of the view
   tool (and once because dprint re-padded a file between my read and
   edit). The rule is explicit; I keep paying the tax anyway.
7. **The daemon raced my drills twice and I only proved the safety
   afterward.** Two sweeps committed while mutations sat in the working
   tree; had a sweep landed between append and restore, a drill line would
   be in master history. It did not (`git log -S` verified), but I ran the
   drills before checking whether the daemon was mid-cycle — the safe order
   is the reverse.
8. **Worked ≈40 minutes against a lying LSP and restarted it only at the
   end.** gopls screamed phantom syntax errors (`rune literal not
   terminated`, stale line numbers) on every tool result from the first
   docs_test.go write onward. I correctly trusted `go vet`/`go test` as
   truth (the AGENTS rule worked), but I let the noise pollute every
   diagnostic payload for the rest of the session instead of restarting
   gopls the moment the compiler contradicted it. The restart was an
   afterthought; it should have been the second action after the first
   contradiction.

## e) WHAT WE SHOULD IMPROVE

1. **Treat handoff summaries as hypotheses, not state.** Every session
   should open with the same three commands (`git status --branch`, `git
   log origin/master..HEAD`, `date`) before trusting any prose about repo
   state — including mine. The 20:32 summary was written honestly and was
   still wrong within two hours, because the daemon kept moving the tree.
2. **Never compose edit anchors from memory; batch size 1 for anchored
   edits.** Every failed or misfired edit this session (§d.2, §d.5, §d.6)
   came from an anchor typed from recall or from a file dprint/daemon had
   touched since my read. View immediately before edit, even for files read
   minutes earlier — the tree moves under you by design here.
3. **Restart the LSP on first compiler contradiction, not at session end.**
   Propose extending the AGENTS tool-trust bullet: when diagnostics and
   `go vet`/`go test` disagree, run the compiler AND restart the LSP in the
   same breath; stale diagnostics are noise that erodes attention for real
   findings.
4. **Archive edits get a stricter rule than code edits: render-class
   repairs may reformat, never add.** The §d.3 near-miss becomes a written
   convention: in `docs/status/archived/`, fixing pairing/backticks/
   padding is legitimate; supplying missing table content is not, even when
   the "obvious" value seems inferable.
5. **Prove-fail drills should check for daemon activity first** (`git
   status` immediately before mutating, immediately after restoring), and
   the restore step should end with `git log -S` on the injected string —
   cheap, and it converts "probably fine" into verified.
6. **Prefer measure-then-build for any gate over a hand-written corpus.**
   The Python-prototype-first move (§a.2) is why the verdict-recognizer
   needed zero false-positive triage after landing: the 16 gesture-rule
   items were classified from real archive phrasing, not imagined forms.
   Repeat this shape for the next gate (planning-archive markers, §f.13).
7. **The static gate is a floor, not the oracle.** Its blind spots are
   documented (§b.2); the render-diff remains the ground truth for
   "does this render 1:1 on GitHub". Keep the two tools conceptually
   separate in docs so nobody mistake the cheap gate for completeness.

## f) Up to 50 things we should get done next

Release-ordered first, then gates, then hygiene. Tags: `[PUSH]` · `[GATE]`
· `[DOC]` · `[CI]` · `[OWNER]` · `[PROCESS]`.

1. `[PUSH]` **Owner call: push the 16 pending commits** (docs + gates,
   `082842a..e83a509`) — everything is gate-green at tip; the only remote
   risk is ordinary.
2. `[CI]` **Verify the first green run carrying the `tools/` vet step**
   after that push (`gh run list`/`view`), then retire the T47 observe row.
3. `[GATE]` **Graduate the GFM render-diff into `scripts/verify-render.sh`**
   (per-file GitHub API check, `gh`-auth-gated, documented as local-only —
   never CI), so the deep-verify stops living in `/tmp`.
4. `[GATE]` **Table-driven self-tests for the docs-gate scanner edges:**
   `\|` cell splitting, double-backtick runs, fence detection, unclosed
   run preservation, lazy-continuation item segmentation.
5. `[GATE]` **Extend marker gating to `docs/planning/archived/*.md`** —
   three fully-annotated plans currently stand outside the marker gate;
   the §-letter scheme matches, so the same checker applies.
6. `[CI]` **Add `check-compare-links.sh` as a CI step** (lint job, after
   actionlint) — today it only runs locally and at release time; a poisoned
   link should fail the push, not the release.
7. `[DOC]` **T48: one canonical coverage recipe** across
   README/CONTRIBUTING/FEATURES plus the full CONTRIBUTING drift sweep.
8. `[PROCESS]` **T49: prune AGENTS gotchas below 20** — the budget must be
   refilled before the next gotcha lands; this session added none (the
   cache warning went into an existing bullet), and the pressure remains.
9. `[DOC]` **T50: SECURITY.md posture audit** against the current surface
   (nested tools module, dprint action, Dependabot groups).
10. `[PROCESS]` **T51: safe probe fixtures for guard drills** — a
    gitignored scratch convention so future prove-fail drills never touch
    live workflow/doc files even briefly.
11. `[DOC]` **T52: index/backlog conventions** — State-cell format for the
    status index, planning-file row conventions, whether
    `docs/planning/` earns its own README.
12. `[DOC]` **T53: sweep archived plans for dangling `done at` hashes**
    (the known dangling `9c08595` in the 2026-09-16 plan; find others).
13. `[GATE]` **T54: assert the living docs' stated Go version matches
    `go.mod`**, and decide the go-error-family surface-contract question.
14. `[OWNER]` **T42: Dependabot watchlist** — observe the post-merge
    rebase on the next actions-group PR, then run the `/tools` bump flow
    end-to-end.
15. `[CI]` **Confirm CI green on the batch once pushed** (runs for
    `e83a509`-tip), closing the broken-but-green exposure window the
    unpushed pile represents.
16. `[OWNER]` **Daemon push policy, now with evidence:** the daemon
    committed 16 times between 18:08 and 21:28 without pushing once —
    either its push step is wedged or push-on-commit was never its
    behavior. Either answer changes how every session ends (§g.1).
17. `[PROCESS]` **Adopt the §d.4 archive-edit rule in writing** (one
    sentence in `docs/status/README.md`'s annotation convention: format
    repairs only, never content) so the near-miss becomes policy, not
    anecdote.
18. `[PROCESS]` **Adopt the open-with-commands ritual** (§e.1) in AGENTS:
    status verification before summary trust at session start.
19. `[GATE]` **Decide the marker gate's exemption story for narrative
    numbered items** — today §a/§d/§e are exempt by section letter; if a
    future report puts narrative items inside a gated letter, the gate
    will (correctly) demand annotation or a conscious recognizer change.
    Write that policy down before it first bites.
20. `[DOC]` **Record the gesture-set extension procedure** next to the
    verdict patterns in `docs_test.go` comments are already the mechanism;
    add one line to `docs/status/README.md`: "new verdict phrasing failing
    the gate is fixed by extending the recognizer deliberately."
21. `[CI]` **Once topology is decided: golangci-lint for `tools/` in CI** —
    requires verifying `working-directory` (or equivalent) input at the
    action's pinned SHA and updating `actionInputAllowlist` in the same
    change.
22. `[OWNER]` **Tools-pin topology answer** (ROADMAP Open question) — it
    gates §f.21 and shapes where T47's gate lives long-term.
23. `[OWNER]` **dprint pinning posture** (ROADMAP Open question) — the
    CI action pins 0.57.4 while local runs float on nixpkgs; a mismatch
    class is still possible (drift between the two formatters).
24. `[PROCESS]` **Flake-watch: one `-shuffle=on -count=50` run this week**
    over the suite to chase the once-seen, never-reproduced 13:19-era
    transient FAIL (suspected file-reading test racing git operations).

## g) QUESTIONS I CAN **NOT** FIGURE OUT MYSELF

1. **The daemon stopped pushing at 18:08 but never stopped committing —
   which behavior is intended, and should I treat unpushed piles as
   normal?** I cannot see the daemon's config, runtime, or logs from
   inside the repo. If push-on-commit was the design, it is wedged and 16
   commits (including today's CI-affecting `tools/` vet step) are
   invisible to CI until you push or fix it; if committing without pushing
   is the new policy, I will stop flagging it — but then every session
   ends with an unverified-CI window, and I would like your standing
   instruction: may I push gate-green doc/tooling commits myself when the
   pile grows past a threshold, or is push always yours?
2. **Should the GFM render-diff become a permanent repo script
   (`scripts/verify-render.sh`, requires your `gh` auth, never runs in
   CI), or is the static docs gate sufficient canon and the diff stays a
   disposable deep-verify?** I cannot decide how much verification
   redundancy you want for markdown rendering fidelity; the honest
   trade-off is one authenticated command vs. a small blind spot the
   static gate cannot see (cross-item span pairing, cross-line code
   spans).
3. **Do you want the docs-gate scanners themselves unit-tested
   (table-driven edge cases, ≈150 lines of test code guarding tooling that
   guards docs), or is drill-proven-plus-live-corpus sufficient forever?**
   The gate currently fails loudly on corpus drift, but a regression in
   its own parser (e.g. the `\|` splitter) would fail silently-green; only
   you can say whether that second-order risk justifies the tests.
