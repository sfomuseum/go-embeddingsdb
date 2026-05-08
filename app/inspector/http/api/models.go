package api

import (
	"encoding/json"
	"net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

type ModelsHandlerOptions struct {
	Client client.Client
}

func ModelsHandler(opts *ModelsHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		custom_opts := make([]options.Option, 0)

		provider, err := sanitize.GetString(req, "provider")

		if err != nil {
			logger.Error("Failed to derive provider parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if provider != "" {
			custom_opts = append(custom_opts, options.NewProviderOption(provider))
		}

		models, err := opts.Client.Models(ctx, custom_opts...)

		if err != nil {
			logger.Error("Failed to retrieve models", "provider", provider, "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(models)

		if err != nil {
			logger.Error("Failed to encode models", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
