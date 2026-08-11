// token_search calls the current search endpoint through the raw API. It is a
// runnable bridge until a typed search service is added.
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
	client := birdeye.NewClient(os.Getenv("BIRDEYE_API_KEY"), birdeye.WithChain(birdeye.ChainSolana))
	var data map[string]any
	_, err := client.Do(context.Background(), http.MethodGet, "/defi/v3/search", map[string]string{"keyword": "SOL"}, "", nil, &data)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v\n", data)
}
