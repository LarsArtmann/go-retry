package retry

import (
	"math/rand/v2"
	"time"
)

// Option overrides one piece of [Config] for a single [Do] or [DoWithValue]
// call. Options make new capabilities addable without ever changing Config's
// public field shape, so struct-literal callers keep compiling forever.
//
// Options are applied left-to-right after the caller's Config is copied and
// before Validate runs — a later option wins over both the field and earlier
// options, and an option that sets an invalid value produces the same
// Rejection error the field would. Nil options in the slice are ignored,
// matching how nil callbacks are treated as "not set". The matching-Config
// mirror options (WithIsRetryable, WithOnRetry, …) and the options-only
// capabilities (WithJitter, WithRandomSource) are one-line constructors; see
// their docs.
type Option func(*Config)

// WithIsRetryable overrides [Config.IsRetryable] for one call: the predicate
// deciding whether an error triggers a retry. See the Config field for
// semantics, including the nil-default behavior.
func WithIsRetryable(f func(error) bool) Option {
	return func(c *Config) { c.IsRetryable = f }
}

// WithDelayFunc overrides [Config.DelayFunc] for one call: the per-attempt
// backoff override (e.g. honoring a server-provided Retry-After). See the
// Config field for the 0-falls-back-to-exponential semantics.
func WithDelayFunc(f func(attempt int, err error) time.Duration) Option {
	return func(c *Config) { c.DelayFunc = f }
}

// WithOnRetry overrides [Config.OnRetry] for one call: the callback fired
// after each failed retryable attempt, before the sleep.
func WithOnRetry(f func(attempt int, delay time.Duration, err error)) Option {
	return func(c *Config) { c.OnRetry = f }
}

// WithExhausted overrides [Config.OnExhausted] for one call: the callback
// fired once after all attempts have failed (never on context end).
func WithExhausted(f func(attempts int, err error)) Option {
	return func(c *Config) { c.OnExhausted = f }
}

// JitterStrategy selects how jitter is applied to the computed backoff
// delay. The zero value is [JitterAdditive] — the package's default since
// v0.1.0 — so callers who never pass options see identical delays.
type JitterStrategy int

const (
	// JitterAdditive adds a random value of up to 50% of the capped
	// exponential delay on top, then caps the sum at MaxDelay. The actual
	// wait is therefore in [base, min(base*1.5, MaxDelay)] — never above
	// MaxDelay.
	JitterAdditive JitterStrategy = iota

	// JitterNone disables jitter entirely: the delay is the pure capped
	// exponential
	//
	//	min(InitialDelay * Multiplier^(n-1), MaxDelay)
	//
	// Deterministic, which makes it the right choice for tests that assert
	// exact delay sequences, for previews, and for callers who need
	// reproducible timing.
	JitterNone
)

// WithJitter overrides the jitter strategy for one call. See
// [JitterStrategy] for the available strategies; the zero-value default is
// [JitterAdditive].
func WithJitter(s JitterStrategy) Option {
	return func(c *Config) { c.jitterStrategy = s }
}

// WithRandomSource overrides the randomness used for jitter for one call.
// Pass a [rand.Source] (e.g. a seeded rand.NewPCG) to make delays
// reproducible — the foundation for exact jittered-delay assertions in
// tests. A nil source (the default) uses the package-global
// math/rand/v2 generator.
//
// A source is consumed sequentially by the retry loop of one call. Do not
// share one non-thread-safe source across concurrently running calls; wrap
// it in a locking source if you must.
func WithRandomSource(src rand.Source) Option {
	return func(c *Config) { c.randSource = src }
}

// applyOptions applies opts to config left-to-right, skipping nil options.
func applyOptions(config *Config, opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}
}
