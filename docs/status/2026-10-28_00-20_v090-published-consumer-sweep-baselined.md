# Status 2026-10-28 00:20 — v0.9.0 published end-to-end, consumer sweep baselined

Session scope: the owner answered the standing g-3 question with one
instruction — "Publish the new version!" — and this session executed exactly
that: the full `go-release` lifecycle for **v0.9.0** (the first release since
v0.8.0, 2026-10-08), followed by the start of the T77 consumer sweep via
`go-ecosystem-upgrade`. This report covers only that run and what was noticed
along the way.

## a) FULLY DONE

1. **Release assessment and version determination.** 43 commits since v0.8.0;
   the `[Unreleased]` CHANGELOG carried two user-visible changes (terminal
   errors now surface the last attempt's error-family code —
   `retry.go` (`promoteLastAttemptCode`); the `go 1.26` pin / go-error-family
   v0.11.0 restore) plus repo-internal entries. Verdict: MINOR bump —
   v0.9.0, matching T77. Confirmed no CHANGELOG drift for v0.8.0 (footer
   compare links complete).
2. **CHANGELOG cut.** `[Unreleased]` → `[0.9.0] - 2026-10-27`; empty
   `[Unreleased]` placeholders restored in this repo's own Added/Changed/Fixed
   format (verified against the v0.8.0 cut); footer links updated
   (`[Unreleased]: compare/v0.9.0...HEAD`, `[0.9.0]: compare/v0.8.0...v0.9.0`).
   Landed in daemon commit `b392900`.
3. **go.mod hygiene (release Phase 3).** No `replace` directives, no
   `00010101` pseudo-version sentinels, module path correct,
   `go mod tidy` + `go mod verify` clean with zero diff; `go 1.26` pin and
   go-error-family v0.11.0 intact in the tagged tree (verified via
   `git show v0.9.0:go.mod`).
4. **Full pre-tag gate battery, all green:** `gofmt -l .` clean;
   `go vet ./...` + `go -C tools vet ./...`; `go test ./... -race -count=10`
   (ok, 21.4 s); `golangci-lint run ./...` — 0 issues; coverage
   `go test -cover ./...` — 100.0 % of statements (CI floor 95 %);
   `./scripts/check-docs.sh` green twice (before and after the annotation
   edits). Note: the session-long LSP warnings on `retry_test.go`
   (golines/gci "not properly formatted") were **stale** — the actual
   golangci-lint run and `gofmt` both clean.
5. **Release-time govulncheck, root + `tools/`.** First scan under the
   default local toolchain go1.26.8 reported 5 stdlib vulnerabilities (all
   `net/http`, all "Fixed in: go1.26.9") — toolchain-level, not
   module-level. Re-ran with `GOTOOLCHAIN=go1.26.9`: **0 vulnerabilities
   affecting code** in both modules (1 not-called informational each). The
   `go 1.26` pin was not touched (any 1.26.x compiles it; CI's
   govulncheck-action resolves the latest 1.26.x from `go-version-file:
   go.mod`). govulncheck + actionlint reinstalled from the `tools/` module
   pins after one transient install failure.
6. **Tag and push, with the daemon race handled.** The daemon committed the
   CHANGELOG cut plus the parallel docs-health leftovers into `b392900`
   ("auto-commit 25 changed file(s)") and pushed master before I could;
   CI run `40586981343` on exactly `b392900`: **success** (49 s). Annotated
   tag `v0.9.0` created at `b392900` (tag object `16e332f`; message in the
   v0.8.0 house style), `git tag --points-at` and tagged-tree `go.mod`
   verified, then pushed.
7. **Compare-links chicken-and-egg solved by ordering.** The doc battery was
   run **after** creating the local tag (`check-compare-links.sh` requires
   both tags of each pair to exist locally); all tag pairs resolved.
8. **Post-push verification (Phase 6).** `proxy.golang.org` serves
   `v0.9.0.info` with `Origin.Hash` exactly `b392900b34ed…`; clean-module
   `go get github.com/larsartmann/go-retry@v0.9.0` resolves and pulls
   go-error-family v0.11.0 with no checksum errors; a consumer smoke program
   in a fresh `/tmp/release-verify` module imports `retry`, builds, and runs
   (prints the `ErrExhausted` sentinel). CI green on the tag commit.
9. **GitHub Release v0.9.0** created (`--latest`, matching v0.8.0 precedent,
   not `--prerelease`): headline = last-attempt code promotion, then pin
   restore, BuildFlow-config removal with the upstream issue link, art-dupl
   internal note. https://github.com/LarsArtmann/go-retry/releases/tag/v0.9.0
10. **Annotate-as-you-land (Phase 8).** The 22:08 report's `§f` rows 3, 4, 21
    and `§g` question 3 resolved inline (`v0.9.0` tagged at `b392900`);
    `TODO_LIST.md` T77 rewritten to the sweep-only remainder with
    release-done evidence; dprint-formatted; doc battery re-run green;
    daemon committed as `2b29b72`.
11. **Consumer sweep started per `go-ecosystem-upgrade`.** Phase 0–1
    complete: middleware module pins `go-retry v0.8.0` + go-error-family
    v0.12.0 on `go 1.27.2` (MVS will keep v0.12.0; compatible direction,
    upgrade); the committed go.work `use` list re-confirmed to NOT include
    go-retry (so `GOWORK=off go get` in `middleware/` is the right bump
    shape, never a temp go.work); no `replace` in middleware/go.mod (one
    stale trailing comment "flightrecorder is not yet tagged" noticed,
    deliberately not touched — not this bump's scope); retry usage is
    `middleware/retry.go` (`NewRetry`, `MessageAdapter`) and **no**
    `retry.exhausted`/`ErrExhausted`/`errorfamily.Code` assertions exist in
    middleware or the three ritual suites — the v0.9.0 code-promotion
    behavior change is assessed low-risk for this consumer.
    **Baseline green:** `go build ./...` + `go test -race ./...` in
    `middleware/` — ok, 4.515 s (go-cqrs-lite at `654d2f530`).

## b) PARTIALLY DONE

1. **T77 consumer sweep — baselined, bump not yet applied.** Done: skills
   loaded, consumer mapped, risk assessed, baseline green (see a.11).
   Remaining: `GOWORK=off go get github.com/larsartmann/go-retry@v0.9.0` in
   `middleware/` (then grep go.mod — never trust `go get` output), tidy if
   needed, the four ritual suites with `-race` (`middleware`,
   `commandlifecycle`, `integration`, `example/taskmanager`) through the
   committed go.work, commit in go-cqrs-lite, ROADMAP bump-trace row here,
   close T77. Blocker: none — the owner's report-and-wait order interrupted
   mid-protocol. Effort: M (suites are the long pole; baseline took seconds,
   integration may not).
2. **pkg.go.dev documentation page — triggered, not confirmed.** Both
   `pkg.go.dev/fetch/...@v0.9.0` and the version page returned 404 at
   ~23:50 CET (indexing lag is normal; the proxy had indexed within ~2 min of
   the tag push). Remaining: one bounded re-check that the page renders.
   Not blocking: proxy + clean-dir `go get` + consumer build already prove
   the publish. Effort: S.

## c) NOT STARTED

1. **go-paperless adoption of v0.9.0.** Named in T77's rationale as awaiting
   the code-promotion fix; not inspected this session (explicitly out of the
   "don't research unrelated stuff" scope). Priority: owner call — see g.2.
2. **The rest of the open backlog** (T78–T94, the 22:08 report's remaining
   `§f` rows 5–20, ROADMAP open questions g-1/g-2). Untouched this session by
   design; the live `TODO_LIST.md` remains the source. Priority: as ranked
   there.
3. **HARVEST of this report's §f** into TODO_LIST/ROADMAP — deliberately not
   done pending instruction (same posture as the 22:08 report).

## d) TOTALLY FUCKED UP

Nothing destructive shipped — the release is verified sound at every layer.
The honest fuckups, all recovered:

1. **First govulncheck scan ran under the stale local toolchain (go1.26.8)**
   and produced a release-blocking-looking report of 5 stdlib CVEs. Root
   cause: I reached for the scan before asking "which toolchain will CI
   scan with". Cost: one alarm cycle and a background re-scan. Mitigation
   used: `GOTOOLCHAIN=go1.26.9` override. The real fix is process: see e.1.
2. **Transient `go -C tools install` failure** ("no required module provides
   package") on the first govulncheck install, plus one shell shape
   (`go install … 2>&1; echo exit=$?`) blocked by the permission hook.
   Retried with the AGENTS-documented combined install form — worked. No
   damage; ~3 min lost.
3. **Process slip: attempted the TODO_LIST edit before Viewing the file**
   (the edit tool correctly refused). Re-read, then edited. No damage, but it
   is exactly the class of shortcut the repo rules exist to prevent.
4. **Release-order luck, not control.** My plan was commit → push master →
   watch CI → tag. The daemon pushed master inside that window; I verified
   CI green on `b392900` _after the fact_ and before pushing the tag. The
   safety check held (CI was green on the exact tag commit), but the
   sequence was daemon-shaped, not deliberate. If the daemon's push had been
   red, tagging would have depended on catching it in that verification.
5. **`buildflow --build-mode dev` was not run this session** (it is in the
   AGENTS Session Ritual; the previous session ran it green). The Go/doc/CI
   batteries covered the tagged content, and post-`.buildflow.yml`-removal
   the default pipeline is the surface that matters — but the ritual was not
   executed in full, and I only noticed while writing this report.
6. **pkg.go.dev step left dangling** (see b.2) — I moved on to the sweep
   instead of bounding the wait; it remains an unverified claim in the
   release checklist.

## e) WHAT WE SHOULD IMPROVE

1. **Release-time toolchain policy for govulncheck.** The local toolchain
   trails the patch train (1.26.8 vs 1.26.9 today); a release scan under it
   manufactures stdlib CVE noise. Fix: pin the release ritual to scan under
   the latest 1.26.x (`GOTOOLCHAIN=go1.26.<latest> govulncheck ./...` root +
   `tools/`), or update the local toolchain at release time. Cheap, prevents
   a repeat of d.1.
2. **Pin the discovered release ordering in AGENTS.md.** This session had to
   derive: cut CHANGELOG → Go battery → commit (daemon) → **local annotated
   tag** → `check-docs.sh` (compare-links needs the local tag) → verify CI
   green on the tag commit → push tag → proxy/`go get` checks → GitHub
   release → annotate-as-you-land. None of that order (especially the local
   tag before the doc battery) is written down. One paragraph in the Session
   Ritual closes it.
3. **Consumer inventory for go-retry.** The sweep's Phase 2 enumeration was
   from memory + T77 (middleware today, go-paperless claimed). A short
   "known consumers" list in AGENTS.md or ROADMAP would make future sweeps
   complete by construction instead of by recollection.
4. **pkg.go.dev verification with a bounded retry** (or drop the step — the
   proxy and `go get` already prove the publish). Today it is a 404 and a
   shrug.
5. **LSP noise protocol.** Stale golines/gci diagnostics on `retry_test.go`
   persisted the entire session against a clean `golangci-lint run`. Either
   restart the LSP when diagnostics contradict a green CLI run, or note in
   AGENTS that golangci-lint CLI is the truth and the LSP view is advisory.
6. **Split bundled TODO rows at harvest time.** T77 bundled the release and
   the sweep; the owner's "publish" instruction then legally covered only
   half of one row. Bundles make scoping requests ambiguous — harvest
   should split them (the 22:08 harvest kept it bundled; this session had to
   re-scope mid-task).
7. **go-cqrs-lite housekeeping (their repo, noticed during baseline):** the
   stale trailing comment "Workspace-local replace. flightrecorder is not
   yet tagged" in `middleware/go.mod` describes a replace that no longer
   exists. One-line cleanup for a session in that repo.
8. **Persist release notes in-repo.** The GitHub notes body lives only on
   GitHub (composed from CHANGELOG at release time by policy — fine), but
   the /tmp drafts are ephemeral; if the policy ever wants an audit trail,
   a `docs/drafts/` release-notes file per cut is the pattern already used
   for issue bodies.

## f) Things to get done next (ranked)

| #  | Task                                                                                                                              | Size | Impact / note                                        |
| -- | --------------------------------------------------------------------------------------------------------------------------------- | ---- | ---------------------------------------------------- |
| 1  | Execute the middleware bump: `GOWORK=off go get go-retry@v0.9.0` in `middleware/`, grep go.mod after, tidy if needed              | S    | High — the sweep's core step; baseline already green |
| 2  | Run the four ritual suites `-race` (middleware, commandlifecycle, integration, example/taskmanager) through the committed go.work | M    | High — the sweep's verification gate                 |
| 3  | Land the sweep commit in go-cqrs-lite (daemon or explicit per owner) and append the ROADMAP bump-trace row here                   | S    | High — closes T77's remainder                        |
| 4  | Re-check the pkg.go.dev page for v0.9.0 (bounded retry); record verdict in the 22:08 report row 21 annotation if it needed more   | S    | Medium — closes b.2                                  |
| 5  | Owner: extend the v0.9.0 sweep to go-paperless / other consumers? (g.2)                                                           | S    | High if yes                                          |
| 6  | Write the release-ordering paragraph into AGENTS.md Session Ritual (e.2)                                                          | S    | Medium — institutionalizes today's discovered order  |
| 7  | Add the release-time govulncheck toolchain rule to AGENTS.md (e.1)                                                                | S    | Medium — prevents d.1's noise cycle                  |
| 8  | Consumer inventory list for go-retry in AGENTS.md/ROADMAP (e.3)                                                                   | S    | Medium                                               |
| 9  | HARVEST this report's §f + the 22:08 report's rows 5–20 via docs-health (verify-before-routing)                                   | S    | Medium — the standing deferred HARVEST               |
| 10 | Decide `TestBuildflowConfigStaysAbsent` guard (22:08 row 10; decision already recorded in ROADMAP)                                | S    | Medium                                               |
| 11 | T85 remainder: file the BuildFlow dry-mode `.golangci.yml`-write issue with a clean repro (verify-before-filing first)            | M    | High — unexplained dry-run writes erode trust        |
| 12 | Watch BuildFlow#77; on fix: verify at source, strip interim-exposure notes from the docs                                          | S    | High — closes the go-pin interim exposure            |
| 13 | T78: CI red-master alarm (required checks / watchdog) — 12-day red master proven real                                             | M    | High                                                 |
| 14 | Owner: go-pin endgame (hold `go 1.26` vs deliberate 1.27 re-pin sweep) — standing g-1                                             | S    | Critical for posture, gates #20-class work           |
| 15 | Owner: SECURITY.md SLA numbers — keep/adjust/drop — standing g-2                                                                  | S    | Critical (public doc, agent-invented numbers)        |
| 16 | T81: `check-compare-links.sh` as a CI lint-job step (fail the push, not the release)                                              | S    | Medium                                               |
| 17 | T79: table-driven self-tests for the docs-gate scanner edges                                                                      | M    | Medium                                               |
| 18 | T80: extend marker gating to `docs/planning/archived/*.md`                                                                        | S    | Medium                                               |
| 19 | T54: guard that living docs' stated Go version matches go.mod                                                                     | S    | Medium                                               |
| 20 | T83: re-run `BenchmarkComputeDelay`, refresh FEATURES numbers                                                                     | S    | Medium                                               |
| 21 | T92: dead-directive experiment (flip one accept-hash byte, watch art-dupl)                                                        | S    | Medium                                               |
| 22 | T94: decide warnings-gating stance for clone-hash drift                                                                           | S    | Medium                                               |
| 23 | T42: Dependabot watchlist — execute the /tools bump flow end-to-end on the next actions PR                                        | S    | Low                                                  |
| 24 | T51: safe probe fixtures for guard drills (gitignored scratch dir)                                                                | S    | Low                                                  |
| 25 | T52: index/backlog conventions for planning files                                                                                 | S    | Low                                                  |
| 26 | T53: sweep archived plans for dangling `done at` hashes                                                                           | S    | Low                                                  |
| 27 | ROADMAP: jitter strategies Full/Equal/Decorrelated (library feature work)                                                         | L    | Medium — the declared follow-ups                     |
| 28 | go-cqrs-lite cleanup: remove the stale flightrecorder replace comment in middleware/go.mod (their repo)                           | S    | Low                                                  |
| 29 | Watch consumer Dependabots for go-retry v0.9.0 PRs (fleet propagation signal)                                                     | S    | Low                                                  |
| 30 | Consider LSP-vs-CLI truth rule in AGENTS.md (e.5)                                                                                 | S    | Low                                                  |

## g) Questions I can NOT figure out myself

1. **Local toolchain policy:** may I update this machine's Go to the latest
   1.26.x (1.26.9 today) so release-time scans run clean by default — or
   should the `GOTOOLCHAIN` override stay the ritual (e.g. because the nix
   or system install is intentionally frozen)? I scanned correctly with the
   override, but I cannot know whether the local 1.26.8 is pinned on
   purpose.
2. **Sweep scope:** T77's rationale names go-paperless as awaiting this
   fix. Should the v0.9.0 adoption sweep extend beyond go-cqrs-lite's
   `middleware/` — to go-paperless and any other consumers you know of —
   and if a repo is not on this machine, where does it live? (I did not
   search the fleet, per your no-unrelated-research order.)
3. **Sweep commit authorship in go-cqrs-lite:** let that repo's auto-commit
   daemon absorb the middleware pin bump (simplest, matches how this repo
   works), or make an explicit reviewed commit there? And does the bumped
   middleware pin need a middleware-module release tag on their side for
   you to consider T77 "done", or is the on-master pin enough?

_Report written 2026-10-28 00:20 CET. Waiting for instructions._
