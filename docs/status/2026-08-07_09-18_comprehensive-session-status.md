# Go Retry Status Report

**Snapshot:** 2026-08-07 09:18 CEST  
**Repository:** `github.com/larsartmann/go-retry`  
**Branch:** `master`  
**Release baseline:** `v0.2.0` tag points at `5a693f9`  
**Current HEAD:** `10a5566`  
**Working tree:** `TODO_LIST.md` modified and uncommitted; all other observed changes are committed by the repository's auto-commit daemon.

## Executive summary

The session successfully completed the previously open T1-T5 work: retry-policy interoperability, concurrency coverage, fuzz coverage, CI, API cross-links, and a security policy. The code is healthy under the local verification gate: race tests repeated ten times, golangci-lint, go vet, and 100% statement coverage all pass.

The biggest failure was process discipline, not runtime correctness: the first report claimed the work was complete while `TODO_LIST.md` still had open items and the new CI file did not appear in the visible final diff because the auto-commit daemon had already committed it. The final TODO cleanup is currently uncommitted, and CI has not been observed running on GitHub because nothing was pushed during this session.

## a) FULLY DONE

### 1. Retry-policy interoperability

- Added `retry.FromPolicy(errorfamily.RetryPolicy) Config`.
- Maps `MaxAttempts`, `MinDelay` to `InitialDelay`, and `MaxDelay`.
- Retains this library's default multiplier, retry predicate, and unset hooks.
- Evidence: `config.go:62`, tests in `retry_test.go`; commit `692d95c`; full suite and coverage pass.
- Scope: additive API, no breaking change.

### 2. Concurrent retry isolation coverage

- Added a 100-goroutine concurrent invocation test.
- Each invocation independently retries once and verifies its own call count.
- Evidence: `retry_test.go:66` area; commit `e840c7d`; `go test ./... -race -count=10` passed.
- Scope: behavioral guarantee only; no production implementation change required.

### 3. Fuzz coverage for `ComputeDelay`

- Added fuzz seeds for ordinary, zero-cap, overflow, and near-maximum inputs.
- Invalid attempts must return an error; valid attempts must not return negative delays.
- Evidence: `retry_test.go` fuzz target; commit `e840c7d`; seed execution passes.
- Scope: fuzz target exists, but extended-duration fuzzing was not run.

### 4. Hierarchical-error migration verification

- Searched this package and found no `errors.As` or `errors.AsType` usage.
- The package uses `errors.Is`; no migration was needed.
- The dependency already uses Go 1.26 `errors.AsType` internally.
- Evidence: source search and dependency inspection; TODO removed.

### 5. Continuous integration workflow

- Added push and pull-request workflows.
- Test job runs `go test ./... -race`.
- Lint job uses Go version from `go.mod` and `golangci/golangci-lint-action@v9` with v2.12.2.
- Uses read-only repository permissions.
- Evidence: `.github/workflows/ci.yml`; local commands pass.
- Scope: workflow is present locally, not remotely verified.

### 6. Public API documentation cross-links

- `Backoff` links to `ComputeDelay`.
- `ComputeDelay` links to `Backoff`.
- README now states both return `(time.Duration, error)` and reject attempts below 1.
- Evidence: `retry.go:98`, `retry.go:112`, `README.md:76`; commit `10a5566`.

### 7. Security policy

- Added `SECURITY.md` with supported-version guidance and private GitHub advisory reporting.
- Explicitly asks reporters not to disclose secrets or sensitive data.
- Evidence: `SECURITY.md`; commit `10a5566`.

### 8. Documentation and release consistency

- Updated CHANGELOG, FEATURES, README, and TODO summary for the new work.
- The prior v0.2.0 panic fixes remain intact.
- Evidence: commits `5a693f9`, `d2a5c6f`, `10a5566`.

### 9. Verification

All passed locally:

```text
go test ./... -race -count=10       PASS
golangci-lint run ./...            PASS, 0 issues
go vet ./...                       PASS
go test -coverprofile=/tmp/cov.out ./... PASS, 100.0%
go tool cover -func=/tmp/cov.out   100.0% every function
```

## b) PARTIALLY DONE

### 1. Release publication

- Works now: signed local `v0.2.0` exists and points to the intended release commit.
- Missing: the release/tag was not pushed during this session, so GitHub and pkg.go.dev publication were not verified.
- Blocker: explicit publication instruction is still required before pushing.
- Effort: S.

### 2. CI validation

- Works now: workflow syntax and commands are locally represented; local test/lint commands pass.
- Missing: GitHub Actions has not actually executed, and action compatibility, permissions, cache behavior, and Go 1.26 runner availability are unverified remotely.
- Blocker: repository must be pushed, then the workflow run inspected.
- Effort: S.

### 3. Fuzzing

- Works now: fuzz target compiles and executes seed corpus.
- Missing: no time-bounded fuzz campaign was run, no corpus directory was retained, and no fuzzing job is in CI.
- Blocker: none; this is prioritization.
- Effort: S-M.

### 4. TODO list finalization

- Works now: all listed T1-T5 work was implemented and the open sections were removed.
- Missing: `TODO_LIST.md` is currently modified in the working tree and is not part of the latest commit shown by `git log`.
- Blocker: commit/autocommit timing.
- Effort: S.

### 5. Documentation line references

- Works now: major references were updated.
- Missing: this session did not perform a complete link-and-line-reference audit across every documentation file.
- Blocker: none; requires deliberate docs verification.
- Effort: S-M.

## c) NOT STARTED

These were not requested as implementation work in this session and have no code changes here:

1. Push `master` and tag `v0.2.0` to the remote.
2. Verify the GitHub release page renders correctly.
3. Verify pkg.go.dev indexes the release.
4. Add a coverage threshold check to CI.
5. Add a dedicated fuzzing job or scheduled fuzz campaign.
6. Add dependency vulnerability scanning with `govulncheck`.
7. Add secret scanning with `gitleaks`.
8. Add release automation or tag-triggered release packaging.
9. Add changelog automation.
10. Add API compatibility checking for future releases.
11. Add benchmarks to CI with a regression budget.
12. Add examples for `FromPolicy`.
13. Add an explicit example of invalid-attempt handling.
14. Add a test for non-transient family policy conversion semantics.
15. Add a test documenting how callers should configure a non-retryable policy.
16. Add a `go doc` or documentation build check.
17. Add external consumer compilation tests.
18. Add a compatibility test for the v0.2.0 breaking signature.
19. Revisit whether `FromPolicy` should map only timing/attempt fields or also derive retryability from a family.
20. Investigate whether `FromPolicy` should reject `RetryPolicy{MaxAttempts: 1}` as a valid no-retry configuration rather than relying on `Multiplier` and delays.
21. Add a project-level policy for auto-commit daemon coordination.
22. Add a release checklist document.
23. Add a post-release smoke-test script.
24. Add GitHub issue templates.
25. Add a pull-request template.
26. Add a code of conduct if the public project needs one.
27. Add supported Go-version testing beyond the pinned Go 1.26.5.
28. Add race testing to the release checklist explicitly.
29. Add a test for context cancellation while the callback itself is executing.
30. Add a test for callback panics and document that callback panics are caller-owned.
31. Add a test for nil `AttemptFunc` behavior or explicitly document its contract.
32. Add a test for `NaN` and negative/zero multipliers through the public API.
33. Add a test for `MaxDelay < InitialDelay` validation policy.
34. Add a deterministic jitter injection option if reproducible delay sequences become important.
35. Add a configurable jitter factor if the fixed 50% behavior proves limiting.
36. Add deadline-aware retry budgeting if callers need it.
37. Add a benchmark for `Do`, not only `ComputeDelay`.
38. Add allocation assertions to performance tests.
39. Add a dependency update policy.
40. Add release notes for consumers migrating from v0.1.x.
41. Add a migration section to README for the `(duration, error)` API change.
42. Add a generated API documentation check.
43. Add static analysis for documentation links.
44. Add a fuzz corpus retention policy.
45. Add a test proving no shared mutable configuration is introduced by future changes.
46. Add a test for hook ordering and exact callback arguments.
47. Add a test for timer cleanup under cancellation and normal completion.
48. Add a test for very large valid `MaxAttempts` behavior with a canceled context.
49. Add a test for error-family code stability across wrapped errors.
50. Add a formal v1.0 readiness checklist.

## d) TOTALLY FUCKED UP!

### 1. The first completion claim was too strong

The prior response said the work was complete while `TODO_LIST.md` still had open TODO entries at that point. That was inaccurate status reporting. Severity: medium process failure, not a runtime failure. Root cause: relying on the auto-commit daemon and not re-checking the complete working tree/document state before declaring completion. Mitigation: this report identifies the discrepancy; the TODO cleanup is now visible as a working-tree modification.

### 2. The new CI workflow was not visible in the final status inspection

`git status` showed only `TODO_LIST.md`, while the workflow existed in repository history. This is not necessarily data loss, but it made the handoff ambiguous and the final response failed to mention the auto-commit boundary clearly. Severity: medium traceability failure. Root cause: no final `git log --all --stat` reconciliation after automatic commits. Mitigation: inspect commit history, not only working-tree diff, before reporting.

### 3. The release was not actually published

The local tag exists, but remote publication was intentionally not performed. Calling the project released without clarifying “local only” would be misleading. Severity: high for users expecting availability, none for local code safety. Workaround: push the branch and tag, then verify GitHub and pkg.go.dev.

### 4. CI was not remotely verified

A locally valid-looking YAML file is not evidence that GitHub accepted and ran it. Severity: medium. Root cause: no push and no remote workflow inspection. Mitigation: publish only after a real Actions run succeeds.

### 5. The fuzz claim is narrower than a fuzz campaign

The fuzz target's seed corpus ran, but this session did not execute a sustained fuzz run. Describing fuzzing as fully proven would overstate confidence. Severity: low-medium. Mitigation: run `go test -fuzz=FuzzComputeDelayNeverPanics` for a bounded duration and preserve any useful corpus.

## e) WHAT WE SHOULD IMPROVE

### 1. Close the status loop before declaring done

Always run `git status --short --untracked-files=all`, `git diff`, and recent history after auto-commit activity. Report committed, staged, and uncommitted work separately.

### 2. Separate “implemented” from “published”

Use explicit states: local implementation, committed, tagged, pushed, GitHub release verified, pkg.go.dev verified. A tag alone is not a public release.

### 3. Treat CI as unverified until it runs remotely

Local execution validates commands, not Actions metadata, runner support, permissions, or action versions.

### 4. Make API migration discoverable

The README mentions the new return type, but a dedicated migration section would better protect v0.1.x consumers from compile failures.

### 5. Improve `FromPolicy` semantics documentation

The function maps policy fields but does not encode a family into `IsRetryable`, because `RetryPolicy` contains no family value. This limitation should be explicit in godoc and examples.

### 6. Add CI quality gates beyond race and lint

At minimum add vet, coverage threshold, and a short fuzz smoke test. Later add govulncheck and compatibility checks.

### 7. Reduce flaky timing dependence

The cancellation test uses a real 5-second delay and a goroutine sleeping 10ms. It passed repeatedly, but deterministic clock/timer injection would make this more robust.

### 8. Revisit test naming and API evidence

The new concurrency test is strong for race detection but does not independently prove all shared-state properties. Add assertions around callback isolation and configuration mutation behavior if those become supported guarantees.

### 9. Resolve the modernizer diagnostic deliberately

The LSP reports `retry_test.go:653` using `b.N` where Go 1.26 supports `b.Loop()`. The lint command passes, but the repository still has one IDE diagnostic. Either modernize the benchmark or document why retaining `b.N` is intentional.

### 10. Keep living docs and historical reports distinct

The TODO list should contain only actionable open work. Status reports should preserve the snapshot, while later sessions annotate them rather than silently rewriting history.

## f) Up to 50 things we should get done next

Ranked by impact, then effort.

| # | Task | Impact | Effort | Category |
|---:|---|---|---|---|
| 1 | Commit the final `TODO_LIST.md` cleanup after reviewing the daemon's latest history. | Critical | S | Cleanup |
| 2 | Push `master` and signed tag `v0.2.0` only when publication is authorized. | Critical | S | Release |
| 3 | Verify the GitHub Actions workflow on an actual push/PR run. | Critical | S | Infrastructure |
| 4 | Verify GitHub release rendering and pkg.go.dev indexing for `v0.2.0`. | High | S | Release |
| 5 | Add a v0.1.x to v0.2.0 migration section documenting the new error returns. | High | S | Documentation |
| 6 | Add `go vet ./...` to CI. | High | S | Quality |
| 7 | Add a CI coverage threshold and fail below the agreed floor. | High | S | Quality |
| 8 | Run a bounded `FuzzComputeDelayNeverPanics` campaign. | High | S | Quality |
| 9 | Add a short fuzz smoke job to CI or a scheduled workflow. | Medium | S-M | Infrastructure |
| 10 | Add `govulncheck` to CI after verifying availability for Go 1.26.5. | High | S-M | Security |
| 11 | Add secret scanning for pull requests. | Medium | S | Security |
| 12 | Decide whether the single policy converter should remain additive or evolve into a richer family-aware API. | High | M | API design |
| 13 | Document that `FromPolicy` cannot infer retryability from `RetryPolicy` alone. | High | S | Documentation |
| 14 | Add an `ExampleFromPolicy` godoc example. | Medium | S | Documentation |
| 15 | Add invalid-attempt and invalid-config examples. | Medium | S | Documentation |
| 16 | Add external consumer compile tests for the public API. | High | M | Compatibility |
| 17 | Add a v0.2.0 API compatibility snapshot/check. | Medium | M | Quality |
| 18 | Resolve the `b.N` versus `b.Loop()` modernizer diagnostic. | Low | S | Cleanup |
| 19 | Add hook-order and exact-argument tests. | Medium | S-M | Quality |
| 20 | Add timer cleanup tests where practical without brittle timing. | Medium | M | Quality |
| 21 | Add tests for `NaN`, infinity, negative, and zero multiplier boundary behavior. | High | S | Quality |
| 22 | Decide and document policy for `MaxDelay < InitialDelay`. | Medium | S | API design |
| 23 | Add tests for very large valid attempt counts with cancellation. | Medium | S | Quality |
| 24 | Add stable error-code assertions for all validation branches. | Medium | S | Quality |
| 25 | Add a release checklist covering tag, push, Actions, GitHub, and pkg.go.dev. | High | S | Process |
| 26 | Add release automation for signed annotated tags if the repository workflow supports it. | Medium | M | Release |
| 27 | Add dependency update automation with review constraints. | Medium | M | Maintenance |
| 28 | Add supported-version testing if compatibility beyond Go 1.26 is desired. | Medium | M | CI |
| 29 | Add a benchmark for end-to-end `Do`. | Low | S | Performance |
| 30 | Establish benchmark regression thresholds before enforcing them. | Low | M | Performance |
| 31 | Add allocation regression checks for delay calculation. | Low | S-M | Performance |
| 32 | Decide whether deterministic RNG injection belongs in the API. | Medium | M | API design |
| 33 | Decide whether configurable jitter belongs in the API. | Medium | M | API design |
| 34 | Add deadline-aware retry budgeting if real consumers need it. | Medium | L | Feature |
| 35 | Add an explicit nil `AttemptFunc` contract and test. | Medium | S | API design |
| 36 | Decide whether callback panics are caller-owned and document that boundary. | Low | S | Documentation |
| 37 | Add callback isolation tests for concurrent `Do` calls. | Medium | S-M | Quality |
| 38 | Add documentation-link validation. | Low | M | Quality |
| 39 | Add generated API documentation checks. | Low | M | Documentation |
| 40 | Add GitHub issue templates for bugs and feature requests. | Low | S | Infrastructure |
| 41 | Add a pull-request template containing the race/lint/vet checklist. | Low | S | Infrastructure |
| 42 | Add a public release note for panic fixes and the breaking API change. | High | S | Documentation |
| 43 | Add consumer guidance on `MaxAttempts` counting the initial call. | Medium | S | Documentation |
| 44 | Add property tests for delay upper bounds under unusual valid configurations. | Medium | S-M | Quality |
| 45 | Define fuzz corpus retention and minimization policy. | Low | S | Process |
| 46 | Add a formal v1.0 readiness checklist. | Medium | M | Planning |
| 47 | Reassess whether the flat single-package boundary remains appropriate. | Low | M | Architecture |
| 48 | Add a compatibility policy for error-family dependency upgrades. | Medium | S-M | Maintenance |
| 49 | Review GitHub Actions pinning and supply-chain trust policy. | High | S | Security |
| 50 | Perform a focused public API review before the next release. | High | M | Quality |

## g) Questions I cannot determine from the repository

1. Should the local `v0.2.0` release be pushed to `origin` now, or remain local until a separate release approval?
2. Should this library promise support for Go versions other than the currently pinned Go 1.26.5?
3. Should `FromPolicy` remain a minimal field mapper, or should the public API eventually accept family context and derive retryability too?

## Current handoff

No further code changes were made after the verified implementation and documentation work. The only visible uncommitted change is the final `TODO_LIST.md` cleanup. The repository was not pushed. This report is a snapshot, not a release confirmation.
