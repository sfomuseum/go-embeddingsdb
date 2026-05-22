package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/crypto"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var key_uri string
	var pswd_uri string

	var verbose bool

	fs := flagset.NewFlagSet("emit")

	fs.StringVar(&key_uri, "key-uri", "", "...")
	fs.StringVar(&pswd_uri, "password-uri", "", "...")

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

	key, err := crypto.LoadKey(ctx, key_uri, pswd_uri)

	if err != nil {
		log.Fatalf("Failed to load key, %v", err)
	}

	signer, err := crypto.NewSigner(key)

	if err != nil {
		log.Fatalf("Failed to create signer, %v", err)
	}

	for rec, err := range parquet.Iterate(ctx, uris...) {

		if err != nil {
			log.Fatalf("Iterator yield an error, %v", err)
		}

		a, err := crypto.SignRecordWithSigner(ctx, signer, rec)

		if err != nil {
			log.Fatalf("Failed to sign record %s, %v", rec.Key(), err)
		}

		log.Println(string(a))
	}
}
