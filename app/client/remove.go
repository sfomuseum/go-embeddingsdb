package client

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
	"github.com/sfomuseum/go-flags/flagset"
	_ "github.com/sfomuseum/go-flags/multi"
)

func RemoveRecord(ctx context.Context, args []string) {

	var client_uri string
	var provider string
	var depiction_id string
	var model string
	var verbose bool

	// var database_uris multi.MultiString
	// var dimensions multi.MultiInt

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve.")
	fs.StringVar(&depiction_id, "depiction-id", "", "The unique depiction ID associated with the record to retrieve.")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the model associated with the record to retrieve.")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	// fs.Var(&database_uris, "database-uri", "...")
	// fs.Var(&dimensions, "dimensions", "...")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for removing a record from a gRPC EmbeddingsDB \"service\".\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "record")
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

	req := &embeddingsdb.RemoveRecordRequest{
		Provider:    provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	opts := make([]options.Option, 0)

	/*
		for _, d := range dimensions {
			o := options.NewDimensionsOption(d)
			opts = append(opts, o)
		}
	*/

	err = cl.RemoveRecord(ctx, req, opts...)

	if err != nil {
		log.Fatalf("Failed to remove record, %v", err)
	}
}
