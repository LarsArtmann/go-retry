# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish.

---

## P1

## P2

### T23 — Consumer sweep: bump `go-cqrs-lite/middleware/v4` to the latest release

Post-release propagation was skipped for v0.6.0; the v1.0 API-freeze claim is
stronger once ≥1 real consumer pins the current release. Run the
`go-ecosystem-upgrade` flow. **Gated on the owner's answer** to `ROADMAP.md`
→ Open questions (cross-repo writes). Evidence: tag `v0.6.0`, `go.mod`.
Source: 2026-09-14 report §f.2, §f.50, §g.3.

### T24 — Fuzz crash drill: exercise the failure path

The fuzz workflow's success path is runner-proven, but the crash-artifact
upload + minimization path has never executed. Plant a temporary panic on a
throwaway branch, `workflow_dispatch` `fuzz.yml`, and verify the SHA-named
artifact uploads with `-fuzzminimizetime 5m` inside the 45-minute budget.
Evidence: `.github/workflows/fuzz.yml`. Source: 2026-09-14 report §f.9, §b.3.

## P3

### T26 — Add `workflow_dispatch` to `ci.yml`

Manual re-run surface without empty commits; parity with `fuzz.yml`.
Evidence: `.github/workflows/ci.yml` (`on:`). Source: 2026-09-14 report
§f.13.

### T27 — Godoc examples for `Backoff` and `ComputeDelay`

Completes the docs-site precondition "every entry point exampled" (`Do` /
`DoWithValue` are done). Evidence: `ROADMAP.md` (docs-site preconditions).
Source: 2026-09-14 report §f.16.

### T28 — CHANGELOG compare-link guard

A tiny script or test validating every footer compare-link resolves. All 7
verified by hand on 2026-09-16; keep it that way mechanically. Source:
2026-09-14 report §f.21.

### T29 — Link-rot sweep over README/CONTRIBUTING external links

dprint checks formatting, not 404s. Source: 2026-09-14 report §f.22.

### T32 — Coverage floor: raise 95% → 99%?

Local coverage sits at 100% and the sync test guards test-integrity; decide
whether the floor's deliberate headroom is still worth four points. Source:
2026-09-14 report §f.46.

### T33 — Verification watchlist (one-glance checks)

- The next Dependabot actions-group PR should arrive rebased on the merged
  pins (post-merge auto-rebase behavior unobserved). Source: 2026-09-14
  report §f.10.
- Push twice rapidly to one PR branch and observe the superseded run
  actually cancel (concurrency group). Source: 2026-09-14 report §f.14.

---

New findings enter through the docs-health HARVEST route (`TODO_LIST.md` ←
status reports / session discoveries); long-term bets and owner questions
live in `ROADMAP.md`.
