# Birdeye Go SDK

![BirdEye Golang SDK](https://i.postimg.cc/FFqTsjMG/birdeye-golang-github.jpg)

A small, dependable Go client for the [Birdeye public API](https://docs.birdeye.so/reference/birdeye-api-getting-started). It is built for applications that need current token prices and chart data without hiding HTTP details, swallowing API errors, or making surprise retries.

Maintained by [Igor Sazonov](https://github.com/tigusigalpa) · [sovletig@gmail.com](mailto:sovletig@gmail.com). This is an independent community project, not an official Birdeye SDK.

## Install

```bash
go get github.com/tigusigalpa/birdeye-go
```

Go 1.22 or later is required.

## Start here

Put your key in the environment, not in source control:

```bash
export BIRDEYE_API_KEY="..."
```

```go
package main

import (
    "context"
    "fmt"
    "os"

    birdeye "github.com/tigusigalpa/birdeye-go"
    "github.com/tigusigalpa/birdeye-go/price"
)

func main() {
    client := birdeye.NewClient(os.Getenv("BIRDEYE_API_KEY"), birdeye.WithChain(birdeye.ChainSolana))
    quote, err := client.Price.GetPrice(context.Background(), "So11111111111111111111111111111111111111112", price.PriceOptions{})
    if err != nil { panic(err) }
    fmt.Println(quote.Value)
}
```

Birdeye receives `X-API-KEY` on every request. `WithChain` supplies a default `x-chain`; each request options struct also has a `Chain` field for a one-off override. The client deliberately does not guess whether your Birdeye plan has access to an endpoint—Birdeye remains the authority and returns an ordinary API error when access is denied.

## What is included

| Area | Methods | Birdeye reference |
|---|---|---|
| Spot prices | Single price; multi-price (GET and POST) | [Price & OHLCV](https://docs.birdeye.so/reference/price-ohlcv) |
| Historical prices | Point-in-time and time-series price | [Historical price](https://docs.birdeye.so/reference/get-defi-history_price) |
| Candles | V3 token/pair OHLCV and base/quote OHLCV | [OHLCV V3](https://docs.birdeye.so/reference/get-defi-v3-ohlcv) |
| Rolling activity | Single and batch price-volume snapshots | [Price volume](https://docs.birdeye.so/reference/get-defi-price_volume-single) |

The detailed route-to-method map is in [docs/endpoints.md](docs/endpoints.md). When Birdeye has not published a stable field schema, those methods return `price.RawObject`, preserving each response field as `json.RawMessage` instead of inventing a brittle model.

The legacy `/defi/ohlcv` and `/defi/ohlcv/pair` endpoints are deprecated upstream, so this package intentionally does not wrap them. Wallets, transactions, token lists, Perps, blockchain data, x402, and WebSockets are not yet mapped as typed services.

## Examples

The runnable examples read `BIRDEYE_API_KEY` themselves:

```bash
go run ./examples/get_price
go run ./examples/get_ohlcv
go run ./examples/multi_price
go run ./examples/error_handling
go run ./examples/token_search
go run ./examples/wallet_pnl # also requires BIRDEYE_WALLET
```

## Errors and retries

Use `errors.Is` for common HTTP outcomes, or `errors.As` to inspect the original Birdeye error. `BirdeyeError` carries the HTTP status, Birdeye error code when supplied, request ID, message, and a size-limited raw response body.

```go
var apiErr *transport.BirdeyeError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.HTTPStatus, apiErr.Code, apiErr.RequestID)
}
```

GET requests retry only on HTTP 429 and transient network errors. Retries use bounded exponential backoff with jitter and honour `Retry-After`. POST requests are never retried automatically. Configure the policy with `birdeye.WithRetryPolicy(...)`, or pass `birdeye.NoRetry()` to turn it off.

## Any endpoint, today

`Client.Do` is an escape hatch for new Birdeye routes. It still handles authentication, chain selection, envelope decoding, errors, and safe retry behaviour.

```go
var data map[string]any
_, err := client.Do(ctx, "GET", "/defi/token_overview", map[string]string{"address": address}, "", nil, &data)
```

For endpoint-specific headers such as `x-perp`, use `DoWithHeaders`. It cannot replace the client's API key.

```go
_, err := client.DoWithHeaders(ctx, "GET", "/perps/v1/token/list", nil, "", http.Header{"x-perp": {"true"}}, nil, &data)
```

## Development checks

Tests are hand-authored `httptest` fixtures and never call Birdeye or consume Compute Units.

```bash
go test ./...
go test -race ./...
go vet ./...
gofmt -l .
```

## Security

Do not commit a real `BIRDEYE_API_KEY`. `.env.example` is safe to copy; keep your own `.env` private. The default client never logs your key or response bodies.

## License

MIT. See [LICENSE](LICENSE).
