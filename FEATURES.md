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
  `retry.go` (`Do`).
- **Value-returning retries (`DoWithValue[T]`)** — generic companion to `Do`
  for retries that produce a value: returns the successful attempt's result
  with a nil error, and the zero value with the same error `Do` would produce
  on any failure (non-retryable, exhaustion, context end). `retry.go`
  (`DoWithValue`), `retry_test.go`.
- **Exponential backoff with additive jitter, hard-capped** — delay for
  attempt `n` is `min(InitialDelay * Multiplier^(n-1) + jitter, MaxDelay)`
  where jitter is up to 50% of the capped exponential delay; the returned
  value never exceeds `MaxDelay` (20 000-sample regression test).
  `retry.go` (`Backoff`, `ComputeDelay`).
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
  backoff." `config.go` (`DelayFunc` field), `retry.go` (`Do`).
- **Retry-policy interoperability** — `FromPolicy` maps an
  `errorfamily.RetryPolicy` into this package's `Config`, preserving the
  dependency's advisory attempt and delay defaults while retaining this
  package's multiplier and hooks. `config.go` (`FromPolicy`).
- **Panic-proof delay computation** — the internal `computeDelay` (`retry.go`)
  is hardened so no input combination can panic or
  return a negative duration: zero/unset `MaxDelay` degrades to "no growth
  beyond initial", sub-2ns delays skip jitter, and `math.Pow` overflow
  saturates to `MaxDelay` instead of wrapping. Proven by a matrix property test
  and a fuzz target.
- **Context endings distinguished during backoff** — if the context ends
  while waiting, `Do` branches on `ctx.Err()`: an expired deadline returns
  an error matching `ErrDeadlineExceeded` (also unwrapping to
  `context.DeadlineExceeded`), an explicit cancel returns `ErrCanceled`
  (unwrapping to `context.Canceled`). Both keep the last `fn` error in the
  chain. `retry.go` (`awaitBackoff`, `contextEnded`).

### Configuration

- **`Config` with sensible defaults** — `MaxAttempts: 3`, `InitialDelay: 100ms`,
  `MaxDelay: 5s`, `Multiplier: 2.0`. `config.go` (defaults;
  `DefaultConfig`).
- **Config validation** — `Validate()` rejects `MaxAttempts < 1`,
  `InitialDelay <= 0`, `MaxDelay <= 0`, and `Multiplier <= 1` with
  `Rejection`-family errors. `config.go` (`Validate`).
- **Pluggable retryable predicate** — `Config.IsRetryable func(error) bool`;
  when `nil`, `Do` substitutes `errorfamily.IsRetryable`. `config.go`
  (`IsRetryable` field), `retry.go` (`Do`).

### Observability hooks

- **Per-attempt callback** — `Config.OnRetry(attempt, delay, err)` fires after a
  failed attempt, before sleeping, only when more attempts remain. When
  `DelayFunc` is set, `OnRetry` receives the `DelayFunc`-computed delay, not
  the exponential one. `config.go` (`OnRetry` field), `retry.go`
  (`awaitBackoff`).
- **Exhaustion callback** — `Config.OnExhausted(attempts, err)` fires once after
  all attempts fail, receiving the exact last error by identity — and never
  when a context end terminates the loop. `config.go` (`OnExhausted` field),
  `retry.go` (`Do`).

### Error model (`error-family` integration)

- **Classified sentinel errors** — `ErrExhausted`, `ErrCanceled`, and
  `ErrDeadlineExceeded` are `Infrastructure`-family (retry exhaustion /
  context endings = downstream concern); validation failures are
  `Rejection`-family (bad caller input). `retry.go` (`ErrExhausted`,
  `ErrCanceled`, `ErrDeadlineExceeded`), `config.go` (`Validate`).
- **Cause chaining** — the exhaustion error wraps the last `fn` error via
  `WithCause`, so `errors.Is(err, lastFnErr)` holds; cancel/deadline errors
  wrap both the context error and the last `fn` error (Go 1.20 multi-`%w`).
  Nested loops are fail-closed: `ErrExhausted` is `Infrastructure`, which
  the default `IsRetryable` predicate does not retry, so an outer loop
  treats an inner loop's exhaustion as terminal. `retry.go` (`Do`,
  `contextEnded`).
- **Stable error codes** — `retry.exhausted`, `retry.canceled`,
  `retry.deadline`, `retry.invalid_max_attempts`,
  `retry.invalid_initial_delay`, `retry.invalid_max_delay`,
  `retry.invalid_multiplier`, `retry.invalid_attempt`. `retry.go`;
  `config.go` (`Validate`).

### Testing guarantees

- **100% statement coverage** — every line of `retry.go` and `config.go` is
  exercised by the test suite. `retry_test.go`.
- **Concurrent isolation** — 100-goroutine test proves `Do` invocations share
  no mutable state (`-race` clean). `retry_test.go`
  (`TestDo_ConcurrentInvocationsShareNoMutableState`).
- **No-panic property test** — sweeps `initial x maxDelay x multiplier x
attempt` to prove `computeDelay` cannot panic or return negative for any
  reachable input. `retry_test.go` (`TestComputeDelay_NeverPanicsAcrossMatrix`).
- **Nested loops are fail-closed (pinned, both directions)** — an outer `Do`
  makes exactly one attempt when an inner loop returns `ErrExhausted`, because
  `Infrastructure` is not retryable by default
  (`TestDo_NestedRetriesAreFailClosed`); a deliberate `IsRetryable` override
  re-enables amplification, pinned as `outer(3) × inner(3) = 9` attempts
  (`TestDo_NestedRetriesAmplifyWhenOverridden`). `retry_test.go`.
- **`OnExhausted` never fires on context end (pinned)** — cancellation and
  deadline termination return without the exhaustion callback. `retry_test.go`
  (`TestDo_OnExhaustedNotCalledOnCancel`,
  `TestDo_OnExhaustedNotCalledOnDeadline`).
- **Fuzz target** — `FuzzComputeDelayNeverPanics` with seeds for ordinary,
  zero-cap, overflow, and near-`MaxInt64` inputs. The seed corpus is also
  committed in `testdata/fuzz/FuzzComputeDelayNeverPanics/`, and a scheduled
  CI workflow fuzzes daily for 30 minutes. `retry_test.go`, `testdata/fuzz/`,
  `.github/workflows/fuzz.yml`.
- **Behavioral guarantees** — `OnRetry` not called after the final failure;
  a pre-canceled context yields `ErrCanceled`; a deadline exceeded during
  backoff yields `ErrDeadlineExceeded` matching `context.DeadlineExceeded`
  and not `ErrCanceled`; `OnExhausted` receives the exact last error by
  identity; `DelayFunc` receives the error and its delay propagates to
  `OnRetry`. `retry_test.go`.

### Documentation & developer experience

- **Runnable godoc examples** — `ExampleDo` (success path),
  `ExampleDo_customIsRetryable` (custom predicate), `ExampleDo_delayFunc`
  (server-provided Retry-After), `ExampleFromPolicy` (error-family
  policy → `Config`), and `ExampleDoWithValue` (value-returning API) are
  deterministic, carry `// Output:` comments, and render on `pkg.go.dev`.
  `retry_test.go`.
- **Backoff benchmark** — `BenchmarkComputeDelay` documents the hot-path cost
  (~20–35 ns/op depending on machine load; 0 allocations — the jitter path
  allocates nothing). `retry_test.go`.
- **Committed lint config** — `.golangci.yml` (v2) enables the standard
  defaults plus ~100 extra linters (`gosec`, `mnd`, `exhaustruct_v5` among
  them); the in-source `//nolint:` markers are verified live.
  `.golangci.yml`.
- **Domain glossary** — `docs/DOMAIN_LANGUAGE.md` defines the retry and
  `error-family` vocabulary and the `retry.<event>` code table.
- **CI workflow** — `.github/workflows/ci.yml` runs `go vet`,
  `go test ./... -race`, and a `govulncheck` vulnerability scan, lints via
  golangci-lint (version pinned in `.github/workflows/ci.yml`), and enforces a 95%
  coverage floor on every push and pull request; each job carries a
  10-minute timeout and pushes to the same ref cancel superseded runs.
  Verified green on real runners for the current tip (run 34755167105,
  2026-09-13).

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
  of the capped exponential delay, with the sum hard-capped at `MaxDelay`"
  (`retry.go`, `computeDelay`). A `Config.JitterFactor` (or
  `Jitter: none | additive | full | equal`) would let callers disable jitter
  for deterministic tests or tune spread. **Deferred (2026-08-08, reaffirmed
  2026-08-22):** `DelayFunc` already covers the custom-delay escape hatch
  (compute pure exponential in the callback for zero jitter); jitter config
  will land with the options-pattern migration (see `ROADMAP.md` v1.0
  section). The v0.4.0 cap fix makes the additive strategy contract-safe —
  the delay can never exceed `MaxDelay` — which removes the correctness
  pressure to switch strategies but does not by itself justify a new
  `Config` field. Tradeoff: another field to validate and freeze.
- **Deterministic RNG option** — `ComputeDelay` uses `math/rand/v2` globally
  (`retry.go` imports); a pluggable `rand` source would make delay sequences
  reproducible in tests without sampling-based assertions (the existing
  `TestBackoff_IncreasesExponentially` works around this by testing the formula,
  not sampled values). **Decided (2026-09-13):** lands as `WithRandomSource`
  with the options-pattern migration — see the `ROADMAP.md` v1.0 section.
- **Deadline-aware attempt budgeting** — currently `MaxAttempts` is the only
  budget; a caller with a hard deadline cannot ask `Do` to stop retrying when
  the remaining context budget is too small for another attempt. Possibly out of
  scope (callers can cancel the context), but worth a decision.
  **Decided (2026-09-13): stay count-based** — deadline budgeting is
  documented as a `MaxDelay`-setting recipe, not new semantics; see the
  `ROADMAP.md` composition-primitives section.
- **Composable neighbors in this package?** — circuit-breaker / bulkhead
  primitives are intentionally NOT here (they belong closer to the caller or in
  `go-cqrs-lite/middleware`). Document the boundary explicitly in `doc.go` if
  asked again, rather than adding code.
