package www

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aaronland/go-http/v4/slog"
	"github.com/aaronland/go-http/v4/sanitize"	
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
	embeddingsdb_http "github.com/sfomuseum/go-embeddingsdb/http"
)

type RecordHandlerOptions struct {
	Database  database.Database
	Templates *template.Template
	MaxResults int32	
}

type RecordHandlerVars struct {
	Record  *embeddingsdb.Record
	Similar []*embeddingsdb.SimilarRecord
	Models []string
	Providers []string
	SimilarProvider string
}

func RecordHandler(opts *RecordHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("record")

	if t == nil {
		return nil, fmt.Errorf("Failed to load 'record' template")
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		record, err := embeddingsdb_http.GetRecordFromRequest(req, opts.Database)

		if err != nil {
			logger.Error("Failed to get database record", "error", err)
			http.Error(rsp, "Not found", http.StatusNotFound)
			return
		}

		model, _ := sanitize.GetString(req, "model")		
		
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
			similar_req.SimilarProvider = &similar_provider
		}
		
		similar, err := opts.Database.SimilarRecords(ctx, similar_req)
		
		if err != nil {
			logger.Error("Failed to retrieve similar records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

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
		
		vars := RecordHandlerVars{
			Record: record,
			Similar: similar,
			Models: models,
			Providers: providers,
			SimilarProvider: similar_provider,
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
