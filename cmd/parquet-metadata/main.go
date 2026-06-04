package main

import (
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
	var keyvalue bool

	fs := flagset.NewFlagSet("import")

	fs.BoolVar(&keyvalue, "key-value", false, "Only display key-value metadata.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Print JSON-encoded metadata for a Parquet file to STDOUT\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	args := fs.Args()

	if len(args) != 1 {
		log.Fatal("Invalid arguments.")
	}

	path := args[0]

	r, err := os.Open(path)

	if err != nil {
		log.Fatalf("Failed to open file for reading, %v", err)
	}

	defer r.Close()

	var meta any

	if keyvalue {

		kv, err := parquet.KeyValueMetadata(r)

		if err != nil {
			log.Fatalf("Failed to read key value metadata, %v", err)
		}

		meta = kv

	} else {

		m, err := parquet.Metadata(r)

		if err != nil {
			log.Fatalf("Failed to read metadata, %v", err)
		}

		meta = m
	}

	enc := json.NewEncoder(os.Stdout)
	err = enc.Encode(meta)

	if err != nil {
		log.Fatalf("Failed to encode metadata, %v", err)
	}
}
