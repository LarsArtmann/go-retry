# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md` (`[Unreleased]`); long-term/unbounded ideas live in `ROADMAP.md`;
questions that need a human decision live in `ROADMAP.md` → Open questions.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish or blocked.

---

## P1 — Code quality (fix on sight)

### T2. Modernize benchmark to `b.Loop()`

`BenchmarkComputeDelay` uses `for range b.N` (Go pre-1.24 idiom). gopls flags
`retry_test.go` as modernizable to `b.Loop()`. One-line change.
**Evidence:** `retry_test.go` (`BenchmarkComputeDelay`); gopls `bloop`
diagnostic.

## P2 — Developer experience

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
(`FuzzComputeDelayNeverPanics`).

---

## Done (harvested into CHANGELOG `[0.4.0]`)

- ~~T1. Reduce `Do` cyclomatic complexity below the lint threshold~~ —
  `awaitBackoff`/`nextDelay` extraction, complexity 13 → below 12.
- ~~T3. Add `DelayFunc` to the README configuration table~~ — already
  present after the 2026-08-08 docs rebuild; verified during the v0.4.0
  docs pass.
- ~~T4. Add godoc example for `DelayFunc`~~ — `ExampleDo_delayFunc`.
- ~~T5. Add godoc example for `FromPolicy`~~ — `ExampleFromPolicy`.
