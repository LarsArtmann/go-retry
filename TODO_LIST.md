# Todo List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish.

## Work-item → plan mapping (for the next ANNOTATE pass)

The 2026-09-16 Pareto plan
(`docs/planning/2026-09-16_17-35_pareto-plan-doc-currency-options-migration-and-dual-release.md`)
absorbed T22–T33 and status-report §f items into M-tasks. Resolution status
after execution (2026-09-16):

| Item                               | Plan task  | Resolved                            |
| ---------------------------------- | ---------- | ----------------------------------- |
| T22 options migration (v1.0 track) | M4–M8      | shipped in v0.7.0                   |
| T23 consumer sweep                 | M9         | middleware/v4 on v0.7.0, pushed     |
| T24 fuzz crash drill               | M10.4–10.7 | artifact + budget verified          |
| T25 next-cut decision              | M3.1       | v0.6.1 then v0.7.0, both cut        |
| T26 workflow_dispatch              | M10.1      | landed                              |
| T27 godoc examples                 | M11        | ExampleBackoff, ExampleComputeDelay |
| T28 compare-link guard             | M12.1      | script, proven failing              |
| T29 link-rot sweep                 | M12.3      | all 6 links alive                   |
| T30 local govulncheck              | M3.5       | in pre-tag ritual                   |
| T31 release-notes skeleton         | M3.9       | in CONTRIBUTING                     |
| T32 coverage floor 95→99           | M15.1      | decided: keep 95 (ROADMAP)          |
| T33a Dependabot rebase             | M10.9      | unobservable this window            |
| T33b concurrency cancel            | M10.8      | superseded runs cancelled           |
| T33c tidy no-diff at tag           | M3.10      | verified at v0.6.1                  |

---

Nothing else is open. New findings enter through the docs-health HARVEST route
(`TODO_LIST.md` ← status reports / session discoveries); long-term bets and
owner questions live in `ROADMAP.md`.
