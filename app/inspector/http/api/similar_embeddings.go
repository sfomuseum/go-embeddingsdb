package api

import (
	"encoding/json/v2"
	"net/http"

	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

type SimilarWithEmbeddingsHandlerOptions struct {
	Client client.Client
}

func SimilarWithEmbeddingsHandler(opts *SimilarWithEmbeddingsHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		logger := slog.LoggerWithRequest(req, nil)
		ctx := req.Context()

		var similar_req *embeddingsdb.SimilarRecordsRequest

		err := json.UnmarshalRead(req.Body, &similar_req)

		if err != nil {
			logger.Error("Failed to parse request", "error", err)
			http.Error(rsp, "Failed to parse request", http.StatusBadRequest)
			return
		}

		// is valid model w/ sync once here...

		custom_opts := make([]options.Option, 0)

		similar, err := opts.Client.SimilarRecords(ctx, similar_req, custom_opts...)

		if err != nil {
			logger.Error("Failed to retrieve similar records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		err = json.MarshalWrite(rsp, similar)

		if err != nil {
			logger.Error("Failed to encode results", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
