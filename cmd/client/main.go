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
	case "query-by-id":
		queryById(args[2:])
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
	fmt.Fprintf(os.Stderr, "* query-by-id [options]\n")
	flag.PrintDefaults()

	os.Exit(1)
}

func record(args []string) {

	var client_uri string
	var provider string
	var depiction_id string
	var model string

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A valid sfomuseum/go-mobileclip.EmbeddingsClient URI.")
	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve.")
	fs.StringVar(&depiction_id, "depiction-id", "", "The unique depiction ID associated with the record to retrieve.")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the model associated with the record to retrieve.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving a record from a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "record")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	ctx := context.Background()

	cl, err := embeddingsdb.NewEmbeddingsDBClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	rsp, err := cl.GetRecord(ctx, provider, depiction_id, model)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}

func queryById(args []string) {

	var client_uri string
	var provider string
	var depiction_id string
	var model string
	var similar_provider string

	fs := flagset.NewFlagSet("record")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A valid sfomuseum/go-mobileclip.EmbeddingsClient URI.")
	fs.StringVar(&provider, "provider", "", "The name of the provider associated with the record to retrieve to establish embeddings to compare.")
	fs.StringVar(&depiction_id, "depiction_id", "", "The unique depiction ID associated with the record to retrieve to establish embeddings to compare.")
	fs.StringVar(&model, "model", "apple/mobileclip_s0", "The name of the model associated with the record to retrieve to establish embeddings to compare.")
	fs.StringVar(&similar_provider, "similar_provider", "", "The name of the provider to limit similar record queries to. If empty then all the records for the model chosen will be queried.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Command-line tool for retrieving records similar to the embeddings for a specific record stored in a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n\n", "query-by-id")
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	fs.Parse(args)

	ctx := context.Background()

	cl, err := embeddingsdb.NewEmbeddingsDBClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new embeddings client, %v", err)
	}

	rsp, err := cl.SimilarRecordsById(ctx, provider, depiction_id, model)

	if err != nil {
		log.Fatalf("Failed to get record, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.Encode(rsp)
}
