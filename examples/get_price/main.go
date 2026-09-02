// Example: fetch a single token's real-time price.
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

	// Wrapped SOL.
	result, err := client.Price.GetPrice(context.Background(), "So11111111111111111111111111111111111111112", price.Options{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SOL price: $%.4f (updated %s)\n", result.Value, result.UpdateHumanTime)
}
