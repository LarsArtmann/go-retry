# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md` (`[Unreleased]`); long-term/unbounded ideas live in `ROADMAP.md`;
questions that need a human decision live in `ROADMAP.md` → Open questions.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish or blocked.

_Recently completed (now in `CHANGELOG.md` `[0.1.0]`): README rewrite,
`.golangci.yml`, godoc `Example*` functions, `BenchmarkComputeDelay`, coverage
workflow in CONTRIBUTING, `docs/DOMAIN_LANGUAGE.md`, Keep-a-Changelog compare
links._

---

## P1 — Correctness hardening

### T6. Fix three reachable `Int64N` panics in `computeDelay`

**Highest-priority item in this file.** Found 2026-08-07 while evaluating
`go-retry` for adoption in `go-sse`; all three reproduced with runnable
programs, not inferred. Root cause is `retry.go:137-139`:

```go
delay += time.Duration(rand.Int64N(int64(delay) / 2))
```

`rand.Int64N` panics on a non-positive argument, and three distinct paths
reach it:

| #  | Trigger                                                                                                                                                                                                                    | Reachable via                                                    |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| B1 | **`MaxDelay` is never validated.** `Config.Validate()` checks `MaxAttempts`, `InitialDelay`, and `Multiplier` but not `MaxDelay`. Unset `MaxDelay` gives `min(delay, 0) == 0` → `Int64N(0)` → panic.                       | `Do` with any `Config` literal that omits `MaxDelay`.            |
| B2 | **Sub-2 ns delays.** Any `delay < 2ns` makes `int64(delay)/2 == 0` → panic. `InitialDelay: 1` currently passes `Validate()`.                                                                                               | `Do`, `Backoff`, `ComputeDelay`.                                 |
| B3 | **`math.Pow` overflow.** Out-of-range `float64 → time.Duration` yields `INT64_MIN` on amd64; `min(negative, MaxDelay)` keeps the negative → panic. **Plain `DefaultConfig()` panics at attempt 38.**                       | `Do` with `MaxAttempts >= 38`, or a large `Multiplier`.          |

Observed:

```text
Config{MaxAttempts:3,InitialDelay:10ms,Mult:2}   PANIC: invalid argument to Int64N
overflow via Multiplier=10, MaxAttempts=15       PANIC: invalid argument to Int64N
DefaultConfig() overflows at attempt=38
```

A library retry loop must **never** panic — it sits in the failure path, so a
panic here converts a recoverable downstream blip into a process crash.

**Verified fix** (validated against an 84,000-case matrix of
`initial × maxDelay × multiplier × attempt`, asserting no panic, no negative
delay, and respect for the documented `MaxDelay + 50%` bound):

```go
func computeDelay(initial, maxDelay time.Duration, multiplier float64, attempt int) time.Duration {
	if initial <= 0 {
		return 0
	}
	if maxDelay <= 0 { // B1: treat an unset cap as "no growth beyond initial"
		maxDelay = initial
	}

	scaled := float64(initial) * math.Pow(multiplier, float64(attempt-1))

	// B3: compare in float space before converting, so an out-of-range
	// value saturates to maxDelay instead of wrapping to INT64_MIN.
	var delay time.Duration
	if scaled >= float64(maxDelay) || math.IsInf(scaled, 1) || math.IsNaN(scaled) {
		delay = maxDelay
	} else {
		delay = min(time.Duration(scaled), maxDelay)
	}
	if delay <= 0 {
		return 0
	}

	half := int64(delay) / 2
	if half <= 0 { // B2: delay < 2ns has no room for jitter
		return delay
	}

	jitter := time.Duration(rand.Int64N(half))
	if delay > math.MaxInt64-jitter { // saturate rather than wrap
		return math.MaxInt64
	}

	return delay + jitter
}
```

Also required alongside the fix:

- Add a `MaxDelay` check to `Config.Validate()` (reject `<= 0` with a
  `retry.invalid_max_delay` Rejection), **or** document that zero means
  "cap at `InitialDelay`" — pick one and test it.
- Land the `ComputeDelay` fuzz target already listed in T1; it is exactly the
  test that would have caught B2 and B3.

Note that 100% statement coverage did not catch any of these — every panicking
path executes the same three lines that the existing tests already cover with
benign inputs. This is a coverage-vs-input-domain gap, not a missing-line gap.

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
