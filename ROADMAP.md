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

The current release is **v0.6.0** (tagged 2026-09-13). The path to v1.0 is an
**API-stability promise**, not a feature list. Open questions to resolve before
v1.0:

- **Is `Config`'s shape final?** The `WORTH_CONSIDERING` items in `FEATURES.md`
  (configurable jitter factor, deterministic RNG) would add fields. Decide
  before freezing, or commit to adding new behavior via options-style
  extension so the struct can stay compatible.

  **Decision (2026-08-08): defer configurable jitter.** The existing `DelayFunc`
  field already provides a full escape hatch — a caller who needs pure
  exponential (zero jitter) can compute `initial * mult^(n-1)` in the callback.
  The hardcoded additive jitter (up to 50%) is a sound default that prevents
  thundering-herd problems without requiring callers to opt in. Adding a
  `Jitter` field now would prematurely freeze the API shape before the planned
  options-pattern migration, and doing it properly requires deciding on jitter
  _strategy_ (none / additive / full / equal / decorrelated), not just a numeric
  factor. Defer until options-based configuration, where `WithJitter(...)` can
  land without breaking existing struct literals.

- **Is `AttemptFunc(ctx, attempt)` the signature callers want?** Some retry
  libraries pass the previous error back into `fn`; this one does not. Worth a
  deliberate decision, not an accident.
- **Public API surface audit** — confirm every exported symbol
  (`Do`, `DoWithValue`, `Config`, `DefaultConfig`, `FromPolicy`, `Backoff`,
  `ComputeDelay`, `AttemptFunc`, `ResultFunc`, `ErrExhausted`, `ErrCanceled`,
  `ErrDeadlineExceeded`) and every exported `Config`
  field (`MaxAttempts`, `InitialDelay`, `MaxDelay`, `Multiplier`, `IsRetryable`,
  `DelayFunc`, `OnRetry`, `OnExhausted`) is one callers should depend on, and
  that nothing exported is leaking an implementation detail.

## Raw ideas (unscoped)

- **`OnSuccess(attempts)` hook.** Would fire once when the loop succeeds
  after retries, receiving the attempt count — the symmetric bookend to
  `OnExhausted` (metrics: an "attempts until success" histogram).
  **Deferred (2026-08-22):** no consumer has asked for it; every caller
  that needs the count can already derive it (`attempt == n` on the
  successful `fn` invocation, or a wrapper around `fn`); and each new
  `Config` hook widens the API surface that must survive the planned
  options-pattern migration. Revisit when a concrete consumer exists.
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
  depends on `go-error-family v0.10.0` (`go.mod`) and leans on
  `errorfamily.IsRetryable` as its default retry predicate. As that library
  evolves, document which `go-retry` versions support which `go-error-family`
  majors.
- **CI hardening ideas (unscoped).** A periodic `-race` fuzz short-run
  (throughput vs concurrency-bug tradeoff); pinning the govulncheck action's
  internal `go install …@latest` posture; an auto-PR loop that lands fuzz
  crashers into `testdata/fuzz/` automatically; `go test -shuffle=on` as a
  cheap order-dependency detector.
- **Public documentation site.** Other LarsArtmann libraries use the Astro +
  Starlight + Firebase Hosting pattern (see the `website-launch` skill). A
  rendered docs site is plausible once the API is stable. Godoc examples now
  exist (`ExampleDo`, `ExampleDo_customIsRetryable` — see `CHANGELOG.md`
  `[0.1.0]`); the remaining precondition is API stability (see v1.0 bar
  above). Not before.

## Explicit non-goals

- **No CQRS message types** in this package (lives in `go-cqrs-lite/middleware`).
- **No OpenTelemetry SDK dependency** (same reason; `OnRetry`/`OnExhausted` are
  the integration seam).
- **No built-in HTTP/database retry helpers.** This is a primitive, not a
  batteries-included toolkit; callers bring their own `AttemptFunc`.

## Open questions

Unresolved decisions that need a human (they are _not_ TODO tasks). They block
parts of the docs/release flow, so they live here rather than rotting in a
status report.

- **0.x releases: GitHub full release or prerelease?** The `go-release` skill
  defaults 0.x to prereleases, but practice in these repos is full releases
  (wise-go v0.9.0; go-retry v0.4.0 and v0.5.0 were published as full
  releases, following the owner's demonstrated preference). Confirm full
  releases for 0.x going forward so the skill default stops fighting
  practice. (Raised in
  `docs/status/archived/2026-08-22_01-20_go-retry-v0.4.0-hardening-executed.md`,
  Q3.)

- **Where do release-notes bodies live?** **Decided (2026-09-13): GitHub-only.**
  Release bodies are composed at release time from the matching `CHANGELOG.md`
  section (the `go-release` skill flow — a curated user-focused summary, not a
  copy). No `docs/releases/` directory: the CHANGELOG is the single in-repo
  copy and GitHub Releases is the presentation layer; a third copy would
  drift. (Resolves TODO_LIST T14.)

_Decided (kept for the record): the repo deliberately uses raw `go` /
`golangci-lint` commands instead of the LarsArtmann `flake.nix` convention —
see `AGENTS.md` → Commands. Do not invent nix targets._

_Decided (kept for the record, 2026-09-13): no delay-sequence table in the
README. Proposed repeatedly, never demanded; the formula is already documented
in three places, and a static table would be a fourth number to keep in sync
with `computeDelay`. Reopen only if a consumer actually asks._
