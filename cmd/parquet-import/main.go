package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
	_ "github.com/sfomuseum/go-flags/multi"
)

func main() {

	var client_uri string
	// var database_uris multi.MultiString

	start := int64(0)
	end := int64(0)

	var verbose bool

	fs := flagset.NewFlagSet("import")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A registered sfomuseum/go-embeddingsdb/client.Client URI.")
	// fs.Var(&database_uris, "database-uri", "...")
	fs.Int64Var(&start, "start", 0, "Starting offset for importing records. If '0' then records will be imported from the first record onwards.")
	fs.Int64Var(&end, "end", 0, "Ending offset for import records. If '0' then records will be imported up to and including the last record.")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Import parquet-encoded embeddingsdb records from one or more files or HTTP(S) URIs and add them to an embeddingsdb instance.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file(N) parquet_file(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	logger := slog.Default()
	ctx := context.Background()

	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new client, %v", err)
	}

	cl_closefunc := func() {

		ctx := context.Background()
		err := cl.Close(ctx)

		if err != nil {
			logger.Error("Failed to close", "error", err)
			log.Fatalf("Failed to close client, %v", err)
		}
	}

	defer cl_closefunc()

	uris := fs.Args()

	count, err := parquet.ImportWithRange(ctx, cl, start, end, uris...)

	if err != nil {
		log.Fatal(err)
	}

	logger.Info("Imported all records", "count", count)
}
