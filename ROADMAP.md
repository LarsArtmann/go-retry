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

The current release is **v0.7.0** (tagged 2026-09-16). The path to v1.0 is an
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

  **Scope call (2026-09-16, landing `WithJitter`):** `JitterAdditive` (zero
  value = the historical default) and `JitterNone` (pure capped exponential,
  the deterministic escape hatch) ship with the options migration.
  `Full`/`Equal`/`Decorrelated` are recorded as follow-ups — no constants are
  exported for unimplemented strategies (an unknown strategy value today
  falls back to additive, never panics). Revisit when a consumer asks.

- **Is `AttemptFunc(ctx, attempt)` the signature callers want?** Some retry
  libraries pass the previous error back into `fn`; this one does not. Worth a
  deliberate decision, not an accident.

  **Decision (2026-09-13): keep `AttemptFunc(ctx, attempt) error`.** Survey of
  the mainstream Go retry APIs (verified against source): avast/retry-go v5
  uses `RetryableFunc func()` (bare); cenkalti/backoff v5 uses `Operation[T]`
  and its own docs warn "the operation receives no context — capture ctx
  inside the operation if you want cancellation to abort an in-flight
  attempt"; sethvargo/go-retry uses `RetryFunc func(ctx) error`. Against that
  field, this package's signature is a deliberate small superset: passing
  `ctx` avoids cenkalti's documented gotcha entirely, and passing `attempt`
  (1-based) powers the common gate-on-try pattern (`if attempt == 1 { fast
  path }`, attempt-aware logging) without forcing every caller through
  closure counters. Passing the previous `err` into `fn` was considered and
  **rejected**: retryability is `Config.IsRetryable`'s job (single decision
  point), and `OnRetry`/`DelayFunc` already observe the error where it
  belongs. Revisit only if a concrete consumer demonstrates a need that
  `IsRetryable` + callbacks cannot express.
- **Public API surface audit — DONE (2026-09-13, v1.0 track).** Every exported
  symbol walked for purpose, leak-check, doc quality, and need (`go doc -all`
  matched the listed surface exactly — no undocumented exports):

  - [x] `Do` — core loop; doc covers attempt count, retryable semantics,
        context endings, exhaustion. Freeze.
  - [x] `DoWithValue[T]` — value wrapper; zero-value-on-failure guarantee
        documented and pinned. Freeze.
  - [x] `Config` — 8 fields, all domain-meaningful; no clock/rand/logger
        leaks. Struct-literal compat is the v1 promise. Freeze.
  - [x] `Config.Validate` — Rejection-family errors, codes glossary-synced.
        Freeze.
  - [x] `DefaultConfig` — value receiver, valid by construction. Freeze.
  - [x] `FromPolicy` — **flagged, accepted:** signature exposes
        `errorfamily.RetryPolicy` (the one intentional dependency-type leak,
        consistent with `doc.go`'s integration stance); v1 compat tracking
        must follow `RetryPolicy`'s shape. Freeze with note.
  - [x] `Backoff` — config-facing delay incl. `DelayFunc`; `(duration, error)`
        contract. Freeze.
  - [x] `ComputeDelay` — dependency-free pure function; deliberate export.
        Freeze.
  - [x] `AttemptFunc` — signature decision tracked separately (below). Freeze
        pending that record.
  - [x] `ResultFunc[T]` — generic counterpart to `AttemptFunc`. Freeze.
  - [x] `ErrExhausted` / `ErrCanceled` / `ErrDeadlineExceeded` —
        Infrastructure sentinels; code+family identity, unwrap semantics
        documented. Freeze.

  Verdict: **no symbol to remove, rename, or deprecate before v1.0**; the
  only tracked coupling is `FromPolicy`'s `RetryPolicy` parameter.

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

  **Landed (2026-09-16, v0.7.0).** The `Option` tail, the four callback
  mirrors, `WithJitter` (`Additive`/`None`), and `WithRandomSource` shipped
  exactly per this design. Options live in `options.go`; follow-up strategies
  (`Full`/`Equal`/`Decorrelated` jitter) are recorded in the jitter scope
  call below.

  **Design (2026-09-13, v1.0 track).** Goals: grow behavior (jitter strategy,
  RNG source, future hooks) without ever breaking `Config` struct literals;
  keep zero-config usage idiomatic. Non-goals: replacing `Config`, runtime
  reconfiguration, builder chaining as the primary style. **Compat contract:
  the entry points grow a variadic tail — `Do(ctx, config, fn, opts ...Option)`
  and `DoWithValue(ctx, config, fn, opts ...Option)` — which is a purely
  additive, source-compatible change.** Options override the matching
  `Config` field for that call when set; the fields keep working forever, so
  nothing forces migration. Planned `With*` inventory: `WithIsRetryable`,
  `WithDelayFunc`, `WithOnRetry`, `WithExhausted` (mirrors of the four
  callbacks), plus options-only capabilities that never become fields:
  `WithJitter(strategy)` (none / additive / full / equal / decorrelated —
  closes the twice-deferred jitter question) and `WithRandomSource`
  (deterministic RNG seam, see below). New behavior lands options-only so
  `Config`'s 8-field shape survives v1.0 unchanged.

  **Type skeleton (proposed 2026-09-16, from the 2026-09-14 report §f.25 —
  removes the implementer's first decision; override here if a better shape
  emerges):**

  ```go
  // Option overrides one piece of Config for a single call. Mutator shape:
  // Config stays the single source of truth, options compose left-to-right,
  // and an option is a one-line method.
  type Option func(*Config)

  func WithIsRetryable(f func(error) bool) Option { return func(c *Config) { c.IsRetryable = f } }
  func WithDelayFunc(f func(attempt int, err error) time.Duration) Option { return func(c *Config) { c.DelayFunc = f } }
  func WithOnRetry(f func(attempt int, delay time.Duration, err error)) Option { return func(c *Config) { c.OnRetry = f } }
  func WithExhausted(f func(attempts int, err error)) Option { return func(c *Config) { c.OnExhausted = f } }

  // Options-only capabilities — never Config fields.
  type JitterStrategy int

  const (
  	JitterAdditive JitterStrategy = iota // today's hardcoded default
  	JitterNone
  	JitterFull
  	JitterEqual
  	JitterDecorrelated
  )

  func WithJitter(s JitterStrategy) Option        { return func(c *Config) { c.jitterStrategy = s } }
  func WithRandomSource(src rand.Source) Option   { return func(c *Config) { c.randSource = src } }
  ```

  The unexported fields (`jitterStrategy`, `randSource`) keep `Config`'s
  public 8-field shape frozen; `Do`/`DoWithValue` apply `opts` after the
  caller's `Config` and before `Validate`.

  **Deterministic RNG decision (2026-09-13) — LANDED 2026-09-16 as
  `WithRandomSource(rand.Source)` (math/rand/v2; nil = global generator).**
  Resolves the `WORTH_CONSIDERING` item: a pluggable randomness source lands as
  `WithRandomSource(rand.Source)` **with the options migration**, not as a
  `Config` field and not as a package-level test seam. A test seam (mutable
  package var) is rejected — it is shared mutable state under `-race` and
  parallel tests; "never" is rejected because jitter is the package's only
  nondeterminism and delay-comparison tests already paid for it (see
  `TestBackoff_IncreasesExponentially`'s formula-not-samples pattern).
- **Composition primitives (documented, not coded here).** Circuit-breaker,
  bulkhead, and deadline-budgeting are intentionally absent. The roadmap
  question is whether `go-retry` should ship thin **adapters** that compose
  with a caller-chosen breaker, or stay purely a loop and leave composition
  entirely to the caller / to `go-cqrs-lite`. Lean: stay pure, document the
  pattern.

  **Deadline-aware budgeting decision (2026-09-13).** Resolves the
  `WORTH_CONSIDERING` item: stay **count-based**, no time-budget semantics.
  Today a deadline only terminates the loop (`ErrDeadlineExceeded`, wait
  interrupted); attempts are not sized against the remaining ctx budget.
  Auto-fitting delays to the deadline would make the number of attempts
  environment- and clock-dependent, breaking the `MaxAttempts` contract
  ("exactly N calls" is documented and pinned by tests). Callers with a
  deadline budget should set `MaxDelay` below their remaining time — that
  recipe goes in the README instead of new semantics.
- **Version-compatibility matrix with `go-error-family`.** This package
  depends on `go-error-family` (version pinned in `go.mod`) and leans on
  `errorfamily.IsRetryable` as its default retry predicate. As that library
  evolves, document which `go-retry` versions support which `go-error-family`
  majors.

  **Matrix (2026-09-13; rows updated 2026-09-17).** The `go-error-family` API
  surface `go-retry` compiles against (extracted from the `go.mod` pin —
  v0.10.0 at audit time, v0.10.1 since the Dependabot bump; grep-verified):
  `NewInfrastructure`, `NewRejection`, `NewTransient` (tests),
  `WrapInfrastructure`, `IsRetryable`, `Classify` (tests), the
  `Transient`/`Rejection` family constants (tests), and the `RetryPolicy`
  type consumed by `FromPolicy`.

  | go-retry         | go-error-family | Notes                                                                         |
  | ---------------- | --------------- | ----------------------------------------------------------------------------- |
  | v0.1.0–v0.3.1    | v0.9.x          | pre-audit; unverified                                                         |
  | v0.4.0           | v0.10.0         | family/code contract settled                                                  |
  | v0.5.0           | v0.10.0         | —                                                                             |
  | v0.6.0           | v0.10.0         | surface above; minor bumps of go-error-family within v0.x are accepted ad hoc |
  | v0.6.1           | v0.10.1         | patch bump via Dependabot (2026-09-16); surface unchanged, gates green        |
  | v0.7.0           | v0.10.1         | options surface added; consumes no new go-error-family symbol                 |

  Rule: a go-error-family **major** (post-v1) or any change to the surface
  above requires a go-retry minor bump and a new matrix row; the `RetryPolicy`
  shape is the riskiest coupling (see the API-audit `FromPolicy` note).

  Who bumps go-error-family (trace, 2026-09-16): Dependabot's weekly gomod
  watcher opens the PR; the owner (or an agent following the AGENTS
  verify-then-merge flow) SHA-verifies each action/gomod change, runs the
  gates, and merges when green. No one edits the pin by hand.
- **CI hardening ideas (unscoped).** A periodic `-race` fuzz short-run
  (throughput vs concurrency-bug tradeoff); pinning the govulncheck action's
  internal `go install …@latest` posture; an auto-PR loop that lands fuzz
  crashers into `testdata/fuzz/` automatically; `go test -shuffle=on` as a
  cheap order-dependency detector.

  **Decisions (2026-09-13).**
  - `-shuffle=on`: **adopted** on the CI test job (`go test -shuffle=on
    ./... -race`); a local trial (5 shuffled runs + 4 fixed seeds) found no
    order dependency.
  - `-race` fuzz: **rejected for the daily campaign** — measured throughput
    locally is ~19k execs/s with `-race` vs ~400k/s without (~20×), and the
    loop is single-goroutine; caller-side races are covered by the
    always-on `-race` test job. Revisit only if a concurrency bug class
    appears that tests miss.
  - govulncheck `@latest`: **accepted, consciously.** Verified from the
    pinned action source (`action.yml` at our SHA): the install step is
    `go install golang.org/x/vuln/cmd/govulncheck@latest`. The scanner
    binary does not ship in the artifact, the vulnerability DB is live
    regardless, and the action is the official Go-team one at a SHA we pin.
    Self-pinning would trade automatic scanner freshness for a stale-scanner
    failure mode.
  - Auto-PR crash-corpus loop: **design now, build when the first crasher
    appears.** Sketch: the fuzz job already uploads `testdata/fuzz/` as the
    `fuzz-crash-corpus-<sha>` artifact on failure; an auto-PR workflow
    (triggered on that failure) would extract new corpus entries, commit
    them plus a distilled seed to a branch, and open a PR. Cost that keeps
    it deferred: it needs `contents: write` + `pull-requests: write`
    permissions (today the workflows are `contents: read`), and crasher
    triage is inherently human — the corpus-seeds test keeps the manual
    path cheap.
- **CI hardening ideas (unscoped, harvested 2026-09-16 from the 2026-09-14
  report §f).**
  - Remote-action input-allowlist test: the 6 pinned actions' inputs checked
    against a table — would have caught the 2026-09-13 `namee:` typo that
    actionlint cannot see. Weigh maintenance against the gate gap.
  - Concurrency-group shape: `ci-${{ github.ref }}` currently separates the
    tag ref from master even at identical SHAs; sharing one group would
    cancel duplicate runs for the same commit. Needs a deliberate call.
  - Release workflow (tag → `gh release`) vs the manual `go-release` skill
    discipline: automation trades the skill's gate ceremony for speed —
    choose deliberately, not by drift.
  - Verified-in-practice (2026-09-13): `setup-go` with the default
    `go-version-input: stable` coexists with `go-version-file`; the manifest
    lag caveat lives in `AGENTS.md` → Gotchas.
- **Public documentation site.** Other LarsArtmann libraries use the Astro +
  Starlight + Firebase Hosting pattern (see the `website-launch` skill). A
  rendered docs site is plausible once the API is stable. Godoc examples now
  exist (`ExampleDo`, `ExampleDo_customIsRetryable` — see `CHANGELOG.md`
  `[0.1.0]`); the remaining precondition is API stability (see v1.0 bar
  above). Not before.

  **Preconditions checklist (2026-09-13), tied to the v1.0 freeze:**
  1. v1.0.0 tag cut (API-stability promise in force) — hard gate.
  2. Public-API audit verdicts all "freeze" (done 2026-09-13, see the v1.0
     bar above).
  3. Godoc examples cover every entry point — **done**: all seven exported
     entry points carry output-pinned examples (`ExampleBackoff` /
     `ExampleComputeDelay` landed post-v0.7.0).
  4. Launch content decision: demo video per the `website-launch` pattern
     (owner call).
  5. Docs-site content source chosen (README-derived vs dedicated pages) —
     decide at launch time, not before.

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

- **0.x releases: GitHub full release or prerelease?** **Decided (asked
  2026-09-16, owner): full releases for 0.x** (v0.6.1 next, then v0.7.0);
  v1.0 remains a separate later decision.

- **Daemon push policy.** **Decided (asked 2026-09-16, owner): the daemon
  keeps committing AND pushing.** Agent commits amend over daemon races only
  while unpushed, exactly as AGENTS documents.

- **Status-report archive retention.** **Decided (asked 2026-09-16, owner):
  keep all forever** — `docs/status/archived/` grows unbounded by design.

- **Third local tool: checked-in `tools.go` or two-tool convention?**
  **Decided (asked 2026-09-16, owner): the `tools.go` pattern is blessed** for
  local-only tools (govulncheck, formatter); landing it is queued (plan M2/M15
  follow-through), and `actionlint` stays `go run`-pinned until then.

- **Cross-repo consumer bumps: proactive or on-request?** **Decided (asked
  2026-09-16, owner): authorized** — after v0.7.0 ships, bump
  `go-cqrs-lite/middleware/v4` to it (the M9 sweep) without a further ask.

- **`.config/metadata.yaml`: keep or remove?** **Decided (asked 2026-09-16,
  owner): keep.** The external writer owns the file; the AGENTS do-not-edit
  note stands as the permanent convention, not a stopgap.

- **Where do release-notes bodies live?** **Decided (2026-09-13): GitHub-only.**
  Release bodies are composed at release time from the matching `CHANGELOG.md`
  section (the `go-release` skill flow — a curated user-focused summary, not a
  copy). No `docs/releases/` directory: the CHANGELOG is the single in-repo
  copy and GitHub Releases is the presentation layer; a third copy would
  drift. (Resolves TODO_LIST T14.)

_Decided (kept for the record): the repo deliberately uses raw `go` /
`golangci-lint` commands instead of the LarsArtmann `flake.nix` convention —
see `AGENTS.md` → Commands. Do not invent nix targets._

_Decided (asked 2026-09-16, owner): the relaxed `go 1.26` directive in
`go.mod` is intentional — stay on it; do not re-pin to a patch version._

_Reaffirmed (2026-09-17): an external writer re-pinned the directive to
`go 1.27.1`; it was reverted to `go 1.26` (no dependency or code needs 1.27 —
the sole dependency declares `go 1.26`) and `TestModuleGoDirectiveStaysPinned`
now guards the directive, proven failing on drift before it landed._

_Decided (kept for the record, 2026-09-13): no delay-sequence table in the
README. Proposed repeatedly, never demanded; the formula is already documented
in three places, and a static table would be a fourth number to keep in sync
with `computeDelay`. Reopen only if a consumer actually asks._

_Decided (2026-09-16, plan M15): keep the CI coverage floor at 95%. The floor
is a tripwire against catastrophic drift, not a coverage maximizer; local
coverage has held 100% across the repo's life and the Session Ritual enforces
it, while the four points of headroom protect the panic-proof defensive style
(untestable guard branches must not turn CI red). Revisit only if local
coverage ever dips below 100._

_Decided (2026-09-16, plan M15): supply-chain-triggered fuzzing is deferred.
The daily campaign plus the corpus↔seeds guard already bound the risk; a
trigger-on-dependency-PR fuzz would need `pull-requests: write` and secrets
on untrusted branches — cost exceeds benefit until a real supply-chain event
occurs. (Resolves the 2026-09-14 report §f.42 question.)_

_Idea seeds (2026-09-16): pin dprint in a checked-in `tools.go` so the
markdown format gate works offline (owner blessed the tools.go pattern);
a bump-trace audit trail — one line in each go-error-family-bump commit
linking the Dependabot PR and the SHA-verification evidence._
