package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/aaronland/go-http/v4/server"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/http/api"
	"github.com/sfomuseum/go-embeddingsdb/http/www"
	"github.com/sfomuseum/go-embeddingsdb/www/static"
	"github.com/sfomuseum/go-embeddingsdb/www/templates/html"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var server_uri string
	var database_uri string

	var max_results int

	var verbose bool

	fs := flagset.NewFlagSet("inspect")

	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "...")
	fs.StringVar(&database_uri, "database-uri", "", "...")

	fs.IntVar(&max_results, "max-results", 20, "...")
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

	static_handler := http.FileServerFS(static.FS)

	mux.Handle("/css/", static_handler)
	mux.Handle("/javascript/", static_handler)

	record_opts := &www.RecordHandlerOptions{
		Database:   db,
		Templates:  t,
		MaxResults: int32(max_results),
	}

	record_handler, err := www.RecordHandler(record_opts)

	if err != nil {
		log.Fatalf("Failed to create new record handler, %v", err)
	}

	mux.Handle("/record/{provider}/{depiction_id}/", record_handler)

	api_embeddings_opts := &api.EmbeddingsHandlerOptions{
		Database: db,
	}

	api_embeddings_handler, err := api.EmbeddingsHandler(api_embeddings_opts)

	if err != nil {
		log.Fatalf("Failed to create new API embeddings handler, %v", err)
	}

	mux.Handle("/api/embeddings/{provider}/{depiction_id}/", api_embeddings_handler)

	list_opts := &www.ListHandlerOptions{
		Database:   db,
		Templates:  t,
	}

	list_handler, err := www.ListHandler(list_opts)

	if err != nil {
		log.Fatalf("Failed to create new list handler, %v", err)
	}

	mux.Handle("/", list_handler)
	
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
