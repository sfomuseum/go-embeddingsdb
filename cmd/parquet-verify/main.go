package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"

	_ "github.com/duckdb/duckdb-go/v2"

	"github.com/sfomuseum/go-embeddingsdb/crypto"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

func main() {

	var signatures multi.MultiString
	var verbose bool
	var public_key_uri string

	fs := flagset.NewFlagSet("emit")

	fs.Var(&signatures, "signature", "...")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")
	fs.StringVar(&public_key_uri, "public-key-uri", "", "...")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "\n")
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

	db, err := sql.Open("duckdb", "")

	if err != nil {
		log.Fatalf("Failed to open DuckDB, %v", err)
	}

	defer db.Close()

	verifier, err := crypto.LoadVerificationHandler(ctx, public_key_uri)

	if err != nil {
		log.Fatalf("Failed to create verification handler, %v", err)
	}

	sigs := make([]string, len(signatures))

	for i := 0; i < len(signatures); i++ {
		sigs[i] = fmt.Sprintf("'%s'", signatures[i])
	}

	str_sigs := strings.Join(sigs, ",")

	uris := fs.Args()

	for rec, err := range parquet.Iterate(ctx, uris...) {

		if err != nil {
			log.Fatalf("Iterator yield an error, %v", err)
		}

		rec_hash, err := rec.Hash()

		if err != nil {
			log.Fatalf("Failed to hash record, %v", err)
		}

		q := fmt.Sprintf("SELECT record_signature FROM read_parquet(%s) WHERE record_hash = ?", str_sigs)

		row := db.QueryRowContext(ctx, q, rec_hash)

		var record_sig string

		err = row.Scan(&record_sig)

		if err != nil {
			log.Fatalf("Failed to scan signaure, %v", err)
		}

		ok, err := crypto.VerifyRecordSignatureWithVerifier(ctx, verifier, rec, []byte(record_sig))

		if err != nil {
			log.Fatalf("Failed to verify record, %v", err)
		}

		if !ok {
			slog.Error("Fail", "record", rec)
		}
	}

}
