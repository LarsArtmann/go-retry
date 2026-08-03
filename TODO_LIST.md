# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md` (`[Unreleased]`); long-term/unbounded ideas live in `ROADMAP.md`;
questions that need a human decision live in `ROADMAP.md` → Open questions.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish or blocked.

_Recently completed (now in `CHANGELOG.md` `[Unreleased]`): README rewrite,
`.golangci.yml`, godoc `Example*` functions, `BenchmarkComputeDelay`, coverage
workflow in CONTRIBUTING, `docs/DOMAIN_LANGUAGE.md`._

---

## P1 — Correctness hardening

### T1. Close remaining behavioral-guarantee gaps in the test suite

Coverage is 100% by statement, and the core behavioral guarantees are now
asserted (`OnRetry` is not called after the final attempt; a pre-canceled
context yields `ErrCanceled`; `OnExhausted` receives the exact last error by
identity — shipped, see `CHANGELOG.md` `[Unreleased]`). Two exploratory gaps
remain (harvested from `docs/status/2026-08-03_21-21_*.md` §f.13-20):

- **Concurrent `Do` invocations share no mutable state** — the global
  `math/rand/v2` is safe, but prove it with a parallel stress test.
- **Fuzz `ComputeDelay`** for numeric edges (huge `attempt`/`multiplier`
  overflow, negative-ish durations, `Multiplier` just above 1) — currently only
  spot-checked (`retry.go:114`).

### T2. Verify the `hierarchical-errors` migration applies (or close it)

Go 1.26+ offers generic `errors.AsType[T]`. `go-retry` itself uses only
`errors.Is` (not `errors.As`), so this is likely a **no-op here** — but its
dependency `go-error-family` may use `errors.As`. Confirm there is nothing to
migrate in this package and close the item, or note the finding against
`go-error-family`. (See the `hierarchical-errors` skill.)

## P2 — Polish

### T3. Cross-link `Backoff` and `ComputeDelay` doc comments

Each refers to the same formula; add a one-line "See [ComputeDelay] for the
raw-parameter variant" / "See [Backoff] for the Config-based variant" so godoc
readers find both (`retry.go:97`, `retry.go:108`).

### T4. Add a `SECURITY.md`

The `LICENSE` is proprietary with a reporting contact (`git@lars.software`),
but a dedicated `SECURITY.md` is the conventional place for vulnerability
reporting policy. Low effort.

## P3 — Blocked on a git remote (see ROADMAP → Open questions)

These are not actionable until a remote is published (the module path
`github.com/larsartmann/go-retry` and a signed `v0.1.0` tag imply "will be
published", but `git remote -v` is empty — see `ROADMAP.md` → Open questions).

### T5. Minimal CI

No `.github/workflows/` exists. Once a remote exists, add a workflow running
`go test ./... -race` and `golangci-lint run ./...` on push/PR.

### T6. Keep-a-Changelog compare links

`CHANGELOG.md` intentionally omits `[Unreleased]`/version compare links today
(there is no remote URL to build them from — see the note in `CHANGELOG.md`).
Add the standard footer links when a remote is published.
