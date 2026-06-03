package main

/*

> go run -race cmd/parquet-verify/main.go -public-key-uri file:///usr/local/sfomuseum/go-embeddingsdb/work/sfomuseum-debug.pub -signature sig-test.parquet /usr/local/data/embeddings/sfomuseum-collection-1152-siglip2-naflex-22060423.parquet

*/

import (
	"context"
	"log"

	"github.com/sfomuseum/go-embeddingsdb/app/parquet/verify"
)

func main() {

	ctx := context.Background()
	err := verify.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run verify tool, %v", err)
	}
}
