package www

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
	embeddingsdb_http "github.com/sfomuseum/go-embeddingsdb/http"
)

type RecordHandlerOptions struct {
	Database  database.Database
	Templates *template.Template
}

type RecordHandlerVars struct {
	Record  *embeddingsdb.Record
	Similar []*embeddingsdb.SimilarRecord
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

 		model := req.PathValue("model")
		
		similar_req := &embeddingsdb.SimilarRecordsRequest{
			Embeddings: record.Embeddings,
			Model:      model,
		}

		similar, err := opts.Database.SimilarRecords(ctx, similar_req)
		
		if err != nil {
			logger.Error("Failed to retrieve similar records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		vars := RecordHandlerVars{
			Record: record,
			Similar: similar,
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
