package inspector

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aaronland/go-http/v4/server"
	"github.com/sfomuseum/go-embeddings"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/http/api"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/http/www"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/www/static"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/www/templates/html"
	"github.com/sfomuseum/go-flags/flagset"
)

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	logger := slog.Default()

	db, err := database.NewDatabase(ctx, database_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new database, %w", err)
	}

	defer db.Close(ctx)

	t, err := html.LoadTemplates(ctx)

	if err != nil {
		return fmt.Errorf("Failed to load HTML templates, %w", err)
	}

	mux := http.NewServeMux()

	static_handler := http.FileServerFS(static.FS)

	mux.Handle("/css/", static_handler)
	mux.Handle("/javascript/", static_handler)

	list_opts := &www.ListHandlerOptions{
		Database:      db,
		Templates:     t,
		EnableUploads: enable_uploads,
	}

	list_handler, err := www.ListHandler(list_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new list handler, %w", err)
	}

	mux.Handle("/", list_handler)

	record_opts := &www.RecordHandlerOptions{
		Database:      db,
		Templates:     t,
		MaxResults:    int32(max_results),
		EnableUploads: enable_uploads,
	}

	record_handler, err := www.RecordHandler(record_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new record handler, %w", err)
	}

	mux.Handle("/record/{provider}/{depiction_id}/", record_handler)

	api_embeddings_opts := &api.EmbeddingsHandlerOptions{
		Database: db,
	}

	api_embeddings_handler, err := api.EmbeddingsHandler(api_embeddings_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new API embeddings handler, %w", err)
	}

	mux.Handle("/api/embeddings/{provider}/{depiction_id}/", api_embeddings_handler)

	if enable_uploads {

		emb_cl, err := embeddings.NewEmbedder32(ctx, embeddings_client_uri)

		if err != nil {
			return fmt.Errorf("Failed to create new embeddings client, %w", err)
		}

		upload_opts := &www.UploadHandlerOptions{
			Database:      db,
			Templates:     t,
			EnableUploads: enable_uploads,
		}

		upload_handler, err := www.UploadHandler(upload_opts)

		if err != nil {
			return fmt.Errorf("Failed to create upload handler, %w", err)
		}

		mux.Handle("/upload/", upload_handler)

		api_upload_opts := &api.UploadHandlerOptions{
			Database:         db,
			EmbeddingsClient: emb_cl,
			MaxResults:       int32(max_results),
			MaxUploadSize:    max_upload_size,
		}

		api_upload_handler, err := api.UploadHandler(api_upload_opts)

		if err != nil {
			return fmt.Errorf("Failed to create API upload handler, %w", err)
		}

		mux.Handle("/api/upload/", api_upload_handler)
	}

	s, err := server.NewServer(ctx, server_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new server, %w", err)
	}

	logger.Info("Listen for requests", "address", s.Address())

	err = s.ListenAndServe(ctx, mux)

	if err != nil {
		return fmt.Errorf("Failed to start server, %w", err)
	}

	return nil
}
