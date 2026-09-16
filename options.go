package retry

import "time"

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

// applyOptions applies opts to config left-to-right, skipping nil options.
func applyOptions(config *Config, opts []Option) {
	for _, opt := range opts {
		if opt != nil {
			opt(config)
		}
	}
}
