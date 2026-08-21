package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
)

// ErrExhausted is returned by [Do] when all retry attempts have failed.
// It is classified as Infrastructure because retry exhaustion typically
// indicates a downstream system problem.
var ErrExhausted = errorfamily.NewInfrastructure(
	"retry.exhausted",
	"all retry attempts failed",
)

// ErrCanceled is returned by [Do] when the context is canceled during
// a retry delay. Errors matching it also unwrap to [context.Canceled],
// and the last attempt error remains in the chain.
var ErrCanceled = errorfamily.NewInfrastructure(
	"retry.canceled",
	"retry canceled during backoff delay",
)

// ErrDeadlineExceeded is returned by [Do] when the context deadline is
// exceeded during a retry delay. Errors matching it also unwrap to
// [context.DeadlineExceeded], and the last attempt error remains in the
// chain. The distinction from [ErrCanceled] matters operationally: a
// deadline means the operation was too slow, a cancel means the caller
// shut down — the two are debugged differently.
var ErrDeadlineExceeded = errorfamily.NewInfrastructure(
	"retry.deadline",
	"retry deadline exceeded during backoff delay",
)

// AttemptFunc is the function retried by [Do]. The attempt argument
// starts at 1 (the first attempt) and increments with each retry.
type AttemptFunc func(ctx context.Context, attempt int) error

// Do executes fn with retry logic according to config. It calls fn at
// most config.MaxAttempts times, sleeping with exponential backoff + jitter
// between attempts. If fn returns nil, Do returns nil immediately.
//
// If config.IsRetryable returns false for an error, that error is returned
// immediately without further attempts.
//
// If the context is canceled during a backoff delay, Do returns an error
// wrapping [ErrCanceled] that also unwraps to [context.Canceled]. If the
// context deadline is exceeded instead, the returned error wraps
// [ErrDeadlineExceeded] and unwraps to [context.DeadlineExceeded]. In both
// cases the last attempt error stays in the chain.
//
// If all attempts fail, Do calls config.OnExhausted (if set) and returns
// an error wrapping [ErrExhausted] with the last error as its cause.
func Do(ctx context.Context, config Config, fn AttemptFunc) error {
	if err := config.Validate(); err != nil {
		return err
	}

	isRetryable := config.IsRetryable
	if isRetryable == nil {
		isRetryable = errorfamily.IsRetryable
	}

	var err error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err = fn(ctx, attempt)
		if err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attempt == config.MaxAttempts {
			break
		}

		delay := computeDelay(config.InitialDelay, config.MaxDelay, config.Multiplier, attempt)

		if config.DelayFunc != nil {
			if d := config.DelayFunc(attempt, err); d > 0 {
				delay = d
			}
		}

		if config.OnRetry != nil {
			config.OnRetry(attempt, delay, err)
		}

		timer := time.NewTimer(delay)

		select {
		case <-timer.C:
			timer.Stop()
		case <-ctx.Done():
			timer.Stop()

			return contextEnded(ctx, err)
		}
	}

	if config.OnExhausted != nil {
		config.OnExhausted(config.MaxAttempts, err)
	}

	return errorfamily.WrapInfrastructure(ErrExhausted, "retry.exhausted",
		"all attempts failed").WithCause(err)
}

// contextEnded classifies why the context ended — deadline exceeded or
// explicit cancel — and builds the matching terminal error. The cause chain
// carries both the context error and the last attempt error (Go 1.20 multi-%w),
// so errors.Is reaches either one without losing the other.
func contextEnded(ctx context.Context, lastErr error) error {
	chain := fmt.Errorf("%w; last attempt: %w", ctx.Err(), lastErr)

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errorfamily.WrapInfrastructure(ErrDeadlineExceeded, "retry.deadline",
			"retry deadline exceeded during backoff delay").WithCause(chain)
	}

	return errorfamily.WrapInfrastructure(ErrCanceled, "retry.canceled",
		"retry canceled during backoff delay").WithCause(chain)
}

// Backoff calculates the delay before the next attempt using exponential
// backoff with additive jitter. The delay for attempt n is:
//
//	min(InitialDelay * Multiplier^(n-1) + jitter, MaxDelay)
//
// where jitter is a random value of up to 50% of the capped exponential
// delay. The cap applies to the jittered sum, so the returned value never
// exceeds MaxDelay. Exported so callers can preview or log the planned delay
// without executing the retry loop. See [ComputeDelay] for the
// raw-parameter variant.
//
// attempt must be >= 1; passing a lower value returns a Rejection error.
func Backoff(config Config, attempt int) (time.Duration, error) {
	return ComputeDelay(config.InitialDelay, config.MaxDelay, config.Multiplier, attempt)
}

// ComputeDelay calculates the exponential backoff delay with jitter from raw
// parameters, without requiring a [Config]. The delay for attempt n is:
//
//	min(initial * multiplier^(n-1) + jitter, maxDelay)
//
// where jitter is a random value of up to 50% of the capped exponential
// delay. The cap applies to the jittered sum, so the returned value never
// exceeds maxDelay. attempt must be >= 1; passing a lower value returns a
// Rejection error. See [Backoff] for the Config-based variant.
func ComputeDelay(initial, maxDelay time.Duration, multiplier float64, attempt int) (time.Duration, error) {
	if attempt < 1 {
		return 0, errorfamily.NewRejection(
			"retry.invalid_attempt",
			fmt.Sprintf("attempt must be >= 1, got %d", attempt),
		)
	}

	return computeDelay(initial, maxDelay, multiplier, attempt), nil
}

// computeDelay is the trusted internal computation. It is hardened so that no
// combination of inputs can panic: a retry loop sits on the failure path, so a
// panic here converts a recoverable downstream blip into a process crash.
// Callers must guarantee attempt >= 1; no attempt validation is performed.
func computeDelay(initial, maxDelay time.Duration, multiplier float64, attempt int) time.Duration {
	if initial <= 0 {
		return 0
	}
	// B1: treat an unset cap as "no growth beyond initial" so a missing
	// MaxDelay can never collapse the delay to zero.
	if maxDelay <= 0 {
		maxDelay = initial
	}

	scaled := float64(initial) * math.Pow(multiplier, float64(attempt-1))

	// B3: compare in float space before converting, so an out-of-range value
	// saturates to maxDelay instead of wrapping to INT64_MIN.
	var delay time.Duration
	if scaled >= float64(maxDelay) || math.IsInf(scaled, 1) || math.IsNaN(scaled) {
		delay = maxDelay
	} else {
		delay = min(time.Duration(scaled), maxDelay)
	}

	if delay <= 0 {
		return 0
	}

	// B2: a delay under 2ns has no room for jitter (half == 0); return as-is
	// rather than calling Int64N(0), which panics.
	half := int64(delay) / 2
	if half <= 0 {
		return delay
	}

	// Jitter is added to the capped exponential delay and the sum is capped
	// again: capping before jitter would let real sleeps reach 1.5x MaxDelay
	// while the docs promise a hard cap.
	jitter := time.Duration(rand.Int64N(half)) //nolint:gosec // jitter divisor; weak rand fine

	delayed := delay + jitter
	if delayed < delay || delayed > maxDelay {
		// Overflow wrap or cap exceeded: saturate at the cap.
		return maxDelay
	}

	return delayed
}
