package client

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

func ListRecords(ctx context.Context, args []string) {

	var client_uri string
	var database_uris multi.MultiString
	var start_page int64
	var end_page int64
	var per_page int64
	var verbose bool

	fs := flagset.NewFlagSet("list")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.Var(&database_uris, "database-uri", "...")
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

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	cl, err := client.NewClientWithDatabaseURIs(ctx, client_uri, database_uris...)

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
			log.Fatalf("List records iterator yielded an error, %v", err)
		}

		err := enc.Encode(r)

		if err != nil {
			log.Fatalf("Faied to encode record, %v", err)
		}
	}
}
