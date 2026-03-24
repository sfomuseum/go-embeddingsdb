package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var client_uri string
	var verbose bool

	fs := flagset.NewFlagSet("import")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A registered sfomuseum/go-embeddingsdb/client.Client URI.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Import parquet-encoded embeddingsdb records from one or more files and add them to an embeddingsdb instance.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file(N) parquet_file(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

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

	for _, path := range fs.Args() {

		r, err := os.Open(path)

		if err != nil {
			log.Fatalf("Failed to open %s for reading, %v", path, err)
		}

		defer r.Close()

		_, err = parquet.Import(ctx, cl, r)

		if err != nil {
			log.Fatalf("Failed to export Parquet data for %s, %v", path, err)
		}
	}

}
