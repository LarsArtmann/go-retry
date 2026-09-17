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
(ledgers, retro notes) are deliberately left unmarked. HARVEST routes items
forward with **verify-before-routing** — a harvested item's premise is checked
against code before it becomes a TODO row.

**How to read an archived report** (2026-09-17): `~~struck text~~ done at`hash`` marks the item shipped; `**Won't implement — reason**` and
`NOT-DO/DUPLICATE` mark it closed without shipping; `→ routed to <file>`
means the open work moved to a living doc (follow the pointer, not the
archive); untouched numbered items are still open. Table rows carrying
`Verdict`/`Status` columns express the same verdicts column-wise.

**Strikethrough rendering rules** (learned 2026-09-17 the hard way — three
bug classes left raw `~~` visible on GitHub across 8 archived files, since
repaired; verified via GitHub's GFM API renderer):

1. A closing `~~` must sit **directly after the struck text's last
   character** — a space before it (`text. ~~**— done**`) prevents the closer
   from binding, and the whole span renders literally. Write
   `text.~~ **— done**`.
2. **No single `~` inside a span** (`~20 lines`, `~1 h`): a lone tilde kills
   the pairing. Write `≈20` instead.
3. **Every span must close in the same paragraph** — an opener without its
   closer swallows the following text into the strike (and a `~~` mentioned
   inside backticks is inert, which is fine for convention examples like
   `` `~~item~~` ``).

| Report                                                                                                                                                                                     | Session                                                          | State                                                                                                                                                                                               |
| ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [2026-09-17 13:19 — v0.7.1 release half-cut, interrupted](archived/2026-09-17_13-19_v0.7.1-release-half-cut-interrupted-status.md)                                                         | Release cut interrupted mid-flight; split brain documented       | Fully annotated 2026-09-17 (`§b`–`§g` verdicts; `§a`/`§d`/`§e` narrative left as-is) — the release finished the same day (`082842a`, tag `v0.7.1`); open tails routed to `TODO_LIST.md`/`ROADMAP.md` |
| [2026-09-17 09:36 — T34–T44 executed: tools pin, CI dprint, input-allowlist guard](archived/2026-09-17_09-36_t34-t44-executed-tools-pin-ci-dprint-and-input-allowlist.md)                  | Harvested backlog executed; nested tools module; CI gates        | Fully annotated 2026-09-17 (`§b`–`§g` verdicts); open tails routed to `TODO_LIST.md` (T42, T45–T54) / `ROADMAP.md`                                                                                  |
| [2026-09-17 08:47 — docs-health AUDIT: go-directive fix, full archive, living-doc rebuild](archived/2026-09-17_08-47_docs-health-audit-go-directive-fix-archive-and-living-doc-rebuild.md) | Full 2026-0* audit; go.mod guard; TODO_LIST reseed               | Fully annotated 2026-09-17 (all ~40 `[NEW]` §f items verified, routed, or closed; `§b.1` rendering gap found real and fixed)                                                                        |
| [2026-09-17 05:53 — Pareto plan fully executed: v0.6.1 + v0.7.0 shipped, options migration landed](archived/2026-09-17_05-53_pareto-plan-fully-executed-v061-v070-shipped.md)              | Options migration, dual release, CI parity + fuzz drill          | Fully annotated 2026-09-17 (`§b`–`§g` verdicts; `§a`/`§d`/`§e` narrative left as-is); open tails routed to `TODO_LIST.md` (T34–T44) / `ROADMAP.md`                                                  |
| [2026-09-16 16:55 — docs-health pass 3: annotation, archive & living-doc reseed](archived/2026-09-16_16-55_docs-health-pass3-annotation-archive-and-living-doc-reseed.md)                  | Docs-health audit, TODO_LIST reseed, formatter identified        | Fully annotated 2026-09-17 (`§b`–`§g` verdicts); work traced to the 2026-09-16 Pareto plan                                                                                                          |
| [2026-09-14 03:30 — Pareto plan executed: v0.6.0 released, CI hardened, v1.0 track decided](archived/2026-09-14_03-30_pareto-plan-executed-v060-released-and-v1-track-decided.md)          | M1–M26 execution: v0.6.0 ship, actionlint gate, v1.0 API audit   | Fully annotated (§f.1–50 + §g 2026-09-16; §b/§c backfilled 2026-09-17)                                                                                                                              |
| [2026-09-13 14:48 — docs-health pass 2: annotation, archive & CI hardening](archived/2026-09-13_14-48_docs-health-pass2-annotation-archive-and-ci-hardening.md)                            | CI incident + PR #1 + pass-2 audit                               | Fully annotated (§f.1–50 from the pass; §b/§c/§f/§g leftovers backfilled 2026-09-17)                                                                                                                |
| [2026-09-13 12:57 — T9–T15 execution & brutal self-review](archived/2026-09-13_12-57_t9-t15-execution-status.md)                                                                           | Executed TODO T9–T15 end to end                                  | Fully annotated; open tails routed to `TODO_LIST.md` (T16–T21) / `ROADMAP.md`                                                                                                                       |
| [2026-09-13 12:20 — docs-health audit v0.5: annotation & archive](archived/2026-09-13_12-20_docs-health-audit-v05-annotation-and-archive.md)                                               | Annotated + archived all nine 2026-0* files; rebuilt living docs | Fully annotated; open tails routed to `TODO_LIST.md` / `ROADMAP.md`                                                                                                                                 |
| [2026-08-22 01:20 — v0.4.0 hardening executed](archived/2026-08-22_01-20_go-retry-v0.4.0-hardening-executed.md)                                                                            | Cap fix, deadline/cancel split, T9–T15 harvest                   | Fully annotated (T9–T15 `done at` the 2026-09-13 hashes)                                                                                                                                            |
| [2026-08-08 11:22 — docs-health audit & living-doc rebuild](archived/2026-08-08_11-22_docs-health-audit-and-living-doc-rebuild.md)                                                         | Jitter decision + full audit                                     | Fully annotated                                                                                                                                                                                     |
| [2026-08-08 11:12 — jitter-deferral decision](archived/2026-08-08_11-12_jitter-deferral-decision.md)                                                                                       | Deferred configurable jitter                                     | Fully annotated                                                                                                                                                                                     |
| [2026-08-07 09:18 — comprehensive session status](archived/2026-08-07_09-18_comprehensive-session-status.md)                                                                               | v0.3.x era work                                                  | Fully annotated                                                                                                                                                                                     |
| [2026-08-07 08:39 — backoff validation & unfixed panics (HTML)](archived/2026-08-07_08-39_backoff-validation-and-unfixed-panics.html)                                                      | B1/B2/B3 panic investigation                                     | Fully annotated (Status column); parse-validated 2026-09-13                                                                                                                                         |
| [2026-08-03 22:09 — post-publish cleanup, release & stale docs](archived/2026-08-03_22-09_post-publish-cleanup-release-and-stale-docs.md)                                                  | v0.1.0 aftermath                                                 | Fully annotated                                                                                                                                                                                     |
| [2026-08-03 21:48 — MIT switch & public GitHub launch](archived/2026-08-03_21-48_mit-switch-and-public-github-launch.md)                                                                   | Public launch                                                    | Fully annotated                                                                                                                                                                                     |
| [2026-08-03 21:21 — docs-health audit & self-review](archived/2026-08-03_21-21_docs-health-audit-and-self-review.md)                                                                       | First docs audit                                                 | Fully annotated                                                                                                                                                                                     |

Planning snapshots live in [`docs/planning/archived/`](../planning/archived/)
(three files: the 2026-08-21 cap-and-semantics hardening plan, the
2026-09-13 CI-trust/v1-track Pareto plan, and the 2026-09-16 doc-currency /
options-migration / dual-release Pareto plan — all fully annotated).
