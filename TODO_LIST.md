# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md` (`[Unreleased]`); long-term/unbounded ideas live in `ROADMAP.md`;
questions that need a human decision live in `ROADMAP.md` → Open questions.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish or blocked.

---

## P1 — Code quality (fix on sight)

### T1. Reduce `Do` cyclomatic complexity below the lint threshold

`Do` has cyclomatic complexity 13 (`golangci-lint` `cyclop` max is 12). Extract
the delay-computation + `DelayFunc`-override + `OnRetry` block (retry.go lines
70-80) into a helper. Mechanical, low-risk, silences the only lint warning.
**Evidence:** `.golangci.yml` (`cyclop.max-complexity: 12`); `retry.go` (`Do`,
line 44); LSP diagnostic on every file read.

### T2. Modernize benchmark to `b.Loop()`

`BenchmarkComputeDelay` uses `for range b.N` (Go pre-1.24 idiom). gopls flags
`retry_test.go:911` as modernizable to `b.Loop()`. One-line change.
**Evidence:** `retry_test.go` (`BenchmarkComputeDelay`, line 911); gopls
`bloop` diagnostic.

## P2 — Developer experience

### T3. Add `DelayFunc` to the README configuration table

The README configuration table (7 rows) omits `DelayFunc`, which shipped in
v0.3.0. Add a row documenting the field and its zero-return fallback semantics.
**Evidence:** `README.md` (Configuration table); `config.go` (`DelayFunc`,
line 48).

### T4. Add godoc example for `DelayFunc`

`ExampleDo` and `ExampleDo_customIsRetryable` exist but there is no example
showing the `DelayFunc` escape hatch (e.g., honoring an HTTP `Retry-After`
header). A runnable `ExampleDo_delayFunc` with `// Output:` would render on
`pkg.go.dev`. **Evidence:** `retry_test.go` (existing examples, line 844);
`config.go` (`DelayFunc`, line 48).

### T5. Add godoc example for `FromPolicy`

No example shows the `errorfamily.RetryPolicy` → `Config` conversion.
**Evidence:** `config.go` (`FromPolicy`, line 75); `retry_test.go` (existing
examples).

### T6. Add `go vet` to CI

The CI workflow runs `go test -race` and `golangci-lint` but not `go vet`.
Add a vet step (or a combined quality job). **Evidence:** `.github/workflows/ci.yml`.

### T7. Add coverage threshold to CI

Coverage is 100% locally but not enforced in CI. Add a coverage step that fails
below an agreed floor. **Evidence:** `.github/workflows/ci.yml`; `CONTRIBUTING.md`
("Current statement coverage is 100%").

## P3 — Polish

### T8. Run a bounded fuzz campaign on `FuzzComputeDelayNeverPanics`

The fuzz target exists with seed corpus but no sustained fuzz run has been
executed. Run `go test -fuzz=FuzzComputeDelayNeverPanics -fuzztime=5m` and
preserve any useful corpus entries. **Evidence:** `retry_test.go`
(`FuzzComputeDelayNeverPanics`, line 522).
