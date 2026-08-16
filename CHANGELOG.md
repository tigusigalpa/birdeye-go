# Changelog

All notable changes to this project are documented in this file.

## [v1.0.4] - 2026-08-11

### Added

- Price & OHLCV methods for historical price series, base/quote OHLCV, and single and batch price-volume snapshots.
- `Client.DoWithHeaders` for raw requests that require endpoint-specific headers such as `x-perp`; it cannot override the configured API key.
- `WithRetries` for changing the maximum attempt count while retaining the SDK's default backoff settings.
- Request ID and Birdeye application error code capture on API errors.
- Runnable raw-request examples for token search and wallet PnL.
- Mock-based coverage for the new price methods, endpoint-specific headers, error metadata, and retry boundaries.

### Changed

- Reworked the README with a clearer quick start, configuration guidance, endpoint status, raw API usage, retry behaviour, and security advice.
- Expanded the endpoint registry to document all ten currently typed Price & OHLCV routes.
- GET retries now occur only for rate limiting (`429`) and transient transport failures. Server (`5xx`) responses are returned immediately.
- Base URLs with a trailing slash are now normalized before a request path is joined.

### Deprecated / Not included

- The upstream-deprecated legacy `/defi/ohlcv` and `/defi/ohlcv/pair` routes remain intentionally unsupported.
- Wallet, transactions, holders, Perps, blockchain data, x402, and WebSocket APIs are not yet exposed as typed services; use `Client.Do` or `Client.DoWithHeaders` in the meantime.

### Verification

- `go test ./...` passed.
- `go vet ./...` passed.
- `gofmt -l .` produced no output.
- `go test -race ./...` requires CGO, which is unavailable in the verification environment.
