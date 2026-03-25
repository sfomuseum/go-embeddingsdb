package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/aaronland/go-http/v4/server"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/http/www"
	"github.com/sfomuseum/go-embeddingsdb/templates/html"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var server_uri string
	var database_uri string
	var verbose bool

	fs := flagset.NewFlagSet("inspect")

	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "...")
	fs.StringVar(&database_uri, "database-uri", "", "...")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	logger := slog.Default()
	ctx := context.Background()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		log.Fatalf("Failed to create new database, %v", err)
	}

	defer db.Close(ctx)

	t, err := html.LoadTemplates(ctx)

	if err != nil {
		log.Fatalf("Failed to load HTML templates, %v", err)
	}

	mux := http.NewServeMux()

	record_opts := &www.RecordHandlerOptions{
		Database:  db,
		Templates: t,
	}

	record_handler, err := www.RecordHandler(record_opts)

	if err != nil {
		log.Fatalf("Failed to create new record handler, %v", err)
	}

	mux.Handle("/record/", record_handler)

	s, err := server.NewServer(ctx, server_uri)

	if err != nil {
		log.Fatalf("Failed to create new server, %v", err)
	}

	logger.Info("Listen for requests", "address", s.Address())

	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}
}
