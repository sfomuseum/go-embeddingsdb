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

	fs := flagset.NewFlagSet("emit")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "...\n")
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

	enc := json.NewEncoder(os.Stdout)

	for _, uri := range uris {

		for rec, err := range parquet.Iterate(ctx, uri) {

			if err != nil {
				log.Fatalf("Iterator yield an error, %v", err)
			}

			err = enc.Encode(rec)

			if err != nil {
				log.Fatalf("Failed to marshal record, %v", err)
			}

		}
	}
}
