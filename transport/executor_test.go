package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestExecutor(t *testing.T, handler http.HandlerFunc, apiKey string) (*Executor, *httptest.Server) {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	exec := NewExecutor(ExecutorConfig{
		BaseURL:     server.URL,
		APIKey:      apiKey,
		RetryPolicy: NoRetry(),
	})
	return exec, server
}

func TestDo_SetsAPIKeyHeader(t *testing.T) {
	var gotHeaders http.Header
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}, "test-key")

	_, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if got := gotHeaders.Get("X-API-KEY"); got != "test-key" {
		t.Errorf("X-API-KEY = %q, want test-key", got)
	}
}

func TestDo_SetsChainHeaderOnlyWhenNonEmpty(t *testing.T) {
	var gotChain string
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		gotChain = r.Header.Get("x-chain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}, "k")

	if _, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotChain != "" {
		t.Errorf("x-chain = %q, want empty when no chain configured", gotChain)
	}

	if _, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "ethereum", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotChain != "ethereum" {
		t.Errorf("x-chain = %q, want ethereum", gotChain)
	}
}

func TestDo_ChainOverrideBeatsDefaultChain(t *testing.T) {
	var gotChain string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotChain = r.Header.Get("x-chain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	defer server.Close()

	exec := NewExecutor(ExecutorConfig{BaseURL: server.URL, APIKey: "k", DefaultChain: "solana", RetryPolicy: NoRetry()})

	if _, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "ethereum", nil, nil); err != nil {
		t.Fatalf("Do: %v", err)
	}
	if gotChain != "ethereum" {
		t.Errorf("x-chain = %q, want ethereum (per-call override)", gotChain)
	}
}

func TestDo_DecodesResultData(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"value":1.23}}`))
	}, "k")

	var result struct {
		Value float64 `json:"value"`
	}
	_, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, &result)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result.Value != 1.23 {
		t.Errorf("Value = %v, want 1.23", result.Value)
	}
}

func TestDo_NullDataLeavesResultZeroValue(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":null}`))
	}, "k")

	result := struct {
		Value float64 `json:"value"`
	}{Value: 99}
	_, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, &result)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if result.Value != 99 {
		t.Errorf("null data should leave result untouched, got %+v", result)
	}
}

func TestDo_SuccessFalseReturnsBirdeyeError(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":false,"message":"invalid address"}`))
	}, "k")

	_, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, nil)
	var birdeyeErr *BirdeyeError
	if !errors.As(err, &birdeyeErr) {
		t.Fatalf("expected *BirdeyeError, got %T: %v", err, err)
	}
	if birdeyeErr.Message != "invalid address" {
		t.Errorf("unexpected message: %+v", birdeyeErr)
	}
}

func TestDo_MapsHTTPStatusSentinel(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"success":false,"message":"invalid key"}`))
	}, "bad-key")

	_, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, nil)
	if !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
}

func TestDo_RetriesOnlyRateLimitedGETRequests(t *testing.T) {
	var getAttempts, postAttempts int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getAttempts++
		} else {
			postAttempts++
		}
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"success":false,"message":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"success":false,"message":"rate limited"}`))
	}))
	defer server.Close()

	exec := NewExecutor(ExecutorConfig{
		BaseURL:     server.URL,
		APIKey:      "k",
		RetryPolicy: &RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, MaxElapsed: time.Second},
	})

	meta, _ := exec.Do(context.Background(), http.MethodGet, "/path", nil, "", nil, nil)
	_, _ = exec.Do(context.Background(), http.MethodPost, "/path", nil, "", map[string]string{"a": "b"}, nil)

	if meta == nil || meta.Attempts != 3 {
		t.Errorf("response attempts = %+v, want 3", meta)
	}
	if getAttempts != 3 {
		t.Errorf("GET attempts = %d, want 3 (retried)", getAttempts)
	}
	if postAttempts != 1 {
		t.Errorf("POST attempts = %d, want 1 (never retried)", postAttempts)
	}
}

func TestDo_DoesNotRetryServerErrors(t *testing.T) {
	attempts := 0
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"success":false,"message":"temporary server failure"}`))
	}, "k")
	exec.cfg.RetryPolicy = &RetryPolicy{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Millisecond, MaxElapsed: time.Second}
	_, err := exec.Do(context.Background(), http.MethodGet, "/path", nil, "", nil, nil)
	if !errors.Is(err, ErrServerError) || attempts != 1 {
		t.Fatalf("server errors must not be retried: attempts=%d err=%v", attempts, err)
	}
}

func TestDo_CapturesRequestIDAndErrorCode(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "req-123")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"success":false,"code":1001,"message":"invalid address"}`))
	}, "k")
	meta, err := exec.Do(context.Background(), http.MethodGet, "/path", nil, "", nil, nil)
	var apiErr *BirdeyeError
	if !errors.As(err, &apiErr) || apiErr.Code != "1001" || apiErr.RequestID != "req-123" || meta.RequestID != "req-123" {
		t.Fatalf("unexpected error metadata: meta=%+v err=%+v", meta, apiErr)
	}
}

func TestDoWithHeaders_AddsEndpointSpecificHeaderWithoutReplacingAPIKey(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-perp"); got != "true" {
			t.Errorf("x-perp = %q, want true", got)
		}
		if got := r.Header.Get("X-API-KEY"); got != "configured-key" {
			t.Errorf("X-API-KEY = %q, want configured-key", got)
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}, "configured-key")
	_, err := exec.DoWithHeaders(context.Background(), http.MethodGet, "/perps/v1/token/list", nil, "", http.Header{"x-perp": {"true"}, "X-API-KEY": {"replacement"}}, nil, nil)
	if err != nil {
		t.Fatalf("DoWithHeaders: %v", err)
	}
}

func TestDo_RespectsRetryAfterHeader(t *testing.T) {
	var attempts int
	var timestamps []time.Time

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timestamps = append(timestamps, time.Now())
		attempts++
		if attempts == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"success":false,"message":"rate limited"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{}}`))
	}))
	defer server.Close()

	exec := NewExecutor(ExecutorConfig{
		BaseURL:     server.URL,
		APIKey:      "k",
		RetryPolicy: &RetryPolicy{MaxAttempts: 2, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond, MaxElapsed: 5 * time.Second},
	})

	_, err := exec.Do(context.Background(), http.MethodGet, "/path", nil, "", nil, nil)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	if len(timestamps) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(timestamps))
	}
	if elapsed := timestamps[1].Sub(timestamps[0]); elapsed < 900*time.Millisecond {
		t.Errorf("expected Retry-After: 1 to delay ~1s, only waited %v", elapsed)
	}
}

func TestDo_RejectsOversizedResponse(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(make([]byte, maxResponseBodyBytes+1))
	}, "k")

	_, err := exec.Do(context.Background(), http.MethodGet, "/defi/price", nil, "", nil, nil)
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("expected ErrResponseTooLarge, got %v", err)
	}
}

func TestMapHTTPStatus_Success(t *testing.T) {
	if err := MapHTTPStatus(200); err != nil {
		t.Errorf("expected nil for 200, got %v", err)
	}
}

func TestMapHTTPStatus(t *testing.T) {
	tests := []struct {
		status int
		want   error
	}{
		{201, nil}, {400, ErrBadRequest}, {401, ErrUnauthorized}, {403, ErrForbidden},
		{404, ErrNotFound}, {429, ErrRateLimited}, {500, ErrServerError}, {418, ErrBadRequest}, {302, nil},
	}
	for _, test := range tests {
		if got := MapHTTPStatus(test.status); !errors.Is(got, test.want) {
			t.Errorf("MapHTTPStatus(%d) = %v, want %v", test.status, got, test.want)
		}
	}
}

func TestNewExecutor_DefaultsAndHelpers(t *testing.T) {
	exec := NewExecutor(ExecutorConfig{})
	if exec.cfg.BaseURL == "" || exec.cfg.HTTPClient == nil || exec.cfg.Clock == nil || exec.cfg.Logger == nil || exec.cfg.RetryPolicy == nil {
		t.Fatalf("defaults were not applied: %+v", exec.cfg)
	}
	if policy := NewDefaultRetryPolicy(); policy.MaxAttempts != 3 || policy.BaseDelay <= 0 || policy.MaxDelay <= 0 {
		t.Fatalf("unexpected default policy: %+v", policy)
	}
	logger := NoopLogger{}
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
}

func TestRetryDelayAndRetryableClassification(t *testing.T) {
	policy := &RetryPolicy{BaseDelay: time.Millisecond, MaxDelay: 4 * time.Millisecond}
	if delay := retryDelay(policy, 1, &ResponseMeta{Headers: http.Header{"Retry-After": {"0"}}}); delay != 0 {
		t.Errorf("retry delay = %s, want 0", delay)
	}
	date := time.Now().Add(time.Second).UTC().Format(http.TimeFormat)
	if delay := retryDelay(policy, 1, &ResponseMeta{Headers: http.Header{"Retry-After": {date}}}); delay <= 0 {
		t.Errorf("HTTP-date retry delay = %s, want positive", delay)
	}
	if delay := retryDelay(policy, 3, &ResponseMeta{Headers: http.Header{"Retry-After": {"invalid"}}}); delay < 0 || delay > policy.MaxDelay {
		t.Errorf("fallback retry delay = %s", delay)
	}
	if !isRetryable(nil, errors.New("connection refused")) {
		t.Error("missing response should be retryable")
	}
	if isRetryable(nil, context.Canceled) || isRetryable(nil, context.DeadlineExceeded) {
		t.Error("cancelled contexts must not be retryable")
	}
	if !isRetryable(&ResponseMeta{HTTPStatus: http.StatusTooManyRequests}, errors.New("rate limited")) {
		t.Error("429 should be retryable")
	}
	if isRetryable(&ResponseMeta{HTTPStatus: http.StatusInternalServerError}, errors.New("server error")) {
		t.Error("500 should not be retryable")
	}
}

func TestDo_QueryFilteringAndInvalidPayloads(t *testing.T) {
	exec, _ := newTestExecutor(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.RawQuery; got != "keep=value" {
			t.Errorf("query = %q, want keep=value", got)
		}
		_, _ = w.Write([]byte(`not-json`))
	}, "k")
	if _, err := exec.Do(context.Background(), http.MethodGet, "/path", map[string]string{"keep": "value", "omit": ""}, "", nil, nil); err == nil {
		t.Fatal("invalid envelope did not return an error")
	}

	if _, err := exec.Do(context.Background(), http.MethodPost, "/path", nil, "", make(chan int), nil); err == nil {
		t.Fatal("unmarshalable body did not return an error")
	}
}
