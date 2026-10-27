# Status Report — 2026-10-27 22:08 — BuildFlow config removed, upstream #77 filed

Session scope: the `test-race` failure triage, the owner's no-`.buildflow.yml`
decision executed end-to-end, and the upstream filing. Based on this session
only. Repo state at write time: gates all green (gofmt, vet root + `tools/`,
`go test ./... -race -count=10`, golangci-lint 0 issues, coverage 100.0%,
check-docs, dprint), go pins intact (`go 1.26` root, `go 1.26.0` tools),
`.buildflow.yml` gone, open issues in this repo: 0, LarsArtmann/BuildFlow#77
OPEN.

## a) FULLY DONE

1. **Pasted `test-race` failure root-caused.** It was the 20:40–20:42
   deletion window: daemon commit `7d19442` deleted `.buildflow.yml`,
   buildflow run `20261027-194005` correctly went red on the two disposition
   guards, restore landed (`6791e1d`). Verified stale, guards green, the
   exact failing step (`buildflow -s test-race`) re-run green — before the
   owner's decision changed the game.
2. **Owner decision executed: no `.buildflow.yml`.** File deleted; the two
   disposition guards and their parser removed from `dispositions_test.go`
   (sentinel, dependency-floor, and lychee guards intact and green).
3. **Four-part empirical verification matrix** (why the file was / was not
   needed), all with the installed binary `49ced3e` (verified equal to
   BuildFlow repo HEAD):
   - `go-mod-update` bumps a pinned go line unconditionally — LIVE repro in a
     scratch module (replace-only dep, `go version` shimmed to 1.27): logged
     `Updating go line 1.26 -> 1.27`, ran `go mod edit -go=1.27`, zero
     dependency changes.
   - art-dupl `//art-dupl:accept` directives suppress standalone: 6 findings,
     all suppressed, findings gate exit 0. The old AGENTS/CONTRIBUTING claim
     ("honored only via the opt-in") was wrong — corrected.
   - `go-mod-normalize` holds back dependency-floor-demanded patch pins
     (`HeldBackByDepFloor`), so the `respect_patch_floor` disposition was
     redundant for this repo's shape.
   - `go-version-auto-configure` is detect/warn-only on patch-form
     directives (no rewrite).
4. **`0f61e18` re-diagnosed.** The 2026-10-15 pin loss was dependency-floor
   forcing (`go get` of errorfamily v0.12.0, which declares `go 1.27`) — the
   `.buildflow.yml` skip was present at that commit and never failed; the
   inherited "go-mod-update did it" attribution was wrong.
5. **Upstream issue filed: LarsArtmann/BuildFlow#77** — pin-blind go-line
   bump, source-verified (`determineGoLineTargets`) + live-reproduced,
   duplicate-searched (distinct from closed #58), voice-checked
   (0 FAIL / 0 WARN), verified landed OPEN. Drafts kept in `docs/drafts/`.
6. **Docs swept (8 files):** AGENTS (line budget held at 219/220), FEATURES,
   CONTRIBUTING (2 spots), engineering-notes (pin-fight section, patch-floor
   paragraph, file table rows), ROADMAP (new _Decided 2026-10-27_ entry with
   the #77 link; daemon question refocused), CHANGELOG (Unreleased Changed
   entry + stale Added wording fixed), `.go-structure-linter.yaml` comment.
7. **Parallel-session bullet corrected** in engineering-notes: the ~21:02
   working-tree deletion was the owner's order, not the daemon; added the
   #77 link; dropped the "go-line rewriters" overstatement.
8. **Hygiene:** coverage ritual gap closed (100.0%), stale LSP diagnostics
   cleared via restart (lint was 0 all along), `/tmp` repro trashed,
   final sweep confirmed no lingering `.buildflow.yml` references in living
   docs.

## b) PARTIALLY DONE

1. **TODO_LIST accuracy pass.** T94 already carries today's art-dupl finding
   (updated by the parallel session), but T60's premise is now stale (binary
   == HEAD verified today) and T85's premise is half-stale (the second
   2026-10-27 deletion was owner-ordered, not a daemon malfunction). Noticed
   at session end; deliberately untouched because the parallel session was
   mid-flight in that file.
2. **HARVEST of this report's section (f)** into TODO_LIST/ROADMAP — deferred
   pending instruction (standing order from the previous session).
3. **Interim pin exposure handled as documentation, not automation.** Every
   doc states the restore ritual; nothing repo-side can prevent the bump
   (fleet-level; that is what BuildFlow#77 asks for). Nuance verified today:
   with the local toolchain at 1.26.8 the bump case cannot even fire — the
   exposure goes live the day the installed toolchain moves past 1.26.

## c) NOT STARTED

1. The three standing owner questions (see g) — they gate release posture.
2. v0.9.0 release pipeline (govulncheck root + `tools/`, CHANGELOG cut,
   annotated tag, GitHub-only notes).
3. Consumer sweep in go-cqrs-lite's `middleware/` module (pin bump + ritual
   suites with `-race`; also check whether its docs/config carry
   `.buildflow.yml` assumptions worth the same treatment).
4. T85's remaining BuildFlow filing: the `--dry-run` that wrote
   `.golangci.yml` (needs its own verify-before-filing repro).
5. BuildFlow#77 follow-up watch (verify the fix when it lands, then strip the
   interim-exposure notes from the docs).

## d) TOTALLY FUCKED UP!

1. **Wrote an incident narrative before source verification.** My first
   ROADMAP edit this session recorded the inherited story ("two more daemon
   incident classes") as fact. Source work later showed `0f61e18` was
   dep-floor forcing and the `.buildflow.yml` deletion was owner-intended.
   This violated the repo's own ritual ("never trust stale status-report
   claims"). Fixed the same session (bullet rewritten twice), but the repo
   briefly carried wrong causal claims, and the first version of today's
   ROADMAP entry would have misinformed the daemon-guard discussion.
2. **Late discovery of the parallel session.** The edit-conflict warning on
   ROADMAP was misread as daemon noise; unfamiliar diffs (README whitespace,
   mass archival renames, TODO_LIST edits) registered only at the final
   sweep. Real collision risk while I was editing the same doc set;
   mitigated only by re-running the battery at the end.
3. **Coverage gate skipped after test changes.** The ritual lists coverage
   whenever tests change; I changed `dispositions_test.go` and ran race +
   lint but omitted `go test -cover` until the wrap-up (now 100.0%).
4. **Skills reloaded late.** `how-to-golang` was not re-viewed before editing
   `dispositions_test.go` (the handoff summary said to reload per session).
   `buildflow`, `verify-before-filing`, and `github-voice` were loaded at the
   right moments; the miss was the Go-policy skill. No damage — style matched
   and gates green — but the process contract slipped.
5. Minor: tolerated stale LSP warnings all session (restarted only at the
   end), and left `docs/drafts/2026-10-27_buildflow-issue-body.md` as a
   second artifact next to the annotated draft.

## e) WHAT WE SHOULD IMPROVE!

1. **Extend the verification ritual to attribution, not just hashes.** Today
   proved a verified hash can carry a wrong causal story. Rule candidate for
   AGENTS: incident narratives cite a mechanism verified at source (or
   reproduce it) before landing in living docs.
2. **Treat edit-conflicts and unfamiliar diffs as parallel-session signals.**
   A mid-session `git log` authorship/mtime check is cheap; discovering the
   second session at the final sweep is not.
3. **Run the ritual as a checklist, not from memory** — coverage was skipped
   precisely because "tests changed" was processed as "race + lint".
4. **Reload handoff-listed skills at session start**, not when remembered.
5. **Falsified TODO premises should get their staleness note in the same
   session** (T60, T85 today) — otherwise the queue lies quietly.
6. **Watch the pin-narrative fan-out.** The no-config decision now touches
   AGENTS, engineering-notes, ROADMAP, CHANGELOG, CONTRIBUTING, FEATURES.
   engineering-notes is the single long-form home; keep every other mention
   a pointer, or they will drift (the wrong "honored only via opt-in" claim
   survived three docs for weeks — found only because today's work forced a
   re-read).

## f) Things to get done next (ranked)

| #                                                                                                                                           | Task                                                                                                                   | Size                                 | Impact / note                                       |
| ------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ------------------------------------ | --------------------------------------------------- |
| 1                                                                                                                                           | Owner: go-pin endgame decision (hold `go 1.26` vs deliberate `1.27` sweep)                                             | S                                    | Critical — gates release posture (g-1)              |
| 2                                                                                                                                           | Owner: SECURITY.md SLA numbers — keep / adjust / drop                                                                  | S                                    | Critical — public doc, agent-invented numbers (g-2) |
| 3 **→ resolved 2026-10-27 22:47: owner ordered the cut — published as `v0.9.0`**                                                            | S                                                                                                                      | Critical (g-3)                       |                                                     |
| 4 **→ resolved 2026-10-27: `v0.9.0` tagged at `b392900`; CI green on the tag commit; GitHub release + proxy + clean-dir `go get` verified** | M                                                                                                                      | High — both changes are user-visible |                                                     |
| 5                                                                                                                                           | Consumer sweep: go-cqrs-lite `middleware/` bump + ritual suites `-race`; check its docs for config assumptions         | M                                    | High                                                |
| 6                                                                                                                                           | File the BuildFlow dry-mode-write issue (T85 remainder) after a clean repro                                            | M                                    | High — unexplained writes erode trust in dry-run    |
| 7                                                                                                                                           | Update TODO_LIST T60 (stale: binary == HEAD since today) and T85 premise (second deletion owner-ordered)               | S                                    | High — queue accuracy                               |
| 8                                                                                                                                           | Watch BuildFlow#77; on fix: verify, then strip interim-exposure notes from the docs                                    | S                                    | High                                                |
| 9                                                                                                                                           | HARVEST this report's (f) into TODO_LIST/ROADMAP (docs-health, verify-before-routing)                                  | S                                    | Medium                                              |
| 10                                                                                                                                          | Decide on a `TestBuildflowConfigStaysAbsent` guard (fail-loud if anything re-creates `.buildflow.yml`)                 | S                                    | Medium — decision already recorded in ROADMAP       |
| 11                                                                                                                                          | Per-push CI `-count=2`/`-count=3` flake net (standing 2026-10-09 question)                                             | S                                    | Medium                                              |
| 12                                                                                                                                          | T54: guard that living docs' stated Go version matches `go.mod`                                                        | S                                    | Medium                                              |
| 13                                                                                                                                          | T83: re-run `BenchmarkComputeDelay` on the current toolchain, refresh FEATURES numbers                                 | S                                    | Medium                                              |
| 14                                                                                                                                          | T92: dead-directive experiment (flip one accept-hash byte, observe art-dupl) — completes today's art-dupl verification | S                                    | Medium                                              |
| 15                                                                                                                                          | T94: decide the warnings-gating stance for clone-hash drift (make it loud)                                             | S                                    | Medium                                              |
| 16                                                                                                                                          | Fleet audit (owner call): which repos carry `.buildflow.yml` dispositions today's findings make unnecessary            | M                                    | Medium — art-dupl opt-in, respect_patch_floor       |
| 17                                                                                                                                          | Verify docs/status/archived index integrity after the parallel archival sweep                                          | S                                    | Low                                                 |
| 18                                                                                                                                          | Fold the `docs/drafts/*issue-body*` extract into the annotated draft (single artifact)                                 | S                                    | Low                                                 |
| 19                                                                                                                                          | Raise the daemon-side-guard open question with BuildFlow/pma, now with the 12-day-red evidence                         | M                                    | Low — fleet-level                                   |
| 20                                                                                                                                          | If pin endgame = 1.27: execute the prepared full doc sweep (list lives in ROADMAP)                                     | M                                    | Blocked by g-1                                      |
| 21 **→ resolved 2026-10-27: proxy + `go get` + consumer build verified at `v0.9.0`; pkg.go.dev indexing pending**                           | S                                                                                                                      | Blocked by release                   |                                                     |
| 22                                                                                                                                          | Revisit SECURITY.md wording once the SLA decision lands                                                                | S                                    | Blocked by g-2                                      |

## g) Questions I can NOT figure out myself

1. **Go-pin endgame:** hold `go 1.26` (and absorb a guard fight every time
   go-error-family ships a `go 1.27` floor), or re-pin to `1.27` deliberately
   with the full doc sweep? (standing question from the 20:43 report, g-1)
2. **SECURITY.md response-time numbers** ("acknowledged within two business
   days, triaged within five" — invented by the agent on a public doc): keep,
   adjust, or drop? (g-2)
3. ~~**v0.9.0 now or batched?** The last-attempt code promotion and the
   config-file removal are both user-visible and sitting in Unreleased.
   (g-3)~~ → resolved 2026-10-27 22:47: owner ordered the cut — published
   as `v0.9.0` (tag `b392900`, CI green on the tag commit).

_Report written 2026-10-27 22:08 CET. Waiting for instructions._
