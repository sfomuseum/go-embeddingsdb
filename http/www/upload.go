package www

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"slices"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	sfom_embeddings "github.com/sfomuseum/go-embeddings"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

type UploadHandlerOptions struct {
	Database         database.Database
	EmbeddingsClient sfom_embeddings.Embedder[float32]
	Templates        *template.Template
	MaxResults       int32
}

type UploadHandlerFormVars struct {
	Models    []string
	Providers []string
}

type UploadHandlerResultsVars struct {
	Similar         []*embeddingsdb.SimilarRecord
	Models          []string
	Providers       []string
	SimilarProvider string
}

func UploadHandler(opts *UploadHandlerOptions) (http.Handler, error) {

	form_t := opts.Templates.Lookup("upload_form")

	if form_t == nil {
		return nil, fmt.Errorf("Failed to load 'upload_form' template")
	}

	results_t := opts.Templates.Lookup("upload_results")

	if results_t == nil {
		return nil, fmt.Errorf("Failed to load 'upload_results' template")
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		models, err := opts.Database.Models(ctx)

		if err != nil {
			logger.Error("Failed to retrieve models", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		providers, err := opts.Database.Providers(ctx)

		if err != nil {
			logger.Error("Failed to retrieve providers", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		switch req.Method {
		case http.MethodGet:

			vars := UploadHandlerFormVars{
				Models:    models,
				Providers: providers,
			}

			err = form_t.Execute(rsp, vars)

			if err != nil {
				logger.Error("Failed to render template", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			return

		case http.MethodPost:

			req.Body = http.MaxBytesReader(rsp, req.Body, 10<<20) // 10 * 1024 * 1024

			err := req.ParseMultipartForm(10 << 20)

			if err != nil {
				logger.Error("Failed to parse form", "error", err)
				http.Error(rsp, "Bad request", http.StatusBadRequest)
				return
			}

			model, err := sanitize.PostString(req, "model")

			if err != nil {
				logger.Error("Failed to derive model from query", "error", err)
				http.Error(rsp, "Bad request", http.StatusBadRequest)
				return
			}

			if !slices.Contains(models, model) {
				logger.Error("Unsupported model parameter", "model", model, "error", err)
				http.Error(rsp, "Bad request", http.StatusBadRequest)
				return
			}

			// Now the file

			r, _, err := req.FormFile("upload")

			if err != nil {
				logger.Error("Failed to read upload", "error", err)
				http.Error(rsp, "Bad request", http.StatusBadRequest)
				return
			}

			defer r.Close()

			im_body, err := io.ReadAll(r)

			if err != nil {
				logger.Error("Failed to read upload body", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			emb_req := &sfom_embeddings.EmbeddingsRequest{
				Body:  im_body,
				Model: model,
			}

			emb_rsp, err := opts.EmbeddingsClient.ImageEmbeddings(ctx, emb_req)

			if err != nil {
				logger.Error("Failed to derive embeddings for upload", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			//

			similar_req := &embeddingsdb.SimilarRecordsRequest{
				Embeddings: emb_rsp.Embeddings(),
				Model:      model,
				MaxResults: &opts.MaxResults,
			}

			similar_provider, err := sanitize.PostString(req, "similar-provider")

			if err != nil {
				logger.Error("Failed to derive similar-provider parameter", "error", err)
				http.Error(rsp, "Bad request", http.StatusBadRequest)
				return
			}

			if similar_provider != "" {

				if !slices.Contains(providers, similar_provider) {
					logger.Error("Unsupported similar-provider parameter", "provider", similar_provider, "error", err)
					http.Error(rsp, "Bad request", http.StatusBadRequest)
					return
				}

				similar_req.SimilarProvider = &similar_provider
			}

			similar, err := opts.Database.SimilarRecords(ctx, similar_req)

			if err != nil {
				logger.Error("Failed to retrieve similar records", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			vars := UploadHandlerResultsVars{
				Similar:         similar,
				Models:          models,
				Providers:       providers,
				SimilarProvider: similar_provider,
			}

			err = results_t.Execute(rsp, vars)

			if err != nil {
				logger.Error("Failed to render template", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			return

		default:
			http.Error(rsp, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
