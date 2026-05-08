package api

import (
	"encoding/json"
	"net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

type ProvidersHandlerOptions struct {
	Client client.Client
}

func ProvidersHandler(opts *ProvidersHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		custom_opts := make([]options.Option, 0)

		model, err := sanitize.GetString(req, "model")

		if err != nil {
			logger.Error("Failed to derive model parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if model != "" {
			custom_opts = append(custom_opts, options.NewModelOption(model))
		}

		providers, err := opts.Client.Providers(ctx, custom_opts...)

		if err != nil {
			logger.Error("Failed to retrieve providers", "model", model, "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(providers)

		if err != nil {
			logger.Error("Failed to encode providers", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
