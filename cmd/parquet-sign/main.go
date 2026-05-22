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

	pswd := []byte("foobar")

	k, err := crypto.NewKey("alice", "alice@alice.com", pswd)

	if err != nil {
		log.Fatal(err)
	}

	is_locked, err := k.IsLocked()

	if err != nil {
		log.Fatal(err)
	}

	if is_locked {

		k, err = k.Unlock(pswd)

		if err != nil {
			log.Fatal(err)
		}
	}

	signer, err := crypto.NewSigner(k)

	if err != nil {
		log.Fatal(err)
	}

	for rec, err := range parquet.Iterate(ctx, uris...) {

		if err != nil {
			log.Fatalf("Iterator yield an error, %v", err)
		}

		a, err := crypto.SignRecordWithSigner(ctx, signer, rec)

		if err != nil {
			log.Fatal(err)
		}

		log.Println(string(a))
	}
}
