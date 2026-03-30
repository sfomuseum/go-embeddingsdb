package main

// This code needs to be refactored in to something more manageable.
// Unsure what that is yet.

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	_ "github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
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
	case "models":
		models(args[2:])
	case "list":
		listRecords(args[2:])
	case "providers":
		providers(args[2:])
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
	fmt.Fprintf(os.Stderr, "* list [options]\n")
	fmt.Fprintf(os.Stderr, "* models [options]\n")
	fmt.Fprintf(os.Stderr, "* providers [options]\n")
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

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
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

func listRecords(args []string) {

	var client_uri string
	var start_page int64
	var end_page int64
	var per_page int64
	var verbose bool

	fs := flagset.NewFlagSet("list")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")

	fs.Int64Var(&start_page, "start-page", 1, "The initial page of results to emit.")
	fs.Int64Var(&end_page, "end-page", -1, "The maximum page number of results to emit. If -1 then this flag will be ignored and all the results (remaining after -start-page * -per-page) will be returned.")
	fs.Int64Var(&per_page, "per-page", 10, "The number of records to include in each paginated result set.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Paginated list of all the records in an embeddingsdb database emitted to STDOUT as line-separated JSON.")
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

	enc := json.NewEncoder(os.Stdout)

	list_opts := client.DefaultListRecordsOptions()
	list_opts.PerPage = per_page
	list_opts.StartPage = start_page
	list_opts.EndPage = end_page

	for r, err := range client.ListRecords(ctx, cl, list_opts) {

		if err != nil {
			log.Fatalf("List records iterator yielded an error, %v", "error", err)
		}

		err := enc.Encode(r)

		if err != nil {
			log.Fatalf("Faied to encode record, %v", err)
		}
	}
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

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
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

func models(args []string) {

	var client_uri string
	var providers multi.MultiString

	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.Var(&providers, "provider", "Zero or more providers to limit model selection by.")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving the unique list of models stored in a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "models")
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

	models, err := cl.Models(ctx, providers...)

	if err != nil {
		log.Fatalf("Failed to get models, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(models)
}

func providers(args []string) {

	var client_uri string
	var verbose bool

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving the unique list of providers stored in a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "models")
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

	providers, err := cl.Providers(ctx)

	if err != nil {
		log.Fatalf("Failed to get models, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(providers)
}
