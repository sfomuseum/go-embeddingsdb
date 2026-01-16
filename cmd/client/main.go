package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	args := os.Args

	if len(args) < 2 {
		usage()
	}

	cmd := args[1]

	switch cmd {
	case "-h":
		usage()
	case "record":
		record(args[2:])
	case "similar-by-id":
		similarById(args[2:])
	default:
		slog.Warn("Unsupported command", "command", cmd)
		usage()
	}
}

func usage() {

	fmt.Fprintf(os.Stderr, "Command-line tool for interacting with a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
	fmt.Fprintf(os.Stderr, "Usage:\n\t%s [command] [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Valid commands are:\n")
	fmt.Fprintf(os.Stderr, "* record [options]\n")
	fmt.Fprintf(os.Stderr, "* similar-by-id [options]\n")
	flag.PrintDefaults()

	os.Exit(1)
}

func record(args []string) {

	var client_uri string
	var provider string
	var depiction_id string
	var model string
	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A valid sfomuseum/go-mobileclip.EmbeddingsClient URI.")
	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve.")
	fs.StringVar(&depiction_id, "depiction-id", "", "The unique depiction ID associated with the record to retrieve.")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the model associated with the record to retrieve.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving a record from a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "record")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	ctx := context.Background()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	req := &embeddingsdb.GetRecordRequest{
		Provider:    provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	rsp, err := cl.GetRecord(ctx, req)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}

func similarById(args []string) {

	var client_uri string
	var provider string
	var depiction_id string
	var model string
	var similar_provider string
	var max_results int
	var max_distance float64

	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A valid sfomuseum/go-mobileclip.EmbeddingsClient URI.")
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

	ctx := context.Background()

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	req := &embeddingsdb.SimilarRecordsByIdRequest{
		Provider:    provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	if similar_provider != "" {
		req.SimilarProvider = &similar_provider
	}

	if max_distance > 0 {
		d := float32(max_distance)
		req.MaxDistance = &d
	}

	if max_results > 0 {
		r := int32(max_results)
		req.MaxResults = &r
	}

	rsp, err := cl.SimilarRecordsById(ctx, req)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}
