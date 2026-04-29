package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

func Models(ctx context.Context, args []string) {

	var client_uri string
	var providers multi.MultiString

	// var database_uris multi.MultiString
	// var dimensions multi.MultiInt

	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.Var(&providers, "provider", "Zero or more providers to limit model selection by.")

	// fs.Var(&database_uris, "database-uri", "...")
	// fs.Var(&dimensions, "dimensions", "...")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving the unique list of models stored in a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "models")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	opts := make([]options.Option, 0)

	for _, p := range providers {
		opts = append(opts, options.NewProviderOption(p))
	}

	/*
		for _, d := range dimensions {
			opts = append(opts, options.NewDimensionsOption(d))
		}
	*/

	models, err := cl.Models(ctx, opts...)

	if err != nil {
		log.Fatalf("Failed to get models, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(models)
}
