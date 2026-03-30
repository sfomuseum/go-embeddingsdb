package inspector

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aaronland/go-http/v4/server"
	"github.com/sfomuseum/go-embeddings"
	inspector_http "github.com/sfomuseum/go-embeddingsdb/app/inspector/http"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/http/api"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/http/www"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/www/static"
	"github.com/sfomuseum/go-embeddingsdb/app/inspector/www/templates/html"
	"github.com/sfomuseum/go-embeddingsdb/client"
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

	cl, err := client.NewClient(ctx, client_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new client, %w", err)
	}

	t, err := html.LoadTemplates(ctx)

	if err != nil {
		return fmt.Errorf("Failed to load HTML templates, %w", err)
	}

	uris, err := inspector_http.DefaultURIs(uri_prefix)

	if err != nil {
		return fmt.Errorf("Failed to derive default URIs, %w", err)
	}

	mux := http.NewServeMux()

	static_handler := http.FileServerFS(static.FS)

	if uri_prefix != "" {
		static_handler = http.StripPrefix(uri_prefix, static_handler)
	}

	mux.Handle(uris.CSS, static_handler)
	mux.Handle(uris.JavaScript, static_handler)

	list_opts := &www.ListHandlerOptions{
		Client:        cl,
		Templates:     t,
		URIs:          uris,
		EnableUploads: enable_uploads,
	}

	list_handler, err := www.ListHandler(list_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new list handler, %w", err)
	}

	logger.Debug("Register list handler", "uri", uris.List)
	mux.Handle(uris.List, list_handler)

	record_opts := &www.RecordHandlerOptions{
		Client:        cl,
		Templates:     t,
		MaxResults:    int32(max_results),
		EnableUploads: enable_uploads,
		URIs:          uris,
	}

	record_handler, err := www.RecordHandler(record_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new record handler, %w", err)
	}

	logger.Debug("Register record handler", "uri", uris.RecordWithVars)
	mux.Handle(uris.RecordWithVars, record_handler)

	api_embeddings_opts := &api.EmbeddingsHandlerOptions{
		Client: cl,
	}

	api_embeddings_handler, err := api.EmbeddingsHandler(api_embeddings_opts)

	if err != nil {
		return fmt.Errorf("Failed to create new API embeddings handler, %w", err)
	}

	logger.Debug("Register API embeddings handler", "uri", uris.APIEmbeddingsWithVars)
	mux.Handle(uris.APIEmbeddingsWithVars, api_embeddings_handler)

	if enable_uploads {

		emb_cl, err := embeddings.NewEmbedder32(ctx, embeddings_client_uri)

		if err != nil {
			return fmt.Errorf("Failed to create new embeddings client, %w", err)
		}

		upload_opts := &www.UploadHandlerOptions{
			Client:        cl,
			Templates:     t,
			EnableUploads: enable_uploads,
			URIs:          uris,
		}

		upload_handler, err := www.UploadHandler(upload_opts)

		if err != nil {
			return fmt.Errorf("Failed to create upload handler, %w", err)
		}

		logger.Debug("Register upload handler", "uri", uris.Upload)
		mux.Handle(uris.Upload, upload_handler)

		api_upload_opts := &api.UploadHandlerOptions{
			Client:           cl,
			EmbeddingsClient: emb_cl,
			MaxResults:       int32(max_results),
			MaxUploadSize:    max_upload_size,
		}

		api_upload_handler, err := api.UploadHandler(api_upload_opts)

		if err != nil {
			return fmt.Errorf("Failed to create API upload handler, %w", err)
		}

		logger.Debug("Register API upload handler", "uri", uris.APIUpload)
		mux.Handle(uris.APIUpload, api_upload_handler)
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
