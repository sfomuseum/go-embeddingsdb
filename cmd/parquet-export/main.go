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
	var verbose bool

	fs := flagset.NewFlagSet("parquet")

	fs.StringVar(&database_uri, "database-uri", "", "...")
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

	wr := os.Stdout

	err = parquet.Export(ctx, db, wr)

	if err != nil {
		log.Fatalf("Failed to create Parquet export, %v", err)
	}
}
