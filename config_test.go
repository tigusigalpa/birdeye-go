package birdeye

import (
	"net/http"
	"testing"
	"time"
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
