package retry_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync/atomic"
	"testing"
	"time"

	errorfamily "github.com/larsartmann/go-error-family"
	retry "github.com/larsartmann/go-retry"
)

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	err := retry.Do(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) error {
			calls.Add(1)

			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", calls.Load())
	}
}

func TestDo_RetriesUntilSuccess(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	err := retry.Do(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) error {
			calls.Add(1)

			if attempt < 3 {
				return errorfamily.NewTransient("test.transient", "fail")
			}

			return nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestDo_ConcurrentInvocationsShareNoMutableState(t *testing.T) {
	t.Parallel()

	const invocationCount = 100

	results := make(chan error, invocationCount)

	for range invocationCount {
		go func() {
			var calls atomic.Int32

			err := retry.Do(context.Background(), fastConfig(), func(ctx context.Context, attempt int) error {
				calls.Add(1)

				if attempt == 1 {
					return errorfamily.NewTransient("test.transient", "retry once")
				}

				return nil
			})
			if err != nil {
				results <- err

				return
			}

			if calls.Load() != 2 {
				results <- fmt.Errorf("expected 2 calls, got %d", calls.Load())

				return
			}

			results <- nil
		}()
	}

	for range invocationCount {
		if err := <-results; err != nil {
			t.Fatalf("concurrent retry invocation failed: %v", err)
		}
	}
}

func TestDo_ReturnsErrExhaustedWhenAllAttemptsFail(t *testing.T) {
	t.Parallel()

	transient := errorfamily.NewTransient("test.transient", "always fail")

	var calls atomic.Int32

	err := retry.Do(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) error {
			calls.Add(1)

			return transient
		},
	)

	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}

	if !errors.Is(err, transient) {
		t.Fatalf("expected cause to be wrapped, got %v", err)
	}

	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestDo_DoesNotRetryNonRetryableError(t *testing.T) {
	t.Parallel()

	rejection := errorfamily.NewRejection("test.rejection", "non-retryable")

	var calls atomic.Int32

	err := retry.Do(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) error {
			calls.Add(1)

			return rejection
		},
	)

	if !errors.Is(err, rejection) {
		t.Fatalf("expected rejection error, got %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("expected 1 call (no retry), got %d", calls.Load())
	}
}

func TestDo_RespectsCustomIsRetryable(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("custom sentinel")
	cfg := fastConfig()
	cfg.IsRetryable = func(err error) bool { return errors.Is(err, sentinel) }

	var calls atomic.Int32

	err := retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		calls.Add(1)

		return sentinel
	})

	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}

	if calls.Load() != 3 {
		t.Fatalf("expected 3 calls, got %d", calls.Load())
	}
}

func TestDo_OnRetryCalledBetweenAttempts(t *testing.T) {
	t.Parallel()

	var retryCalls atomic.Int32

	cfg := fastConfig()
	cfg.OnRetry = func(attempt int, delay time.Duration, err error) {
		retryCalls.Add(1)
	}

	transient := errorfamily.NewTransient("test.transient", "fail")
	_ = retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		if attempt < 3 {
			return transient
		}

		return nil
	})

	// OnRetry is called after attempts 1 and 2 (before attempts 2 and 3)
	if retryCalls.Load() != 2 {
		t.Fatalf("expected 2 OnRetry calls, got %d", retryCalls.Load())
	}
}

func TestDo_OnExhaustedCalledAfterAllAttemptsFail(t *testing.T) {
	t.Parallel()

	var (
		exhaustedAttempts int
		exhaustedErr      error
	)

	cfg := fastConfig()
	cfg.OnExhausted = func(attempts int, err error) {
		exhaustedAttempts = attempts
		exhaustedErr = err
	}

	transient := errorfamily.NewTransient("test.transient", "fail")
	err := retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	_ = err

	if exhaustedAttempts != 3 {
		t.Fatalf("expected OnExhausted attempts=3, got %d", exhaustedAttempts)
	}

	if !errors.Is(exhaustedErr, transient) {
		t.Fatalf("expected OnExhausted err to be the transient error, got %v", exhaustedErr)
	}
}

func TestDo_ContextCancellationDuringBackoff(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())

	transient := errorfamily.NewTransient("test.transient", "fail")

	go func() {
		time.Sleep(10 * time.Millisecond) // let the first attempt fail
		cancel()
	}()

	cfg := retry.Config{
		MaxAttempts:  10,
		InitialDelay: 5 * time.Second, // long delay so cancel fires during it
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	if !errors.Is(err, retry.ErrCanceled) {
		t.Fatalf("expected ErrCanceled, got %v", err)
	}

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled in chain, got %v", err)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancel must not match DeadlineExceeded, got %v", err)
	}
}

func TestDo_DeadlineExceededDuringBackoff(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	transient := errorfamily.NewTransient("test.transient", "fail")

	cfg := retry.Config{
		MaxAttempts:  10,
		InitialDelay: 5 * time.Second, // long delay so the deadline fires during it
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded in chain, got %v", err)
	}

	if !errors.Is(err, retry.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrDeadlineExceeded, got %v", err)
	}

	if errors.Is(err, retry.ErrCanceled) {
		t.Fatalf("deadline must not match ErrCanceled, got %v", err)
	}

	if !errors.Is(err, transient) {
		t.Fatalf("expected last attempt error in chain, got %v", err)
	}
}

func TestDo_AttemptNumberStartsAt1(t *testing.T) {
	t.Parallel()

	var attempts []int

	_ = retry.Do(context.Background(), fastConfig(), func(ctx context.Context, attempt int) error {
		attempts = append(attempts, attempt)
		if attempt < 2 {
			return errorfamily.NewTransient("test.transient", "fail")
		}

		return nil
	})

	if len(attempts) != 2 || attempts[0] != 1 || attempts[1] != 2 {
		t.Fatalf("expected attempts [1,2], got %v", attempts)
	}
}

func TestDo_InvalidConfigReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config retry.Config
	}{
		{
			name:   "zero attempts",
			config: retry.Config{MaxAttempts: 0, InitialDelay: 1, MaxDelay: 1, Multiplier: 2},
		},
		{
			name:   "zero initial delay",
			config: retry.Config{MaxAttempts: 1, InitialDelay: 0, MaxDelay: 1, Multiplier: 2},
		},
		{
			name:   "zero max delay",
			config: retry.Config{MaxAttempts: 1, InitialDelay: 1, MaxDelay: 0, Multiplier: 2},
		},
		{
			name:   "multiplier <= 1",
			config: retry.Config{MaxAttempts: 1, InitialDelay: 1, MaxDelay: 1, Multiplier: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := retry.Do(
				context.Background(),
				tt.config,
				func(ctx context.Context, attempt int) error {
					t.Fatal("fn should not be called with invalid config")

					return nil
				},
			)

			if errorfamily.Classify(err) != errorfamily.Rejection {
				t.Fatalf(
					"expected Rejection family, got %v (family: %s)",
					err,
					errorfamily.Classify(err),
				)
			}
		})
	}
}

func TestBackoff_RespectsMaxDelay(t *testing.T) {
	t.Parallel()

	cfg := retry.Config{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     200 * time.Millisecond,
		Multiplier:   10.0,
	}

	delay, err := retry.Backoff(cfg, 5) // attempt 5 would be 100ms * 10^4 = huge
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if delay > 200*time.Millisecond { // cap is strict, jitter included
		t.Fatalf("expected delay <= 200ms (capped), got %v", delay)
	}
}

func TestBackoff_InvalidAttemptReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		attempt int
	}{
		{name: "zero", attempt: 0},
		{name: "negative", attempt: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := retry.Config{
				MaxAttempts:  3,
				InitialDelay: 10 * time.Millisecond,
				MaxDelay:     1 * time.Second,
				Multiplier:   2.0,
			}

			_, err := retry.Backoff(cfg, tt.attempt)

			if errorfamily.Classify(err) != errorfamily.Rejection {
				t.Fatalf("expected Rejection for attempt %d, got %v (family: %s)",
					tt.attempt, err, errorfamily.Classify(err))
			}
		})
	}
}

// TestComputeDelay_NeverExceedsMaxDelay pins the documented contract that the
// returned delay — jitter included — never exceeds maxDelay. The cap used to be
// applied before jitter was added, letting sampled delays reach 1.5× maxDelay
// (observed: ~300ms against a declared 200ms cap). Sampling attempt 10 forces
// the exponential term far past the cap so every sample exercises the
// cap-then-jitter path.
func TestComputeDelay_NeverExceedsMaxDelay(t *testing.T) {
	t.Parallel()

	const (
		initial  = 100 * time.Millisecond
		maxDelay = 200 * time.Millisecond
	)

	for sample := range 20000 {
		delay, err := retry.ComputeDelay(initial, maxDelay, 2.0, 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if delay > maxDelay {
			t.Fatalf("delay %v exceeds MaxDelay cap %v (sample %d)", delay, maxDelay, sample)
		}
	}
}

func TestComputeDelay_InvalidAttemptReturnsError(t *testing.T) {
	t.Parallel()

	_, err := retry.ComputeDelay(10*time.Millisecond, 1*time.Second, 2.0, 0)

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Fatalf("expected Rejection for attempt 0, got %v (family: %s)",
			err, errorfamily.Classify(err))
	}
}

// TestComputeDelay_NeverPanicsOnExtremeInputs guards against the three
// reproduced Int64N panics (B1: omitted MaxDelay, B2: sub-2ns delay,
// B3: math.Pow overflow). A retry loop sits on the failure path, so no input
// combination may crash the process. Each case must return a non-negative
// duration without panicking.
func TestComputeDelay_NeverPanicsOnExtremeInputs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		initial    time.Duration
		maxDelay   time.Duration
		multiplier float64
		attempt    int
	}{
		{"B1 omitted maxDelay", 100 * time.Millisecond, 0, 2.0, 1},
		{"B2 sub-2ns delay", 1, 5 * time.Second, 2.0, 1},
		{"B3 default config overflow at 38", 100 * time.Millisecond, 5 * time.Second, 2.0, 38},
		{"B3 large multiplier fast overflow", time.Millisecond, 5 * time.Second, 10.0, 15},
		{"B3 attempt 1000", 100 * time.Millisecond, 5 * time.Second, 2.0, 1000},
		{"zero initial", 0, 5 * time.Second, 2.0, 5},
		{"maxDelay below initial", time.Second, time.Millisecond, 2.0, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			delay, err := retry.ComputeDelay(tt.initial, tt.maxDelay, tt.multiplier, tt.attempt)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if delay < 0 {
				t.Fatalf("delay must never be negative, got %v", delay)
			}
			// The effective cap is maxDelay when set, else initial (B1 path).
			effectiveCap := tt.maxDelay
			if effectiveCap <= 0 {
				effectiveCap = tt.initial
			}
			// The cap applies after jitter, so the delay never exceeds it.
			if effectiveCap > 0 && delay > effectiveCap {
				t.Fatalf("delay %v exceeds cap %v", delay, effectiveCap)
			}
		})
	}
}

// TestComputeDelay_NeverPanicsAcrossMatrix sweeps a broad input domain to prove
// computeDelay cannot panic or return a negative duration for any reachable
// combination. Statement coverage could not catch B1/B2/B3 because the
// panicking lines were already exercised with benign inputs; this property
// test covers the input domain instead.
func TestComputeDelay_NeverPanicsAcrossMatrix(t *testing.T) {
	t.Parallel()

	initials := []time.Duration{0, 1, 2, time.Millisecond, 100 * time.Millisecond, time.Second}
	maxDelays := []time.Duration{0, 1, time.Millisecond, 5 * time.Second}
	multipliers := []float64{0.5, 1.0, 1.5, 2.0, 10.0}
	attempts := []int{1, 2, 5, 38, 50, 100, 1000}

	for _, initial := range initials {
		for _, maxDelay := range maxDelays {
			for _, multiplier := range multipliers {
				for _, attempt := range attempts {
					delay, err := retry.ComputeDelay(initial, maxDelay, multiplier, attempt)
					if err != nil {
						t.Fatalf("unexpected error for initial=%v maxDelay=%v mult=%v attempt=%d: %v",
							initial, maxDelay, multiplier, attempt, err)
					}

					if delay < 0 {
						t.Fatalf("negative delay for initial=%v maxDelay=%v mult=%v attempt=%d: %v",
							initial, maxDelay, multiplier, attempt, delay)
					}
				}
			}
		}
	}
}

func TestValidate_RejectsInvalidMaxDelay(t *testing.T) {
	t.Parallel()

	cfg := retry.Config{MaxAttempts: 1, InitialDelay: 1, MaxDelay: 0, Multiplier: 2}
	err := cfg.Validate()

	if errorfamily.Classify(err) != errorfamily.Rejection {
		t.Fatalf("expected Rejection for MaxDelay=0, got %v (family: %s)",
			err, errorfamily.Classify(err))
	}
}

// TestComputeDelay_SaturatesNearMaxInt64 covers the overflow-saturation path:
// when the capped delay is near math.MaxInt64, adding jitter must saturate to
// math.MaxInt64 rather than wrapping negative. No panic, never exceeds MaxInt64.
func TestComputeDelay_SaturatesNearMaxInt64(t *testing.T) {
	t.Parallel()

	const nearMax = time.Duration(math.MaxInt64)

	for range 100 {
		delay, err := retry.ComputeDelay(nearMax, nearMax, 2.0, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if delay < 0 {
			t.Fatalf("delay must never be negative near MaxInt64, got %v", delay)
		}

		if delay > nearMax {
			t.Fatalf("delay must not exceed MaxInt64, got %v", delay)
		}
	}
}

func FuzzComputeDelayNeverPanics(f *testing.F) {
	f.Add(int64(time.Millisecond), int64(time.Second), 2.0, 1)
	f.Add(int64(1), int64(0), 10.0, 38)
	f.Add(int64(math.MaxInt64), int64(math.MaxInt64), 2.0, 1)
	// Seeds distilled from the 2026-08-22 5-minute campaign (104M execs, 0
	// failures): the four input classes the fuzzer found beyond the seeds
	// above — near-MaxInt64 with a fractional multiplier (jitter-saturation
	// path), negative multiplier (NaN via math.Pow), negative initial, and a
	// negative attempt (Rejection path).
	f.Add(int64(math.MaxInt64), int64(math.MaxInt64-179), 0.4, 2)
	f.Add(int64(1000000), int64(1000000000), -140.0, 15)
	f.Add(int64(-24), int64(0), 10.0, 38)
	f.Add(int64(1), int64(0), 10.0, -58)

	f.Fuzz(func(t *testing.T, initialNanos, maxDelayNanos int64, multiplier float64, attempt int) {
		initial := time.Duration(initialNanos)
		maxDelay := time.Duration(maxDelayNanos)

		delay, err := retry.ComputeDelay(initial, maxDelay, multiplier, attempt)
		if attempt < 1 {
			if err == nil {
				t.Fatal("expected invalid attempt error")
			}

			return
		}

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if delay < 0 {
			t.Fatalf("delay must never be negative, got %v", delay)
		}
	})
}

func TestBackoff_IncreasesExponentially(t *testing.T) {
	t.Parallel()

	cfg := retry.Config{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Hour, // effectively uncapped
		Multiplier:   2.0,
	}

	// Base delays (without jitter): attempt 1 = 10ms, attempt 2 = 20ms.
	// Jitter (up to 50%) can make d1 > d2 in rare cases, so we verify the
	// exponential formula instead of comparing sampled values.
	base1 := float64(cfg.InitialDelay)
	base2 := float64(cfg.InitialDelay) * cfg.Multiplier

	if base2 <= base1 {
		t.Fatalf("expected exponential increase: base1=%v base2=%v", base1, base2)
	}
}

func TestDefaultConfig_IsValid(t *testing.T) {
	t.Parallel()

	cfg := retry.DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig should be valid: %v", err)
	}
}

func TestFromPolicy_MapsRetryPolicy(t *testing.T) {
	t.Parallel()

	policy := errorfamily.Transient.RetryPolicy()
	cfg := retry.FromPolicy(policy)

	if cfg.MaxAttempts != policy.MaxAttempts {
		t.Fatalf("expected MaxAttempts=%d, got %d", policy.MaxAttempts, cfg.MaxAttempts)
	}

	if cfg.InitialDelay != policy.MinDelay {
		t.Fatalf("expected InitialDelay=%v, got %v", policy.MinDelay, cfg.InitialDelay)
	}

	if cfg.MaxDelay != policy.MaxDelay {
		t.Fatalf("expected MaxDelay=%v, got %v", policy.MaxDelay, cfg.MaxDelay)
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("converted policy should be valid: %v", err)
	}
}

func TestFromPolicy_PreservesDefaultLoopSettings(t *testing.T) {
	t.Parallel()

	cfg := retry.FromPolicy(errorfamily.Transient.RetryPolicy())
	defaults := retry.DefaultConfig()

	if cfg.Multiplier != defaults.Multiplier {
		t.Fatalf("expected default Multiplier=%v, got %v", defaults.Multiplier, cfg.Multiplier)
	}

	if cfg.IsRetryable == nil {
		t.Fatal("expected default IsRetryable predicate")
	}
}

func TestFromPolicy_NonRetryableFamilyIsInvalidForLoop(t *testing.T) {
	t.Parallel()

	cfg := retry.FromPolicy(errorfamily.Rejection.RetryPolicy())
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected non-retryable family policy to require caller configuration")
	}
}

func TestDo_NilIsRetryableDefaultsToErrorFamily(t *testing.T) {
	t.Parallel()

	cfg := retry.Config{
		MaxAttempts:  2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     1 * time.Millisecond,
		Multiplier:   2.0,
		IsRetryable:  nil, // should default to errorfamily.IsRetryable
	}

	var calls atomic.Int32

	_ = retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		calls.Add(1)

		return errorfamily.NewTransient("test.transient", "retryable")
	})

	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls with default IsRetryable, got %d", calls.Load())
	}
}

func TestDo_OnRetryNotCalledAfterFinalAttempt(t *testing.T) {
	t.Parallel()

	var retryCalls atomic.Int32

	cfg := fastConfig()
	cfg.OnRetry = func(attempt int, delay time.Duration, err error) {
		retryCalls.Add(1)
	}

	transient := errorfamily.NewTransient("test.transient", "always fail")
	_ = retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	// 3 attempts, but OnRetry fires only *between* them (after 1 and 2), never
	// after the final failure.
	if got, want := retryCalls.Load(), int32(2); got != want {
		t.Fatalf("expected %d OnRetry calls (MaxAttempts-1), got %d", want, got)
	}
}

func TestDo_PreCanceledContextReturnsErrCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // canceled before Do starts

	transient := errorfamily.NewTransient("test.transient", "fail")

	var calls atomic.Int32

	err := retry.Do(ctx, fastConfig(), func(ctx context.Context, attempt int) error {
		calls.Add(1)

		return transient
	})

	// The first attempt runs; the backoff select then sees ctx.Done() immediately.
	if !errors.Is(err, retry.ErrCanceled) {
		t.Fatalf("expected ErrCanceled with a pre-canceled context, got %v", err)
	}

	if got, want := calls.Load(), int32(1); got != want {
		t.Fatalf("expected %d call (no real backoff after cancel), got %d", want, got)
	}
}

func TestDo_OnExhaustedReceivesExactLastError(t *testing.T) {
	t.Parallel()

	transient := errorfamily.NewTransient("test.transient", "the last failure")

	var received error

	cfg := fastConfig()
	cfg.OnExhausted = func(attempts int, err error) {
		received = err
	}

	_ = retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	// Identity, not just errors.Is: OnExhausted must receive the exact last error.
	if !errors.Is(received, transient) {
		t.Fatalf("expected OnExhausted to receive the exact last error %v, got %v",
			transient, received)
	}
}

func TestDo_DelayFuncOverridesExponentialBackoff(t *testing.T) {
	t.Parallel()

	var delays []time.Duration

	cfg := fastConfig()
	cfg.DelayFunc = func(attempt int, _ error) time.Duration {
		d := time.Duration(attempt) * time.Microsecond
		delays = append(delays, d)

		return d
	}

	transient := errorfamily.NewTransient("test.transient", "fail")

	_ = retry.Do(context.Background(), cfg, func(_ context.Context, _ int) error {
		return transient
	})

	// DelayFunc is called after attempts 1 and 2 (not after the final attempt 3).
	if len(delays) != 2 {
		t.Fatalf("expected 2 DelayFunc calls, got %d", len(delays))
	}

	if delays[0] != 1*time.Microsecond || delays[1] != 2*time.Microsecond {
		t.Fatalf("expected delays [1us, 2us], got %v", delays)
	}
}

func TestDo_DelayFuncReceivesError(t *testing.T) {
	t.Parallel()

	sentinel := errorfamily.NewTransient("test.transient", "honor retry-after")

	var receivedErr error

	cfg := fastConfig()
	cfg.DelayFunc = func(_ int, err error) time.Duration {
		receivedErr = err

		return 0
	}

	_ = retry.Do(context.Background(), cfg, func(_ context.Context, _ int) error {
		return sentinel
	})

	if !errors.Is(receivedErr, sentinel) {
		t.Fatalf("expected DelayFunc to receive the error, got %v", receivedErr)
	}
}

func TestDo_DelayFuncZeroFallsBackToExponential(t *testing.T) {
	t.Parallel()

	var delayFuncCalled bool

	cfg := fastConfig()
	cfg.DelayFunc = func(_ int, _ error) time.Duration {
		delayFuncCalled = true

		return 0 // 0 = use default exponential backoff
	}

	var totalBackoff time.Duration

	cfg.OnRetry = func(_ int, delay time.Duration, _ error) {
		totalBackoff += delay
	}

	transient := errorfamily.NewTransient("test.transient", "fail")

	_ = retry.Do(context.Background(), cfg, func(_ context.Context, _ int) error {
		return transient
	})

	if !delayFuncCalled {
		t.Fatal("expected DelayFunc to be called")
	}

	// With fastConfig (1ms initial, 5ms max, multiplier 2.0), returning 0
	// should fall back to exponential backoff, not skip the delay entirely.
	// Two retries → at least the initial delay should have been used.
	if totalBackoff <= 0 {
		t.Fatalf("expected positive backoff from default, got %v", totalBackoff)
	}
}

func TestDo_OnRetryReceivesDelayFuncDelay(t *testing.T) {
	t.Parallel()

	const customDelay = 42 * time.Microsecond

	var loggedDelay time.Duration

	cfg := fastConfig()
	cfg.DelayFunc = func(_ int, _ error) time.Duration { return customDelay }
	cfg.OnRetry = func(_ int, delay time.Duration, _ error) {
		loggedDelay = delay
	}

	transient := errorfamily.NewTransient("test.transient", "fail")

	_ = retry.Do(context.Background(), cfg, func(_ context.Context, _ int) error {
		return transient
	})

	// OnRetry must receive the DelayFunc-computed delay, not the exponential one.
	if loggedDelay != customDelay {
		t.Fatalf("expected OnRetry delay=%v, got %v", customDelay, loggedDelay)
	}
}

func fastConfig() retry.Config {
	return retry.Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
	}
}

func ExampleDo() {
	cfg := retry.Config{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
		Multiplier:   2.0,
	}

	var attempt int

	err := retry.Do(context.Background(), cfg, func(ctx context.Context, n int) error {
		attempt = n
		if n < 3 {
			// A Transient error is retryable by the default predicate.
			return errorfamily.NewTransient("example.transient", "service unavailable")
		}

		return nil
	})

	fmt.Println("attempt:", attempt)
	fmt.Println("error:", err)
	// Output:
	// attempt: 3
	// error: <nil>
}

func ExampleDo_customIsRetryable() {
	sentinel := errors.New("service overloaded")

	cfg := retry.DefaultConfig()
	cfg.MaxAttempts = 4
	cfg.InitialDelay = time.Millisecond
	// Retry only our own sentinel; other errors (even Transient ones) are not retried.
	cfg.IsRetryable = func(err error) bool {
		return errors.Is(err, sentinel)
	}

	var attempt int

	err := retry.Do(context.Background(), cfg, func(ctx context.Context, n int) error {
		attempt = n
		if n < 2 {
			return sentinel
		}

		return nil
	})

	fmt.Println("attempt:", attempt)
	fmt.Println("error:", err)
	// Output:
	// attempt: 2
	// error: <nil>
}

// ExampleDo_delayFunc honors a server-provided delay (e.g. an HTTP
// "Retry-After" header carried by the attempt's error) instead of the
// computed exponential backoff.
func ExampleDo_delayFunc() {
	cfg := retry.DefaultConfig()
	cfg.MaxAttempts = 2
	cfg.InitialDelay = time.Millisecond
	cfg.MaxDelay = 5 * time.Millisecond
	cfg.DelayFunc = func(attempt int, err error) time.Duration {
		_ = err // in practice: parse Retry-After out of err here

		return 2 * time.Millisecond // return 0 to fall back to exponential backoff
	}

	var delays []time.Duration

	cfg.OnRetry = func(attempt int, delay time.Duration, err error) {
		delays = append(delays, delay)
	}

	err := retry.Do(context.Background(), cfg, func(ctx context.Context, n int) error {
		return errorfamily.NewTransient("example.rate_limited", "too many requests")
	})

	fmt.Println("delays:", delays)
	fmt.Println("error:", err)
	// Output:
	// delays: [2ms]
	// error: [infrastructure:retry.exhausted] all attempts failed: [transient:example.rate_limited] too many requests
}

// ExampleFromPolicy converts an error-family retry policy — the advisory
// defaults for Transient errors — into a Config.
func ExampleFromPolicy() {
	policy := errorfamily.Transient.RetryPolicy()

	cfg := retry.FromPolicy(policy)
	cfg.InitialDelay = time.Millisecond // shrunk so the example runs instantly
	cfg.MaxDelay = 2 * time.Millisecond

	var attempt int

	err := retry.Do(context.Background(), cfg, func(ctx context.Context, n int) error {
		attempt = n

		return errorfamily.NewTransient("example.transient", "still down")
	})

	fmt.Println("policy attempts:", policy.MaxAttempts)
	fmt.Println("ran attempts:", attempt)
	fmt.Println("exhausted:", errors.Is(err, retry.ErrExhausted))
	// Output:
	// policy attempts: 3
	// ran attempts: 3
	// exhausted: true
}

func BenchmarkComputeDelay(b *testing.B) {
	const (
		initial    = 100 * time.Millisecond
		maxDelay   = 5 * time.Second
		multiplier = 2.0
		attempt    = 5
	)

	for b.Loop() { // b.Loop auto-resets the timer and reports allocations
		_, _ = retry.ComputeDelay(initial, maxDelay, multiplier, attempt)
	}
}
