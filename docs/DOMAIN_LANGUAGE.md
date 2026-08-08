# Domain Language

The ubiquitous vocabulary of `go-retry`. Terms are used consistently across
code, doc comments, errors, and docs; use them verbatim rather than inventing
synonyms. Code is the source of truth — this glossary mirrors it.

## Retry vocabulary

- **Attempt** — one execution of the caller's `AttemptFunc(ctx, attempt)`. The
  counter **starts at 1** (the first call), not 0 (`retry.go` `AttemptFunc`,
  line 30; `Do`, line 56).
- **`MaxAttempts`** — the **total** number of attempts, _including the first
  call_. `MaxAttempts: 3` means 1 call + 2 retries = 3 invocations, not 3
  retries on top of the first call (`config.go:20-22`).
- **Backoff** — the delay inserted _before_ a retry. Computed as
  `InitialDelay * Multiplier^(attempt-1)`, capped at `MaxDelay`
  (`retry.go` `Backoff`, line 114; `ComputeDelay`, line 125).
- **`DelayFunc`** — an optional callback that overrides the backoff delay for a
  single attempt. Receives the attempt number and the error from the failed
  attempt, so callers can honor server-provided delays (e.g. HTTP `Retry-After`)
  or implement custom strategies. A return `> 0` overrides; `0` means "use the
  default exponential backoff." (`config.go` `DelayFunc`, line 48).
- **Jitter** — random noise added _on top of_ the capped backoff delay, up to
  50% of it. **Additive, not symmetric**: the actual wait lands in
  `[delay, delay * 1.5]`, never below the base delay (`retry.go` `computeDelay`,
  lines 167-177).
- **Exhaustion** — the state where all `MaxAttempts` have been spent without a
  success. `Do` then fires `OnExhausted` and returns an error wrapping
  `ErrExhausted` (`retry.go` `Do`, lines 96-101).
- **Cancellation** — the caller's `context.Context` is canceled _during_ a
  backoff delay. `Do` returns an error wrapping `ErrCanceled` (with the last
  operation error as its cause), distinct from a bare `context.Canceled`
  (`retry.go` `Do`, lines 82-93).
- **Retryable** — an error qualifies for another attempt. Decided by
  `Config.IsRetryable`; when `nil`, `errorfamily.IsRetryable` is used
  (`retry.go` `Do`, lines 49-52). Non-retryable errors short-circuit out of `Do`
  immediately.

## The `error-family` vocabulary (load-bearing)

`go-retry` borrows its entire error model from
[`go-error-family`](https://github.com/larsartmann/go-error-family) (`v0.10.0`).
Every error carries a behavioral **Family** and a stable string **code**.

- **Family** — an error's behavioral profile, used here to answer "should I
  retry?" and "whose fault is it?". Defined as `errorfamily.Family` (an `int`
  enum). `go-retry` uses three of the six families:

  | Family             | Meaning in `go-retry`                             | Retryable? | Used for                                  |
  | ------------------ | ------------------------------------------------- | ---------- | ----------------------------------------- |
  | **Transient**      | Temporary failure; system's fault.                | **Yes**    | The default retryable error (tests, ops). |
  | **Rejection**      | Bad caller input; user's fault. No state changed. | No         | `Config.Validate()` failures.             |
  | **Infrastructure** | The system cannot serve / downstream problem.     | No         | `ErrExhausted`, `ErrCanceled` outcomes.   |

- **`IsRetryable(err) bool`** — the default retry predicate. Returns `true` iff
  `Classify(err) == Transient` (`classify.go:67`). Substituted by `Do` when
  `Config.IsRetryable` is `nil`.
- **`Classify(err) Family`** — returns an error's family, defaulting to
  `Transient` for unknown errors (fail-open for retry) and `Rejection` for
  `nil` (`classify.go:61`). Used in tests to assert the family of validation
  errors.
- **`NewTransient / NewRejection / NewInfrastructure(code, msg)`** — construct
  a fresh error of that family (`constructors.go`). `go-retry`'s sentinels are
  built with `NewInfrastructure`; validation errors with `NewRejection`.
- **`WrapInfrastructure(sentinel, code, msg).WithCause(err)`** — chains the
  last operation error behind a sentinel so `errors.Is(err, sentinel)` **and**
  `errors.Is(err, lastFnErr)` both hold. This is how exhaustion/cancellation
  preserve the original cause (`retry.go` `Do`, lines 89-91, 100-101).

## Error code convention

Every error carries a machine-readable string code. `go-retry`'s codes follow
`retry.<snake_case_event>`:

| Code                          | Family         | Source                          |
| ----------------------------- | -------------- | ------------------------------- |
| `retry.exhausted`             | Infrastructure | `retry.go:17` (`ErrExhausted`)  |
| `retry.canceled`              | Infrastructure | `retry.go:24` (`ErrCanceled`)   |
| `retry.invalid_max_attempts`  | Rejection      | `config.go:88` (`Validate`)     |
| `retry.invalid_initial_delay` | Rejection      | `config.go:95` (`Validate`)     |
| `retry.invalid_max_delay`     | Rejection      | `config.go:102` (`Validate`)    |
| `retry.invalid_multiplier`    | Rejection      | `config.go:109` (`Validate`)    |
| `retry.invalid_attempt`       | Rejection      | `retry.go:128` (`ComputeDelay`) |

These codes are part of the public contract — callers may switch on them. New
codes must follow the `retry.<event>` pattern.

## Out of scope (deliberately not in this vocabulary)

These terms do **not** belong to `go-retry`'s domain. They live in
`go-cqrs-lite/middleware/v4` and are intentionally absent here (`doc.go`):
**Command**, **Event**, **StreamID**, **MessageAdapter**, **dead-letter entry**,
**OpenTelemetry span**. If a concept needs any of these, it does not belong in
this package.
