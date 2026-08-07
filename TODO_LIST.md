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

## P2 — Infrastructure

The public remote is live at <https://github.com/LarsArtmann/go-retry>.

## P3 — Polish

_None currently._
