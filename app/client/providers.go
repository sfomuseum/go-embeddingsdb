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

func Providers(ctx context.Context, args []string) {

	var client_uri string
	var models multi.MultiString
	var verbose bool

	// var dimensions multi.MultiInt
	// var database_uris multi.MultiString

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.Var(&models, "model", "Zero or more models to limit model selection by.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	// fs.Var(&dimensions, "dimensions", "...")
	// fs.Var(&database_uris, "database-uri", "...")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving the unique list of providers stored in a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
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

	/*
		for _, d := range dimensions {
			opts = append(opts, options.NewDimensionsOption(d))
		}
	*/

	for _, m := range models {
		opts = append(opts, options.NewModelOption(m))
	}

	providers, err := cl.Providers(ctx, opts...)

	if err != nil {
		log.Fatalf("Failed to get models, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(providers)
}
