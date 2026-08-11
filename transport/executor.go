package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBodyBytes = 10 << 20

// Executor is the shared HTTP transport for the Birdeye REST API. It
// attaches X-API-KEY/x-chain headers, decodes Birdeye's {success, data}
// response envelope, captures response metadata, and retries idempotent
// (GET) calls per its RetryPolicy, honoring a Retry-After response
// header when present.
type Executor struct {
	cfg ExecutorConfig
}

// NewExecutor creates an Executor. Unset fields in cfg fall back to safe
// defaults (SystemClock, NoopLogger, NewDefaultRetryPolicy,
// &http.Client{}, the production BaseURL).
func NewExecutor(cfg ExecutorConfig) *Executor {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://public-api.birdeye.so"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 15 * time.Second}
	}
	if cfg.Clock == nil {
		cfg.Clock = SystemClock{}
	}
	if cfg.Logger == nil {
		cfg.Logger = NoopLogger{}
	}
	if cfg.RetryPolicy == nil {
		cfg.RetryPolicy = NewDefaultRetryPolicy()
	}
	return &Executor{cfg: cfg}
}

// Do issues a request and decodes Birdeye's envelope's "data" field into
// result (nil to discard it). chain overrides ExecutorConfig.DefaultChain
// for this call only; pass "" to use the client default. body is
// JSON-marshaled for POST/PUT/PATCH and ignored for GET/DELETE.
func (e *Executor) Do(ctx context.Context, method, path string, query map[string]string, chain string, body interface{}, result interface{}) (*ResponseMeta, error) {
	endpoint := path
	if len(query) > 0 {
		values := url.Values{}
		for k, v := range query {
			if v != "" {
				values.Set(k, v)
			}
		}
		if encoded := values.Encode(); encoded != "" {
			endpoint += "?" + encoded
		}
	}

	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("birdeye: marshal request body: %w", err)
		}
	}

	if chain == "" {
		chain = e.cfg.DefaultChain
	}

	retryable := strings.EqualFold(method, http.MethodGet)
	policy := e.cfg.RetryPolicy
	maxAttempts := 1
	if retryable && policy != nil && policy.MaxAttempts > 1 {
		maxAttempts = policy.MaxAttempts
	}

	start := e.cfg.Clock.Now()
	var lastErr error
	var lastMeta *ResponseMeta

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			delay := retryDelay(policy, attempt-1, lastMeta)
			if policy.MaxElapsed > 0 && e.cfg.Clock.Now().Sub(start)+delay > policy.MaxElapsed {
				break
			}
			if policy.OnRetry != nil {
				policy.OnRetry(attempt-1, delay, lastErr)
			}
			select {
			case <-ctx.Done():
				return lastMeta, ctx.Err()
			case <-time.After(delay):
			}
		}

		meta, err := e.attempt(ctx, method, endpoint, chain, bodyBytes, result)
		if meta != nil {
			meta.Attempts = attempt
		}
		lastMeta, lastErr = meta, err
		if err == nil {
			return meta, nil
		}
		if !retryable || !isRetryable(meta, err) {
			return meta, err
		}
	}

	return lastMeta, lastErr
}

// retryDelay prefers a Retry-After header from the previous response
// (seconds or an HTTP-date, per RFC 7231) over the policy's own
// backoff/jitter calculation.
func retryDelay(policy *RetryPolicy, attempt int, lastMeta *ResponseMeta) time.Duration {
	if lastMeta != nil && lastMeta.Headers != nil {
		if ra := lastMeta.Headers.Get("Retry-After"); ra != "" {
			if secs, err := strconv.Atoi(ra); err == nil && secs >= 0 {
				return time.Duration(secs) * time.Second
			}
			if when, err := http.ParseTime(ra); err == nil {
				if d := time.Until(when); d > 0 {
					return d
				}
			}
		}
	}
	return policy.delayForAttempt(attempt)
}

func isRetryable(meta *ResponseMeta, err error) bool {
	var netErr interface{ Temporary() bool }
	if errors.As(err, &netErr) && netErr.Temporary() {
		return true
	}
	if meta == nil {
		// No response at all (connection error, timeout) is retryable.
		return true
	}
	return meta.HTTPStatus == http.StatusTooManyRequests || meta.HTTPStatus >= 500
}

func (e *Executor) attempt(ctx context.Context, method, endpoint, chain string, bodyBytes []byte, result interface{}) (*ResponseMeta, error) {
	fullURL := e.cfg.BaseURL + endpoint

	var reader io.Reader
	if len(bodyBytes) > 0 {
		reader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), fullURL, reader)
	if err != nil {
		return nil, fmt.Errorf("birdeye: build request: %w", err)
	}
	req.Header.Set("X-API-KEY", e.cfg.APIKey)
	if chain != "" {
		req.Header.Set("x-chain", chain)
	}
	if len(bodyBytes) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	e.cfg.Logger.Debug("birdeye: request", "method", method, "path", endpoint, "chain", chain)

	resp, err := e.cfg.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("birdeye: do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBodyBytes+1))
	if err != nil {
		return nil, fmt.Errorf("birdeye: read response body: %w", err)
	}
	if len(respBody) > maxResponseBodyBytes {
		return nil, fmt.Errorf("%w (%d bytes)", ErrResponseTooLarge, maxResponseBodyBytes)
	}

	meta := &ResponseMeta{
		HTTPStatus: resp.StatusCode,
		RawBody:    respBody,
		Headers:    resp.Header,
		Attempts:   1,
	}

	var envelope struct {
		Success bool            `json:"success"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	decodeErr := json.Unmarshal(respBody, &envelope)
	if decodeErr == nil {
		meta.Success = envelope.Success
		meta.Message = envelope.Message
	}

	if sentinel := MapHTTPStatus(resp.StatusCode); sentinel != nil {
		return meta, fmt.Errorf("%w: %w", sentinel, &BirdeyeError{HTTPStatus: resp.StatusCode, Success: envelope.Success, Message: envelope.Message, Raw: respBody})
	}

	if decodeErr != nil {
		return meta, fmt.Errorf("birdeye: decode response envelope: %w", decodeErr)
	}

	if !envelope.Success {
		return meta, &BirdeyeError{HTTPStatus: resp.StatusCode, Success: false, Message: envelope.Message, Raw: respBody}
	}

	if result != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, result); err != nil {
			return meta, fmt.Errorf("birdeye: decode response data: %w", err)
		}
	}

	return meta, nil
}
