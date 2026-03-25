package api

import (
	"encoding/json"
	"net/http"

	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb/database"
	embeddingsdb_http "github.com/sfomuseum/go-embeddingsdb/http"
)

type SimilarHandlerOptions struct {
	Database database.Database
}

func SimilarHandler(opts *SimilarHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.LoggerWithRequest(req, nil)

		similar, err := embeddingsdb_http.GetSimilarRecordsFromRequest(req, opts.Database)

		if err != nil {
			logger.Error("Failed to get similar records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(similar)

		if err != nil {
			logger.Error("Failed to encode results", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
