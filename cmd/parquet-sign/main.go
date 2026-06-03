package main

import (
	"context"
	"log"

	"github.com/sfomuseum/go-embeddingsdb/app/parquet/sign"
)

func main() {

	ctx := context.Background()
	err := sign.Run(ctx)

	if err != nil {
		log.Fatal("Failed to run sign application, %v", err)
	}
}
