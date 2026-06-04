package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var verbose bool

	fs := flagset.NewFlagSet("import")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Gather embeddingsdb statistics from one or more Parquet files and write to STDOUT as JSON-encoded data..\n")
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

	uris := fs.Args()

	stats, err := parquet.GatherStatistics(ctx, uris...)

	if err != nil {
		log.Fatalf("Failed to merge files, %v", err)
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(stats)

	if err != nil {
		log.Fatalf("Failed to encode statistics, %v", err)
	}
}
