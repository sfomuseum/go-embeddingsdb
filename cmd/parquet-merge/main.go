package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

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

	logger := slog.Default()
	ctx := context.Background()

	wr, err := parquet.NewWriter(ctx, output)

	if err != nil {
		log.Fatalf("Failed to create new parquet writer, %v", err)
	}

	for _, path := range fs.Args() {

		switch {
		case strings.HasPrefix(path, "http"):

			uri, err := url.Parse(path)

			if err != nil {
				log.Fatalf("Failed to parse %s as URL, %v", path, err)
			}

			logger.Debug("Merge remote data", "url", uri.String())

			_, err = parquet.MergeRemote(ctx, wr, uri)

			if err != nil {
				log.Fatalf("Failed to merge remote data from %s, %v", uri.String(), err)
			}

		default:

			r, err := os.Open(path)

			if err != nil {
				log.Fatalf("Failed to open %s for reading, %v", path, err)
			}

			defer r.Close()

			logger.Debug("Merge data", "path", path)
			_, err = parquet.Merge(ctx, wr, r)

			if err != nil {
				log.Fatalf("Failed to merge Parquet data for %s, %v", path, err)
			}
		}
	}

	err = wr.Close()

	if err != nil {
		log.Fatal(err)
	}
}
