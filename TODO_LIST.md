# Todo List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish. The **Source** column is the harvest trail back into the
point-in-time report a task came from, so the next ANNOTATE pass is a lookup,
not archaeology.

## Open work

| #   | Task                                                                                                                                                                    | P  | Evidence / why                                                                                                                | Source                     |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -- | ----------------------------------------------------------------------------------------------------------------------------- | -------------------------- |
| T34 | Add a `dprint check` step to the CI lint job and pin the version, so markdown drift from non-session writers fails on master instead of only locally                    | P2 | `.github/workflows/ci.yml` has no dprint step; the formatter is documented and gated only in the local ritual                 | 2026-09-17 `§f.6` / `§c.2` |
| T35 | Land a checked-in `tools.go` pinning `actionlint`, `govulncheck`, and the dprint reference, replacing the three ad-hoc `go run`/`nix run` mechanisms                    | P2 | Owner blessed the `tools.go` pattern (`ROADMAP.md` → Open questions); the two-tool convention is currently stretched to three | 2026-09-17 `§f.7` / `§c.1` |
| T36 | Add a remote-action input-allowlist test covering every pinned `uses:` action, catching the 2026-09-13 `namee:` class that actionlint's schema check cannot see         | P2 | `AGENTS.md` → Gotchas documents the actionlint gap; weigh maintenance against the gate gap                                    | 2026-09-17 `§f.18`         |
| T37 | Re-verify at the next release cut that pkg.go.dev renders `ExampleBackoff` and `ExampleComputeDelay` (they landed after v0.7.0, so they are not on the tagged page yet) | P2 | Post-tag doc commits `4c48d13`; known to miss the v0.7.0 train                                                                | 2026-09-17 `§f.11`/`§f.23` |
| T38 | Add an `ExampleDo_withOptions` godoc example for the options tail                                                                                                       | P3 | README's options snippet is verified but has no output-pinned godoc counterpart                                               | 2026-09-17 `§f.22`         |
| T39 | Sweep `errors.Is`/`errors.As` sites against the `go-error-modernization` skill and migrate where `errors.AsType[E]` is the better match                                 | P3 | `retry.go` doc comments already name `errors.AsType`; Go 1.26 toolchain                                                       | 2026-09-17 `§f.25`         |
| T40 | Record a bump-trace audit trail: one line per `go-error-family` bump linking the Dependabot PR and the SHA-verification evidence                                        | P3 | The v0.10.1 bump landed with no in-repo record of the decider                                                                 | 2026-09-17 `§f.14`         |
| T41 | Sweep the remaining `go-cqrs-lite` consumers (`commandlifecycle`, `integration`, `example/taskmanager`) and verify their suites on the current `go-retry`               | P3 | They hold indirect pins; their own builds MVS-resolve, so this is a verification pass, not a bump                             | 2026-09-17 `§f.16`         |
| T42 | Dependabot watchlist: observe the post-merge auto-rebase when the next actions-group PR opens, then retire this row                                                     | P3 | No Dependabot PR was open during the v0.7.0 window; behavior still unobserved                                                 | 2026-09-17 `§f.15`         |
| T43 | Add a `.gitignore` pattern for scratch verification test files                                                                                                          | P3 | Scratch tests were created and deleted by hand during the fuzz drill; a pattern prevents an accidental commit                 | 2026-09-17 `§f.47`         |
| T44 | Link `scripts/check-compare-links.sh` from CONTRIBUTING's Development block and add the annotate-as-you-land ritual there                                               | P3 | Both exist in `AGENTS.md`/the ritual text but not where contributors look                                                     | 2026-09-17 `§f.32`/`§f.44` |

Nothing else is open. New findings enter through the docs-health HARVEST route
(`TODO_LIST.md` ← status reports / session discoveries); long-term bets and
owner questions live in `ROADMAP.md`.
