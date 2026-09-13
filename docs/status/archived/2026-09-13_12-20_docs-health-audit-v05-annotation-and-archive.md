# Status Report — 2026-09-13 12:20 CEST

_Session focus: docs-health AUDIT (BUILD + HARVEST + VERIFY) + ANNOTATE + ARCHIVE
over all nine 2026-0\* historical files, plus a full living-doc truth pass._
**Format note:** written as `.md` per explicit user instruction (skill default is
HTML dashboard — override honored, not propagated).
**Baseline at session start:** master `f1059a6` (clean, v0.5.0 released 2026-09-06).
**End state:** master `afd2f96` (daemon-committed), tree clean, all gates green.

---

## a) FULLY DONE — verified, not claimed

1. **All nine 2026-0\* historical files read in full** (8 × `docs/status/`, 1 ×
   `docs/planning/`, including the HTML report), plus every living doc
   (README, AGENTS, FEATURES, TODO_LIST, ROADMAP, CHANGELOG, DOMAIN_LANGUAGE,
   CONTRIBUTING) and the code (retry.go, config.go, doc.go, retry_test.go,
   go.mod, .golangci.yml, ci.yml).
2. **Quality gate executed, not assumed:** `go test ./... -race` green,
   `-race -count=10` green, `go vet` clean, `golangci-lint` **0 issues**,
   coverage **100.0%** re-verified after changes.
3. **Three new pinning tests shipped** — the docs promised two guarantees that
   nothing enforced (the exact "docs-ahead-of-tests" inversion the 2026-08-22
   report confessed to):
   - `TestDo_NestedRetriesAreFailClosed` (outer loop makes exactly 1 attempt
     on inner `ErrExhausted`) — pins the README + godoc nesting claim.
   - `TestDo_OnExhaustedNotCalledOnCancel` / `...OnDeadline` — pins that the
     exhaustion callback never fires on context end.
4. **Living docs fixed in place (8 files):**
   - `DoWithValue` (v0.5.0's headline) was missing from README, FEATURES,
     AGENTS, and ROADMAP's v1.0 API-audit symbol list — added everywhere.
   - `docs/DOMAIN_LANGUAGE.md` gained the missing `retry.deadline` code-table
     row (the split brain flagged on 2026-08-22 and never fixed), a new
     **Deadline** vocabulary term, and `ErrDeadlineExceeded` in the
     Infrastructure family row.
   - ~16 drifted `file:line` citations re-anchored (DOMAIN_LANGUAGE, FEATURES)
     to `function (line N)` form.
   - AGENTS.md: Go 1.26.7, file table + control flow now include
     `DoWithValue`/`ResultFunc`/`contextEnded`, error-family map includes
     `ErrDeadlineExceeded`, `//nolint` marker claim corrected to `gosec`-only,
     "two long-delay context tests" phrasing, temporal `(v0.4.0)` markers
     trimmed, and the twice-litigated **jitter-deferral decision** recorded as
     a gotcha ("do not re-propose a `Jitter` field").
   - CONTRIBUTING.md: same corrections + a new **Fuzzing** section with the
     campaign command (2026-08-22 f.17 closed).
   - README: new "Retries that produce a value" section; benchmark claim
     updated in FEATURES (~32 ns/op measured, 0 allocs).
5. **HARVEST executed:** TODO_LIST rebuilt from scratch — the 8-item
   struck-through "Done" section (structural decay; done items belong in
   CHANGELOG) deleted, **7 verified open items** added (T9–T15) with evidence
   and source citations from the 2026-08-22 report §f plus twice-dropped items
   (`govulncheck` had been dropped by three consecutive sessions → T15).
6. **ROADMAP repaired:** stale "fuzzing" raw idea graduated to TODO T12 (the
   fuzz target shipped in v0.3.0); the already-decided `flake.nix` open
   question resolved with a decision record; the genuine **0.x-release-policy**
   question added to Open questions; v1.0 API audit list now includes
   `DoWithValue`/`ResultFunc`.
7. **CHANGELOG `[Unreleased]`:** two Added entries for the new pinning tests
   (with the "was documented, enforced by nothing" rationale).
8. **ANNOTATE — every numbered item in all nine files resolved inline:**
   ~290 verdicts (`~~item~~ done at <hash>` / `done (evidence)` /
   `**Won't implement — reason**`), hashes verified via `git cat-file` /
   `git log -S` before citing. Used the skill's canonical scripts with
   mandatory dry-runs; table-row runs shape-checked. The HTML report got a
   full **Status column** (all 20 Next-Tasks rows + 3 questions resolved).
   Self-resolving narrative sections (a/d/e records, guardrails) deliberately
   left unmarked per the skill's SKIP/"so what?" rule.
9. **ARCHIVE — all nine files `git mv`'d** to `docs/status/archived/` (8) and
   `docs/planning/archived/` (1). `docs/status/` and `docs/planning/` now hold
   only `archived/`. Dangling references to the moved files fixed in
   TODO_LIST and ROADMAP.
10. **pkg.go.dev verified live:** v0.5.0 renders with `Do`, `DoWithValue`,
    `ErrDeadlineExceeded`, and the examples (closes 2026-08-22 b.1/f.4 and
    2026-08-03 C8). The dead `/mnt/buildcache` mount verified **alive**
    (86G free — closes 2026-08-22 d.5). Go Report Card verified **sunset**
    (closes C9 as Won't-implement).
11. **Internal link check:** every relative link target across all living
    docs resolves (first checker run was itself buggy — filename-prefixed
    captures — rewritten and re-run clean).

---

## b) PARTIALLY DONE

1. ~~**TODO_LIST harvested but not executed** — T9–T15 are written with
   evidence; none is implemented yet.~~ → done (executed the same day; see
   `archived/2026-09-13_12-57_t9-t15-execution-status.md` + `CHANGELOG.md`),
2. **ROADMAP v1.0 bar updated but the API audit itself not performed** — the
   symbol list is now complete (`Do`, `DoWithValue`, `ResultFunc`, …); walking
   each symbol is still open. → Still open by design, tracked in
   `ROADMAP.md` (v1.0 section).
3. **AGENTS.md gotcha list at 16 entries** — within the 15–20 cap, but two
   entries were added this session; next addition should trigger a pruning
   pass. → Standing rule; count after the same-day docs-health pass: 18 of 20.
4. ~~**CHANGELOG `[Unreleased]`** has only the two test entries; no Fixed
   section. Curation happens at the next tag.~~ → routed to `TODO_LIST.md`
   (T21, owner-gated release cut).
5. **CONTRIBUTING claims verified — except one, verified only during this
   report's writing:** the "`mnd` and `exhaustruct` are excluded from
   `_test.go`" claim was asserted without reading `.golangci.yml` past line
   200 during the audit; it is in fact true (`.golangci.yml:229-232`), but the
   verification order was wrong (see d.4). → done (whole `.golangci.yml` read
   in the same-day docs-health pass; claim re-verified).
6. **Annotation variant fidelity** — routed-to-ROADMAP/FEATURES items were
   marked `done (tracked in …)` because the canonical script lacks a
   NOT-DO/DUPLICATE kind; the catalog's `**NOT-DO/DUPLICATE — …**` form would
   have been semantically exact (~100 items affected). → done (the same-day
   pass uses explicit hand-written variants — `routed to …`, `**Won't
   implement**` — recorded in `docs/status/README.md`).
7. **Planned doc reorganization (2026-08-22 f.19, release-notes home)** —
   still undecided, now tracked as TODO T14. → done (T14 decided 2026-09-13:
   GitHub-only; `ROADMAP.md`).

---

## c) NOT STARTED

1. ~~**`ExampleDoWithValue` godoc example** — pkg.go.dev renders `DoWithValue`
   with no example; the value-returning API is the newest surface and deserves
   the same treatment as `Do`.~~ → done (same-day docs-health pass:
   `ExampleDoWithValue` with `// Output:`; README snippet now its verified
   form).
2. ~~**`docs/status/archived/` index** — file 3's f.50 suggested an index for
   old reports; I archived them but built no index/README, and AGENTS.md does
   not mention the archive convention.~~ → done (same-day pass:
   `docs/status/README.md` index + `AGENTS.md` pointer).
3. ~~**`exhaustruct` → `exhaustruct_v5` migration** — golangci-lint emits a
   deprecation warning ("deprecated since v2.13.0, replaced by
   exhaustruct_v5") on every run; observed this session, not actioned.~~ →
   done (same-day pass at `23192cd`; lint 0 issues, warning gone).
4. **Remote CI verification for today's commits** — local gates green; no
   GitHub Actions run observed (the recurring 2026-08-07 lesson: local ≠
   remote). → done for the ci.yml side (run 34755167105 green for `691744b`,
   3/3 jobs, 2026-09-13); first fuzz-workflow run → `TODO_LIST.md` (T16).
5. ~~**Fuzz smoke on the committed tree** — CONTRIBUTING now documents the fuzz
   command; this session never ran it (only seeds + unit tests + benchmark).~~
   → done (same-day pass: 20 s campaign, 8.5M execs, 0 failures).
6. ~~**HTML render validation** — the archived HTML report received 24 hand
   edits (Status column + resolutions) without re-rendering.~~ → done
   (same-day pass: parse-validated, all tags balanced).
7. **wise-go adoption spike** (other repo) — the actual payoff of the v0.4.0
   hardening; untouched here by design. → Out of scope for this repo (lives
   in wise-go).
8. **All v1.0-track work** — API audit, `AttemptFunc` signature decision,
   options-pattern migration design, version-compatibility matrix. → Tracked
   in `ROADMAP.md` (v1.0 section + raw ideas).
9. ~~**TODO T9–T15 execution** (identity assert, `contextEnded` dedup, checkout
   bump, scheduled fuzz, corpus, release-notes home, govulncheck).~~ → done
   (same day; see `archived/2026-09-13_12-57_t9-t15-execution-status.md`).

---

## d) TOTALLY FUCKED UP (honest ledger)

1. **The new README `DoWithValue` snippet is unverified pseudo-code.**
   ~~`fetchUser(ctx, id)` and `*User` exist nowhere; the snippet was never
   compiled or run.~~ → done (same-day pass: `ExampleDoWithValue` written and
   run; README now carries the verified form).
2. **TODO_LIST was rebuilt with bare line-number citations** — the exact
   doomed-citation anti-pattern. → done (superseded in the same-day pass:
   TODO_LIST rebuilt citation-free and living docs moved to function-name-only
   citations; `AGENTS.md` gotcha added).
3. **~100 routed items annotated as `done (tracked in …)`.** → done (same-day
   pass uses explicit variants; convention recorded in
   `docs/status/README.md`).
4. **Verified CONTRIBUTING's lint-exclusion claim only after writing this
   report.** → done (whole-config reads practiced in the same-day pass).
5. **24 hand edits to the HTML report with zero render validation.** → done
   (parse-validated 2026-09-13: all tags balanced).
6. **Hash-to-row mapping taken on trust for fine-grain planning rows.** →
   accepted as historical (rows verified at parent-task level; per-row re-diff
   not planned — the parent hashes are cited).
7. **Documented fuzz, ran no fuzz.** → done (20 s fuzz smoke run in the
   same-day pass: 8.5M execs, 0 failures).
8. **Half of file-3's f.50 suggestion implemented.** → done (archive index
   built in the same-day pass: `docs/status/README.md`).

---

## e) WHAT WE SHOULD IMPROVE

1. ~~**Every README code snippet must compile.** …~~ → done (`ExampleDoWithValue`
   is now the README snippet's source of truth; remaining snippets are real,
   compiling code).
2. ~~**Function-name citations only, everywhere, including TODO_LIST.**~~ →
   done (adopted repo-wide in the same-day pass; `AGENTS.md` gotcha bans
   line-number citations in prose docs).
3. **Extend the annotation tooling, don't bend it.** The canonical scripts
   support h/v/p/w; routing needs a fifth kind (d = NOT-DO/DUPLICATE). →
   **Won't implement (for now)** — explicit hand-written variants cover the
   need; revisit only if batch passes recur.
4. ~~**Read the whole config file before asserting claims about it.**~~ → done
   (practiced in the same-day pass).
5. ~~**Validate structured artifacts after structural edits** (HTML: parse or
   headless-render; markdown tables: the scripts' shape check — which worked).
   Edit-success ≠ render-success.~~ → done (HTML parse-validated in the
   same-day pass).
6. ~~**Run the gates you document, in the session you document them.** The
   fuzz command landed in CONTRIBUTING while its author skipped it (d.7).~~ →
   done (fuzz gate run in the same-day pass).
7. ~~**State the annotation-classification policy in the health report itself.**
   The a/d/e SKIP rationale lived in tool output and my head; the report's
   reader had to infer it.~~ → done (classification policy recorded in
   `docs/status/README.md`).
8. ~~**Close the remote loop.** Every docs pass should end with "CI green on
   GitHub for the tip commit" or an explicit "not verified" line — the
   2026-08-07 report's §e.3 lesson, still unlearned in practice (c.4).~~ →
   done (CI green on GitHub for tip `691744b`, run 34755167105; first fuzz
   run explicitly not verified → `TODO_LIST.md` T16).
9. ~~**Archive needs a door sign.** `archived/` without an index or an AGENTS.md
   pointer relies on grep; one AGENTS.md line ("annotated snapshots live in
   docs/status/archived/") would prevent future sessions from treating the
   archive as missing history.~~ → done (both: `AGENTS.md` pointer +
   `docs/status/README.md`).

---

## f) Up to 50 things to get done next

Ranked by impact → effort. Items already tracked in TODO_LIST keep their IDs;
most P2/P3 tail items are ROADMAP fuel, not commitments.

**P1 — close this session's gaps:**

1. ~~Convert the README `DoWithValue` snippet into `ExampleDoWithValue` with
   `// Output:` and paste the verified form back into README (fixes d.1 + c.1).~~
   → done (same-day pass).
2. ~~Add the amplification-side nesting test: outer loop with a custom
   `IsRetryable` that retries `Infrastructure` DOES multiply attempts — pins
   the "override deliberately" half of the README nesting paragraph.~~ → done
   at `23192cd` (`TestDo_NestedRetriesAmplifyWhenOverridden`, outer 3 × inner 3
   = 9).
3. ~~Replace TODO_LIST's bare line citations with function-name citations (d.2).~~
   → done (superseded: TODO_LIST rebuilt citation-free; convention is
   function-name-only, `AGENTS.md` gotcha).
4. ~~Add an archive pointer + annotation convention to AGENTS.md (one line) and
   a minimal `docs/status/README.md` index (fixes c.2/d.8).~~ → done (same-day
   pass: both).
5. ~~Run a 1-minute fuzz smoke on the committed tree (d.7).~~ → done (20 s
   campaign, 8.5M execs, 0 failures).
6. ~~Verify CI is green on GitHub for `afd2f96` (c.4).~~ → done (better: green
   for the newer tip `691744b` — run 34755167105, 3/3 jobs, 2026-09-13).
7. ~~Validate the annotated HTML renders (parse or headless open) (d.5).~~ →
   done (parse-validated: all tags balanced).
8. ~~Execute TODO T9 — strengthen the non-retryable assert to `err != rejection`
   identity.~~ → done at `a89ad17` (`//nolint:errorlint` with reason).
9. ~~Execute TODO T10 — dedupe `contextEnded`'s code/message pairs.~~ → done at
   `a89ad17` (const block; wrapper message unified on the sentinel wording).
10. ~~Migrate `exhaustruct` → `exhaustruct_v5` in `.golangci.yml` (kills the
    deprecation warning on every run).~~ → done at `23192cd` (incl. CI
    golangci bump to v2.13.2, required for the new linter).

**P2 — CI & tooling:**

11. ~~TODO T11 — `actions/checkout` v4 → v5/v6 (SHA re-pin).~~ → done at
    `9d7efa1` (v6, `d23441a4…`); runner-verified via run 34755167105.
12. ~~TODO T15 — `govulncheck` step in CI.~~ → done at `9d7efa1`
    (`032d4551… # v1`); runner-verified via run 34755167105.
13. ~~TODO T12 — scheduled fuzz job in CI.~~ → done at `9d7efa1`
    (`fuzz.yml`); zero runs so far → `TODO_LIST.md` (T16).
14. ~~TODO T13 — committed `testdata/fuzz/` corpus.~~ → done at `9d7efa1`
    (7 corpus files).
15. ~~TODO T14 — decide release-notes home (repo `docs/releases/` vs
    GitHub-only).~~ → done at `3ec60b0` (decided: GitHub-only;
    `ROADMAP.md`).
16. ~~Verify dependabot.yml actually covers the pinned action SHAs + golangci
    action (content never read this session).~~ → done (read 2026-09-13:
    weekly gomod + github-actions watchers as documented; but zero PRs ever
    opened → `TODO_LIST.md` T17).
17. ~~Confirm local golangci-lint version vs CI's pinned v2.12.2 don't drift
    findings.~~ → done at `23192cd` (aligned: local v2.13.2 = CI pin).
18. ~~Document the status-report annotation/archive convention in CONTRIBUTING
    (what future sessions should do with `docs/status/`).~~ → done
    (deliberately NOT in CONTRIBUTING — status reports are AI-session
    artifacts, not contributor-facing; the convention lives in
    `docs/status/README.md` + an `AGENTS.md` pointer).
19. ~~Next release: keep CHANGELOG section order Added→Changed→Fixed (0.4.0's
    Fixed-before-Added sin must not recur).~~ → done (standing order held in
    `[0.5.0]` and `[Unreleased]`).
20. ~~Re-check pkg.go.dev rendering after the next tag (lag window).~~ →
    folded into `TODO_LIST.md` (T21).

**P3 — documentation durability:**

21. ~~Sweep all remaining bare line-number citations repo-wide into
    `function (line N)` form (2026-08-08_11-22 f.14, now truly finishable).~~
    → done, differently and better (same-day pass: the `function (line N)`
    form itself rotted within hours of being written — living docs now cite
    function names only, no line numbers anywhere).
22. **Prune AGENTS.md gotchas when the 20-row cap hits (16 now).** → Standing
    rule; 18 of 20 after the same-day pass.
23. ~~README: mark any remaining illustrative snippets as such, or make them
    examples.~~ → done (the one pseudo-code snippet now mirrors
    `ExampleDoWithValue`; the rest is real code).
24. ~~Add "Was this report resolved?" pointer line to the archive index (each
    archived file is fully annotated; say so at a glance).~~ → done (index
    carries a State column).
25. **ROADMAP: document the circuit-breaker/bulkhead composition recipe when
    composition questions return.** → already tracked (`ROADMAP.md`
    composition primitives).
26. ~~Consider a one-page delay-sequence table in README (200ms cap example) —
    repeatedly proposed, repeatedly rejected; decide once.~~ → decided
    2026-09-13: **Won't implement** — the formula is already documented in
    three places; a static table would be a fourth number to keep in sync
    (`ROADMAP.md` decision record).

**P4 — v1.0 track (ROADMAP-owned):**

27. ~~Public API surface audit (all 12 symbols).~~ → tracked (`ROADMAP.md`
    v1.0 section).
28. ~~`AttemptFunc(ctx, attempt)` signature decision (pass previous error?).~~ →
    tracked (`ROADMAP.md` v1.0 section).
29. ~~Options-pattern migration design (`WithOnRetry`, `WithJitter`, …).~~ →
    tracked (`ROADMAP.md` raw ideas).
30. ~~Version-compatibility matrix with `go-error-family`.~~ → tracked
    (`ROADMAP.md` raw ideas).
31. ~~Deterministic-RNG option decision (FEATURES WORTH_CONSIDERING).~~ →
    tracked (`FEATURES.md` WORTH_CONSIDERING).
32. ~~Deadline-aware attempt-budgeting decision (FEATURES WORTH_CONSIDERING).~~
    → tracked (`FEATURES.md` WORTH_CONSIDERING).
33. ~~Docs website (Astro/Starlight/Firebase) — post-API-freeze only.~~ →
    tracked (`ROADMAP.md` raw ideas).
34. ~~v1.0 readiness checklist + freeze.~~ → tracked (`ROADMAP.md` v1.0
    section).

**P5 — wise-go (other repo, the actual payoff):**

35. ~~Failsafe→go-retry adoption spike (M10 link).~~ → out of scope here
    (wise-go repo).
36. ~~wise-go v1.0.0 tag (owner-gated).~~ → out of scope here (wise-go repo).
37. ~~wise-go sandbox integration tests (API-key-gated).~~ → out of scope here
    (wise-go repo).
38. ~~wise-go CI re-enable (nix/GOEXPERIMENT).~~ → out of scope here (wise-go
    repo).

**P6 — owner/environment:**

39. ~~Confirm 0.x release policy (ROADMAP Open question → g.1 below).~~ →
    still open, owner-gated (`ROADMAP.md` → Open questions).
40. ~~Confirm archive retention policy (keep forever vs prune — g.2 below).~~ →
    default decided: keep forever (files are small and fully annotated);
    index at `docs/status/README.md`. Owner may still override.
41. ~~/mnt/buildcache`verified alive; drop any remaining`/tmp` cache
    workarounds in shells.~~ → done (verified in this session itself).
42. ~~Confirm annotation-variant policy for future passes (g.3 below).~~ →
    decided 2026-09-13: explicit hand-written variants (`done at`, `routed
    to …`, `**Won't implement**`); the h/v/p/w scripts stay for batch
    table/prose runs (recorded in `docs/status/README.md`).

**Explicitly closed — do not re-open:**

43. Configurable jitter factor — deferred twice (2026-08-08, 2026-08-22);
    recorded in AGENTS.md as a gotcha.
44. `OnSuccess(attempts)` hook — deferred (no consumer; derivable).
45. Go Report Card — service sunset; nothing to do.
46. `errors.AsType` migration — no-op (package uses `errors.Is` only).
47. Flake.nix — decided: raw Go commands (ROADMAP decision record).
48. Phantom `Attempt` type — Won't-implement (validation errors shipped).
49. `MaxAttempts: 1` dedicated test — covered by success/exhaustion paths.
50. Badges/CODEOWNERS/issue-templates/PR-templates — solo-maintainer repo;
    Won't-implement until external contributors appear.

---

## g) Questions I can NOT figure out myself

1. ~~**0.x release policy — full release or prerelease?** The `go-release` skill
   defaults 0.x to prereleases; wise-go v0.9.0 and go-retry v0.4.0/v0.5.0 were
   all published as full releases following your demonstrated preference. I
   recorded the question in ROADMAP → Open questions. Confirm "full releases
   for 0.x going forward" and I'll close the question for good.~~ → still
   open, owner-gated (`ROADMAP.md` → Open questions).
2. ~~**Archive retention:** should `docs/status/archived/` + the planning
   archive be kept forever (they're small, fully annotated), or pruned after
   N releases? And do you want the one-line AGENTS.md pointer + tiny index
   README (f.4), or is grep-only access fine for you?~~ → default decided:
   keep forever + index (`docs/status/README.md`); owner can override.
3. ~~**Annotation vocabulary going forward:** are `done (tracked in ROADMAP.md
   …)` markers acceptable for re-homed items, or should future passes use the
   strictly-correct `NOT-DO/DUPLICATE` form (requires extending the skill's
   annotate scripts with a new kind)? I used the former ~100 times this pass
   for tooling simplicity.~~ → decided 2026-09-13: explicit hand-written
   variants (`routed to …` / `**Won't implement**`); recorded in
   `docs/status/README.md`.

---

_Point-in-time snapshot; stale by design. Follow-ups belong in TODO_LIST /
ROADMAP, not here. The auto-commit daemon will pick this file up; no manual
commit per harness rules._
