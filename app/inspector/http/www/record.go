package www

import (
	"fmt"
	"html/template"
	"net/http"
	"slices"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb"
	inspector_http "github.com/sfomuseum/go-embeddingsdb/app/inspector/http"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

type RecordHandlerOptions struct {
	Database      database.Database
	Templates     *template.Template
	MaxResults    int32
	EnableUploads bool
}

type RecordHandlerVars struct {
	Record          *embeddingsdb.Record
	Similar         []*embeddingsdb.SimilarRecord
	Models          []string
	Providers       []string
	SimilarProvider string
	EnableUploads   bool
}

func RecordHandler(opts *RecordHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("record")

	if t == nil {
		return nil, fmt.Errorf("Failed to load 'record' template")
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

		record, err := inspector_http.GetRecordFromRequest(req, opts.Database)

		if err != nil {
			logger.Error("Failed to get database record", "error", err)
			http.Error(rsp, "Not found", http.StatusNotFound)
			return
		}

		model, _ := sanitize.GetString(req, "model")

		if !slices.Contains(models, model) {
			logger.Error("Unsupported model parameter", "model", model, "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		similar_req := &embeddingsdb.SimilarRecordsRequest{
			Embeddings: record.Embeddings,
			Model:      model,
			MaxResults: &opts.MaxResults,
		}

		similar_provider, err := sanitize.GetString(req, "similar-provider")

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

		vars := RecordHandlerVars{
			Record:          record,
			Similar:         similar,
			Models:          models,
			Providers:       providers,
			SimilarProvider: similar_provider,
			EnableUploads:   opts.EnableUploads,
		}

		err = t.Execute(rsp, vars)

		if err != nil {
			logger.Error("Failed to render template", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
