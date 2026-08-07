# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md` (`[Unreleased]`); long-term/unbounded ideas live in `ROADMAP.md`;
questions that need a human decision live in `ROADMAP.md` → Open questions.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish or blocked.

_Recently completed (now in `CHANGELOG.md` `[0.2.0]`): panic-proof
`computeDelay` (B1/B2/B3), `MaxDelay` validation, no-panic matrix test,
`Backoff`/`ComputeDelay` error-return signature, `FromPolicy` interoperability,
concurrent retry coverage, fuzz seeds, CI, API cross-links, and security policy._

_Previously (`[0.1.0]`): README rewrite, `.golangci.yml`, godoc `Example*`
functions, `BenchmarkComputeDelay`, coverage workflow in CONTRIBUTING,
`docs/DOMAIN_LANGUAGE.md`, Keep-a-Changelog compare links._

---

## P1 — Correctness hardening

### T7. Reconcile `Config` with `errorfamily.RetryPolicy`

`go-error-family` (this package's only dependency) already ships a retry
parameter type in its `retry.go`:

```go
type RetryPolicy struct {
	MaxAttempts int
	MinDelay    time.Duration
	MaxDelay    time.Duration
}
func (f Family) RetryPolicy() RetryPolicy
```

`go-retry` uses that package's `IsRetryable` but **ignores `RetryPolicy`**,
defining a competing `Config` whose fields overlap with different names
(`InitialDelay` vs `MinDelay`). Any consumer of both imports two types that
mean the same thing — a split brain across two libraries by the same author.

Decide and document one direction:

- **A:** Add `retry.FromPolicy(errorfamily.RetryPolicy) Config` so the family's
  advisory defaults feed the loop. Cheap, additive, no breaking change.
- **B:** Deprecate `RetryPolicy` upstream in `go-error-family`, since its own
  doc comment already says "the library does not implement the retry loop."

Raised 2026-08-07 by the `go-sse` adoption review, which declined adoption
partly on this unresolved overlap.

### T1. Close remaining behavioral-guarantee gaps in the test suite

Coverage is 100% by statement, and the core behavioral guarantees are now
asserted (`OnRetry` is not called after the final attempt; a pre-canceled
context yields `ErrCanceled`; `OnExhausted` receives the exact last error by
identity — shipped, see `CHANGELOG.md` `[0.1.0]`). Two exploratory gaps
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

## P2 — Infrastructure

The public remote is live at <https://github.com/LarsArtmann/go-retry>.

### T3. Minimal CI

No `.github/workflows/` exists. Add a workflow running
`go test ./... -race` and `golangci-lint run ./...` on push/PR.

## P3 — Polish

### T4. Cross-link `Backoff` and `ComputeDelay` doc comments

Each refers to the same formula; add a one-line "See [ComputeDelay] for the
raw-parameter variant" / "See [Backoff] for the Config-based variant" so godoc
readers find both (`retry.go:97`, `retry.go:108`).

### T5. Add a `SECURITY.md`

The `LICENSE` is MIT; a dedicated `SECURITY.md` is the conventional place for
vulnerability reporting policy. Low effort.
