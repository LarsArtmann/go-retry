# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md`; long-term/unbounded ideas live in `ROADMAP.md`; questions that
need a human decision live in `ROADMAP.md` → Open questions.

Priority: **P1** = high impact, do first; **P2** = valuable, not blocking;
**P3** = polish.

Harvested from
`docs/status/2026-08-22_01-20_go-retry-v0.4.0-hardening-executed.md` (§f) and
verified against the current code on 2026-09-13. T1–T8 (complexity, benchmark,
README `DelayFunc`, godoc examples, CI vet + coverage floor, fuzz campaign)
shipped in v0.4.0/v0.5.0 — see `CHANGELOG.md`.

---

## P1 — correctness & semantics

### T9. Return the typed non-retryable error by identity

`Do` returns a non-retryable error directly (`retry.go:91`), and
`TestDo_DoesNotRetryNonRetryableError` pins it with `errors.Is`
(`retry_test.go:156`). Strengthen the assertion to `err != rejection`
(identity), pinning that the typed error is never re-wrapped.
Effort: S.

### T10. Deduplicate the terminal-error code/message pairs

`contextEnded` repeats the sentinel code + message strings already carried by
`ErrCanceled` and `ErrDeadlineExceeded` (`retry.go:36-39` and `retry.go:47-50`
vs `retry.go:194-199`). Extract constants or derive from the sentinels so a
code rename cannot drift the two sites apart. Effort: S.

## P2 — CI & tooling

### T11. Bump `actions/checkout` v4 → v5/v6

`.github/workflows/ci.yml:14` pins checkout v4 by SHA while `setup-go` is
already v6; the v4 action runs on deprecated Node. Effort: S.

### T12. Scheduled fuzz job in CI

`FuzzComputeDelayNeverPanics` exists (`retry_test.go:592`) and a 5-minute
local campaign ran clean (104M+ execs, 0 failures). A scheduled workflow
would keep exploring the input domain without a human remembering to run it.
Graduated from `ROADMAP.md` → Raw ideas. Effort: S-M.

### T15. `govulncheck` in CI

Dropped by three consecutive sessions (2026-08-07 §c.6, 2026-08-08 §f.35,
missing from the 2026-08-22 list). Add a `govulncheck ./...` step to the
`test` job in `.github/workflows/ci.yml`. Effort: S.

## P3 — polish

### T13. Committed fuzz corpus (`testdata/fuzz/`)

The seven `f.Add` seeds live in code only; a committed corpus directory
survives cache loss and accumulates campaign discoveries. Effort: S.

### T14. Decide where release-notes bodies live

Release notes are currently composed ad hoc (the v0.4.0 body was drafted in
`/tmp`). Pick once: repo `docs/releases/` or GitHub-only. Effort: S.
