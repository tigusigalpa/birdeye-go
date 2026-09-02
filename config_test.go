package birdeye

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tigusigalpa/birdeye-go/transport"
)

func TestNewClient_AppliesOptions(t *testing.T) {
	httpClient := &http.Client{Timeout: 3 * time.Second}
	client := NewClient("test-key",
		WithBaseURL("https://example.test/"),
		WithChain(ChainEthereum),
		WithHTTPClient(httpClient),
		WithRetries(2),
	)

	if client.Price == nil {
		t.Fatal("Price service was not initialized")
	}
	if client.cfg.APIKey != "test-key" || client.cfg.BaseURL != "https://example.test/" {
		t.Fatalf("unexpected client config: %+v", client.cfg)
	}
	if client.cfg.DefaultChain != ChainEthereum || client.cfg.HTTPClient != httpClient {
		t.Fatalf("options were not retained: %+v", client.cfg)
	}
	if client.cfg.RetryPolicy.MaxAttempts != 2 {
		t.Errorf("MaxAttempts = %d, want 2", client.cfg.RetryPolicy.MaxAttempts)
	}
}

func TestWithRetries_ClampsNonPositiveValues(t *testing.T) {
	client := NewClient("test-key", WithRetries(0))
	if client.cfg.RetryPolicy.MaxAttempts != 1 {
		t.Errorf("MaxAttempts = %d, want 1", client.cfg.RetryPolicy.MaxAttempts)
	}
}

func TestNewClient_AppliesAllConfigurationOptions(t *testing.T) {
	policy := &RetryPolicy{MaxAttempts: 4}
	clock := transport.SystemClock{}
	logger := transport.NoopLogger{}
	client := NewClient("test-key",
		WithTimeout(7*time.Second),
		WithClock(clock),
		WithLogger(logger),
		WithRetryPolicy(policy),
	)

	if client.cfg.HTTPClient.Timeout != 7*time.Second {
		t.Errorf("timeout = %s, want 7s", client.cfg.HTTPClient.Timeout)
	}
	if client.cfg.Clock != clock || client.cfg.Logger != logger || client.cfg.RetryPolicy != policy {
		t.Fatalf("configuration options were not retained: %+v", client.cfg)
	}
}

func TestClientDoAndDoWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "test-key" {
			t.Errorf("missing API key")
		}
		if r.URL.Path == "/second" && r.Header.Get("x-extra") != "value" {
			t.Errorf("missing custom header")
		}
		_, _ = w.Write([]byte(`{"success":true,"data":{"ok":true}}`))
	}))
	defer server.Close()

	client := NewClient("test-key", WithBaseURL(server.URL), WithRetryPolicy(NoRetry()))
	var first, second struct {
		OK bool `json:"ok"`
	}
	if _, err := client.Do(context.Background(), http.MethodGet, "/first", nil, "", nil, &first); err != nil || !first.OK {
		t.Fatalf("Do() = %+v, %v", first, err)
	}
	if _, err := client.DoWithHeaders(context.Background(), http.MethodGet, "/second", nil, "", http.Header{"x-extra": {"value"}}, nil, &second); err != nil || !second.OK {
		t.Fatalf("DoWithHeaders() = %+v, %v", second, err)
	}
}
