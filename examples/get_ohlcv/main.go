// Example: fetch v3 OHLCV candles for a token.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	birdeye "github.com/tigusigalpa/birdeye-go"
	"github.com/tigusigalpa/birdeye-go/price"
)

func main() {
	apiKey := os.Getenv("BIRDEYE_API_KEY")
	if apiKey == "" {
		log.Fatal("set BIRDEYE_API_KEY in your environment first")
	}

	client := birdeye.NewClient(apiKey, birdeye.WithChain(birdeye.ChainSolana))

	now := time.Now()
	page, err := client.Price.GetOHLCVv3(context.Background(), price.TokenOHLCVOptions{
		Address:  "So11111111111111111111111111111111111111112",
		Type:     price.Interval1H,
		TimeFrom: now.Add(-24 * time.Hour).Unix(),
		TimeTo:   now.Unix(),
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, candle := range page.Items {
		fmt.Printf("%d  o=%.4f h=%.4f l=%.4f c=%.4f v=%.2f\n", candle.UnixTime, candle.O, candle.H, candle.L, candle.C, candle.V)
	}
}
