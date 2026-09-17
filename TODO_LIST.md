# Todo List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish. The **Source** column is the harvest trail back into the
point-in-time report a task came from, so the next ANNOTATE pass is a lookup,
not archaeology. Last harvested: 2026-09-17 (docs-health AUDIT pass over the
three 2026-09-17 reports).

## Open work

| #   | Task                                                                                                                                                                    | P  | Evidence / why                                                                | Source                     |
| --- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -- | ----------------------------------------------------------------------------- | -------------------------- |
| T42 | Dependabot watchlist: observe the post-merge auto-rebase when the next actions-group PR opens, and execute the `/tools` bump flow end-to-end (verify, merge, append the ROADMAP bump-trace row — designed but unexercised), then retire this row | P3 | No Dependabot PR was open during the v0.7.0/v0.7.1 windows; behavior still unobserved | 2026-09-17 `§f.15`/`§f.6`  |
| T45 | Marker-completeness gate: script asserting every `docs/status/archived/*.md` numbered item carries a verdict (run it over the existing archive); fold in the index State-cell recheck, the HTML-report strikethrough design call, and a guard for the annotation convention | P1 | Archive marker checks are hand-run; the 2026-09-17 rendering bugs (3 strikethrough classes) shipped because no gate read the archive | 2026-09-17 08:47 `§f.2`/`§f.21`; 09:36 `§f.10`; 13:19 `§f.12` |
| T46 | `scripts/check-docs.sh`: one command orchestrating dprint + marker gate + compare-links; add it to the AGENTS Session Ritual once it exists | P1 | The doc ritual is a manual command sequence; each gate is 2s-class cheap but skipped under time pressure (the red-master lesson) | 2026-09-17 08:47 `§f.19`/`§f.20`; 13:19 `§f.13` |
| T47 | Gate the `tools/` module: `go -C tools vet ./...` (+ golangci-lint) in the CI lint job and the Session Ritual — root `./...` skips the nested module, so it stands outside every quality gate | P2 | `tools/tools.go` compiles and its binaries run, but nothing lints or vets it | 2026-09-17 09:36 `§b.5`/`§f.4` |
| T48 | Canonical coverage recipe: one decision across README/CONTRIBUTING/FEATURES (AGENTS names `go test -cover ./...` as canonical; CONTRIBUTING generates `reports/coverage.out`), plus a full CONTRIBUTING drift sweep | P3 | Two recipes coexist; drift class already bit once | 2026-09-17 08:47 `§f.17`/`§f.18`; 09:36 `§f.18` |
| T49 | AGENTS gotcha prune to <20 rows (standing policy: the budget must be refilled before the next gotcha lands) | P3 | At the 20-row cap since 2026-09-17 | 2026-09-17 08:47 `§f.10`; 09:36 `§f.13`; 13:19 `§f.15` |
| T50 | SECURITY.md posture audit against the current dependency/CI surface (not read beyond a grep in any recent pass) | P3 | Supply-chain surface grew (tools module, dprint action, Dependabot groups) | 2026-09-17 08:47 `§f.30`; 09:36 `§f.17`; 13:19 `§f.20` |
| T51 | Safe probe fixtures for guard drills: gitignored scratch-workflow convention or a fixture dir, so proving guards fail never touches live `.github/workflows/` | P3 | The `namee:` probe lived seconds in the live workflows dir (race window) | 2026-09-17 09:36 `§f.8` |
| T52 | Index/backlog conventions: TODO_LIST "last harvested" staleness signal (done this pass — keep it current), a State convention for planning-file rows, the standing Verdict/Status column format, and whether `docs/planning/` gets its own README index | P3 | Hand-maintained columns rot; conventions are undocumented | 2026-09-17 08:47 `§f.23`/`§f.27`/`§f.28`/`§f.32` |
| T53 | Annotate archived plans' `done at` hashes that point at unreachable pre-amend objects (the 2026-09-16 plan's M2 cited dangling `9c08595`; sweep for others) | P3 | Hash citations must resolve; dangling ones weaken the archive | 2026-09-17 08:47 `§f.22` |
| T54 | Extend the go-directive guard: assert the living docs' stated Go version matches `go.mod` (the test pins go.mod only), and decide whether the guard also asserts the go-error-family surface contract (ROADMAP rule) | P3 | Docs could drift from go.mod with only the module file guarded | 2026-09-17 08:47 `§f.3`/`§f.45` |

Nothing else is open. New findings enter through the docs-health HARVEST route
(`TODO_LIST.md` ← status reports / session discoveries) with
verify-before-routing; long-term bets and owner questions live in `ROADMAP.md`.
