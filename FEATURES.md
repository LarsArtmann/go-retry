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
  up to 50% of the capped delay. `retry.go:104` (`Backoff`),
  `retry.go:114` (`ComputeDelay`).
- **Backoff is previewable without running the loop** — `Backoff(config, n)` and
  the dependency-free `ComputeDelay(...)` are exported so callers can log/preview
  the planned delay. `retry.go:104`, `retry.go:114`.
- **Context cancellation during backoff** — if `ctx` is canceled while waiting,
  `Do` returns an error wrapping `ErrCanceled` (with the last `fn` error as its
  cause), using a `select` on `timer.C` vs `ctx.Done()`. `retry.go:75-86`.

### Configuration

- **`Config` with sensible defaults** — `MaxAttempts: 3`, `InitialDelay: 100ms`,
  `MaxDelay: 5s`, `Multiplier: 2.0`. `config.go:10-15` (defaults),
  `config.go:49` (`DefaultConfig`).
- **Config validation** — `Validate()` rejects `MaxAttempts < 1`,
  `InitialDelay <= 0`, and `Multiplier <= 1` with `Rejection`-family errors.
  `config.go:60-83`.
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
  `retry.invalid_multiplier`. `retry.go:16,22`; `config.go:63,69,76`.

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
  of the delay" (`retry.go:121`). A `Config.JitterFactor` (or `Jitter: none |
  additive | full`) would let callers disable jitter for deterministic tests or
  tune spread. Tradeoff: another `Config` field to validate.
- **Deterministic RNG option** — `ComputeDelay` uses `math/rand/v2` globally
  (`retry.go:6`); a pluggable `rand` source would make delay sequences
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
