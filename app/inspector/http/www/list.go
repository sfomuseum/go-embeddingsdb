package www

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	inspector_http "github.com/sfomuseum/go-embeddingsdb/app/inspector/http"
	"github.com/sfomuseum/go-embeddingsdb/client"
)

type ListHandlerOptions struct {
	Client        client.Client
	Templates     *template.Template
	EnableUploads bool
	URIs          *inspector_http.URIs
}

type ListHandlerVars struct {
	Records         []*embeddingsdb.Record
	Pagination      pagination.Results
	Models          []string
	Providers       []string
	CurrentModel    string
	CurrentProvider string
	EnableUploads   bool
	URIs            *inspector_http.URIs
}

func ListHandler(opts *ListHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("list")

	if t == nil {
		return nil, fmt.Errorf("Failed to load 'list' template")
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		models, err := opts.Client.Models(ctx)

		if err != nil {
			logger.Error("Failed to retrieve models", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		providers, err := opts.Client.Providers(ctx)

		if err != nil {
			logger.Error("Failed to retrieve providers", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		pg_opts, err := countable.NewCountableOptions()

		if err != nil {
			logger.Error("Failed to create pagination options", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		pg_opts.Pointer(int64(1))

		page, err := sanitize.GetInt64(req, "page")

		if err != nil {
			logger.Error("Failed to derive page query parameter", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		if page != 0 {
			pg_opts.Pointer(page)
		}

		filters := make([]*client.ListRecordsFilter, 0)

		model, err := sanitize.GetString(req, "model")

		if err != nil {
			logger.Error("Failed to derive model parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if model != "" {

			f := &client.ListRecordsFilter{
				Column: "model",
				Value:  model,
			}

			filters = append(filters, f)
		}

		provider, err := sanitize.GetString(req, "provider")

		if err != nil {
			logger.Error("Failed to derive provider parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if provider != "" {

			f := &client.ListRecordsFilter{
				Column: "provider",
				Value:  provider,
			}

			filters = append(filters, f)
		}

		records, pg_rsp, err := opts.Client.ListRecords(ctx, pg_opts, filters...)

		if err != nil {
			logger.Error("Failed to list records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		vars := ListHandlerVars{
			Records:         records,
			Pagination:      pg_rsp,
			Models:          models,
			CurrentModel:    model,
			CurrentProvider: provider,
			Providers:       providers,
			EnableUploads:   opts.EnableUploads,
			URIs:            opts.URIs,
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
