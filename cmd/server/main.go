package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/sfomuseum/go-embeddingsdb/server"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var server_uri string
	var verbose bool

	fs := flagset.NewFlagSet("server")

	fs.StringVar(&server_uri, "server-uri", "grpc://localhost:8081?database-uri=null://", "A registered sfomuseum/go-embeddingsdb/server.EmbeddingsDBServer URI.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

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
