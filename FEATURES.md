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
  `retry.go` (`Do`, line 44).
- **Exponential backoff with additive jitter** — delay for attempt `n` is
  `InitialDelay * Multiplier^(n-1)`, capped at `MaxDelay`, plus random jitter
  up to 50% of the capped delay. `retry.go` (`Backoff`, line 114;
  `ComputeDelay`, line 125).
- **Backoff is previewable without running the loop** — `Backoff(config, n)` and
  the dependency-free `ComputeDelay(...)` are exported (both return
  `(time.Duration, error)`; an `attempt < 1` yields a `Rejection` error) so
  callers can log/preview the planned delay. The two doc comments cross-link
  to each other. `retry.go` (`Backoff`, `ComputeDelay`).
- **Custom delay override (`DelayFunc`)** — `Config.DelayFunc` optionally
  overrides the exponential backoff for a single attempt. It receives the
  attempt number and the error from the failed attempt, so callers can honor
  server-provided delays (e.g. HTTP `Retry-After`) or implement custom backoff
  strategies. A return `> 0` overrides; `0` means "use the default exponential
  backoff." `config.go` (`DelayFunc` field, line 48), `retry.go` (`Do`).
- **Retry-policy interoperability** — `FromPolicy` maps an
  `errorfamily.RetryPolicy` into this package's `Config`, preserving the
  dependency's advisory attempt and delay defaults while retaining this
  package's multiplier and hooks. `config.go` (`FromPolicy`, line 75).
- **Panic-proof delay computation** — the internal `computeDelay`
  (`retry.go`, line 140) is hardened so no input combination can panic or
  return a negative duration: zero/unset `MaxDelay` degrades to "no growth
  beyond initial", sub-2ns delays skip jitter, and `math.Pow` overflow
  saturates to `MaxDelay` instead of wrapping. Proven by a matrix property test
  and a fuzz target.
- **Context cancellation during backoff** — if `ctx` is canceled while waiting,
  `Do` returns an error wrapping `ErrCanceled` (with the last `fn` error as its
  cause), using a `select` on `timer.C` vs `ctx.Done()`. `retry.go` (`Do`,
  lines 82-93).

### Configuration

- **`Config` with sensible defaults** — `MaxAttempts: 3`, `InitialDelay: 100ms`,
  `MaxDelay: 5s`, `Multiplier: 2.0`. `config.go` (defaults, lines 10-15;
  `DefaultConfig`, line 62).
- **Config validation** — `Validate()` rejects `MaxAttempts < 1`,
  `InitialDelay <= 0`, `MaxDelay <= 0`, and `Multiplier <= 1` with
  `Rejection`-family errors. `config.go` (`Validate`, line 85).
- **Pluggable retryable predicate** — `Config.IsRetryable func(error) bool`;
  when `nil`, `Do` substitutes `errorfamily.IsRetryable`. `config.go`
  (`IsRetryable` field, line 35), `retry.go` (`Do`, lines 49-52).

### Observability hooks

- **Per-attempt callback** — `Config.OnRetry(attempt, delay, err)` fires after a
  failed attempt, before sleeping, only when more attempts remain. When
  `DelayFunc` is set, `OnRetry` receives the `DelayFunc`-computed delay, not
  the exponential one. `config.go` (`OnRetry` field, line 53), `retry.go`
  (`Do`, lines 78-80).
- **Exhaustion callback** — `Config.OnExhausted(attempts, err)` fires once after
  all attempts fail, receiving the exact last error by identity. `config.go`
  (`OnExhausted` field, line 58), `retry.go` (`Do`, lines 96-98).

### Error model (`error-family` integration)

- **Classified sentinel errors** — `ErrExhausted` and `ErrCanceled` are
  `Infrastructure`-family (retry exhaustion / cancellation = downstream
  concern); validation failures are `Rejection`-family (bad caller input).
  `retry.go` (`ErrExhausted`, line 16; `ErrCanceled`, line 23), `config.go`
  (`Validate`).
- **Cause chaining** — exhaustion and cancellation errors wrap the last `fn`
  error via `WithCause`, so `errors.Is(err, lastFnErr)` holds. `retry.go`
  (`Do`, lines 89-91, 100-101).
- **Stable error codes** — `retry.exhausted`, `retry.canceled`,
  `retry.invalid_max_attempts`, `retry.invalid_initial_delay`,
  `retry.invalid_max_delay`, `retry.invalid_multiplier`,
  `retry.invalid_attempt`. `retry.go` (lines 17, 24, 129); `config.go`
  (lines 88, 95, 102, 109).

### Testing guarantees

- **100% statement coverage** — every line of `retry.go` and `config.go` is
  exercised by the test suite. `retry_test.go`.
- **Concurrent isolation** — 100-goroutine test proves `Do` invocations share
  no mutable state (`-race` clean). `retry_test.go`
  (`TestDo_ConcurrentInvocationsShareNoMutableState`).
- **No-panic property test** — sweeps `initial x maxDelay x multiplier x
  attempt` to prove `computeDelay` cannot panic or return negative for any
  reachable input. `retry_test.go` (`TestComputeDelay_NeverPanicsAcrossMatrix`).
- **Fuzz target** — `FuzzComputeDelayNeverPanics` with seeds for ordinary,
  zero-cap, overflow, and near-`MaxInt64` inputs. `retry_test.go`.
- **Behavioral guarantees** — `OnRetry` not called after the final failure;
  pre-canceled context yields `ErrCanceled`; `OnExhausted` receives the exact
  last error by identity; `DelayFunc` receives the error and its delay
  propagates to `OnRetry`. `retry_test.go`.

### Documentation & developer experience

- **Runnable godoc examples** — `ExampleDo` (success path) and
  `ExampleDo_customIsRetryable` (custom predicate) are deterministic, carry
  `// Output:` comments, and render on `pkg.go.dev`. `retry_test.go`.
- **Backoff benchmark** — `BenchmarkComputeDelay` documents the hot-path cost
  (~18 ns/op, 0 allocations; the jitter path allocates nothing).
  `retry_test.go`.
- **Committed lint config** — `.golangci.yml` (v2) enables the default linters
  plus `gosec`, `mnd`, `exhaustruct`; the in-source `//nolint:` markers are
  verified live. `.golangci.yml`.
- **Domain glossary** — `docs/DOMAIN_LANGUAGE.md` defines the retry and
  `error-family` vocabulary and the `retry.<event>` code table.
- **CI workflow** — `.github/workflows/ci.yml` runs `go test ./... -race` and
  `golangci-lint` on every push and pull request.

## PARTIALLY_FUNCTIONAL

_None._

## BROKEN

_None._

## PLANNED

_None in code or docs._ (See `WORTH_CONSIDERING` below for candidate work, and
`TODO_LIST.md` for committed short-term work.)

## WORTH_CONSIDERING

These are uncommitted ideas — no design, no code. They are candidates for
graduation into `TODO_LIST.md` once scoped.

- **Configurable jitter factor** — jitter is currently hardcoded to "up to 50%
  of the delay" (`retry.go`, `computeDelay`, line 172). A `Config.JitterFactor`
  (or `Jitter: none | additive | full`) would let callers disable jitter for
  deterministic tests or tune spread. **Deferred (2026-08-08):** `DelayFunc`
  already covers the custom-delay escape hatch (compute pure exponential in the
  callback for zero jitter); jitter config will land with the options-pattern
  migration (see `ROADMAP.md` v1.0 section). Tradeoff: another `Config` field
  to validate.
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
