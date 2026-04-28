package main

import (
	"context"
	"log"

	"github.com/sfomuseum/go-embeddingsdb/app/client"
)

func main() {

	ctx := context.Background()
	err := client.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run client, %v", err)
	}
}
