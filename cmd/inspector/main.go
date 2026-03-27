package main

import (
	"context"
	"log"

	"github.com/sfomuseum/go-embeddingsdb/app/inspector"
)

func main() {

	ctx := context.Background()
	err := inspector.Run(ctx)

	if err != nil {
		log.Fatalf("Failed to run inspector application, %v", err)
	}
}
