# Endpoint Registry

Every implemented endpoint, its HTTP method, exact Birdeye route, SDK method, and the official documentation URL it was built against. This is a hand-maintained list, not generated — keep it in sync when adding endpoints.

| SDK Method | HTTP Method | Route | Official Docs |
|---|---|---|---|
| `Client.Price.GetPrice` | GET | `/defi/price` | https://docs.birdeye.so/reference/get-defi-price |
| `Client.Price.GetMultiPrice` | GET | `/defi/multi_price` | https://docs.birdeye.so/reference/get-defi-multi_price |
| `Client.Price.GetMultiPricePOST` | POST | `/defi/multi_price` | https://docs.birdeye.so/reference/post-defi-multi_price |
| `Client.Price.GetHistoricalPriceByUnixTime` | GET | `/defi/historical_price_unix` | https://docs.birdeye.so/reference/get-defi-historical_price_unix |
| `Client.Price.GetOHLCVv3` | GET | `/defi/v3/ohlcv` | https://docs.birdeye.so/reference/get-defi-v3-ohlcv |
| `Client.Price.GetOHLCVv3Pair` | GET | `/defi/v3/ohlcv/pair` | https://docs.birdeye.so/reference/get-defi-v3-ohlcv-pair |

## Deferred (not yet implemented)

Every other endpoint family listed in the original brief — legacy OHLCV (`/defi/ohlcv`, `/defi/ohlcv/pair`, `/defi/ohlcv/base_quote`), price-volume, liquidity OHLC/history, token/pair stats (overview, metadata, market data, trade data, liquidity, price stats), token/market lists, transactions, wallet/net-worth/PnL, balance/transfer, holders, alltime trades, blockchain metadata, creation/trending, meme, security, search/utils, smart money, fee data, DEX/protocol metrics, wallet identity, Perps Data API, Blockchain Data API, x402, and every WebSocket subscription — is **not implemented in this checkpoint**. This is an initial Price & OHLCV foundation; see the root README's Status section for what's planned next. No endpoint here was marked complete without a typed method, a mock-based test, and this table row, per the project's completeness policy.
