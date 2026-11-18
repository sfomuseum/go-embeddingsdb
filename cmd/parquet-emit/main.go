package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"github.com/sfomuseum/go-parquet"
)

func main() {

	var verbose bool
	var to_strip multi.MultiString

	fs := flagset.NewFlagSet("emit")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")
	fs.Var(&to_strip, "strip", "Zero or more record properties to \"strip\" (remove values for)")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Emit embeddingsdb records in a Parquet as JSON-encoded data.\n")
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

	var wr io.Writer
	wr = os.Stdout

	// Add more writer options here...

	enc := json.NewEncoder(wr)

	for rec, err := range parquet.Iterate[embeddingsdb.Record](ctx, uris...) {

		if err != nil {
			log.Fatalf("Iterator yield an error, %v", err)
		}

		for _, k := range to_strip {

			switch k {
			case "embeddings":
				rec.Embeddings = make([]float32, 0)
			case "model":
				rec.Model = ""
			case "created":
				rec.Created = 0
			default:
				slog.Warn("Unsupported key to strip", "key", k)
			}
		}

		err = enc.Encode(rec)

		if err != nil {
			log.Fatalf("Failed to marshal record, %v", err)
		}

	}
}
