# Birdeye Go SDK

![Birdeye Golang SDK](https://i.postimg.cc/FFqTsjMG/birdeye-golang-github.jpg)

[![CI](https://github.com/tigusigalpa/birdeye-go/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/tigusigalpa/birdeye-go/actions/workflows/ci.yml)
[![Tests](https://github.com/tigusigalpa/birdeye-go/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/tigusigalpa/birdeye-go/actions/workflows/test.yml)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green?style=flat-square)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/tigusigalpa/birdeye-go?style=flat-square)](https://goreportcard.com/report/github.com/tigusigalpa/birdeye-go)
[![CodeQL](https://github.com/tigusigalpa/birdeye-go/actions/workflows/codeql.yml/badge.svg?branch=main)](https://github.com/tigusigalpa/birdeye-go/actions/workflows/codeql.yml)
[![Codecov](https://codecov.io/gh/tigusigalpa/birdeye-go/graph/badge.svg)](https://codecov.io/gh/tigusigalpa/birdeye-go)
[![GitHub Release](https://img.shields.io/github/v/release/tigusigalpa/birdeye-go?style=flat-square)](https://github.com/tigusigalpa/birdeye-go/releases)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue?style=flat-square&logo=go)](https://pkg.go.dev/github.com/tigusigalpa/birdeye-go)

An idiomatic, context-first Go client for the [Birdeye public API](https://docs.birdeye.so/reference/birdeye-api-getting-started). It currently focuses on the price and OHLCV workflows that most trading dashboards, portfolio tools, and market monitors need first.

The SDK handles authentication, JSON envelopes, request metadata, typed API errors, and conservative retries. It also includes a raw request API, so a new Birdeye endpoint never has to block your work.

Maintained by [Igor Sazonov](https://github.com/tigusigalpa) · [sovletig@gmail.com](mailto:sovletig@gmail.com). This is an independent community project, not an official Birdeye SDK.

## Contents

- [Requirements and installation](#requirements-and-installation)
- [Configure the client](#configure-the-client)
- [Quick start](#quick-start)
- [Common recipes](#common-recipes)
- [Supported API surface](#supported-api-surface)
- [Errors and retries](#errors-and-retries)
- [Use an endpoint before it is typed](#use-an-endpoint-before-it-is-typed)
- [Examples, testing, and security](#examples-testing-and-security)

## Requirements and installation

This module requires Go **1.22+**.

```bash
go get github.com/tigusigalpa/birdeye-go
```

Create an API key in Birdeye, then store it outside the repository. On macOS/Linux:

```bash
export BIRDEYE_API_KEY="your-key"
```

On PowerShell:

```powershell
$env:BIRDEYE_API_KEY = "your-key"
```

Use `.env.example` as a reminder of the variable name. Do not commit a populated `.env` file.

## Configure the client

`NewClient` is inexpensive and does not perform network I/O. The API key is sent as `X-API-KEY` for every request. Most Birdeye DeFi endpoints also use `x-chain`; set a default once and override it only where needed.

```go
client := birdeye.NewClient(
    os.Getenv("BIRDEYE_API_KEY"),
    birdeye.WithChain(birdeye.ChainSolana),
    birdeye.WithTimeout(10*time.Second),
    birdeye.WithRetries(2),
)
```

Available options:

| Option | Purpose |
|---|---|
| `WithChain(chain)` | Default `x-chain`, such as `birdeye.ChainSolana` or `birdeye.ChainEthereum`. |
| `WithTimeout(duration)` | Timeout for the default HTTP client. |
| `WithHTTPClient(client)` | Bring your own `*http.Client`, for proxies, tracing, or custom transports. |
| `WithBaseURL(url)` | Point at a local mock server or another compatible Birdeye host. |
| `WithRetries(attempts)` | Set the maximum number of attempts while keeping the default backoff policy. |
| `WithRetryPolicy(policy)` | Fully control retry timing and the `OnRetry` hook. |
| `WithLogger(logger)` | Receive structured request lifecycle logs. Never log secrets in your logger. |

Every request options struct has a `Chain` field. It wins over the client's default:

```go
quote, err := client.Price.GetPrice(ctx, token, price.PriceOptions{
    Chain: birdeye.ChainEthereum,
})
```

The SDK does not validate plan access or supported chains locally. Birdeye's response is authoritative, which means your application receives a normal typed error when an endpoint is not available on its subscription.

## Quick start

Fetch the current USD price of wrapped SOL:

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"

    birdeye "github.com/tigusigalpa/birdeye-go"
    "github.com/tigusigalpa/birdeye-go/price"
)

func main() {
    client := birdeye.NewClient(
        os.Getenv("BIRDEYE_API_KEY"),
        birdeye.WithChain(birdeye.ChainSolana),
    )

    quote, err := client.Price.GetPrice(
        context.Background(),
        "So11111111111111111111111111111111111111112",
        price.PriceOptions{},
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("$%.4f (24h: %.2f%%)\n", quote.Value, quote.PriceChange24h)
}
```

## Common recipes

The snippets in this section are intended to live in a helper that returns an
`error`; replace `return err` with the error-handling approach used by your
application when placing them directly in `main`.

### Ask for liquidity with a price

Optional booleans and numeric filters use pointers so that `false` and `0` can be sent deliberately.

```go
includeLiquidity := true
minimumLiquidity := 25_000.0

quote, err := client.Price.GetPrice(ctx, token, price.PriceOptions{
    IncludeLiquidity: &includeLiquidity,
    CheckLiquidity:   &minimumLiquidity,
})
if err != nil { return err }

fmt.Println("price:", quote.Value, "liquidity:", quote.Liquidity)
```

### Load a watchlist in one request

Use the GET variant for a short list. A returned `nil` map value means that Birdeye has no price data for that address; it is not an SDK error.

```go
prices, err := client.Price.GetMultiPrice(ctx, []string{
    "So11111111111111111111111111111111111111112", // SOL
    "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
}, price.PriceOptions{})
if err != nil { return err }

for address, quote := range prices {
    if quote == nil {
        fmt.Printf("%s: no price data\n", address)
        continue
    }
    fmt.Printf("%s: $%.6f\n", address, quote.Value)
}
```

For longer lists, call `GetMultiPricePOST`; it puts the same comma-separated addresses in a JSON request body and is never automatically retried.

### Draw a 24-hour candle chart

V3 OHLCV supports fine intervals, count mode, padding, and outlier control. Use Unix seconds for `TimeFrom` and `TimeTo`.

```go
now := time.Now()
page, err := client.Price.GetOHLCVv3(ctx, price.TokenOHLCVOptions{
    Address:  token,
    Type:     price.Interval1H,
    TimeFrom: now.Add(-24 * time.Hour).Unix(),
    TimeTo:   now.Unix(),
    Currency: price.CurrencyUSD,
})
if err != nil { return err }

for _, candle := range page.Items {
    fmt.Printf("%s close=$%.4f volume=%.2f\n",
        time.Unix(candle.UnixTime, 0).UTC().Format(time.RFC3339),
        candle.C,
        candle.V,
    )
}
```

For one pool rather than a token-wide aggregation, use `GetOHLCVv3Pair` with `price.PairOHLCVOptions`. For an aggregated base/quote market, use `GetOHLCVBaseQuote`.

### Get an exact historical checkpoint

```go
point, err := client.Price.GetHistoricalPriceByUnixTime(ctx, token, price.HistoricalPriceOptions{
    UnixTime: time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC).Unix(),
})
if err != nil { return err }

fmt.Printf("historical value: $%.6f\n", point.Value)
```

To request a time series instead, use `GetHistoricalPriceSeries`. Birdeye does not publish a stable response-field schema for that endpoint, so it returns `price.RawObject`: a map of field name to `json.RawMessage`. This preserves all data without making invented type guarantees.

```go
series, err := client.Price.GetHistoricalPriceSeries(ctx, price.HistoricalSeriesOptions{
    Address: token, AddressType: "token", Type: price.Interval1H,
    TimeFrom: from.Unix(), TimeTo: to.Unix(),
})
if err != nil { return err }

var items []map[string]any
if err := json.Unmarshal(series["items"], &items); err != nil { return err }
```

### Add rolling volume to a token card

```go
snapshot, err := client.Price.GetPriceVolume(ctx, token, price.PriceVolumeOptions{
    Type: "24h",
})
if err != nil { return err }

// The field schema is deliberately preserved rather than guessed.
for field, raw := range snapshot {
    fmt.Printf("%s: %s\n", field, raw)
}
```

`GetMultiPriceVolume` is the batch POST equivalent. Like all POST requests in this SDK, it is sent once to avoid repeating an unsafe operation.

## Supported API surface

| Area | Methods | Response |
|---|---|---|
| Spot prices | `GetPrice`, `GetMultiPrice`, `GetMultiPricePOST` | Typed `PriceData` |
| Historical prices | `GetHistoricalPriceByUnixTime`, `GetHistoricalPriceSeries` | Typed point / `RawObject` series |
| Candles | `GetOHLCVv3`, `GetOHLCVv3Pair`, `GetOHLCVBaseQuote` | Typed V3 candles / `RawObject` |
| Rolling activity | `GetPriceVolume`, `GetMultiPriceVolume` | `RawObject` |

See [docs/endpoints.md](docs/endpoints.md) for every route, HTTP method, SDK method, and official Birdeye documentation link.

The upstream-deprecated `/defi/ohlcv` and `/defi/ohlcv/pair` routes are intentionally unsupported. Wallets, transactions, holders, token lists, Perps, blockchain data, x402, and WebSockets are not yet mapped as typed services, but can be called through the raw API below.

## Errors and retries

Use `errors.Is` for the action your application should take. Use `errors.As` when you need diagnostic details for logging or support.

```go
_, err := client.Price.GetPrice(ctx, token, price.PriceOptions{})
switch {
case err == nil:
    // success
case errors.Is(err, transport.ErrUnauthorized):
    return fmt.Errorf("check BIRDEYE_API_KEY: %w", err)
case errors.Is(err, transport.ErrForbidden):
    return fmt.Errorf("endpoint is unavailable for this Birdeye plan: %w", err)
case errors.Is(err, transport.ErrRateLimited):
    return fmt.Errorf("slow down and retry later: %w", err)
default:
    var apiErr *transport.BirdeyeError
    if errors.As(err, &apiErr) {
        log.Printf("Birdeye request %s failed: status=%d code=%s message=%q",
            apiErr.RequestID, apiErr.HTTPStatus, apiErr.Code, apiErr.Message)
    }
    return err
}
```

The default retry policy makes up to three attempts for **GET** requests after HTTP `429` or a transient transport failure. Backoff is bounded, jittered, and respects `Retry-After`. The SDK never retries POST, PUT, PATCH, or DELETE requests on its own. Disable retries with `birdeye.NoRetry()`:

```go
client := birdeye.NewClient(apiKey, birdeye.WithRetryPolicy(birdeye.NoRetry()))
```

## Use an endpoint before it is typed

`Client.Do` is the low-level escape hatch. It still adds the API key and chain header, serializes query parameters, decodes Birdeye's `data` envelope, and returns the same typed errors.

```go
var data map[string]any
_, err := client.Do(
    ctx,
    http.MethodGet,
    "/defi/token_overview",
    map[string]string{"address": token},
    "", // use the client default chain
    nil,
    &data,
)
```

Some endpoint families require an extra header. `DoWithHeaders` supports this without allowing the configured API key to be replaced:

```go
_, err := client.DoWithHeaders(
    ctx,
    http.MethodGet,
    "/perps/v1/token/list",
    nil,
    "",
    http.Header{"x-perp": {"true"}},
    nil,
    &data,
)
```

## Examples, testing, and security

All examples are executable and read `BIRDEYE_API_KEY` from the environment:

```bash
go run ./examples/get_price
go run ./examples/get_ohlcv
go run ./examples/multi_price
go run ./examples/error_handling
go run ./examples/token_search
go run ./examples/wallet_pnl # also requires BIRDEYE_WALLET
```

Tests use local `httptest` servers and hand-written responses—no test spends Birdeye Compute Units:

```bash
go test ./...
go test -race ./... # requires CGO
go vet ./...
gofmt -l .
```

Every push to `main` and every pull request runs the same formatting check,
`go vet`, race-enabled test suite, coverage collection, and package/example
build in [GitHub Actions](.github/workflows/ci.yml).

Never commit a real `BIRDEYE_API_KEY`. Keep production keys in your deployment platform's secret store, rotate them if exposed, and avoid printing full upstream response bodies in application logs.

## License

MIT. See [LICENSE](LICENSE).
