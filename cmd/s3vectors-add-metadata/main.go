package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/sfomuseum/go-embeddingsdb/database/s3vectors"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var cfg_uri string
	var model string
	var provider string
	var verbose bool

	fs := flagset.NewFlagSet("metadata")
	fs.StringVar(&cfg_uri, "cfg-uri", "", "...")
	fs.StringVar(&provider, "provider", "", "...")
	fs.StringVar(&model, "model", "", "...")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	cl, err := s3vectors.NewDynamoDBClient(ctx, cfg_uri)

	if err != nil {
		log.Fatal(err)
	}

	err = cl.SetupTables(ctx)

	if err != nil {
		log.Fatal(err)
	}

	err = cl.AddModelProviderMetadata(ctx, model, provider)

	if err != nil {
		log.Fatal(err)
	}
}
