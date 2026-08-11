# Birdeye Go SDK

An idiomatic Go client for the [Birdeye](https://birdeye.so) crypto market data API (`https://public-api.birdeye.so`), written from scratch against Birdeye's official documentation — not generated code.

**Author:** Igor Sazonov — [sovletig@gmail.com](mailto:sovletig@gmail.com) — [github.com/tigusigalpa](https://github.com/tigusigalpa)

**Package:** a matching PHP/Laravel SDK is available at `tigusigalpa/birdeye-php`.

---

## Status

**This is an early, honest checkpoint, not a finished library.** Only the **Price & OHLCV** family is implemented and tested: single/multi-token real-time price, v3 OHLCV candles (token and pair), and historical price by Unix timestamp. Every other documented family — token/pair stats, token/market lists, transactions, wallet/net-worth/PnL, balance/transfer, holders, Perps Data API, Blockchain Data API, x402, and WebSocket subscriptions — is **not yet implemented**. See [docs/endpoints.md](docs/endpoints.md) for the exact, hand-maintained list of what's covered, with a direct Birdeye documentation link per method.

We'd rather ship a small, correct surface than a large, half-tested one. If you need broader coverage today, use the raw request escape hatch (`Client.Do`) to call any endpoint this SDK hasn't mapped yet.

---

## Why this exists

Building each endpoint by hand, one at a time, against Birdeye's real documentation — with typed requests/responses, tests, and a docs entry — trades coverage speed for correctness: what's here is verified against the docs, not guessed. Where a response field's exact shape couldn't be confirmed, this SDK does not fabricate it.

---

## Install

```bash
go get github.com/tigusigalpa/birdeye-go
```

Requires Go 1.22 or newer.

---

## Authentication and configuration

Read `BIRDEYE_API_KEY` from an environment variable or your own secret store — never hardcode it:

```go
client := birdeye.NewClient(os.Getenv("BIRDEYE_API_KEY"))
```

Birdeye requires the `X-API-KEY` header on every request; `NewClient` sends it automatically. Some endpoint families also require an `x-chain` header (Birdeye defaults to `"solana"` server-side when it's omitted). Set a client-wide default with `WithChain`, or override it per call via each method's `Options.Chain` field:

```go
client := birdeye.NewClient(apiKey, birdeye.WithChain(birdeye.ChainEthereum))

// Override just this one call:
result, err := client.Price.GetPrice(ctx, address, price.PriceOptions{Chain: birdeye.ChainSolana})
```

Data accessibility (which endpoints/chains you can call) is determined by your Birdeye plan (Standard/Lite/Starter/Premium/Business/Enterprise). This SDK does **not** validate plan access client-side — a request Birdeye rejects for your plan returns a normal API error (`transport.ErrForbidden`), same as any other error.

---

## Quick start

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
	client := birdeye.NewClient(os.Getenv("BIRDEYE_API_KEY"), birdeye.WithChain(birdeye.ChainSolana))

	result, err := client.Price.GetPrice(context.Background(), "So11111111111111111111111111111111111111112", price.PriceOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("price:", result.Value)
}
```

Runnable examples: [examples/](examples/) — `get_price`, `get_ohlcv`, `multi_price`, `error_handling`.

---

## Supported services

| Service | Docs |
|---|---|
| `Client.Price` | [Price & OHLCV overview](https://docs.birdeye.so/reference/price-ohlcv) |

Full per-method mapping: [docs/endpoints.md](docs/endpoints.md).

---

## Error handling and retries

Every error can be matched with `errors.Is` against a sentinel (`transport.ErrUnauthorized`, `transport.ErrForbidden`, `transport.ErrRateLimited`, `transport.ErrNotFound`, `transport.ErrServerError`, ...) or unwrapped with `errors.As` into a `*transport.BirdeyeError` for the exact HTTP status, message, and raw response body Birdeye sent — nothing is silently dropped.

```go
var birdeyeErr *transport.BirdeyeError
if errors.As(err, &birdeyeErr) {
	fmt.Println(birdeyeErr.HTTPStatus, birdeyeErr.Message)
}
```

GET requests retry automatically (bounded exponential backoff with full jitter, honoring a `Retry-After` response header when present) via a conservative default `RetryPolicy` (3 attempts, 250ms-5s backoff, 20s max elapsed). POST requests are **never** auto-retried. Override with `WithRetryPolicy`, or disable retries entirely with `birdeye.NoRetry()`.

---

## Raw request escape hatch

Call any Birdeye endpoint — including ones this SDK hasn't mapped to a typed method yet — without waiting for an SDK update:

```go
var result map[string]any
_, err := client.Do(ctx, http.MethodGet, "/defi/token_overview", map[string]string{"address": addr}, "", nil, &result)
```

---

## Testing

```bash
go test ./...     # unit tests (offline, httptest-backed, no API key required)
go vet ./...
gofmt -l .
```

No test or example in this repository consumes Birdeye Compute Units — all tests run against `httptest` mock servers with hand-authored fixture responses.

`go test -race` requires a C compiler (cgo); if your environment doesn't have one, plain `go test ./...` still runs the full suite.

---

## Security notice

This is an unofficial, community-maintained client. Never commit a real `BIRDEYE_API_KEY` — use `.env.example` as a template and keep your actual `.env` out of version control (already gitignored here). This SDK never logs your API key or full response bodies.

---

## Compatibility

Pre-1.0: breaking changes may happen between minor versions while coverage is being built out. Not affiliated with Birdeye.

## License

MIT. See [LICENSE](LICENSE).

## Author

Igor Sazonov — [@tigusigalpa](https://github.com/tigusigalpa) — sovletig@gmail.com

## Links

- [Birdeye API documentation](https://docs.birdeye.so/reference/birdeye-api-getting-started)
- [Repository](https://github.com/tigusigalpa/birdeye-go)
