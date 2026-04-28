package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

func GetRecord(ctx context.Context, args []string) {

	var client_uri string
	var database_uris multi.MultiString
	var provider string
	var depiction_id string
	var model string
	var dimensions multi.MultiInt
	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.Var(&database_uris, "database-uri", "...")
	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve.")
	fs.StringVar(&depiction_id, "depiction-id", "", "The unique depiction ID associated with the record to retrieve.")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the model associated with the record to retrieve.")
	fs.Var(&dimensions, "dimensions", "...")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving a record from a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "record")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cl, err := client.NewClientWithDatabaseURIs(ctx, client_uri, database_uris...)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	cl_closefunc := func() {

		ctx := context.Background()
		err := cl.Close(ctx)

		if err != nil {
			slog.Error("Failed to close", "error", err)
			log.Fatalf("Failed to close client, %v", err)
		}
	}

	defer cl_closefunc()

	req := &embeddingsdb.GetRecordRequest{
		Provider:    provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	opts := make([]options.Option, 0)

	for _, d := range dimensions {
		o := options.NewDimensionsOption(d)
		opts = append(opts, o)
	}

	rsp, err := cl.GetRecord(ctx, req, opts...)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}
