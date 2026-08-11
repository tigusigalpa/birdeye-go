// Example: handle a Birdeye API error gracefully using the typed error
// hierarchy (sentinel errors + *transport.BirdeyeError for detail).
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	birdeye "github.com/tigusigalpa/birdeye-go"
	"github.com/tigusigalpa/birdeye-go/price"
	"github.com/tigusigalpa/birdeye-go/transport"
)

func main() {
	apiKey := os.Getenv("BIRDEYE_API_KEY")
	if apiKey == "" {
		log.Fatal("set BIRDEYE_API_KEY in your environment first")
	}

	client := birdeye.NewClient(apiKey, birdeye.WithChain(birdeye.ChainSolana))

	// A syntactically invalid address should trigger a 4xx from Birdeye.
	_, err := client.Price.GetPrice(context.Background(), "not-a-real-address", price.PriceOptions{})
	if err == nil {
		fmt.Println("unexpectedly succeeded")
		return
	}

	switch {
	case errors.Is(err, transport.ErrRateLimited):
		fmt.Println("rate limited — back off and retry later")
	case errors.Is(err, transport.ErrUnauthorized):
		fmt.Println("check your BIRDEYE_API_KEY")
	case errors.Is(err, transport.ErrForbidden):
		fmt.Println("this endpoint isn't available on your Birdeye plan")
	default:
		var birdeyeErr *transport.BirdeyeError
		if errors.As(err, &birdeyeErr) {
			fmt.Printf("birdeye error: http=%d message=%q\n", birdeyeErr.HTTPStatus, birdeyeErr.Message)
		} else {
			fmt.Printf("request failed: %v\n", err)
		}
	}
}
