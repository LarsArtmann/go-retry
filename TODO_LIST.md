# TODO List

Short-term, **actionable** open work for `go-retry`. Each item is bounded and
cites its evidence. This file lists open work only — completed items move to
`CHANGELOG.md` (`[Unreleased]`); long-term/unbounded ideas live in `ROADMAP.md`;
questions that need a human decision live in `ROADMAP.md` → Open questions.

Priority uses a simple Pareto ranking: **P1** = high impact, do first;
**P2** = valuable but not blocking; **P3** = nice-to-have polish or blocked.

---

## P3 — Polish

### T8. Run a bounded fuzz campaign on `FuzzComputeDelayNeverPanics`

The fuzz target exists with seed corpus but no sustained fuzz run has been
executed. Run `go test -fuzz=FuzzComputeDelayNeverPanics -fuzztime=5m` and
preserve any useful corpus entries. **Evidence:** `retry_test.go`
(`FuzzComputeDelayNeverPanics`).

---

## Done (harvested into CHANGELOG `[Unreleased]` / shipped directly)

- ~~T1. Reduce `Do` cyclomatic complexity below the lint threshold~~ —
  `awaitBackoff`/`nextDelay` extraction, complexity 13 → below 12
  (shipped in v0.4.0).
- ~~T2. Modernize benchmark to `b.Loop()`~~ — `BenchmarkComputeDelay`
  migrated (see `[Unreleased]`).
- ~~T3. Add `DelayFunc` to the README configuration table~~ — already
  present after the 2026-08-08 docs rebuild; verified during the v0.4.0
  docs pass.
- ~~T4. Add godoc example for `DelayFunc`~~ — `ExampleDo_delayFunc`
  (shipped in v0.4.0).
- ~~T5. Add godoc example for `FromPolicy`~~ — `ExampleFromPolicy`
  (shipped in v0.4.0).
- ~~T6. Add `go vet` to CI~~ — vet step in the `test` job.
- ~~T7. Add coverage threshold to CI~~ — `coverage` job with a 95% floor.
