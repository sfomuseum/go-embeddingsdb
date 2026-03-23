package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var database_uri string
	var input string
	var verbose bool

	fs := flagset.NewFlagSet("parquet")

	fs.StringVar(&database_uri, "database-uri", "", "...")
	fs.StringVar(&input, "input", "", "...")

	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	defer db.Close(ctx)

	r, err := os.Open(input)

	if err != nil {
		log.Fatalf("Failed to open input, %v", err)
	}

	defer r.Close()

	err = parquet.Import(ctx, db, r)

	if err != nil {
		log.Fatalf("Failed to create Parquet export, %v", err)
	}

}
