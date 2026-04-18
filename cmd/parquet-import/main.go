package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		log.Fatalf("Failed to create new client, %v", err)
	}

	cl_closefunc := func() {

		ctx := context.Background()
		err := cl.Close(ctx)

		if err != nil {
			slog.Error("Failed to close", "error", err)
			log.Fatalf("Failed to close client, %v", err)
		}
	}

	defer cl_closefunc()

	for _, path := range fs.Args() {

		switch {
		case strings.HasPrefix(path, "http"):

			uri, err := url.Parse(path)

			if err != nil {
				log.Fatalf("Failed to parse %s as URL, %v", path, err)
			}

			logger.Debug("Import remote data", "url", uri.String())

			_, err = parquet.ImportRemote(ctx, cl, uri)

			if err != nil {
				log.Fatalf("Failed to import remote data from %s, %v", uri.String(), err)
			}

		default:

			r, err := os.Open(path)

			if err != nil {
				log.Fatalf("Failed to open %s for reading, %v", path, err)
			}

			defer r.Close()

			logger.Debug("Import data", "path", path)
			_, err = parquet.Import(ctx, cl, r)

			if err != nil {
				log.Fatalf("Failed to import Parquet data for %s, %v", path, err)
			}
		}
	}

}
