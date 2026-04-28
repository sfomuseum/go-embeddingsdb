package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

func SimilarRecordsById(ctx context.Context, args []string) {

	var client_uri string
	var database_uris multi.MultiString
	var provider string
	var depiction_id string
	var model string
	var similar_provider string
	var max_results int
	var max_distance float64

	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.Var(&database_uris, "database-uri", "...")
	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve to establish embeddings to compare.")
	fs.StringVar(&depiction_id, "depiction-id", "", "The unique depiction ID associated with the record to retrieve to establish embeddings to compare.")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the model associated with the record to retrieve to establish embeddings to compare.")
	fs.StringVar(&similar_provider, "similar-provider", "", "The name of the provider to limit similar record queries to. If empty then all the records for the model chosen will be queried.")
	fs.IntVar(&max_results, "max-results", 0, "The maximum number of results to return in a query. This will override defaults established by the server.")
	fs.Float64Var(&max_distance, "max-distance", 0, "The maximum distance allowed when querying records. This will override defaults established by the server.")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving records similar to the embeddings for a specific record stored in a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "similar-by-id")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	cl, err := client.NewClientWithDatabaseURIs(ctx, client_uri, database_uris...)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	req := &embeddingsdb.SimilarRecordsByIdRequest{
		Provider:    provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	opts := make([]options.Option, 0)

	if similar_provider != "" {
		o := options.NewSimilarProviderOption(similar_provider)
		opts = append(opts, o)
	}

	if max_distance > 0 {
		d := float32(max_distance)
		o := options.NewMaxDistanceOption(d)
		opts = append(opts, o)
	}

	if max_results > 0 {
		r := int32(max_results)
		o := options.NewMaxResultsOption(r)
		opts = append(opts, o)
	}

	rsp, err := cl.SimilarRecordsById(ctx, req, opts...)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}
