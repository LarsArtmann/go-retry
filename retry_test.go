package retry_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"sync"
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

	if err != rejection { //nolint:errorlint // identity check is deliberate: the typed error must never be re-wrapped
		t.Fatalf("expected the typed rejection returned by identity, got %v", err)
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

func TestDo_NestedRetriesAreFailClosed(t *testing.T) {
	t.Parallel()

	var innerCalls atomic.Int32

	var outerCalls atomic.Int32

	err := retry.Do(context.Background(), fastConfig(), func(ctx context.Context, attempt int) error {
		outerCalls.Add(1)

		return retry.Do(ctx, fastConfig(), func(ctx context.Context, innerAttempt int) error {
			innerCalls.Add(1)

			return errorfamily.NewTransient("inner.always.fails", "inner always fails")
		})
	})

	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}

	if outerCalls.Load() != 1 {
		t.Fatalf("outer loop must make exactly 1 attempt on inner exhaustion (fail-closed), got %d", outerCalls.Load())
	}

	if innerCalls.Load() != 3 {
		t.Fatalf("inner loop must make its own 3 attempts, got %d", innerCalls.Load())
	}
}

func TestDo_NestedRetriesAmplifyWhenOverridden(t *testing.T) {
	t.Parallel()

	var outerCalls atomic.Int32

	var innerCalls atomic.Int32

	cfg := fastConfig()
	// Deliberate override: retry everything except caller-input rejections —
	// including the Infrastructure-family exhaustion the inner loop returns,
	// which the default predicate treats as terminal.
	cfg.IsRetryable = func(err error) bool {
		return errorfamily.Classify(err) != errorfamily.Rejection
	}

	err := retry.Do(context.Background(), cfg, func(ctx context.Context, attempt int) error {
		outerCalls.Add(1)

		return retry.Do(ctx, fastConfig(), func(ctx context.Context, innerAttempt int) error {
			innerCalls.Add(1)

			return errorfamily.NewTransient("inner.always.fails", "inner always fails")
		})
	})

	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}

	if outerCalls.Load() != 3 {
		t.Fatalf("overridden predicate must re-enable amplification: expected 3 outer attempts, got %d",
			outerCalls.Load())
	}

	if innerCalls.Load() != 9 {
		t.Fatalf("each outer attempt must run a full inner loop: expected 9 inner attempts, got %d", innerCalls.Load())
	}
}

func TestDo_OnExhaustedNotCalledOnCancel(t *testing.T) {
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

	var exhaustedCalls atomic.Int32

	cfg.OnExhausted = func(attempts int, err error) {
		exhaustedCalls.Add(1)
	}

	err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	if !errors.Is(err, retry.ErrCanceled) {
		t.Fatalf("expected ErrCanceled, got %v", err)
	}

	if exhaustedCalls.Load() != 0 {
		t.Fatalf("OnExhausted must not fire when cancellation ends the loop, got %d calls", exhaustedCalls.Load())
	}
}

func TestDo_OnExhaustedNotCalledOnDeadline(t *testing.T) {
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

	var exhaustedCalls atomic.Int32

	cfg.OnExhausted = func(attempts int, err error) {
		exhaustedCalls.Add(1)
	}

	err := retry.Do(ctx, cfg, func(ctx context.Context, attempt int) error {
		return transient
	})

	if !errors.Is(err, retry.ErrDeadlineExceeded) {
		t.Fatalf("expected ErrDeadlineExceeded, got %v", err)
	}

	if exhaustedCalls.Load() != 0 {
		t.Fatalf("OnExhausted must not fire when a deadline ends the loop, got %d calls", exhaustedCalls.Load())
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

// seedConstExpressions maps the non-literal constant expressions used in the
// f.Add calls of FuzzComputeDelayNeverPanics to their values, so the sync
// test below can normalize seeds and corpus files into comparable forms.
// A new seed that introduces a new expression fails the sync test with an
// "unmapped expression" message naming it, so extend this table and add the
// corpus file in the same change.
var seedConstExpressions = map[string]int64{
	"time.Millisecond":  int64(time.Millisecond),
	"time.Second":       int64(time.Second),
	"math.MaxInt64":     math.MaxInt64,
	"math.MaxInt64-179": math.MaxInt64 - 179,
}

func TestFuzzCorpusMirrorsSeeds(t *testing.T) {
	t.Parallel()

	seeds := fuzzSeedsFromSource(t)
	if len(seeds) == 0 {
		t.Fatal("found no f.Add seeds in retry_test.go: parser broke or seeds were removed")
	}

	corpusDir := filepath.Join("testdata", "fuzz", "FuzzComputeDelayNeverPanics")

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		t.Fatalf("read corpus dir %s: %v", corpusDir, err)
	}

	corpus := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		values := fuzzCorpusValues(t, filepath.Join(corpusDir, entry.Name()))
		corpus = append(corpus, values)
	}

	sortedSeeds := slices.Sorted(slices.Values(seeds))
	sortedCorpus := slices.Sorted(slices.Values(corpus))

	for _, seed := range sortedSeeds {
		if _, found := slices.BinarySearch(sortedCorpus, seed); !found {
			t.Errorf("seed %v has no corpus file: create one under %s with the same values", seed, corpusDir)
		}
	}

	for _, file := range sortedCorpus {
		if _, found := slices.BinarySearch(sortedSeeds, file); !found {
			t.Errorf("corpus entry %v matches no f.Add seed: distill it into a seed or remove the file", file)
		}
	}
}

// fuzzSeedsFromSource extracts the f.Add argument lists from the fuzz
// function in this package's test file and normalizes them into the same
// canonical form the corpus files use, so both sides compare equal.
func fuzzSeedsFromSource(t *testing.T) []string {
	t.Helper()

	source, err := os.ReadFile("retry_test.go")
	if err != nil {
		t.Fatalf("read retry_test.go: %v", err)
	}

	fAdd := regexp.MustCompile(`f\.Add\((.*)\)`)

	var seeds []string

	for line := range strings.SplitSeq(string(source), "\n") {
		match := fAdd.FindStringSubmatch(line)
		if match == nil {
			continue
		}

		parts := strings.Split(match[1], ", ")

		canonical := make([]string, 0, len(parts))
		for _, part := range parts {
			canonical = append(canonical, canonicalFuzzValue(t, part))
		}

		seeds = append(seeds, strings.Join(canonical, ", "))
	}

	return seeds
}

// fuzzCorpusValues reads one `go test fuzz v1` corpus file and returns its
// normalized value tuple.
func fuzzCorpusValues(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 || lines[0] != "go test fuzz v1" {
		t.Fatalf("%s is not a `go test fuzz v1` corpus file", path)
	}

	values := make([]string, 0, len(lines)-1)
	for _, line := range lines[1:] {
		values = append(values, canonicalFuzzValue(t, line))
	}

	return strings.Join(values, ", ")
}

// canonicalFuzzValue normalizes one typed value (`int64(1)`, `float64(2)`,
// or a constant expression like `int64(time.Millisecond)`) into the minimal
// typed form the corpus files are written in. Untyped literals in the seeds
// (`2.0`, `38`) normalize to the type the fuzz engine infers for them.
func canonicalFuzzValue(t *testing.T, raw string) string {
	t.Helper()

	typed := regexp.MustCompile(`^(int64|int|float64)\((.+)\)$`)

	match := typed.FindStringSubmatch(strings.TrimSpace(raw))
	if match == nil {
		expr := strings.TrimSpace(raw)
		if strings.Contains(expr, ".") {
			return canonicalFloat(t, expr)
		}

		return canonicalInt(t, expr)
	}

	typ, expr := match[1], strings.TrimSpace(match[2])

	if value, known := seedConstExpressions[expr]; known {
		return fmt.Sprintf("%s(%s)", typ, strconv.FormatInt(value, 10))
	}

	switch typ {
	case "float64":
		return canonicalFloat(t, expr)
	case "int":
		return canonicalInt(t, expr)
	default:
		return canonicalInt64(t, expr)
	}
}

func canonicalFloat(t *testing.T, expr string) string {
	t.Helper()

	if value, known := seedConstExpressions[expr]; known {
		return fmt.Sprintf("float64(%s)", strconv.FormatFloat(float64(value), 'g', -1, 64))
	}

	value, err := strconv.ParseFloat(expr, 64)
	if err != nil {
		t.Fatalf("unmapped float expression %q in a fuzz seed: extend seedConstExpressions", expr)
	}

	return fmt.Sprintf("float64(%s)", strconv.FormatFloat(value, 'g', -1, 64))
}

func canonicalInt(t *testing.T, expr string) string {
	t.Helper()

	value, err := strconv.Atoi(expr)
	if err != nil {
		t.Fatalf("unmapped int expression %q in a fuzz seed: extend seedConstExpressions", expr)
	}

	return fmt.Sprintf("int(%s)", strconv.Itoa(value))
}

func canonicalInt64(t *testing.T, expr string) string {
	t.Helper()

	value, err := strconv.ParseInt(expr, 10, 64)
	if err != nil {
		t.Fatalf("unmapped int64 expression %q in a fuzz seed: extend seedConstExpressions", expr)
	}

	return fmt.Sprintf("int64(%s)", strconv.FormatInt(value, 10))
}

func TestBackoff_IncreasesExponentially(t *testing.T) {
	t.Parallel()

	cfg := retry.Config{
		MaxAttempts:  5,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     1 * time.Hour, // effectively uncapped
		Multiplier:   2.0,
	}

	// JitterNone makes delays deterministic, so the exponential formula is
	// verified against real computed delays through the public API (no more
	// re-deriving the formula inside the test).
	var delays []time.Duration

	retry.Do(context.Background(), cfg, func(_ context.Context, _ int) error {
		return errorfamily.NewTransient("test.transient", "fail")
	}, retry.WithJitter(retry.JitterNone), retry.WithOnRetry(func(_ int, d time.Duration, _ error) {
		delays = append(delays, d)
	}))

	// Four sleeps between five attempts: 10ms, 20ms, 40ms, 80ms exactly.
	want := []time.Duration{
		10 * time.Millisecond,
		20 * time.Millisecond,
		40 * time.Millisecond,
		80 * time.Millisecond,
	}
	if len(delays) != len(want) {
		t.Fatalf("got %d delays (%v), want %v", len(delays), delays, want)
	}

	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delay[%d] = %s, want exactly %s", i, delays[i], want[i])
		}
	}
}

func TestWithRandomSource_DeterministicSequence(t *testing.T) {
	t.Parallel()

	run := func() []time.Duration {
		var delays []time.Duration

		retry.Do(context.Background(), fastConfig(), func(_ context.Context, _ int) error {
			return errorfamily.NewTransient("test.transient", "fail")
		},
			retry.WithRandomSource(rand.NewPCG(1, 2)),
			retry.WithOnRetry(func(_ int, d time.Duration, _ error) { delays = append(delays, d) }),
		)

		return delays
	}

	first, second := run(), run()

	if len(first) != 2 {
		t.Fatalf("expected 2 jittered delays, got %d (%v)", len(first), first)
	}

	for i := range first {
		if first[i] != second[i] {
			t.Fatalf("same seed must reproduce the same sequence: first[%d]=%s second[%d]=%s",
				i, first[i], i, second[i])
		}
	}

	// Additive range per attempt: [base, base+50%] with base 1ms, 2ms.
	bounds := [][2]time.Duration{
		{time.Millisecond, 1500 * time.Microsecond},
		{2 * time.Millisecond, 3 * time.Millisecond},
	}
	for i, d := range first {
		if d < bounds[i][0] || d > bounds[i][1] {
			t.Fatalf("jittered delay[%d]=%s outside the additive range [%v, %v]",
				i, d, bounds[i][0], bounds[i][1])
		}
	}
}

func TestWithRandomSource_NilFallsBackToGlobal(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := retry.Do(ctx, fastConfig(), func(_ context.Context, _ int) error {
		return errorfamily.NewTransient("test.transient", "fail")
	}, retry.WithRandomSource(nil))
	if err == nil {
		t.Fatal("expected the context-end error; nil source must not panic")
	}
}

func TestWithRandomSource_ConcurrentCallsShareNoState(t *testing.T) {
	t.Parallel()

	var callGroup sync.WaitGroup

	errs := make(chan error, 32)

	for range 32 {
		callGroup.Go(func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			errs <- retry.Do(ctx, fastConfig(), func(_ context.Context, _ int) error {
				return errorfamily.NewTransient("test.transient", "fail")
			}, retry.WithRandomSource(rand.NewPCG(uint64(rand.Int64()), uint64(rand.Int64()))))
		})
	}

	callGroup.Wait()
	close(errs)

	for err := range errs {
		if err == nil {
			t.Fatal("expected the context-end error from every concurrent call")
		}
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
	// error: [infrastructure:retry.exhausted] all retry attempts failed: [transient:example.rate_limited] too many requests
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

// ExampleDoWithValue retries a call that produces a value: the successful
// attempt's result comes straight back, and any failure yields the zero
// value with the same error [Do] would produce.
func ExampleDoWithValue() {
	cfg := retry.DefaultConfig()
	cfg.MaxAttempts = 3
	cfg.InitialDelay = time.Millisecond

	user, err := retry.DoWithValue(context.Background(), cfg, func(ctx context.Context, attempt int) (string, error) {
		if attempt < 3 {
			// A Transient error is retryable by the default predicate.
			return "", errorfamily.NewTransient("example.transient", "lookup failed")
		}

		return "ada", nil
	})

	fmt.Println("user:", user)
	fmt.Println("error:", err)
	// Output:
	// user: ada
	// error: <nil>
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

func TestDoWithValue_ReturnsValueOnSuccess(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	got, err := retry.DoWithValue(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) (string, error) {
			calls.Add(1)

			if attempt < 2 {
				return "", errorfamily.NewTransient("test.transient", "fail")
			}

			return "payload", nil
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != "payload" {
		t.Fatalf("expected %q, got %q", "payload", got)
	}

	if calls.Load() != 2 {
		t.Fatalf("expected 2 calls, got %d", calls.Load())
	}
}

func TestDoWithValue_ReturnsZeroValueOnExhaustion(t *testing.T) {
	t.Parallel()

	got, err := retry.DoWithValue(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) (int, error) {
			return 42, errorfamily.NewTransient("test.transient", "fail")
		},
	)
	if !errors.Is(err, retry.ErrExhausted) {
		t.Fatalf("expected ErrExhausted, got %v", err)
	}

	if got != 0 {
		t.Fatalf("expected zero value on exhaustion, got %d", got)
	}
}

func TestDoWithValue_ReturnsZeroValueOnNonRetryable(t *testing.T) {
	t.Parallel()

	rejection := errorfamily.NewRejection("test.rejection", "invalid input")

	calls := 0

	got, err := retry.DoWithValue(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) (int, error) {
			calls++

			return 7, rejection
		},
	)
	if !errors.Is(err, rejection) {
		t.Fatalf("expected the rejection to pass through, got %v", err)
	}

	if got != 0 {
		t.Fatalf("expected zero value on non-retryable error, got %d", got)
	}

	if calls != 1 {
		t.Fatalf("expected exactly 1 call, got %d", calls)
	}
}

func TestDoWithValue_DoesNotLeakPartialValueOnLaterFailure(t *testing.T) {
	t.Parallel()

	got, err := retry.DoWithValue(
		context.Background(),
		fastConfig(),
		func(ctx context.Context, attempt int) (string, error) {
			if attempt == 1 {
				return "stale", errorfamily.NewTransient("test.transient", "fail")
			}

			return "", errorfamily.NewTransient("test.transient", "fail again")
		},
	)
	if err == nil {
		t.Fatal("expected an error")
	}

	if got != "" {
		t.Fatalf("expected no partial value from a failed attempt, got %q", got)
	}
}

func TestDo_OptionsTail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		opts        []retry.Option
		wantRetries int
		wantReject  bool
	}{
		{
			name:        "no options preserves struct-literal behavior",
			opts:        nil,
			wantRetries: fastConfig().MaxAttempts,
		},
		{
			name:        "empty options slice behaves identically",
			opts:        []retry.Option{},
			wantRetries: fastConfig().MaxAttempts,
		},
		{
			name:        "nil options are ignored",
			opts:        []retry.Option{nil},
			wantRetries: fastConfig().MaxAttempts,
		},
		{
			name:        "option raising MaxAttempts is honored",
			opts:        []retry.Option{func(c *retry.Config) { c.MaxAttempts = 5 }},
			wantRetries: 5,
		},
		{
			name:       "option setting an invalid value is rejected before the loop",
			opts:       []retry.Option{func(c *retry.Config) { c.MaxAttempts = 0 }},
			wantReject: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var calls atomic.Int32

			err := retry.Do(context.Background(), fastConfig(), func(_ context.Context, _ int) error {
				calls.Add(1)

				return errorfamily.NewTransient("test.transient", "always fails")
			}, tt.opts...)

			if tt.wantReject {
				if err == nil || errorfamily.Classify(err) != errorfamily.Rejection {
					t.Fatalf("expected a Rejection from Validate, got %v", err)
				}

				if got := calls.Load(); got != 0 {
					t.Fatalf("Validate must run before the loop, but fn was called %d times", got)
				}

				return
			}

			if err == nil {
				t.Fatal("expected exhaustion error")
			}

			if got := calls.Load(); got != int32(tt.wantRetries) {
				t.Fatalf("fn ran %d times, want %d", got, tt.wantRetries)
			}
		})
	}
}

func TestDo_OptionOverridesFieldForOneCallOnly(t *testing.T) {
	t.Parallel()

	var optionCalls, fieldCalls atomic.Int32

	onRetryField := func(_ int, _ time.Duration, _ error) { fieldCalls.Add(1) }
	onRetryOption := func(_ int, _ time.Duration, _ error) { optionCalls.Add(1) }

	cfg := fastConfig()
	cfg.OnRetry = onRetryField
	cfg.MaxAttempts = 2 // one retry per call: exactly one OnRetry fire

	custom := retry.Option(func(c *retry.Config) { c.OnRetry = onRetryOption })

	fail := func(_ context.Context, _ int) error {
		return errorfamily.NewTransient("test.transient", "fail")
	}

	if err := retry.Do(context.Background(), cfg, fail, custom); err == nil {
		t.Fatal("expected exhaustion error")
	}

	if optionCalls.Load() == 0 {
		t.Fatal("option callback never fired; option did not override the field")
	}

	if fieldCalls.Load() != 0 {
		t.Fatalf("field callback fired %d times despite option override", fieldCalls.Load())
	}

	optionCallsAfterFirst := optionCalls.Load()

	if err := retry.Do(context.Background(), cfg, fail); err == nil {
		t.Fatal("expected exhaustion error")
	}

	if fieldCalls.Load() == 0 {
		t.Fatal("field callback never fired on the option-free call; the field must keep working")
	}

	if optionCalls.Load() != optionCallsAfterFirst {
		t.Fatalf(
			"option callback fired on the option-free call (%d -> %d); an option must apply to one call only",
			optionCallsAfterFirst,
			optionCalls.Load(),
		)
	}
}

func TestDo_OptionsComposeLeftToRight(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	err := retry.Do(context.Background(), fastConfig(),
		func(_ context.Context, _ int) error {
			attempts.Add(1)

			return errorfamily.NewTransient("test.transient", "fail")
		},
		func(c *retry.Config) { c.MaxAttempts = 5 },
		func(c *retry.Config) { c.MaxAttempts = 2 },
	)
	if err == nil {
		t.Fatal("expected exhaustion error")
	}

	if got := attempts.Load(); got != 2 {
		t.Fatalf("later option must win: fn ran %d times, want 2", got)
	}
}

func TestDoWithValue_OptionsTail(t *testing.T) {
	t.Parallel()

	got, err := retry.DoWithValue(context.Background(), fastConfig(),
		func(_ context.Context, attempt int) (string, error) {
			if attempt < 3 {
				return "", errorfamily.NewTransient("test.transient", "fail twice")
			}

			return "ok", nil
		},
		func(c *retry.Config) { c.MaxAttempts = 3 },
	)
	if err != nil {
		t.Fatalf("DoWithValue() error = %v", err)
	}

	if got != "ok" {
		t.Fatalf("DoWithValue() = %q, want %q", got, "ok")
	}
}

func TestWithIsRetryable(t *testing.T) {
	t.Parallel()

	nonRetryable := errorfamily.NewRejection("test.rejection", "do not retry")

	err := retry.Do(context.Background(), fastConfig(), func(_ context.Context, _ int) error {
		return nonRetryable
	}, retry.WithIsRetryable(func(error) bool { return false }))
	if err != nonRetryable { //nolint:errorlint // identity is the assertion: not wrapped, not retried
		t.Fatalf("non-retryable error must come back unwrapped and unretried, got %v", err)
	}

	var retried atomic.Int32

	err = retry.Do(context.Background(), fastConfig(), func(_ context.Context, _ int) error {
		retried.Add(1)

		return errorfamily.NewRejection("test.rejection", "retry me anyway")
	}, retry.WithIsRetryable(func(error) bool { return true }))
	if err == nil {
		t.Fatal("expected exhaustion error")
	}

	if got := int(retried.Load()); got != fastConfig().MaxAttempts {
		t.Fatalf("WithIsRetryable(true) must drive retries to exhaustion, got %d calls", got)
	}
}

func TestWithDelayFunc(t *testing.T) {
	t.Parallel()

	t.Run("positive return overrides the backoff delay", func(t *testing.T) {
		t.Parallel()

		retry.Do(context.Background(), fastConfig(), func(_ context.Context, _ int) error {
			return errorfamily.NewTransient("test.transient", "fail")
		}, retry.WithDelayFunc(func(_ int, _ error) time.Duration {
			return 42 * time.Millisecond
		}), retry.WithOnRetry(func(_ int, delay time.Duration, _ error) {
			if delay != 42*time.Millisecond {
				t.Fatalf("OnRetry saw delay %s, want the DelayFunc override 42ms", delay)
			}
		}))
	})

	t.Run("zero return falls back to exponential backoff", func(t *testing.T) {
		t.Parallel()

		retry.Do(context.Background(), fastConfig(), func(_ context.Context, _ int) error {
			return errorfamily.NewTransient("test.transient", "fail")
		}, retry.WithDelayFunc(func(_ int, _ error) time.Duration { return 0 }))
	})
}

func TestWithOnRetryAndWithExhausted(t *testing.T) {
	t.Parallel()

	var retries, exhausted atomic.Int32

	err := retry.Do(context.Background(), fastConfig(), func(_ context.Context, attempt int) error {
		if attempt < fastConfig().MaxAttempts {
			return errorfamily.NewTransient("test.transient", "fail early")
		}

		return nil // succeed on the last attempt
	},
		retry.WithOnRetry(func(attempt int, _ time.Duration, _ error) {
			retries.Add(1)

			if attempt != int(retries.Load()) {
				t.Fatalf("OnRetry saw attempt %d out of order (fire #%d)", attempt, retries.Load())
			}
		}),
		retry.WithExhausted(func(int, error) {
			exhausted.Add(1)

			t.Error("WithExhausted must not fire when the loop succeeds")
		}))
	if err != nil {
		t.Fatalf("expected success on the final attempt, got %v", err)
	}

	if got := retries.Load(); got != 2 {
		t.Fatalf("OnRetry fired %d times, want 2 (once per failed attempt, before the sleep)", got)
	}
}

func TestOptions_FieldAndOptionCoexistence(t *testing.T) {
	t.Parallel()

	var fieldRetries, optionRetries atomic.Int32

	cfg := fastConfig()
	cfg.OnRetry = func(_ int, _ time.Duration, _ error) { fieldRetries.Add(1) }

	fail := func(_ context.Context, _ int) error {
		return errorfamily.NewTransient("test.transient", "fail")
	}

	// Option set, field set: option wins for that call.
	_ = retry.Do(context.Background(), cfg, fail, retry.WithOnRetry(func(_ int, _ time.Duration, _ error) {
		optionRetries.Add(1)
	}))

	if optionRetries.Load() == 0 || fieldRetries.Load() != 0 {
		t.Fatalf("option must win when both are set (option=%d field=%d)", optionRetries.Load(), fieldRetries.Load())
	}

	// Field set, no option: field keeps working.
	_ = retry.Do(context.Background(), cfg, fail)

	if fieldRetries.Load() == 0 {
		t.Fatal("field callback must keep working without options")
	}

	// Field unset, option set: option alone drives behavior.
	cfgNoField := fastConfig()
	cfgNoField.OnRetry = nil

	optionOnly := atomic.Int32{}

	_ = retry.Do(context.Background(), cfgNoField, fail, retry.WithOnRetry(func(_ int, _ time.Duration, _ error) {
		optionOnly.Add(1)
	}))

	if optionOnly.Load() == 0 {
		t.Fatal("option must work when the field is unset")
	}
}

func TestWithJitter_NoneIsDeterministic(t *testing.T) {
	t.Parallel()

	var delays []time.Duration

	cfg := fastConfig() // 1ms base, x2 multiplier, 5ms cap

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-canceled: the loop computes one delay, fires OnRetry, then exits

	err := retry.Do(ctx, cfg, func(_ context.Context, _ int) error {
		return errorfamily.NewTransient("test.transient", "fail")
	},
		retry.WithJitter(retry.JitterNone),
		retry.WithOnRetry(func(_ int, delay time.Duration, _ error) { delays = append(delays, delay) }),
	)
	if err == nil {
		t.Fatal("expected the context-end error")
	}

	want := []time.Duration{time.Millisecond} // attempt 1: min(1ms * 2^0, 5ms) = 1ms exactly
	if len(delays) != len(want) {
		t.Fatalf("OnRetry fired %d times (%v), want exactly %v", len(delays), delays, want)
	}

	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delay[%d] = %s, want exactly %s (jitter must be absent under JitterNone)", i, delays[i], want[i])
		}
	}
}

func TestWithJitter_NoneSequenceAcrossAttempts(t *testing.T) {
	t.Parallel()

	var delays []time.Duration

	retry.Do(context.Background(), fastConfig(), func(_ context.Context, attempt int) error {
		return errorfamily.NewTransient("test.transient", "always fail")
	},
		retry.WithJitter(retry.JitterNone),
		retry.WithOnRetry(func(_ int, delay time.Duration, _ error) { delays = append(delays, delay) }),
	)

	// Two sleeps (after attempts 1 and 2 of 3): 1ms, 2ms exactly.
	want := []time.Duration{time.Millisecond, 2 * time.Millisecond}
	if len(delays) != len(want) {
		t.Fatalf("got %d delays (%v), want %v", len(delays), delays, want)
	}

	for i := range want {
		if delays[i] != want[i] {
			t.Fatalf("delay[%d] = %s, want exactly %s", i, delays[i], want[i])
		}
	}
}

func TestWithJitter_NoneHonorsMaxDelayCap(t *testing.T) {
	t.Parallel()

	var delays []time.Duration

	cfg := retry.Config{
		MaxAttempts:  2,
		InitialDelay: 6 * time.Millisecond, // attempt 1 already exceeds the 5ms cap
		MaxDelay:     5 * time.Millisecond,
		Multiplier:   2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _ = retry.DoWithValue(ctx, cfg, func(_ context.Context, _ int) (int, error) {
		return 0, errorfamily.NewTransient("test.transient", "fail")
	},
		retry.WithJitter(retry.JitterNone),
		retry.WithOnRetry(func(_ int, delay time.Duration, _ error) { delays = append(delays, delay) }),
	)

	if len(delays) != 1 || delays[0] != 5*time.Millisecond {
		t.Fatalf("delay = %v, want exactly the 5ms cap (exponential 8ms must saturate)", delays)
	}
}

func TestWithJitter_MatrixAcrossStrategiesNeverPanics(t *testing.T) {
	t.Parallel()

	strategies := []retry.JitterStrategy{
		retry.JitterAdditive,
		retry.JitterNone,
		retry.JitterStrategy(42), // unknown: must fall back to additive, never panic
	}

	initials := []time.Duration{0, 1, 2, time.Millisecond, 100 * time.Millisecond, time.Second}
	maxDelays := []time.Duration{0, 1, time.Millisecond, 5 * time.Second}
	multipliers := []float64{0.5, 1.0, 1.5, 2.0, 10.0}

	for _, strategy := range strategies {
		for _, initial := range initials {
			for _, maxDelay := range maxDelays {
				for _, multiplier := range multipliers {
					var delay time.Duration

					cfg := retry.Config{
						MaxAttempts:  2,
						InitialDelay: initial,
						MaxDelay:     maxDelay,
						Multiplier:   multiplier,
					}

					ctx, cancel := context.WithCancel(context.Background())
					cancel()

					_ = retry.Do(ctx, cfg, func(_ context.Context, _ int) error {
						return errorfamily.NewTransient("test.transient", "fail")
					}, retry.WithJitter(strategy), retry.WithOnRetry(func(_ int, d time.Duration, _ error) {
						delay = d
					}))

					upperBound := maxDelay
					if upperBound <= 0 {
						upperBound = initial // B1: unset cap degrades to initial
					}

					if delay < 0 || delay > upperBound {
						t.Fatalf("strategy=%v initial=%v maxDelay=%v mult=%v: delay %v outside [0, %v]",
							strategy, initial, maxDelay, multiplier, delay, upperBound)
					}
				}
			}
		}
	}
}

func TestWithJitter_AdditiveStaysBounded(t *testing.T) {
	t.Parallel()

	var minDelay, maxDelay time.Duration

	minDelay = time.Hour

	for range 200 {
		var delay time.Duration

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_ = retry.Do(ctx, fastConfig(), func(_ context.Context, _ int) error {
			return errorfamily.NewTransient("test.transient", "fail")
		}, retry.WithJitter(retry.JitterAdditive), retry.WithOnRetry(func(_ int, d time.Duration, _ error) {
			delay = d
		}))

		if delay < minDelay {
			minDelay = delay
		}

		if delay > maxDelay {
			maxDelay = delay
		}
	}

	// Additive range for attempt 1: [1ms, 1.5ms] — jitter only adds, cap is 5ms.
	if minDelay < time.Millisecond {
		t.Fatalf("additive delay dipped below the exponential base: %v", minDelay)
	}

	if maxDelay > 1500*time.Microsecond {
		t.Fatalf("additive delay exceeded base+50%%: %v", maxDelay)
	}
}
