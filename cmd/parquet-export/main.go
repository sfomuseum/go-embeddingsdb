package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var client_uri string
	var output string
	var verbose bool

	fs := flagset.NewFlagSet("export")

	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")

	fs.StringVar(&output, "output", "-", "The path where Parquet-encoded data should be written. If \"-\" then data will be written to STDOUT.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Export embeddingsdb records as Parquet-encoded data.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]", os.Args[0])
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
		log.Fatalf("Failed to create new database, %v", err)
	}

	var wr io.WriteCloser

	switch output {
	case "-":
		wr = os.Stdout
	default:

		w, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Failed to open %s for writing, %v", output, err)
		}

		wr = w
	}

	_, err = parquet.Export(ctx, cl, wr)

	if err != nil {
		log.Fatalf("Failed to create Parquet export, %v", err)
	}

	switch output {
	case "-":
		// pass
	default:
		err = wr.Close()

		if err != nil {
			log.Fatalf("Failed to close %s after writing, %v", output, err)
		}
	}

}
