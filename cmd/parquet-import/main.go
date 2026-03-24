package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var client_uri string
	var input string
	var verbose bool

	fs := flagset.NewFlagSet("parquet")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.StringVar(&input, "input", "", "...")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new client, %v", err)
	}

	r, err := os.Open(input)

	if err != nil {
		log.Fatalf("Failed to open input, %v", err)
	}

	defer r.Close()

	err = parquet.Import(ctx, cl, r)

	if err != nil {
		log.Fatalf("Failed to create Parquet export, %v", err)
	}

}
