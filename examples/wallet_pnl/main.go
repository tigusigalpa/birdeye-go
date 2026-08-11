// wallet_pnl demonstrates the raw request escape hatch for an endpoint that
// has not yet received a typed service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	birdeye "github.com/tigusigalpa/birdeye-go"
)

func main() {
	wallet := os.Getenv("BIRDEYE_WALLET")
	if wallet == "" {
		log.Fatal("set BIRDEYE_WALLET to run this example")
	}
	client := birdeye.NewClient(os.Getenv("BIRDEYE_API_KEY"), birdeye.WithChain(birdeye.ChainSolana))
	var data map[string]any
	_, err := client.Do(context.Background(), http.MethodGet, "/wallet/v2/pnl", map[string]string{"wallet": wallet}, "", nil, &data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", data)
}
