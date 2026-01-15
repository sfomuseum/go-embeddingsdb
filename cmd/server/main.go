package main

import (
	"context"
	"log"
	
	"github.com/sfomuseum/go-flags/flagset"	
	"github.com/sfomuseum/go-embeddingsdb/server"
)

func main() {

	var server_uri string

	fs := flagset.NewFlagSet("server")

	fs.StringVar(&server_uri, "server-uri", "grpc://localhost:8081", "...")

	flagset.Parse(fs)

	ctx := context.Background()

	svr, err := server.NewEmbeddingsDBServer(ctx, server_uri)

	if err != nil {
		log.Fatalf("Failed to create server, %v", err)
	}

	err = svr.ListenAndServe(ctx)

	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}
}
	
