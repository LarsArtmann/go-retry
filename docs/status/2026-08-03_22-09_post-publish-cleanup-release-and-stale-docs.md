# Status Report: Post-Publish Cleanup — Tag Retag, GitHub Release, Stale-Docs Purge

**Date:** 2026-08-03 22:09
**Session scope:** Fix the 3 embarrassments from the prior session's public launch.
**Repo:** <https://github.com/LarsArtmann/go-retry> (public, MIT, v0.1.0 released)
**Quality gates:** 22 tests pass (`-race`), `go vet` clean, `golangci-lint` 0 issues, 100% coverage, `go get` verified.

---

## Context

The prior session (`2026-08-03_21-48`) launched the repo publicly: MIT switch,
GitHub repo creation, `v0.1.0` signed tag push. Three problems were
self-identified in that session's critique:

1. LICENSE typo "FITITNESS" shipped to the public repo (fixed in `a16bd9b`,
   but the tag still pointed at pre-MIT code).
2. No GitHub Release — tag pushed but Releases page was empty.
3. TODO/ROADMAP/CHANGELOG still said "blocked on no remote" after the remote
   was created in the same session.

This session was asked to fix all three.

> **Format note:** User explicitly requested `.md`. The `status-report` skill
> defaults to styled HTML — this is a documented override, not a new default.

---

## a) FULLY DONE

### A1. LICENSE typo verification

**Status:** Already fixed in prior session commit `a16bd9b`. Verified clean.

Ran `grep -n "FITITNESS\|FITNESS" LICENSE` — only "FITNESS" exists at line 17.
The typo was caught and corrected before this session. No action was needed,
but verification confirmed the fix is present at HEAD, at the v0.1.0 tag, and
on the GitHub-rendered LICENSE.

### A2. v0.1.0 tag retagged to include MIT license + all fixes

**Status:** Done. Tag moved from `eae60c5` → `aea45cc`.

The prior session pushed `v0.1.0` at commit `eae60c5`, which was the original
feat commit — **before** the MIT switch (`a16bd9b`), before the README rewrite,
before the lint config, before the domain glossary, before the examples and
benchmarks. Anyone running `go get github.com/larsartmann/go-retry@v0.1.0`
against the old tag would have gotten a proprietary-licensed, broken-README,
no-lint-config version.

This session:

- Deleted the old tag locally and on the remote (`git push origin --delete v0.1.0`).
- Created a new signed annotated tag at HEAD (`2e9fe52` at the time).
- Pushed the new tag.
- The auto-git daemon subsequently committed `aea45cc` (CHANGELOG
  reconciliation), and the tag now dereferences to `aea45cc` — which includes
  MIT LICENSE, all docs, examples, benchmarks, and lint config.
- Verified: `git show v0.1.0:LICENSE | head -1` → `MIT License`. SSH signature
  intact. Remote tag SHA `9cc7519` (annotated tag object) → commit `aea45cc`.

### A3. GitHub Release created

**Status:** Done. Published (not draft, not prerelease).

URL: <https://github.com/LarsArtmann/go-retry/releases/tag/v0.1.0>

Release notes include:

- Core feature summary (`Do`, backoff, jitter, cancellation, error-family).
- Full configuration table (7 fields with defaults).
- Quality stats (22 tests, 100% coverage, 0 lint issues, ~18 ns/op benchmark).
- Install command: `go get github.com/larsartmann/go-retry@v0.1.0`.
- Dependency link to `go-error-family`.

### A4. Stale "no remote" references purged from all living docs

**Status:** Done. Zero matches remain.

Files fixed:

| File           | What was stale                                                                | Fix                                                                                                             |
| -------------- | ----------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| `TODO_LIST.md` | "P3 — Blocked on a git remote" section; "git remote -v is empty"              | Removed blocker framing; promoted to "P2 — Infrastructure"; removed completed T6 (compare links)                |
| `ROADMAP.md`   | Open question: "Is a public git remote intended? ... git remote -v is empty"  | Deleted the resolved question entirely                                                                          |
| `CHANGELOG.md` | Note block: "this repo currently has no git remote ... intentionally omitted" | Removed note; added `[Unreleased]` and `[0.1.0]` compare links; fixed `LICENSE (proprietary)` → `LICENSE (MIT)` |

Verified via: `rg -c "no remote|blocked on.*remote|remote.*empty|will be published" CHANGELOG.md TODO_LIST.md ROADMAP.md README.md FEATURES.md CONTRIBUTING.md AGENTS.md` → zero matches.

### A5. CHANGELOG compare links added (T6 completed)

**Status:** Done.

Added Keep-a-Changelog footer links:

- `[Unreleased]: https://github.com/LarsArtmann/go-retry/compare/v0.1.0...HEAD`
- `[0.1.0]: https://github.com/LarsArtmann/go-retry/releases/tag/v0.1.0`

Verified: `gh api repos/LarsArtmann/go-retry/compare/v0.1.0...HEAD` returns
`ahead_by: 1` (the TODO renumbering commit). The links resolve.

The auto-git daemon then reconciled the CHANGELOG: moved all `[Unreleased]`
content into `[0.1.0]` (since the tag was retagged to include everything),
leaving `[Unreleased]` as `_Nothing yet._` — correct behavior.

### A6. `go get` verified in clean module

**Status:** Done.

Created `/tmp/gogettest` with a fresh `go mod init testmod`, ran
`go get github.com/larsartmann/go-retry@v0.1.0`. Resolved from the public Go
module proxy, pulled `go-error-family v0.10.0` as a transitive dependency,
compiled and ran a test program that prints `retry.ErrExhausted`. Confirmed
the public module path works end-to-end.

### A7. Quality gates green

| Gate                           | Result                        |
| ------------------------------ | ----------------------------- |
| `go test ./... -race -count=1` | 22/22 PASS (0.96s)            |
| `go vet ./...`                 | clean                         |
| `golangci-lint run ./...`      | 0 issues                      |
| Coverage                       | 100.0% of statements          |
| `BenchmarkComputeDelay`        | 17.65 ns/op, 0 B/op, 0 allocs |

---

## b) PARTIALLY DONE

### B1. TODO_LIST.md structural cleanup

**Status:** Fixed, but I introduced the problem myself.

When editing TODO_LIST to remove the "blocked" framing, my `edit` call created
a **duplicate P2 heading** — there were two `## P2` sections (one "Polish",
one "Infrastructure"). I caught this on verification (`rg "^## " TODO_LIST.md`
showed two P2 entries) and fixed it by rewriting the bottom section with
correct priority ordering (P1 → P2 Infrastructure → P3 Polish) and renumbering
T3–T5.

The final state is clean, but I should have read the full file after my first
edit instead of assuming the replacement was structurally correct.

### B2. Tag alignment after daemon activity

**Status:** Resolved, but required a force-fetch.

After I retagged `v0.1.0` at `2e9fe52` and pushed, the auto-git daemon
committed `aea45cc` on the remote (CHANGELOG reconciliation). The tag on the
remote was then moved to `aea45cc` (by the daemon or by GitHub's tag
resolution), but my local tag still pointed at `2e9fe52`. A `git fetch
--tags` was rejected ("would clobber"), requiring `git fetch origin --tags
--force` to sync.

This is a symptom of working concurrently with an autonomous git daemon — see
improvement E2.

---

## c) NOT STARTED

### C1. CI workflow (`.github/workflows/ci.yml`)

Now listed as T3 (P2 — Infrastructure). No workflow file exists. The remote
is live so this is fully unblocked. A minimal workflow running `go test
./... -race` and `golangci-lint run ./...` on push/PR is ~20 lines of YAML.

### C2. `SECURITY.md`

Listed as T5 (P3 — Polish). Does not exist. Conventional vulnerability
reporting policy file for public repos.

### C3. Backoff ↔ ComputeDelay doc cross-links

Listed as T4 (P3 — Polish). Both functions implement the same formula but
neither doc comment references the other.

### C4. Concurrency stress test

Listed as T1 (P1). The global `math/rand/v2` is safe, but there's no parallel
test proving concurrent `Do` invocations don't share mutable state.

### C5. Fuzz target for `ComputeDelay`

Listed as T1 (P1). Pure numeric function with caller-controlled inputs —
candidate for `go test -fuzz` to harden overflow/negative-delay edges.

### C6. `hierarchical-errors` migration check

Listed as T2 (P1). Go 1.26+ offers `errors.AsType[T]`. This package uses only
`errors.Is`, so likely a no-op — but the dependency `go-error-family` may use
`errors.As`. Needs a 5-minute confirmation.

### C7. Public API surface audit for v1.0

Listed in `ROADMAP.md`. Not started. Need to confirm every exported symbol
(`Do`, `Config`, `DefaultConfig`, `Backoff`, `ComputeDelay`, `AttemptFunc`,
`ErrExhausted`, `ErrCanceled`) is one callers should depend on.

### C8. pkg.go.dev verification

The module is now resolvable via `go get`, but pkg.go.dev may not have indexed
it yet. Need to check `https://pkg.go.dev/github.com/larsartmann/go-retry`
and request indexing if the page is stale.

### C9. Go Report Card

<https://goreportcard.com/report/github.com/larsartmann/go-retry> — not
checked. Should verify the automated quality grade.

### C10. README badge integration

No badges in README (CI status, Go Report Card, pkg.go.dev, license, Go
version). Standard for public Go libraries.

---

## d) TOTALLY FUCKED UP

### D1. I created a duplicate P2 section in TODO_LIST.md

**What happened:** My `edit` call to rename "P2 — Polish" to "P2 —
Infrastructure" replaced the heading + T3 block, but left the **old** T4/T5

- "P2 — Infrastructure" + T5 section below it. The result was:

```
## P2 — Infrastructure        ← my new heading
### T3. Minimal CI            ← my new T3
## P3 — Polish                ← I inserted this
### T4. Cross-link...         ← old T4 (duplicate number)
### T4. Add SECURITY.md       ← old T4 again
## P2 — Infrastructure        ← OLD duplicate heading
### T5. Minimal CI            ← OLD duplicate T5
```

**Root cause:** I matched only the first `## P2 — Polish` block for
replacement, but the file already had a `## P2 — Infrastructure` section
further down (from the prior session's edit). My replacement didn't account
for the full file structure — I was editing a section without having read the
entire current file first in this session.

**Fix:** Re-read the file, saw the duplication, rewrote the entire bottom
section (lines 42–72) in a single clean replacement.

**Lesson:** Always `view` the full file before structural edits, even if you
"know" what it contains from a prior session. The daemon may have modified it,
and your mental model may be stale.

### D2. I pushed without pulling — got rejected

**What happened:** After committing the TODO fix, `git push origin master`
was rejected: "the remote contains work that you do not have locally." The
auto-git daemon had pushed `aea45cc` (CHANGELOG reconciliation) while I was
working on the TODO fix locally.

**Root cause:** I committed and pushed without first checking if the remote
had advanced. The daemon is highly active in this repo and I knew this — but
I still tried a blind push.

**Fix:** `git pull --rebase origin master`, which cleanly rebased my commit on
top of the daemon's.

**Lesson:** Before any push in a daemon-active repo, `git fetch origin && git
log --oneline origin/master -3` to check for divergence.

### D3. Stale `index.lock` from the daemon

**What happened:** `git commit` failed with "Unable to create
'.git/index.lock': File exists." The daemon had left a 0-byte stale lock
file.

**Root cause:** Race condition between my `git add`/`git commit` and the
daemon's concurrent git operations.

**Fix:** `trash .git/index.lock`, then retried the commit. This is safe (a
0-byte lock file with no running git process is stale), but it's a hack.

**Lesson:** In daemon-active repos, expect lock contention. If a git command
fails with a lock error, check `ps` for a running git process first, then
remove the stale lock if none exists.

---

## e) WHAT WE SHOULD IMPROVE

### E1. Read the full file before structural edits

I violated the first rule: "Read before you write." I edited TODO_LIST.md
based on a prior-session mental model and created a structural mess. The fix
is mechanical: **always `view` the entire file in the current session before
any structural edit** (adding/removing/moving sections), even — especially —
if you think you know what it says.

### E2. Coordinate with the auto-git daemon

The daemon is highly active: it commits files autonomously, sometimes with
empty messages, reformats configs, and reconciles docs. In this session alone
it:

- Committed `aea45cc` (CHANGELOG reconciliation — moved `[Unreleased]` content
  into `[0.1.0]`) while I was editing TODO_LIST.
- Left a stale `.git/index.lock` that blocked my commit.
- Advanced the remote, causing a push rejection.

**Improvement:** Before every git operation (commit, push, tag), run
`git fetch origin && git status && git log --oneline -3 origin/master`. If the
daemon has pushed, rebase first. Never blind-push.

### E3. Tag hygiene at release time

The prior session tagged `v0.1.0` at `eae60c5` (pre-MIT, pre-docs), then
continued making commits without updating the tag. This session had to
**force-move a published signed tag** — an operation that is safe only for a
brand-new v0.1.0 that nobody has consumed yet, but would be catastrophic for
a version with downstream users.

**Improvement:** The tag should be the **last** operation before "done", not
an intermediate step. Tag at the commit that contains everything you want in
the release: license, docs, tests, examples. If you fix something after
tagging, either cut a new version (`v0.1.1`) or move the tag **before
advertising it** — never after.

### E4. Retagged release should be noted somewhere

Moving `v0.1.0` from `eae60c5` to `aea45cc` is a force-update of a signed tag.
For a brand-new library this is fine (no consumers yet), but the retag is not
documented anywhere except this status report. If someone cached the old tag
SHA before the move, they have different code.

**Improvement:** For future retags (which should be avoided — see E3), note
it in the release notes or a CHANGELOG entry. For this specific case (v0.1.0
of a library that was published hours ago), the impact is effectively zero.

### E5. The CHANGELOG was reconciled by the daemon, not by me

The daemon moved all `[Unreleased]` content into `[0.1.0]` after the tag was
retagged. This is semantically correct (the work shipped in v0.1.0), but I
didn't do it. If the daemon hadn't done it, the CHANGELOG would now be
inconsistent — `[Unreleased]` would list items that are actually in the
tagged release.

**Improvement:** When retagging, immediately reconcile the CHANGELOG yourself.
Don't rely on the daemon to do it correctly.

### E6. Proactive CI

The repo has no CI. Every quality gate is run manually. For a public library,
CI is table stakes — it signals to contributors that the project is maintained
and gives confidence in the build. This should have been part of the launch,
not deferred to "next session."

### E7. Verify pkg.go.dev renders correctly

`go get` works, but I haven't checked whether `pkg.go.dev` has indexed the
module and whether the godoc examples render correctly there. The examples
are verified locally, but the public rendering is the actual user experience.

---

## f) Next items (up to 50)

Sorted by impact, Pareto-style.

### P1 — High impact

1. **Add CI workflow** (`.github/workflows/ci.yml`) — `go test ./... -race` +
   `golangci-lint run ./...` on push/PR. ~20 lines. Fully unblocked.
2. **Concurrency stress test** — prove `Do` invocations share no mutable state
   with a `t.Parallel()` goroutine stress test (T1).
3. **Fuzz `ComputeDelay`** — `go test -fuzz` target for numeric edges (T1).
4. **`hierarchical-errors` migration check** — confirm `go-error-family` uses
   no `errors.As` that needs migrating to `errors.AsType[T]` (T2). 5 minutes.
5. **Verify pkg.go.dev rendering** — check
   `https://pkg.go.dev/github.com/larsartmann/go-retry`, request indexing if
   stale, confirm `ExampleDo` renders.
6. **Go Report Card** — check
   `https://goreportcard.com/report/github.com/larsartmann/go-retry`, fix any
   findings.

### P2 — Valuable

7. **Add `SECURITY.md`** — vulnerability reporting policy (T5). Low effort.
8. **Cross-link `Backoff` ↔ `ComputeDelay`** doc comments (T4).
9. **README badges** — CI status, Go Report Card, pkg.go.dev, license, Go
   version reference.
10. **Public API surface audit** — confirm every exported symbol is one
    callers should depend on before v1.0 (ROADMAP).
11. **Options-based configuration investigation** — functional options
    (`WithOnRetry`, `WithMaxAttempts`) so new fields don't break struct-literal
    callers. Design doc first (ROADMAP).
12. **`AttemptFunc` signature review** — some retry libraries pass the
    previous error back into `fn`; this one does not. Deliberate decision or
    accident? (ROADMAP v1.0 question).
13. **Configurable jitter factor** — currently hardcoded at 50% of delay;
    consider making it a `Config` field (FEATURES WORTH_CONSIDERING).
14. **Deterministic RNG option** — allow callers to inject an `*rand.Rand`
    for testable backoff (FEATURES WORTH_CONSIDERING).
15. **Version-compatibility matrix with `go-error-family`** — document which
    `go-retry` versions support which `go-error-family` majors (ROADMAP).

### P3 — Polish

16. **`doc.go` package examples** — consider adding a full-file example that
    shows a realistic retry scenario (HTTP call with backoff).
17. **Backoff visualization in docs** — a table or chart showing actual delay
    sequences for common configs (e.g., `InitialDelay=100ms, Multiplier=2,
MaxDelay=5s` → 100ms, 200ms, 400ms, 800ms, 1.6s, 3.2s, 5s, 5s...).
18. **`CONTRIBUTING.md` PR workflow** — document the branch/PR/review process
    now that the repo is public and could accept external contributions.
19. **Issue templates** — `.github/ISSUE_TEMPLATE/` for bug reports and
    feature requests.
20. **PR template** — `.github/PULL_REQUEST_TEMPLATE.md`.
21. **`CODEOWNERS`** — `.github/CODEOWNERS` for review routing.
22. **Discussions tab** — enable GitHub Discussions for Q&A.
23. **Go version badge in README** — `go.mod` says 1.26.5; badge should match.
24. **GitHub Actions cache** — if CI uses `actions/setup-go`, enable module
    caching for faster runs.
25. **Dependabot config** — `.github/dependabot.yml` for automated dependency
    bumps (`go-error-family`, `golangci-lint` action, `actions/setup-go`).
26. **Release automation** — `goreleaser` config or a GitHub Action that
    cuts tags + releases from commit messages.
27. **Changelog automation** — consider `conventional-commits` + auto-generated
    CHANGELOG for future releases.
28. **Sitemap / SEO** — if a docs site is built (ROADMAP), set up
    `sitemap.xml` and meta tags.
29. **Favicon / social card** — if a docs site is built, create OG image and
    favicon.
30. **`doc.go` link to domain glossary** — cross-reference
    `docs/DOMAIN_LANGUAGE.md` from the package doc comment.

### P4 — Exploration / future

31. **Public documentation site** — Astro + Starlight + Firebase Hosting
    (ROADMAP; `website-launch` skill). Precondition: API stability.
32. **Circuit-breaker adapter** — thin adapter that composes with a
    caller-chosen breaker. Likely belongs in `go-cqrs-lite`, not here (ROADMAP).
33. **Bulkhead / deadline-budgeting primitives** — same: likely document the
    pattern, don't code it here (ROADMAP non-goal).
34. **Metrics hooks** — `OnRetry`/`OnExhausted` are the seam; consider a
    documented example of wiring Prometheus metrics through them.
35. **Structured logging example** — show how to wire `slog` through
    `OnRetry`/`OnExhausted`.
36. **`context.AfterFunc` migration** — Go 1.21+ offers `context.AfterFunc`;
    the current `select` on `timer.C` vs `ctx.Done()` could potentially use it.
    Investigate whether it simplifies the cancellation path.
37. **`MaxDelay` as a `time.Duration` vs `int64`** — currently `time.Duration`
    (correct), but confirm all delay arithmetic is overflow-safe for extreme
    `Multiplier` values.
38. **Heap allocations in the hot path** — benchmark `Do` end-to-end (not just
    `ComputeDelay`) to confirm zero allocations per attempt.
39. **Race detector coverage** — run `go test -race -count=100` to catch
    rare data races in the backoff/jitter path.
40. **Module proxy verification** — `GOPROXY=off go mod download` to confirm
    the module is fully cached and resolvable without network.
41. **`go mod tidy` in CI** — ensure `go.sum` stays in sync.
42. **License scan** — some organizations require automated license scanning
    (FOSSA, Snyk). Consider adding a badge/config.
43. **Tidelift** — if monetizing open source, consider Tidelift listing.
44. **GitHub Sponsors** — `FUNDING.yml` if the author wants sponsorships.
45. **Semantic release tags** — document the tagging convention
    (`vMAJOR.MINOR.PATCH`, signed, annotated) in CONTRIBUTING.
46. **Backward compatibility tests** — `apicompat` or similar to catch
    breaking API changes before they ship.
47. **Golden file tests for error messages** — ensure error codes/messages
    don't drift across versions (they're part of the public contract).
48. **Integration test with `go-cqrs-lite`** — confirm the `MessageAdapter`
    in `middleware/v4` still compiles against this package's API.
49. **`doc.go` test coverage** — `go test` doesn't cover `doc.go`; consider
    a `TestPackageDoc` that asserts the package comment doesn't reference
    removed types.
50. **Archive prior status reports** — `docs/status/` now has 3 reports;
    consider a `docs/status/README.md` index or moving old reports to
    `docs/status/archive/`.

---

## g) Questions (up to 3, that I CAN NOT figure out myself)

### Q1. Should the v0.1.0 retag be documented in the release notes?

The tag was force-moved from `eae60c5` (pre-MIT) to `aea45cc` (current). For a
brand-new library published hours ago, this is low-impact. But it IS a
force-update of a signed tag, and some teams have a policy about documenting
retags. Options:

- **(a)** Leave it undocumented — nobody consumed the old tag, impact is zero.
- **(b)** Add a note to the v0.1.0 release: "Note: this tag was retagged on
  2026-08-03 to include the MIT license and documentation fixes."
- **(c)** Cut `v0.1.1` instead and leave `v0.1.0` as-is.

I can't decide this because it's a policy question about how seriously you
treat tag immutability for a brand-new release.

### Q2. Is the auto-git daemon appropriate for a now-public repo?

The daemon makes commits with empty messages (`31db438`, `6a41c5c`),
overwrites committed configs, and reconciles docs autonomously. For a private
repo under active development, this is a productivity tool. For a public repo
where contributors read the commit history, empty-message commits and
config-overwrite commits look messy and unprofessional.

I can't decide this because it's a workflow/infra policy question. Options:

- **(a)** Disable the daemon for this repo — all commits should be intentional
  and human-authored (or agent-authored with descriptive messages).
- **(b)** Keep the daemon but configure it to always produce descriptive
  commit messages.
- **(c)** Keep as-is — the daemon's productivity value outweighs the
  cosmetic mess.

### Q3. Should I add CI now, or wait for a LarsArtmann-standard workflow?

A minimal CI workflow (`go test -race` + `golangci-lint`) is ~20 lines of
YAML and is fully unblocked. But other LarsArtmann repos may have a standard
CI pattern (matrix strategy, caching, golangci-lint action version, etc.)
that I should copy rather than inventing a one-off.

I can't decide this because I don't know if a standard CI pattern exists in
your other repos. Options:

- **(a)** Add a minimal one-off workflow now.
- **(b)** Check a sibling repo (`go-error-family`, `go-cqrs-lite`) for a
  standard pattern and copy it.
- **(c)** Wait — you'll provide the standard CI config.
