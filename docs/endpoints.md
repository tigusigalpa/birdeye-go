# Endpoint registry

This is the source-of-truth list for the typed SDK surface. Every row has an offline `httptest` unit test. `RawObject` means Birdeye documents the route and request shape but does not publish a stable response-field schema; the data object is retained as `json.RawMessage` values rather than guessed.

| SDK method | HTTP | API route | Response | Official documentation |
|---|---:|---|---|---|
| `Client.Price.GetPrice` | GET | `/defi/price` | `*Data` | [Price](https://docs.birdeye.so/reference/get-defi-price) |
| `Client.Price.GetMultiPrice` | GET | `/defi/multi_price` | `map[string]*Data` | [Multi price (GET)](https://docs.birdeye.so/reference/get-defi-multi_price) |
| `Client.Price.GetMultiPricePOST` | POST | `/defi/multi_price` | `map[string]*Data` | [Multi price (POST)](https://docs.birdeye.so/reference/post-defi-multi_price) |
| `Client.Price.GetHistoricalPriceByUnixTime` | GET | `/defi/historical_price_unix` | `*HistoricalPrice` | [Historical price by Unix time](https://docs.birdeye.so/reference/get-defi-historical_price_unix) |
| `Client.Price.GetHistoricalPriceSeries` | GET | `/defi/history_price` | `RawObject` | [Historical price](https://docs.birdeye.so/reference/get-defi-history_price) |
| `Client.Price.GetOHLCVv3` | GET | `/defi/v3/ohlcv` | `*TokenOHLCVPage` | [OHLCV V3](https://docs.birdeye.so/reference/get-defi-v3-ohlcv) |
| `Client.Price.GetOHLCVv3Pair` | GET | `/defi/v3/ohlcv/pair` | `*PairOHLCVPage` | [OHLCV V3 pair](https://docs.birdeye.so/reference/get-defi-v3-ohlcv-pair) |
| `Client.Price.GetOHLCVBaseQuote` | GET | `/defi/ohlcv/base_quote` | `RawObject` | [OHLCV base/quote](https://docs.birdeye.so/reference/get-defi-ohlcv-base_quote) |
| `Client.Price.GetPriceVolume` | GET | `/defi/price_volume/single` | `RawObject` | [Price volume](https://docs.birdeye.so/reference/get-defi-price_volume-single) |
| `Client.Price.GetMultiPriceVolume` | POST | `/defi/price_volume/multi` | `RawObject` | [Multi price volume](https://docs.birdeye.so/reference/post-defi-price_volume-multi) |

## Not mapped yet

The SDK currently covers price and OHLCV only. Token/market data, transactions, wallets, balances, holders, identity, Perps, blockchain data, x402, and WebSocket subscriptions remain available through `Client.Do` / `Client.DoWithHeaders` while their request and response contracts are added with tests.

The old `/defi/ohlcv` and `/defi/ohlcv/pair` routes are marked **deprecated** by Birdeye and are intentionally excluded. No beta endpoint, WebSocket subscription, or x402 payment flow is implemented.
