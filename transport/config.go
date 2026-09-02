// Package transport implements the shared HTTP request executor, response
// metadata capture, and error hierarchy used by every Birdeye service.
package transport

import (
	"math/rand"
	"net/http"
	"time"
)

// Clock abstracts time so tests can control timing. Real wall-clock time
// unless a Clock is injected.
type Clock interface {
	Now() time.Time
}

// SystemClock is the real wall clock, used unless a Clock is injected.
type SystemClock struct{}

// Now returns the current wall-clock time.
func (SystemClock) Now() time.Time { return time.Now() }

// Logger is a minimal structured-logging interface. Implementations must
// never log the API key or full response bodies containing sensitive
// wallet data.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// NoopLogger discards everything; the default unless a Logger is injected.
type NoopLogger struct{}

// Debug discards a debug log message.
func (NoopLogger) Debug(string, ...any) {}

// Info discards an informational log message.
func (NoopLogger) Info(string, ...any) {}

// Warn discards a warning log message.
func (NoopLogger) Warn(string, ...any) {}

// Error discards an error log message.
func (NoopLogger) Error(string, ...any) {}

// RetryPolicy controls automatic retries for idempotent (GET) REST calls
// only. POST requests are never auto-retried, since Birdeye's docs do not
// document idempotency keys for them.
type RetryPolicy struct {
	// MaxAttempts is the total number of attempts, including the first.
	// 1 disables retries.
	MaxAttempts int

	// BaseDelay and MaxDelay bound the exponential backoff (with full
	// jitter) between attempts, used when the response carries no
	// Retry-After header.
	BaseDelay time.Duration
	MaxDelay  time.Duration

	// MaxElapsed caps total time spent retrying a single call, including
	// the original attempt.
	MaxElapsed time.Duration

	// OnRetry, if set, is called before each retry attempt (1-indexed,
	// excluding the first attempt) with the delay about to be slept.
	OnRetry func(attempt int, delay time.Duration, err error)
}

// NewDefaultRetryPolicy returns a conservative policy: 3 attempts,
// 250ms-5s exponential backoff with full jitter, 20s max elapsed time.
func NewDefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts: 3,
		BaseDelay:   250 * time.Millisecond,
		MaxDelay:    5 * time.Second,
		MaxElapsed:  20 * time.Second,
	}
}

// NoRetry disables automatic retries entirely.
func NoRetry() *RetryPolicy {
	return &RetryPolicy{MaxAttempts: 1}
}

func (p *RetryPolicy) delayForAttempt(attempt int) time.Duration {
	if p.BaseDelay <= 0 {
		return 0
	}
	backoff := p.BaseDelay << uint(attempt-1) //nolint:gosec
	if backoff <= 0 || backoff > p.MaxDelay {
		backoff = p.MaxDelay
	}
	//nolint:gosec // non-cryptographic jitter is fine for retry pacing
	return time.Duration(rand.Int63n(int64(backoff) + 1))
}

// ExecutorConfig configures an Executor.
//
// Docs: https://docs.birdeye.so/reference/birdeye-api-authentication
type ExecutorConfig struct {
	// APIKey is sent as the X-API-KEY header on every request. Required —
	// Birdeye rejects every endpoint without it.
	APIKey string

	// BaseURL defaults to https://public-api.birdeye.so.
	BaseURL string

	// DefaultChain, if set, is sent as the x-chain header on every
	// request that doesn't specify its own chain override. Birdeye
	// defaults to "solana" server-side when the header is omitted
	// entirely, so leaving this empty is a valid choice, not a bug.
	DefaultChain string

	HTTPClient  *http.Client
	Clock       Clock
	Logger      Logger
	RetryPolicy *RetryPolicy
}
