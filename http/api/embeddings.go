package api

import (
	"encoding/json"
	"net/http"

	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb/database"
	embeddingsdb_http "github.com/sfomuseum/go-embeddingsdb/http"
)

type EmbeddingsHandlerOptions struct {
	Database database.Database
}

func EmbeddingsHandler(opts *EmbeddingsHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.LoggerWithRequest(req, nil)

		record, err := embeddingsdb_http.GetRecordFromRequest(req, opts.Database)

		if err != nil {
			logger.Error("Failed to get database record", "error", err)
			http.Error(rsp, "Not found", http.StatusNotFound)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(record.Embeddings)

		if err != nil {
			logger.Error("Failed to encode record", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
