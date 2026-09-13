# Status reports

Point-in-time session snapshots. **A report is never edited to stay current.**
When its items resolve, it is annotated inline (`~~item~~ done at <hash>` /
`→ routed to <file>` / `**Won't implement — reason**`) and then moved here.
The living docs — `TODO_LIST.md`, `ROADMAP.md`, `CHANGELOG.md`, `FEATURES.md`
— are the only current sources; never read an archived report as a backlog.

**Annotation/archive convention** (decided 2026-09-13): hand-written inline
markers use explicit variants — `done at <short-hash>`, `→ routed to
<destination>`, `**Won't implement — reason**`; the docs-health skill's
`annotate-*.py` scripts (kinds `h`/`v`/`p`/`w`) remain the tool for batch
numbered-table/prose runs. Narrative sections that carry their own resolution
(ledgers, retro notes) are deliberately left unmarked.

| Report                                                                                                                                       | Session                                                          | State                                                                         |
| -------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| [2026-09-13 12:57 — T9–T15 execution & brutal self-review](archived/2026-09-13_12-57_t9-t15-execution-status.md)                             | Executed TODO T9–T15 end to end                                  | Fully annotated; open tails routed to `TODO_LIST.md` (T16–T21) / `ROADMAP.md` |
| [2026-09-13 12:20 — docs-health audit v0.5: annotation & archive](archived/2026-09-13_12-20_docs-health-audit-v05-annotation-and-archive.md) | Annotated + archived all nine 2026-0* files; rebuilt living docs | Fully annotated; open tails routed to `TODO_LIST.md` / `ROADMAP.md`           |
| [2026-08-22 01:20 — v0.4.0 hardening executed](archived/2026-08-22_01-20_go-retry-v0.4.0-hardening-executed.md)                              | Cap fix, deadline/cancel split, T9–T15 harvest                   | Fully annotated (T9–T15 `done at` the 2026-09-13 hashes)                      |
| [2026-08-08 11:22 — docs-health audit & living-doc rebuild](archived/2026-08-08_11-22_docs-health-audit-and-living-doc-rebuild.md)           | Jitter decision + full audit                                     | Fully annotated                                                               |
| [2026-08-08 11:12 — jitter-deferral decision](archived/2026-08-08_11-12_jitter-deferral-decision.md)                                         | Deferred configurable jitter                                     | Fully annotated                                                               |
| [2026-08-07 09:18 — comprehensive session status](archived/2026-08-07_09-18_comprehensive-session-status.md)                                 | v0.3.x era work                                                  | Fully annotated                                                               |
| [2026-08-07 08:39 — backoff validation & unfixed panics (HTML)](archived/2026-08-07_08-39_backoff-validation-and-unfixed-panics.html)        | B1/B2/B3 panic investigation                                     | Fully annotated (Status column); parse-validated 2026-09-13                   |
| [2026-08-03 22:09 — post-publish cleanup, release & stale docs](archived/2026-08-03_22-09_post-publish-cleanup-release-and-stale-docs.md)    | v0.1.0 aftermath                                                 | Fully annotated                                                               |
| [2026-08-03 21:48 — MIT switch & public GitHub launch](archived/2026-08-03_21-48_mit-switch-and-public-github-launch.md)                     | Public launch                                                    | Fully annotated                                                               |
| [2026-08-03 21:21 — docs-health audit & self-review](archived/2026-08-03_21-21_docs-health-audit-and-self-review.md)                         | First docs audit                                                 | Fully annotated                                                               |

Planning snapshots live in [`docs/planning/archived/`](../planning/archived/)
(one file: the 2026-08-21 cap-and-semantics hardening plan, fully annotated).
