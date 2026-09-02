// Package price implements Birdeye's price and OHLCV endpoints: single
// and multi-token real-time price, v3 OHLCV candles (token and pair),
// and historical price lookup by Unix timestamp.
//
// Docs: https://docs.birdeye.so/reference/price-ohlcv
package price

import "encoding/json"

// Interval values accepted by the OHLCV "type" parameter.
//
// Docs: https://docs.birdeye.so/reference/get-defi-v3-ohlcv
const (
	Interval1s  = "1s"
	Interval15s = "15s"
	Interval30s = "30s"
	Interval1m  = "1m"
	Interval3m  = "3m"
	Interval5m  = "5m"
	Interval15m = "15m"
	Interval30m = "30m"
	Interval1H  = "1H"
	Interval2H  = "2H"
	Interval4H  = "4H"
	Interval6H  = "6H"
	Interval8H  = "8H"
	Interval12H = "12H"
	Interval1D  = "1D"
	Interval3D  = "3D"
	Interval1W  = "1W"
	Interval1M  = "1M"
)

// UIAmountMode values accepted by the "ui_amount_mode" parameter
// (Solana-only, per Birdeye's docs).
const (
	UIAmountModeRaw    = "raw"
	UIAmountModeScaled = "scaled"
	UIAmountModeBoth   = "both"
)

// OHLCVMode values accepted by the v3 OHLCV "mode" parameter.
const (
	OHLCVModeRange = "range"
	OHLCVModeCount = "count"
)

// Currency values accepted by the v3 token OHLCV "currency" parameter.
const (
	CurrencyUSD    = "usd"
	CurrencyNative = "native"
)

// Data is a single token's real-time price, returned by GetPrice
// (as a single object) and GetMultiPrice/GetMultiPricePOST (as map
// values, which may be nil per Birdeye's docs — an address with no
// price data decodes as a nil map entry, not an error).
//
// Docs: https://docs.birdeye.so/reference/get-defi-price
type Data struct {
	Value               float64 `json:"value"`
	UpdateUnixTime      int64   `json:"updateUnixTime"`
	UpdateHumanTime     string  `json:"updateHumanTime"`
	PriceChange24h      float64 `json:"priceChange24h"`
	PriceInNative       float64 `json:"priceInNative,omitempty"`
	Liquidity           float64 `json:"liquidity,omitempty"`
	IsScaledUIToken     bool    `json:"isScaledUiToken,omitempty"`
	ScaledValue         float64 `json:"scaledValue,omitempty"`
	Multiplier          float64 `json:"multiplier,omitempty"`
	ScaledPriceInNative float64 `json:"scaledPriceInNative,omitempty"`
}

// Options are the shared optional filters for GetPrice,
// GetMultiPrice, and GetMultiPricePOST.
type Options struct {
	// Chain overrides the client's default x-chain header for this call
	// only ("" to use the client default).
	Chain string

	IncludeLiquidity *bool
	CheckLiquidity   *float64
	// UIAmountMode: UIAmountModeRaw (default)/Scaled/Both. Solana-only.
	UIAmountMode string
}

// HistoricalPrice is a token's price at (or nearest to) a specific Unix
// timestamp, returned by GetHistoricalPriceByUnixTime.
//
// Docs: https://docs.birdeye.so/reference/get-defi-historical_price_unix
type HistoricalPrice struct {
	IsScaledUIToken bool    `json:"isScaledUiToken"`
	Value           float64 `json:"value"`
	UpdateUnixTime  int64   `json:"updateUnixTime"`
	PriceChange24h  float64 `json:"priceChange24h"`
	ScaledValue     float64 `json:"scaledValue,omitempty"`
	Multiplier      float64 `json:"multiplier,omitempty"`
}

// HistoricalPriceOptions are the optional filters for
// GetHistoricalPriceByUnixTime.
type HistoricalPriceOptions struct {
	Chain string

	// UnixTime is Unix seconds, 0-10000000000; zero means "let Birdeye
	// pick" (its documented default, not enforced client-side).
	UnixTime     int64
	UIAmountMode string
}

// Candle is one OHLCV bar for a token, returned within
// TokenOHLCVPage.Items by GetOHLCVv3. Field names are verbatim from
// Birdeye's response (o/h/l/c/v, not open/high/low/close/volume) —
// confirmed via documentation research, not assumed. Scaled* fields are
// pointers because they are only present for Solana scaled UI tokens,
// not merely zero for everything else.
//
// Docs: https://docs.birdeye.so/reference/get-defi-v3-ohlcv
type Candle struct {
	O        float64  `json:"o"`
	H        float64  `json:"h"`
	L        float64  `json:"l"`
	C        float64  `json:"c"`
	V        float64  `json:"v"`
	VUsd     float64  `json:"v_usd"`
	UnixTime int64    `json:"unix_time"`
	Address  string   `json:"address"`
	Type     string   `json:"type"`
	Currency string   `json:"currency"`
	ScaledO  *float64 `json:"scaled_o,omitempty"`
	ScaledH  *float64 `json:"scaled_h,omitempty"`
	ScaledL  *float64 `json:"scaled_l,omitempty"`
	ScaledC  *float64 `json:"scaled_c,omitempty"`
	ScaledV  *float64 `json:"scaled_v,omitempty"`
}

// TokenOHLCVPage is the envelope returned by GetOHLCVv3.
type TokenOHLCVPage struct {
	Items           []Candle `json:"items"`
	IsScaledUIToken bool     `json:"is_scaled_ui_token"`
	Multiplier      float64  `json:"multiplier,omitempty"`
}

// TokenOHLCVOptions are the parameters for GetOHLCVv3. Address and Type
// are required; TimeFrom/TimeTo are required in the default "range"
// Mode (Birdeye documents both as required even though CountLimit-based
// "count" mode implies only one endpoint is truly needed — not
// enforced client-side, since this SDK does not second-guess Birdeye's
// own validation).
type TokenOHLCVOptions struct {
	Chain string

	Address string // required
	Type    string // required; one of the Interval* constants

	TimeFrom int64 // Unix seconds
	TimeTo   int64 // Unix seconds

	Currency     string // CurrencyUSD (default) or CurrencyNative
	Mode         string // OHLCVModeRange (default) or OHLCVModeCount
	CountLimit   int    // 0-5000, default 5000
	Padding      *bool  // default false
	Outlier      *bool  // default true
	UIAmountMode string
}

// PairCandle is one OHLCV bar for a trading pair, returned within
// PairOHLCVPage.Items by GetOHLCVv3Pair. No scaled_* fields are
// documented for pair OHLCV (unlike token OHLCV).
//
// Docs: https://docs.birdeye.so/reference/get-defi-v3-ohlcv-pair
type PairCandle struct {
	O        float64 `json:"o"`
	H        float64 `json:"h"`
	L        float64 `json:"l"`
	C        float64 `json:"c"`
	V        float64 `json:"v"`
	VUsd     float64 `json:"v_usd"`
	Address  string  `json:"address"`
	Type     string  `json:"type"`
	UnixTime int64   `json:"unix_time"`
	Currency string  `json:"currency"`
}

// PairOHLCVPage is the envelope returned by GetOHLCVv3Pair.
type PairOHLCVPage struct {
	Items []PairCandle `json:"items"`
}

// PairOHLCVOptions are the parameters for GetOHLCVv3Pair. Address is the
// pair address (not a token address) and Type are required.
type PairOHLCVOptions struct {
	Chain string

	Address string // required; pair address
	Type    string // required; one of the Interval* constants

	TimeFrom int64 // Unix seconds
	TimeTo   int64 // Unix seconds

	Mode       string // OHLCVModeRange (default) or OHLCVModeCount
	CountLimit int    // 0-5000, default 5000
	Padding    *bool  // default false
	Outlier    *bool  // default true
	Inversion  *bool  // default false
}

// HistoricalSeriesOptions selects a bounded historical price series. AddressType
// must be "token" or "pair"; the API performs final validation.
type HistoricalSeriesOptions struct {
	Chain        string
	Address      string
	AddressType  string
	Type         string
	TimeFrom     int64
	TimeTo       int64
	UIAmountMode string
}

// BaseQuoteOHLCVOptions selects aggregated candles for a base/quote market.
type BaseQuoteOHLCVOptions struct {
	Chain, BaseAddress, QuoteAddress, Type string
	TimeFrom, TimeTo                       int64
	UIAmountMode                           string
}

// VolumeOptions selects a rolling price/volume window.
type VolumeOptions struct {
	Chain, Type, UIAmountMode string
}

// VolumeMultiRequest is the documented request body for the batch
// price-volume endpoint. Type defaults to 24h server-side when empty.
type VolumeMultiRequest struct {
	Addresses    []string
	Type         string
	Chain        string
	UIAmountMode string
}

// RawObject preserves a successful data object whose response fields are not
// published as a stable schema in the reference. Decode Fields into a local
// application type when needed; unknown fields are never discarded.
type RawObject map[string]json.RawMessage
