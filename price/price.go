package price

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/tigusigalpa/birdeye-go/transport"
)

// Client provides Birdeye's price and OHLCV endpoints.
type Client struct {
	executor *transport.Executor
}

// NewClient wires a price.Client to the shared Executor. Not normally
// called directly; use birdeye.NewClient.
func NewClient(executor *transport.Executor) *Client {
	return &Client{executor: executor}
}

func boolQuery(v *bool) string {
	if v == nil {
		return ""
	}
	return strconv.FormatBool(*v)
}

// GetPrice returns a single token's real-time price. data may be null
// per Birdeye's docs (e.g. an address with no tracked price) — a null
// response decodes as a zero-value *PriceData, not a nil pointer, since
// the envelope's outer "success" is still true in that case.
//
// Docs: https://docs.birdeye.so/reference/get-defi-price
func (c *Client) GetPrice(ctx context.Context, address string, opts Options) (*Data, error) {
	query := map[string]string{"address": address}
	if opts.IncludeLiquidity != nil {
		query["include_liquidity"] = boolQuery(opts.IncludeLiquidity)
	}
	if opts.CheckLiquidity != nil {
		query["check_liquidity"] = strconv.FormatFloat(*opts.CheckLiquidity, 'f', -1, 64)
	}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result Data
	if _, err := c.executor.Do(ctx, http.MethodGet, "/defi/price", query, opts.Chain, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetMultiPrice returns real-time prices for up to 100 tokens via the
// GET variant (addresses passed as a comma-separated query param). The
// returned map may contain nil values for addresses Birdeye has no
// price data for.
//
// Docs: https://docs.birdeye.so/reference/get-defi-multi_price
func (c *Client) GetMultiPrice(ctx context.Context, addresses []string, opts Options) (map[string]*Data, error) {
	query := map[string]string{"list_address": strings.Join(addresses, ",")}
	if opts.IncludeLiquidity != nil {
		query["include_liquidity"] = boolQuery(opts.IncludeLiquidity)
	}
	if opts.CheckLiquidity != nil {
		query["check_liquidity"] = strconv.FormatFloat(*opts.CheckLiquidity, 'f', -1, 64)
	}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result map[string]*Data
	if _, err := c.executor.Do(ctx, http.MethodGet, "/defi/multi_price", query, opts.Chain, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetMultiPricePOST is the POST variant of GetMultiPrice — same
// semantics, but addresses are sent as a comma-separated string inside
// the JSON body (confirmed via documentation research: still a
// delimited string, not a JSON array), useful when the address list is
// too long for a query string. IncludeLiquidity/CheckLiquidity/
// UIAmountMode remain query parameters even on this POST variant, per
// Birdeye's documented request shape.
//
// Docs: https://docs.birdeye.so/reference/post-defi-multi_price
func (c *Client) GetMultiPricePOST(ctx context.Context, addresses []string, opts Options) (map[string]*Data, error) {
	query := map[string]string{}
	if opts.IncludeLiquidity != nil {
		query["include_liquidity"] = boolQuery(opts.IncludeLiquidity)
	}
	if opts.CheckLiquidity != nil {
		query["check_liquidity"] = strconv.FormatFloat(*opts.CheckLiquidity, 'f', -1, 64)
	}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	body := struct {
		ListAddress string `json:"list_address"`
	}{ListAddress: strings.Join(addresses, ",")}
	var result map[string]*Data
	if _, err := c.executor.Do(ctx, http.MethodPost, "/defi/multi_price", query, opts.Chain, body, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetHistoricalPriceByUnixTime returns a token's price at (or nearest
// to) a specific Unix timestamp.
//
// Docs: https://docs.birdeye.so/reference/get-defi-historical_price_unix
func (c *Client) GetHistoricalPriceByUnixTime(ctx context.Context, address string, opts HistoricalPriceOptions) (*HistoricalPrice, error) {
	query := map[string]string{"address": address}
	if opts.UnixTime != 0 {
		query["unixtime"] = strconv.FormatInt(opts.UnixTime, 10)
	}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result HistoricalPrice
	if _, err := c.executor.Do(ctx, http.MethodGet, "/defi/historical_price_unix", query, opts.Chain, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOHLCVv3 returns v3 OHLCV candles for a token (1s-1M granularity;
// supports sub-minute intervals unlike the legacy /defi/ohlcv endpoint,
// which this SDK does not implement).
//
// Docs: https://docs.birdeye.so/reference/get-defi-v3-ohlcv
func (c *Client) GetOHLCVv3(ctx context.Context, opts TokenOHLCVOptions) (*TokenOHLCVPage, error) {
	query := map[string]string{
		"address": opts.Address,
		"type":    opts.Type,
	}
	if opts.TimeFrom != 0 {
		query["time_from"] = strconv.FormatInt(opts.TimeFrom, 10)
	}
	if opts.TimeTo != 0 {
		query["time_to"] = strconv.FormatInt(opts.TimeTo, 10)
	}
	if opts.Currency != "" {
		query["currency"] = opts.Currency
	}
	if opts.Mode != "" {
		query["mode"] = opts.Mode
	}
	if opts.CountLimit != 0 {
		query["count_limit"] = strconv.Itoa(opts.CountLimit)
	}
	if opts.Padding != nil {
		query["padding"] = boolQuery(opts.Padding)
	}
	if opts.Outlier != nil {
		query["outlier"] = boolQuery(opts.Outlier)
	}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result TokenOHLCVPage
	if _, err := c.executor.Do(ctx, http.MethodGet, "/defi/v3/ohlcv", query, opts.Chain, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetOHLCVv3Pair returns v3 OHLCV candles for a trading pair (address
// is the pair address, not a token address).
//
// Docs: https://docs.birdeye.so/reference/get-defi-v3-ohlcv-pair
func (c *Client) GetOHLCVv3Pair(ctx context.Context, opts PairOHLCVOptions) (*PairOHLCVPage, error) {
	query := map[string]string{
		"address": opts.Address,
		"type":    opts.Type,
	}
	if opts.TimeFrom != 0 {
		query["time_from"] = strconv.FormatInt(opts.TimeFrom, 10)
	}
	if opts.TimeTo != 0 {
		query["time_to"] = strconv.FormatInt(opts.TimeTo, 10)
	}
	if opts.Mode != "" {
		query["mode"] = opts.Mode
	}
	if opts.CountLimit != 0 {
		query["count_limit"] = strconv.Itoa(opts.CountLimit)
	}
	if opts.Padding != nil {
		query["padding"] = boolQuery(opts.Padding)
	}
	if opts.Outlier != nil {
		query["outlier"] = boolQuery(opts.Outlier)
	}
	if opts.Inversion != nil {
		query["inversion"] = boolQuery(opts.Inversion)
	}
	var result PairOHLCVPage
	if _, err := c.executor.Do(ctx, http.MethodGet, "/defi/v3/ohlcv/pair", query, opts.Chain, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetHistoricalPriceSeries returns a bounded price series for a token or pair.
// The response object's field schema is not published as stable, so its fields
// are returned verbatim in RawObject.
//
// Docs: https://docs.birdeye.so/reference/get-defi-history_price
func (c *Client) GetHistoricalPriceSeries(ctx context.Context, opts HistoricalSeriesOptions) (RawObject, error) {
	query := map[string]string{"address": opts.Address, "address_type": opts.AddressType, "type": opts.Type, "time_from": strconv.FormatInt(opts.TimeFrom, 10), "time_to": strconv.FormatInt(opts.TimeTo, 10)}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result RawObject
	_, err := c.executor.Do(ctx, http.MethodGet, "/defi/history_price", query, opts.Chain, nil, &result)
	return result, err
}

// GetOHLCVBaseQuote returns aggregated legacy-format candles for a base/quote
// market. The legacy token and pair OHLCV routes are deprecated and therefore
// intentionally have no SDK methods.
//
// Docs: https://docs.birdeye.so/reference/get-defi-ohlcv-base_quote
func (c *Client) GetOHLCVBaseQuote(ctx context.Context, opts BaseQuoteOHLCVOptions) (RawObject, error) {
	query := map[string]string{"base_address": opts.BaseAddress, "quote_address": opts.QuoteAddress, "type": opts.Type, "time_from": strconv.FormatInt(opts.TimeFrom, 10), "time_to": strconv.FormatInt(opts.TimeTo, 10)}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result RawObject
	_, err := c.executor.Do(ctx, http.MethodGet, "/defi/ohlcv/base_quote", query, opts.Chain, nil, &result)
	return result, err
}

// GetPriceVolume returns the current price and rolling volume for one token.
// Its response fields are preserved verbatim because Birdeye does not publish
// a stable field schema for this endpoint.
//
// Docs: https://docs.birdeye.so/reference/get-defi-price_volume-single
func (c *Client) GetPriceVolume(ctx context.Context, address string, opts VolumeOptions) (RawObject, error) {
	query := map[string]string{"address": address}
	if opts.Type != "" {
		query["type"] = opts.Type
	}
	if opts.UIAmountMode != "" {
		query["ui_amount_mode"] = opts.UIAmountMode
	}
	var result RawObject
	_, err := c.executor.Do(ctx, http.MethodGet, "/defi/price_volume/single", query, opts.Chain, nil, &result)
	return result, err
}

// GetMultiPriceVolume returns current price and rolling volume for multiple
// tokens. POST requests are deliberately not retried by the shared transport.
//
// Docs: https://docs.birdeye.so/reference/post-defi-price_volume-multi
func (c *Client) GetMultiPriceVolume(ctx context.Context, request VolumeMultiRequest) (RawObject, error) {
	query := map[string]string{}
	if request.UIAmountMode != "" {
		query["ui_amount_mode"] = request.UIAmountMode
	}
	body := struct {
		ListAddress string `json:"list_address"`
		Type        string `json:"type,omitempty"`
	}{ListAddress: strings.Join(request.Addresses, ","), Type: request.Type}
	var result RawObject
	_, err := c.executor.Do(ctx, http.MethodPost, "/defi/price_volume/multi", query, request.Chain, body, &result)
	return result, err
}
