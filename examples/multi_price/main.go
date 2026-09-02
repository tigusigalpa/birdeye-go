// Example: fetch prices for multiple tokens in one call.
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
	apiKey := os.Getenv("BIRDEYE_API_KEY")
	if apiKey == "" {
		log.Fatal("set BIRDEYE_API_KEY in your environment first")
	}

	client := birdeye.NewClient(apiKey, birdeye.WithChain(birdeye.ChainSolana))

	addresses := []string{
		"So11111111111111111111111111111111111111112",  // wrapped SOL
		"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", // USDC
	}
	prices, err := client.Price.GetMultiPrice(context.Background(), addresses, price.Options{})
	if err != nil {
		log.Fatal(err)
	}
	for addr, p := range prices {
		if p == nil {
			fmt.Printf("%s: no price data\n", addr)
			continue
		}
		fmt.Printf("%s: $%.4f\n", addr, p.Value)
	}
}
