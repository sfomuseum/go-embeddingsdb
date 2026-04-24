package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var output string
	var verbose bool

	fs := flagset.NewFlagSet("import")

	fs.StringVar(&output, "output", "-", "The path where Parquet-encoded data should be written. If \"-\" then data will be written to STDOUT.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Merge two or more go-embeddingsdb Parquet files in to a new Parquet file.\n")
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

	wr, err := parquet.NewWriter(ctx, output)

	if err != nil {
		log.Fatalf("Failed to create new parquet writer, %v", err)
	}

	uris := fs.Args()

	if len(uris) < 2 {
		log.Fatalf("A minimum of two files to merge are required")
	}

	count, err := parquet.Merge(ctx, wr, uris...)

	if err != nil {
		log.Fatalf("Failed to merge files, %v", err)
	}

	err = wr.Close()

	if err != nil {
		log.Fatalf("Failed to close new file after writing, %v", err)
	}

	logger.Info("Merged records", "count", count)
}
