# Roadmap

Long-term direction and **raw ideas** that are not yet refined into actionable
tasks. Anything bounded and estimable lives in `TODO_LIST.md` instead; this file
holds the bigger, fuzzier questions. Nothing here is a commitment — items
graduate to `TODO_LIST.md` when they get scoped, and to `CHANGELOG.md` when they
ship.

---

## Direction

`go-retry` is intentionally a **small, dependency-light core**: a retry loop
with exponential backoff and jitter, plus the `error-family` integration. Its
reason to exist is that consumers who only need retry (CLIs, batch jobs, simple
services) can import it **without** pulling in CQRS message types or the
OpenTelemetry SDK (`doc.go:1-9`). Every addition should be judged against that
boundary: if a feature needs CQRS or OTel types, it belongs in
`github.com/larsartmann/go-cqrs-lite/middleware/v4`, not here.

## v1.0 — what is the bar?

v0.1.0 shipped today (2026-08-03, tag `v0.1.0`, signed). The path to v1.0 is an
**API-stability promise**, not a feature list. Open questions to resolve before
v1.0:

- **Is `Config`'s shape final?** The `WORTH_CONSIDERING` items in `FEATURES.md`
  (configurable jitter factor, deterministic RNG) would add fields. Decide
  before freezing, or commit to adding new behavior via options-style
  extension so the struct can stay compatible.
- **Is `AttemptFunc(ctx, attempt)` the signature callers want?** Some retry
  libraries pass the previous error back into `fn`; this one does not. Worth a
  deliberate decision, not an accident.
- **Public API surface audit** — confirm every exported symbol
  (`Do`, `Config`, `DefaultConfig`, `Backoff`, `ComputeDelay`, `AttemptFunc`,
  `ErrExhausted`, `ErrCanceled`) is one callers should depend on, and that
  nothing exported is leaking an implementation detail.

## Raw ideas (unscoped)

- **Options-based configuration.** Migrate optional `Config` behavior
  (`OnRetry`, `OnExhausted`, a future jitter config) to functional options
  (`WithOnRetry(...)`) so new capabilities don't break the struct literal
  callers already have. Large change; needs a concrete migration story.
- **Composition primitives (documented, not coded here).** Circuit-breaker,
  bulkhead, and deadline-budgeting are intentionally absent. The roadmap
  question is whether `go-retry` should ship thin **adapters** that compose
  with a caller-chosen breaker, or stay purely a loop and leave composition
  entirely to the caller / to `go-cqrs-lite`. Lean: stay pure, document the
  pattern.
- **Version-compatibility matrix with `go-error-family`.** This package
  depends on `go-error-family v0.10.0` (`go.mod:5`) and leans on
  `errorfamily.IsRetryable` as its default retry predicate. As that library
  evolves, document which `go-retry` versions support which `go-error-family`
  majors.
- **Public documentation site.** Other LarsArtmann libraries use the Astro +
  Starlight + Firebase Hosting pattern (see the `website-launch` skill). A
  rendered docs site is plausible once the API is stable and examples exist
  (see `TODO_LIST.md` T3). Not before.
- **Fuzzing.** `ComputeDelay` is pure numeric code taking caller-controlled
  inputs; a `go test -fuzz` target could harden the overflow / negative-delay
  edges that the current table tests sample only spot-check. Candidate once the
  function's invariants are written down.

## Explicit non-goals

- **No CQRS message types** in this package (lives in `go-cqrs-lite/middleware`).
- **No OpenTelemetry SDK dependency** (same reason; `OnRetry`/`OnExhausted` are
  the integration seam).
- **No built-in HTTP/database retry helpers.** This is a primitive, not a
  batteries-included toolkit; callers bring their own `AttemptFunc`.
