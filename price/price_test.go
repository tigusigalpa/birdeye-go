package price

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/tigusigalpa/birdeye-go/transport"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	executor := transport.NewExecutor(transport.ExecutorConfig{BaseURL: server.URL, APIKey: "test-key", RetryPolicy: transport.NoRetry()})
	return NewClient(executor)
}

func TestGetPrice(t *testing.T) {
	var gotPath, gotQuery string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"value":123.45,"updateUnixTime":1700000000,"updateHumanTime":"2023-11-14T22:13:20","priceChange24h":1.5}}`))
	})

	price, err := client.GetPrice(context.Background(), "So11111111111111111111111111111111111111112", PriceOptions{})
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}
	if gotPath != "/defi/price" || gotQuery != "address=So11111111111111111111111111111111111111112" {
		t.Errorf("unexpected request: path=%s query=%s", gotPath, gotQuery)
	}
	if price.Value != 123.45 || price.UpdateUnixTime != 1700000000 {
		t.Errorf("unexpected price: %+v", price)
	}
}

func TestGetPrice_SendsOptionalParams(t *testing.T) {
	var gotQuery url.Values
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"value":1}}`))
	})

	includeLiquidity := true
	checkLiquidity := 100.0
	_, err := client.GetPrice(context.Background(), "addr", PriceOptions{
		IncludeLiquidity: &includeLiquidity,
		CheckLiquidity:   &checkLiquidity,
		UIAmountMode:     UIAmountModeScaled,
	})
	if err != nil {
		t.Fatalf("GetPrice: %v", err)
	}
	if gotQuery.Get("include_liquidity") != "true" || gotQuery.Get("check_liquidity") != "100" || gotQuery.Get("ui_amount_mode") != "scaled" {
		t.Errorf("unexpected query: %v", gotQuery)
	}
}

func TestGetMultiPrice_JoinsAddressesAndDecodesNullEntries(t *testing.T) {
	var gotQuery string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"addr1":{"value":1},"addr2":null}}`))
	})

	result, err := client.GetMultiPrice(context.Background(), []string{"addr1", "addr2"}, PriceOptions{})
	if err != nil {
		t.Fatalf("GetMultiPrice: %v", err)
	}
	if gotQuery != "list_address=addr1%2Caddr2" {
		t.Errorf("unexpected query: %s", gotQuery)
	}
	if result["addr1"] == nil || result["addr1"].Value != 1 {
		t.Errorf("unexpected addr1: %+v", result["addr1"])
	}
	if result["addr2"] != nil {
		t.Errorf("expected addr2 to decode as nil, got %+v", result["addr2"])
	}
}

func TestGetMultiPricePOST_SendsCommaStringInBody(t *testing.T) {
	var gotBody map[string]interface{}
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		body := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotBody = body
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"addr1":{"value":1}}}`))
	})

	_, err := client.GetMultiPricePOST(context.Background(), []string{"addr1", "addr2"}, PriceOptions{})
	if err != nil {
		t.Fatalf("GetMultiPricePOST: %v", err)
	}
	if gotBody["list_address"] != "addr1,addr2" {
		t.Errorf("unexpected body: %v", gotBody)
	}
}

func TestGetHistoricalPriceByUnixTime(t *testing.T) {
	var gotQuery url.Values
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"value":50,"updateUnixTime":1700000000,"priceChange24h":0}}`))
	})

	price, err := client.GetHistoricalPriceByUnixTime(context.Background(), "addr", HistoricalPriceOptions{UnixTime: 1700000000})
	if err != nil {
		t.Fatalf("GetHistoricalPriceByUnixTime: %v", err)
	}
	if gotQuery.Get("unixtime") != "1700000000" || price.Value != 50 {
		t.Errorf("unexpected: query=%v price=%+v", gotQuery, price)
	}
}

func TestGetOHLCVv3_DecodesCandles(t *testing.T) {
	var gotQuery url.Values
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"items":[{"o":1,"h":2,"l":0.5,"c":1.5,"v":1000,"v_usd":1500,"unix_time":1700000000,"address":"addr","type":"1H","currency":"usd"}],"is_scaled_ui_token":false}}`))
	})

	page, err := client.GetOHLCVv3(context.Background(), TokenOHLCVOptions{Address: "addr", Type: Interval1H, TimeFrom: 1700000000, TimeTo: 1700003600})
	if err != nil {
		t.Fatalf("GetOHLCVv3: %v", err)
	}
	if gotQuery.Get("type") != "1H" || gotQuery.Get("time_from") != "1700000000" {
		t.Errorf("unexpected query: %v", gotQuery)
	}
	if len(page.Items) != 1 || page.Items[0].C != 1.5 {
		t.Errorf("unexpected page: %+v", page)
	}
}

func TestGetOHLCVv3Pair_DecodesCandles(t *testing.T) {
	var gotPath string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":{"items":[{"o":1,"h":2,"l":0.5,"c":1.5,"v":1000,"v_usd":1500,"address":"pairAddr","type":"1H","unix_time":1700000000,"currency":"usd"}]}}`))
	})

	page, err := client.GetOHLCVv3Pair(context.Background(), PairOHLCVOptions{Address: "pairAddr", Type: Interval1H, TimeFrom: 1700000000, TimeTo: 1700003600})
	if err != nil {
		t.Fatalf("GetOHLCVv3Pair: %v", err)
	}
	if gotPath != "/defi/v3/ohlcv/pair" || len(page.Items) != 1 || page.Items[0].Address != "pairAddr" {
		t.Errorf("unexpected: path=%s page=%+v", gotPath, page)
	}
}
