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

| #   | Task                                                                                                                                                                    | P  | Evidence / why                                                                | Source                     |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -- | ----------------------------------------------------------------------------- | -------------------------- |
| T37 | Re-verify at the next release cut that pkg.go.dev renders `ExampleBackoff` and `ExampleComputeDelay` (they landed after v0.7.0, so they are not on the tagged page yet) | P2 | Post-tag doc commits `4c48d13`; known to miss the v0.7.0 train                | 2026-09-17 `§f.11`/`§f.23` |
| T42 | Dependabot watchlist: observe the post-merge auto-rebase when the next actions-group PR opens, then retire this row                                                     | P3 | No Dependabot PR was open during the v0.7.0 window; behavior still unobserved | 2026-09-17 `§f.15`         |

Nothing else is open. New findings enter through the docs-health HARVEST route
(`TODO_LIST.md` ← status reports / session discoveries); long-term bets and
owner questions live in `ROADMAP.md`.

Executed 2026-09-17 (moved out on completion; details in `CHANGELOG.md` →
`[Unreleased]`): T34 (dprint CI step), T35 (`tools/` pin module), T36
(input-allowlist guard test), T38 (`ExampleDo_withOptions`), T39
(`errors.Is`/`errors.As` sweep — zero migrations, all sites are
sentinel/value matches), T40 (ROADMAP bump trace), T41 (consumer sweep:
`commandlifecycle`, `integration`, `example/taskmanager` suites green against
current master via the go-cqrs-lite workspace), T43 (`.gitignore` scratch
pattern), T44 (CONTRIBUTING rituals + recipes).
