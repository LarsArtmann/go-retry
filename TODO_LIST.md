# Todo List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish. The **Source** column is the harvest trail back into the
point-in-time report a task came from, so the next ANNOTATE pass is a lookup,
not archaeology. Last harvested: 2026-10-08, second pass (execution-session closeout):
T48-T49, T55-T69, T70-T76 retired as completed or dispositioned during the
Pareto-plan execution (release v0.8.0, guards, stress workflow + harness,
consumer sweep, lychee/dprint, erraudit deferred-with-rationale, pre-commit
declined, upstream features verified existing).

## Open work

| #   | Task                                                                                                                                                                                                                                                                   | P  | Evidence / why                                                                        | Source                                               |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -- | ------------------------------------------------------------------------------------- | ---------------------------------------------------- |
| T42 | Dependabot watchlist: observe the post-merge auto-rebase when the next actions-group PR opens, and execute the `/tools` bump flow end-to-end (verify, merge, append the ROADMAP bump-trace row — designed but unexercised), then retire this row                       | P3 | No Dependabot PR was open during the v0.7.0/v0.7.1 windows; behavior still unobserved | 2026-09-17 `§f.15`/`§f.6`                            |
| T51 | Safe probe fixtures for guard drills: gitignored scratch-workflow convention or a fixture dir, so proving guards fail never touches live `.github/workflows/`                                                                                                          | P3 | The `namee:` probe lived seconds in the live workflows dir (race window)              | 2026-09-17 09:36 `§f.8`                              |
| T52 | Index/backlog conventions: a State convention for planning-file rows, the standing Verdict/Status column format, and whether `docs/planning/` gets its own README index                                                                                                | P3 | Hand-maintained columns rot; conventions are undocumented                             | 2026-09-17 08:47 `§f.23`/`§f.27`/`§f.28`/`§f.32`     |
| T53 | Annotate archived plans' `done at` hashes that point at unreachable pre-amend objects (the 2026-09-16 plan's M2 cited dangling `9c08595`; sweep for others)                                                                                                            | P3 | Hash citations must resolve; dangling ones weaken the archive                         | 2026-09-17 08:47 `§f.22`                             |
| T54 | Extend the go-directive guard: assert the living docs' stated Go version matches `go.mod` (the module file is guarded by `TestModuleGoDirectiveStaysPinned`; the docs-match half remains — the go-error-family floor half landed 2026-10-08 in `dispositions_test.go`) | P3 | Docs could drift from go.mod with only the module file guarded                        | 2026-09-17 08:47 `§f.3`/`§f.45`; narrowed 2026-10-08 |
| T60 | Rebuild/reinstall the BuildFlow binary after concurrent sessions settle (`nix build .` + `nix run .#reinstall` + `buildflow doctor`)                                                                                                                                   | P2 | Binary lags its own repo HEAD; advisory-only warning every run                        | 2026-10-08 `§f.14`/`§b2`; plan M10                   |

Nothing else is open. New findings enter through the docs-health HARVEST route
(`TODO_LIST.md` ← status reports / session discoveries) with
verify-before-routing; long-term bets and owner questions live in `ROADMAP.md`.
