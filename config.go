// Package birdeye is an idiomatic Go client for the Birdeye crypto market
// data API (https://public-api.birdeye.so), written from scratch against
// Birdeye's official documentation — not generated code.
//
// Docs: https://docs.birdeye.so/reference/birdeye-api-getting-started
package birdeye

import (
	"context"
	"net/http"
	"time"

	"github.com/tigusigalpa/birdeye-go/price"
	"github.com/tigusigalpa/birdeye-go/transport"
)

// DefaultBaseURL is Birdeye's production REST host.
const DefaultBaseURL = "https://public-api.birdeye.so"

// Chain values accepted by the x-chain header. This list is not
// exhaustive and not enforced client-side — different endpoint families
// support different chain subsets (confirmed via documentation research
// across only a handful of endpoints), so passing an unlisted chain
// string is never blocked locally; let Birdeye's API be the source of
// truth and return an error for an unsupported chain.
//
// Docs: https://docs.birdeye.so/reference/birdeye-api-authentication
const (
	ChainSolana    = "solana"
	ChainEthereum  = "ethereum"
	ChainArbitrum  = "arbitrum"
	ChainAvalanche = "avalanche"
	ChainBSC       = "bsc"
	ChainOptimism  = "optimism"
	ChainPolygon   = "polygon"
	ChainBase      = "base"
	ChainZkSync    = "zksync"
	ChainMonad     = "monad"
	ChainHyperEVM  = "hyperevm"
	ChainAptos     = "aptos"
	ChainFogo      = "fogo"
	ChainMantle    = "mantle"
	ChainMegaETH   = "megaeth"
	ChainRobinhood = "robinhood"
	ChainSui       = "sui"
)

// Re-exported from package transport so callers only need to import
// package birdeye for the common case.
type (
	// Clock supplies the current time to client components that need it.
	Clock = transport.Clock
	// Logger receives structured client lifecycle messages.
	Logger = transport.Logger
	// RetryPolicy configures automatic retry behaviour for eligible requests.
	RetryPolicy = transport.RetryPolicy
)

var (
	// NewDefaultRetryPolicy creates the SDK's conservative default retry policy.
	NewDefaultRetryPolicy = transport.NewDefaultRetryPolicy
	// NoRetry creates a retry policy that makes one attempt only.
	NoRetry = transport.NoRetry
)

// ClientConfig holds every configurable knob of a Client. Zero value is
// unusable — an APIKey is always required; NewClient fills in defaults
// for every other unset field.
type ClientConfig struct {
	APIKey string

	BaseURL      string
	DefaultChain string

	HTTPClient *http.Client
	Timeout    time.Duration

	Clock  Clock
	Logger Logger

	RetryPolicy *RetryPolicy
}

// Option configures a ClientConfig at construction time.
type Option func(*ClientConfig)

// WithBaseURL overrides the REST host. Defaults to DefaultBaseURL.
func WithBaseURL(url string) Option {
	return func(c *ClientConfig) { c.BaseURL = url }
}

// WithChain sets the default x-chain header sent on every request that
// doesn't override it per-call via a method's Options.Chain field.
// Leaving this unset is valid — Birdeye defaults to "solana" server-side
// when the header is omitted entirely.
func WithChain(chain string) Option {
	return func(c *ClientConfig) { c.DefaultChain = chain }
}

// WithHTTPClient uses hc to make API requests. Its timeout and transport are
// preserved, so WithTimeout has no effect when this option is used.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *ClientConfig) { c.HTTPClient = hc }
}

// WithTimeout sets the timeout of the SDK-created HTTP client. Use
// WithHTTPClient to configure a custom client directly.
func WithTimeout(d time.Duration) Option {
	return func(c *ClientConfig) { c.Timeout = d }
}

// WithClock injects a clock used by retry timing. It is mainly useful for
// deterministic tests.
func WithClock(clock Clock) Option {
	return func(c *ClientConfig) { c.Clock = clock }
}

// WithLogger sets the structured logger used by the client. The default logger
// discards all messages.
func WithLogger(logger Logger) Option {
	return func(c *ClientConfig) { c.Logger = logger }
}

// WithRetryPolicy overrides the default conservative retry policy
// (GET requests only; see transport.RetryPolicy). Pass NoRetry() to
// disable automatic retries entirely.
func WithRetryPolicy(policy *RetryPolicy) Option {
	return func(c *ClientConfig) { c.RetryPolicy = policy }
}

// WithRetries changes only the maximum number of attempts while retaining the
// default backoff bounds. Values below one disable retries. GET requests alone
// are eligible; unsafe methods are never retried automatically.
func WithRetries(maxAttempts int) Option {
	return func(c *ClientConfig) {
		policy := transport.NewDefaultRetryPolicy()
		if maxAttempts < 1 {
			maxAttempts = 1
		}
		policy.MaxAttempts = maxAttempts
		c.RetryPolicy = policy
	}
}

func newConfig(apiKey string, opts ...Option) *ClientConfig {
	cfg := &ClientConfig{
		APIKey:      apiKey,
		BaseURL:     DefaultBaseURL,
		Timeout:     15 * time.Second,
		Clock:       transport.SystemClock{},
		Logger:      transport.NoopLogger{},
		RetryPolicy: transport.NewDefaultRetryPolicy(),
	}
	for _, opt := range opts {
		opt(cfg)
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	return cfg
}

// Client is the SDK entry point. Construct with NewClient.
type Client struct {
	cfg      *ClientConfig
	executor *transport.Executor

	// Price groups every implemented price/OHLCV endpoint.
	//
	// Docs: https://docs.birdeye.so/reference/price-ohlcv
	Price *price.Client
}

// NewClient builds a fully wired Client. apiKey is required — Birdeye
// rejects every endpoint without an X-API-KEY header. Performs no
// network I/O.
//
// Docs: https://docs.birdeye.so/reference/birdeye-api-authentication
func NewClient(apiKey string, opts ...Option) *Client {
	cfg := newConfig(apiKey, opts...)

	executor := transport.NewExecutor(transport.ExecutorConfig{
		APIKey:       cfg.APIKey,
		BaseURL:      cfg.BaseURL,
		DefaultChain: cfg.DefaultChain,
		HTTPClient:   cfg.HTTPClient,
		Clock:        cfg.Clock,
		Logger:       cfg.Logger,
		RetryPolicy:  cfg.RetryPolicy,
	})

	return &Client{
		cfg:      cfg,
		executor: executor,
		Price:    price.NewClient(executor),
	}
}

// Do is the raw request escape hatch: call any Birdeye endpoint —
// including ones this SDK hasn't mapped to a typed method yet — without
// waiting for an SDK update. method/path/query/body follow the same
// conventions as every typed service method; chain overrides the
// client's default x-chain header for this call only ("" to use the
// default). result is decoded from the response envelope's "data"
// field (nil to discard it).
func (c *Client) Do(ctx context.Context, method, path string, query map[string]string, chain string, body interface{}, result interface{}) (*transport.ResponseMeta, error) {
	return c.executor.Do(ctx, method, path, query, chain, body, result)
}

// DoWithHeaders is the raw escape hatch for endpoints that require an
// endpoint-specific header, such as x-perp. X-API-KEY is always taken from
// the client configuration and cannot be overridden here.
func (c *Client) DoWithHeaders(ctx context.Context, method, path string, query map[string]string, chain string, headers http.Header, body interface{}, result interface{}) (*transport.ResponseMeta, error) {
	return c.executor.DoWithHeaders(ctx, method, path, query, chain, headers, body, result)
}
