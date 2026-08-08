# Features

Honest inventory of what `go-retry` does, by status. Code is the source of
truth — every entry cites where it lives. Status vocabulary: `FULLY_FUNCTIONAL`
(works, exercised by passing tests), `PARTIALLY_FUNCTIONAL` (ships with known
gaps), `BROKEN` (present but failing), `PLANNED` (no code yet), and
`WORTH_CONSIDERING` (idea, not committed).

_Test status: `go test ./... -race` is green; statement coverage is 100%
(`go test -cover` / `go tool cover -func=reports/coverage.out`)._

---

## FULLY_FUNCTIONAL

### Core retry loop

- **Retry with configurable max attempts** — `Do` calls `AttemptFunc` up to
  `Config.MaxAttempts` times, returning immediately on the first `nil`.
  `retry.go:43` (`func Do`).
- **Exponential backoff with additive jitter** — delay for attempt `n` is
  `InitialDelay * Multiplier^(n-1)`, capped at `MaxDelay`, plus random jitter
  up to 50% of the capped delay. `retry.go:107` (`Backoff`),
  `retry.go:118` (`ComputeDelay`).
- **Backoff is previewable without running the loop** — `Backoff(config, n)` and
  the dependency-free `ComputeDelay(...)` are exported (both return
  `(time.Duration, error)`; an `attempt < 1` yields a `Rejection` error) so
  callers can log/preview the planned delay. `retry.go:107`, `retry.go:118`.
- **Retry-policy interoperability** — `FromPolicy` maps an
  `errorfamily.RetryPolicy` into this package's `Config`, preserving the
  dependency's advisory attempt and delay defaults while retaining this
  package's multiplier and hooks. `config.go`.
- **Panic-proof delay computation** — the internal `computeDelay`
  (`retry.go:133`) is hardened so no input combination can panic or return a
  negative duration: zero/unset `MaxDelay` degrades to "no growth beyond
  initial", sub-2ns delays skip jitter, and `math.Pow` overflow saturates to
  `MaxDelay` instead of wrapping. Proven by a matrix property test.
- **Context cancellation during backoff** — if `ctx` is canceled while waiting,
  `Do` returns an error wrapping `ErrCanceled` (with the last `fn` error as its
  cause), using a `select` on `timer.C` vs `ctx.Done()`. `retry.go:78-87`.

### Configuration

- **`Config` with sensible defaults** — `MaxAttempts: 3`, `InitialDelay: 100ms`,
  `MaxDelay: 5s`, `Multiplier: 2.0`. `config.go:10-15` (defaults),
  `config.go:49` (`DefaultConfig`).
- **Config validation** — `Validate()` rejects `MaxAttempts < 1`,
  `InitialDelay <= 0`, `MaxDelay <= 0`, and `Multiplier <= 1` with
  `Rejection`-family errors. `config.go:60-89`.
- **Pluggable retryable predicate** — `Config.IsRetryable func(error) bool`;
  when `nil`, `Do` substitutes `errorfamily.IsRetryable`. `config.go:35`,
  `retry.go:48-51`.

### Observability hooks

- **Per-attempt callback** — `Config.OnRetry(attempt, delay, err)` fires after a
  failed attempt, before sleeping, only when more attempts remain.
  `config.go:40`, `retry.go:71-73`.
- **Exhaustion callback** — `Config.OnExhausted(attempts, err)` fires once after
  all attempts fail. `config.go:45`, `retry.go:89-91`.

### Error model (`error-family` integration)

- **Classified sentinel errors** — `ErrExhausted` and `ErrCanceled` are
  `Infrastructure`-family (retry exhaustion / cancellation = downstream concern);
  validation failures are `Rejection`-family (bad caller input). `retry.go:15-25`,
  `config.go:62-79`.
- **Cause chaining** — exhaustion and cancellation errors wrap the last `fn`
  error via `WithCause`, so `errors.Is(err, lastFnErr)` holds. `retry.go:82-94`.
- **Stable error codes** — `retry.exhausted`, `retry.canceled`,
  `retry.invalid_max_attempts`, `retry.invalid_initial_delay`,
  `retry.invalid_max_delay`, `retry.invalid_multiplier`,
  `retry.invalid_attempt`. `retry.go:16,22,121`; `config.go:63,69,77,85`.

### Documentation & developer experience

- **Runnable godoc examples** — `ExampleDo` (success path) and
  `ExampleDo_customIsRetryable` (custom predicate) are deterministic, carry
  `// Output:` comments, and render on `pkg.go.dev`. `retry_test.go`.
- **Backoff benchmark** — `BenchmarkComputeDelay` documents the hot-path cost
  (~18 ns/op, 0 allocations; the jitter path allocates nothing). `retry_test.go`.
- **Committed lint config** — `.golangci.yml` (v2) enables the default linters
  plus `gosec`, `mnd`, `exhaustruct`; the in-source `//nolint:` markers are
  verified live. `.golangci.yml`.
- **Domain glossary** — `docs/DOMAIN_LANGUAGE.md` defines the retry and
  `error-family` vocabulary and the `retry.<event>` code table.

## PARTIALLY_FUNCTIONAL

_None._

## BROKEN

_None._

## PLANNED

_None in code or docs._ (See `WORTH_CONSIDERING` below for candidate work, and
`TODO_LIST.md` for committed short-term work.)

## WORTH CONSIDERING

These are uncommitted ideas — no design, no code. They are candidates for
graduation into `TODO_LIST.md` once scoped.

- **Configurable jitter factor** — jitter is currently hardcoded to "up to 50%
  of the delay" (`retry.go:167`). A `Config.JitterFactor` (or `Jitter: none |
  additive | full`) would let callers disable jitter for deterministic tests or
  tune spread. **Deferred (2026-08-08):** `DelayFunc` already covers the
  custom-delay escape hatch (compute pure exponential in the callback for zero
  jitter); jitter config will land with the options-pattern migration (see
  `ROADMAP.md` v1.0 section). Tradeoff: another `Config` field to validate.
- **Deterministic RNG option** — `ComputeDelay` uses `math/rand/v2` globally
  (`retry.go:7`); a pluggable `rand` source would make delay sequences
  reproducible in tests without sampling-based assertions (the existing
  `TestBackoff_IncreasesExponentially` works around this by testing the formula,
  not sampled values).
- **Deadline-aware attempt budgeting** — currently `MaxAttempts` is the only
  budget; a caller with a hard deadline cannot ask `Do` to stop retrying when
  the remaining context budget is too small for another attempt. Possibly out of
  scope (callers can cancel the context), but worth a decision.
- **Composable neighbors in this package?** — circuit-breaker / bulkhead
  primitives are intentionally NOT here (they belong closer to the caller or in
  `go-cqrs-lite/middleware`). Document the boundary explicitly in `doc.go` if
  asked again, rather than adding code.
