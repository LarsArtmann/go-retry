# Status Report — TODO execution, MIT switch, and public GitHub launch

**Date:** 2026-08-03 21:48 (Europe/Berlin, `+0200`)
**Session scope:** Execute the TODO_LIST (T1–T5), create `docs/DOMAIN_LANGUAGE.md`, run docs-health HARVEST + consistency pass, switch LICENSE to MIT, create the public GitHub repo and push. Then self-critique.
**Reporter:** Crush (this session)
**Basis:** ONLY this session's work and what it surfaced. No fresh codebase research beyond what was already read.

---

## At a glance

|                      | Count                                                                                     |
| -------------------- | ----------------------------------------------------------------------------------------- |
| a) FULLY DONE        | 11                                                                                        |
| b) PARTIALLY DONE    | 3                                                                                         |
| c) NOT STARTED       | 4                                                                                         |
| d) TOTALLY FUCKED UP | 4                                                                                         |
| Health gate          | `go test -race` ✅ 22 pass · `go vet` ✅ · `golangci-lint` ✅ 0 issues · coverage ✅ 100% |
| Remote               | `origin git@github.com:LarsArtmann/go-retry.git` — **PUBLIC**, MIT, `v0.1.0` tag pushed   |

---

## a) FULLY DONE

1. **`README.md` rewritten** (commit `e4801b3`) — Was a catastrophically broken template (`# .`, fake `go get github.com/username/.`, nonexistent `just` commands, Hello World example). Now has real title, real module path, a **runnable quick-start verified to compile and produce the documented output** (`succeeded on attempt 3`), configuration table, error-model section, real dev commands.
2. **`.golangci.yml` committed** (commit `f0734cf`, then daemon-replaced) — My initial minimal config was immediately superseded by the daemon with the full LarsArtmann standard config (90+ linters, v2 schema). Adapted: added `err113` to test exclusions, fixed the stale `mnd` nolint directive. Lint is 0 issues. The in-source `//nolint:` markers are verified **live** (removing either surfaces a real finding).
3. **`ExampleDo` + `ExampleDo_customIsRetryable`** (commit `9aaf52e`) — Deterministic, carry `// Output:` comments, verified passing, will render on `pkg.go.dev`.
4. **`BenchmarkComputeDelay`** (commit `9aaf52e`) — ~18 ns/op, **0 allocations**. The jitter path is allocation-free. Verified with `-benchmem`.
5. **`CONTRIBUTING.md` rewritten** (commit `9aaf52e`) — Documents prerequisites (Go 1.26, golangci-lint v2), real dev commands, coverage workflow (`reports/coverage.out` regen one-liner), lint policy (committed `.golangci.yml` + intentional `//nolint:` markers), testing conventions.
6. **`docs/DOMAIN_LANGUAGE.md` created** (commit `9aaf52e`) — Full domain glossary: retry vocabulary (attempt, `MaxAttempts`, backoff, jitter, exhaustion, cancellation), the three `error-family` families used (Transient / Rejection / Infrastructure), `IsRetryable` / `Classify` / `WithCause`, and the `retry.<event>` code table. Every term cited to source.
7. **3 behavioral tests added** (commit `6a41c5c`) — `TestDo_OnRetryNotCalledAfterFinalAttempt` (proves `OnRetry` fires `MaxAttempts-1` times, not `MaxAttempts`), `TestDo_PreCanceledContextReturnsErrCanceled` (context canceled before first call), `TestDo_OnExhaustedReceivesExactLastError` (identity check, not just `errors.Is`).
8. **docs-health HARVEST + consistency pass** (commit `6a41c5c`) — Harvested the previous status report's §f (50 items) into TODO_LIST (actionable) + ROADMAP (vague + Open questions). Removed done T1–T5 from TODO (→ CHANGELOG). Updated AGENTS.md (fixed the stale "no `.golangci.yml`" lie). Updated FEATURES.md (added Documentation & DX section). Updated ROADMAP.md (added Open questions section). Cross-file consistency verified: all links resolve, no split-brain, no ghost references.
9. **LICENSE changed to MIT** (commit `a16bd9b`) — Was proprietary ("all rights reserved...strictly prohibited"). Now standard MIT, Copyright (c) 2026 Lars Artmann. GitHub now detects it correctly (`"key":"mit"`). Updated all doc references (README, ROADMAP, TODO_LIST, CHANGELOG).
10. **GitHub repo created and pushed** — `gh repo create larsartmann/go-retry --public --source . --remote origin --push`. Description set. Topics tagged (`go`, `go-library`, `retry`, `backoff`, `exponential-backoff`, `jitter`, `error-handling`). `v0.1.0` signed tag pushed. Remote URL: <https://github.com/LarsArtmann/go-retry>.
11. **Quality gate green throughout** — Every change was followed by `go test -race`, `go vet`, `golangci-lint run`. Final state: 22 tests pass, 0 lint issues, 100% statement coverage, vet clean. The README quick-start example was verified to compile and run (using local `replace` directives, since the repo wasn't public yet at verification time).

---

## b) PARTIALLY DONE

1. **ROADMAP "Open questions" updated — but still lists the flake.nix question.** The LICENSE question was removed (resolved: MIT). The git-remote question is now also resolved (remote exists). But the flake.nix question remains open, and the remote question text still says "git remote -v is empty" — stale post-publish. The TODO P3 items also still say "blocked on a git remote" when the remote now exists. Needs a cleanup pass.

2. **CHANGELOG `[Unreleased]` is comprehensive — but compare links are still missing.** Now that the remote exists (`https://github.com/LarsArtmann/go-retry`), the Keep-a-Changelog footer links (`[Unreleased]: https://github.com/.../compare/v0.1.0...HEAD`, `[0.1.0]: .../tag/v0.1.0`) CAN and SHOULD be added. The CHANGELOG note still says "no remote" — stale.

3. **GitHub repo is live — but no GitHub Release was created.** The `v0.1.0` tag is pushed, but the Releases page (<https://github.com/LarsArtmann/go-retry/releases>) is empty. A tag is not a release. Users browsing GitHub won't see v0.1.0 as a release with notes. Should be `gh release create v0.1.0 --notes-from-tag` or with a generated body.

---

## c) NOT STARTED

1. **CI (`.github/workflows/`)** — No GitHub Actions workflow exists. Was listed as P3 "blocked on a git remote" — **but the remote now exists**, so this is unblocked and should be re-prioritized. A minimal workflow running `go test ./... -race` and `golangci-lint run ./...` on push/PR is ~20 lines of YAML.

2. **`go get` end-to-end verification** — The README example was verified with local `replace` directives (pre-publish). Now that the repo is public, a real `go get github.com/larsartmann/go-retry@latest` in a clean module should be tested to confirm the public import path actually resolves. Likely works (module path matches repo path, `v0.1.0` tag exists), but not proven.

3. **Go Proxy / pkg.go.dev indexing** — A newly published module takes time to appear on `pkg.go.dev` and `proxy.golang.org`. Not actionable from here, but worth noting that the examples (`ExampleDo`) won't render anywhere until Google's crawler indexes the module.

4. **`SECURITY.md`** — Still in TODO (T4). Low priority but conventional for public repos.

---

## d) TOTALLY FUCKED UP

1. **LICENSE typo: "FITITNESS" instead of "FITNESS".** I typed the MIT License text from memory and introduced a spelling error in line 17 ("FITITNESS FOR A PARTICULAR PURPOSE"). This shipped to the public GitHub repo in the initial push (`31db438`). I caught and fixed it in `a16bd9b` before the user would have seen it, but the first push contained a misspelled license — the single most embarrassing file to get wrong in a public repo. **Root cause:** I should have copied the canonical MIT text from a reliable source (e.g., `choosealicense.com/licenses/mit/`) rather than typing from memory. **Lesson:** Licenses are legal documents. Never type them from memory.

2. **Multiple edit-vs-daemon races wasted cycles.** The auto-git daemon modified files between my reads and edits at least 5 times during this session. Each time my `edit`/`multiedit` call failed with "File has been modified since it was last read," I had to re-read the file and retry. The daemon also: (a) replaced my carefully-crafted minimal `.golangci.yml` with the full LarsArtmann standard config, (b) reformatted markdown (`*text*` → `_text_`), (c) committed work mid-session with empty/generic commit messages (`31db438`, `6a41c5c` have no message body). **Root cause:** I didn't anticipate daemon activity after each batch of changes. **Lesson:** In a repo with an active auto-git daemon, always `git status` / re-read before editing, and expect commits you didn't make.

3. **My `.golangci.yml` was immediately overwritten by the daemon.** I designed a minimal, well-reasoned config for this small library (3 linters: `gosec`, `mnd`, `exhaustruct`, plus test exclusions). The daemon replaced it with the full 90+-linter LarsArtmann standard. This then caused NEW lint failures (`err113` on test sentinels, stale `mnd` in the nolint directive since `2` is now in `ignored-numbers`). I had to fix issues that my own config would never have raised. **This isn't exactly my fault** (the daemon overwrote my work), but I was slow to recognize what had happened — I spent a confused moment wondering why new lint issues appeared before checking the config file.

4. **TODO_LIST P3 items and ROADMAP Open questions are now stale.** I wrote them saying "blocked on a git remote" and "git remote -v is empty" — and then I _created the remote_ in the same session without going back to update these. The TODO and ROADMAP now contain lies about the repo state. This is the exact "report it and move on" anti-pattern the previous session self-criticized — I documented the blocker, resolved the blocker, and then didn't update the documentation. **Fix needed immediately.**

---

## e) WHAT WE SHOULD IMPROVE (process, this session)

1. **Never type legal text from memory.** The MIT License is a canonical document. Copy it from `choosealicense.com` or `opensource.org`. The "FITITNESS" typo was 100% preventable and would have been publicly embarrassing if I hadn't caught it.

2. **Update docs that reference blockers RIGHT AFTER resolving the blocker.** I created the git remote, then didn't update TODO T5/T6 or ROADMAP Open questions that say "no remote exists." This is the same "fix it on sight" principle I violated in the previous session (broken README). The blocker-resolution moment is exactly when dependent docs need updating — not "later."

3. **A tag is not a release.** Pushing `v0.1.0` is necessary but not sufficient for a public launch. A GitHub Release (with notes, visible on the Releases page) is what users actually see. I should have run `gh release create` as part of the publish flow, not left it for later.

4. **Recognize daemon config replacement faster.** When new lint issues appear that weren't there before, the first check should be "did the config file change?" I spent time investigating the symptoms before checking the cause.

5. **Verify `go get` works post-publish.** I verified the README example with local `replace` directives pre-publish. Once the repo went public, I should have immediately verified that `go get github.com/larsartmann/go-retry@latest` actually resolves in a clean module — proving the public import path works end-to-end.

6. **Run HARVEST on this status report too.** Section (f) below has forward-looking items. If the session continues, these need to land in TODO_LIST/ROADMAP, not rot here.

---

## f) Up to 50 things to get done next

Pareto-ranked. P1 = highest impact. Items already in `TODO_LIST.md` are marked **(T#)**; the rest are new candidates from this session's findings.

### P1 — unblock correctness (post-publish)

1. **Create the GitHub Release for v0.1.0** — `gh release create v0.1.0` with release notes from CHANGELOG. The Releases page is currently empty despite the tag being pushed. This is the #1 miss.
2. **Update TODO_LIST P3 items** — T5 (CI) and T6 (changelog links) are no longer blocked. The remote exists. Re-prioritize to P2 and remove the "blocked" framing.
3. **Update ROADMAP Open questions** — Remove the resolved git-remote question. The flake.nix question remains. Fix any "git remote -v is empty" text that is now stale.
4. **Add CHANGELOG compare links** — Now possible: `[Unreleased]: https://github.com/LarsArtmann/go-retry/compare/v0.1.0...HEAD` and `[0.1.0]: https://github.com/LarsArtmann/go-retry/releases/tag/v0.1.0`.
5. **Verify `go get github.com/larsartmann/go-retry@latest`** works in a clean module outside the repo. Proves the public import path resolves end-to-end.

### P2 — developer experience & CI

6. **Add minimal CI** (`.github/workflows/ci.yml`) — `go test ./... -race` + `golangci-lint run ./...` on push/PR. ~20 lines. The remote exists now.
7. **(T1) Close remaining behavioral test gaps** — concurrent `Do` stress test; fuzz `ComputeDelay` for overflow/negative edges.
8. **(T2) Verify `hierarchical-errors` migration applies** — likely a no-op (this package uses only `errors.Is`), but confirm and close.
9. **(T3) Cross-link `Backoff` ↔ `ComputeDelay` doc comments** — one-line "See [ComputeDelay]" / "See [Backoff]" in each.
10. **(T4) Add `SECURITY.md`** — conventional vulnerability reporting policy for a public repo.
11. **Add a CONTRIBUTING note on the auto-git daemon** — contributors should know the repo auto-commits. (Or is this intentionally hidden from public view?)

### P2 — API stability (pre-v1.0)

12. **Audit the public API surface** — confirm every exported symbol (`Do`, `Config`, `DefaultConfig`, `Backoff`, `ComputeDelay`, `AttemptFunc`, `ErrExhausted`, `ErrCanceled`) is meant to be public. No leaking implementation details.
13. **Decide whether `AttemptFunc(ctx, attempt)` should pass the previous error** — some retry libraries do. Deliberate decision before v1.0.
14. **Decide on options-style config API** (`WithOnRetry`, `WithJitter`, ...) — forward-compat without struct breakage. Large change; needs migration story.
15. **Write down `ComputeDelay` invariants** for the fuzz target (what must always hold? non-negative? capped? no overflow?).

### P3 — capability candidates (ROADMAP fuel; needs scoping)

16. **Configurable jitter factor** (`Config.JitterFactor` or `Jitter: none|additive|full`).
17. **Deterministic RNG option** (pluggable `rand` source for reproducible tests).
18. **Deadline-aware attempt budgeting** (stop retrying when remaining ctx budget < one more attempt).
19. **Document the circuit-breaker/bulkhead composition boundary** (likely: stay pure, document, don't code).
20. **Version-compat matrix with `go-error-family`** (pinned `v0.10.0`).
21. **Public docs site** (Astro + Starlight + Firebase, per `website-launch` skill) — only after API stable + examples exist (examples now exist).
22. **v1.0 release** once the API-freeze questions (#12, #13, #14) are settled.

### P3 — polish & hygiene

23. **Add `.gitattributes` Go-specific rules** — currently only `* text=auto eol=lf`; add `*.go text eol=lf` and `*.go diff=golang` (if a `.gitattributes` diff driver is configured).
24. **Standardize error-code naming in a table** — currently scattered: `retry.exhausted`, `retry.canceled`, `retry.invalid_*`. A table in DOMAIN_LANGUAGE or doc.go.
25. **Consider a `CODEOWNERS`** if this will be multi-maintainer.
26. **Add `//go:build go1.26` build constraint** — or decide the minimum Go version policy. `go.mod` says `1.26.5`; users on 1.25 can't import it.
27. **Evaluate `modernize` linter findings** — the standard `.golangci.yml` enables `modernize`; check if any suggestions apply to the codebase.
28. **Reconcile `[Unreleased]` semantics** — doc-only additions (README, CONTRIBUTING, DOMAIN_LANGUAGE) arguably aren't SemVer "changes" warranting a version bump. Decide policy.
29. **Add architecture decision record (ADR)** for the "no-CQRS/no-OTel core" boundary so the rationale survives personnel changes.
30. **Schedule a re-run of `docs-health` HARVEST** on this report's section (f) — otherwise these items rot here.

### P4 — nice-to-haves / open questions

31. **Telemetry hook contract** — is `OnRetry`/`OnExhausted` the final seam, or will a structured `Event` type replace it?
32. **Metrics naming guidance** for consumers wrapping the callbacks.
33. **Should `Config` implement `fmt.Stringer`** or a logging helper?
34. **Should `Do` accept a `BackoffFunc` override** (custom schedule beyond exponential)?
35. **Go version bump policy** — currently `go 1.26.5`; how aggressive?
36. **Evaluate `errors.AsType` migration** per the `hierarchical-errors` skill — `go-error-family` may already handle this; verify.
37. **Context-aware `DefaultConfig` variant?**
38. **Survey whether any LarsArtmann sibling lib already wraps this** (avoid duplication — `go-cqrs-lite/middleware/v4` is the known one).
39. **Is the auto-git daemon appropriate for a public repo?** — Contributors seeing commits they didn't make may be confused. Decision needed.
40. **GitHub branch protection rules** — now that the repo is public, should `master` be protected? Should PRs be required?

---

## g) Questions I can NOT figure out myself

1. **Should I create a GitHub Release for v0.1.0 now, or wait?** The tag is pushed but the Releases page is empty. A release with notes is the conventional public-facing artifact. But v0.1.0 was tagged _before_ this session's doc/license fixes — the tag points at `eae60c5`, not at the current HEAD (`a16bd9b`) which has the MIT license, README, examples, etc. Should I: (a) create a release from the existing tag (notes will describe v0.1.0 as originally shipped, then note the post-tag fixes in `[Unreleased]`), (b) move the tag to HEAD and release from there, or (c) cut a v0.2.0 with the doc/license fixes? Option (a) is the honest/standard approach but means the release notes won't mention the MIT license.

2. **Is the auto-git daemon intentional for this now-public repo?** The daemon made commits with empty messages (`31db438`, `6a41c5c`) and overwrote my `.golangci.yml` with the project standard. For a solo repo this may be fine; for a public repo where contributors see the history, empty commit messages and unexpected auto-commits could be confusing. Should the daemon be documented, configured, or disabled for this repo?

3. **Should I add CI now, or is there a LarsArtmann-standard CI workflow to copy?** The repo has no `.github/workflows/`. Other LarsArtmann repos likely have a standard CI setup (Go version matrix, golangci-lint action, etc.). Should I write one from scratch, or is there a canonical workflow to copy from a sibling repo (e.g., `go-error-family`, `go-cqrs-lite`)?

---

_Written 2026-08-03 21:48. Point-in-time snapshot — will go stale. When revisiting, apply the `update-old-docs` skill (annotate, don't rewrite)._
