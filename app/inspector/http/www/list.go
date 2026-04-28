package www

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/aaronland/go-pagination/cursor"
	"github.com/sfomuseum/go-embeddingsdb"
	inspector_http "github.com/sfomuseum/go-embeddingsdb/app/inspector/http"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

type ListHandlerOptions struct {
	Client       client.Client
	Templates    *template.Template
	EnableSearch bool
	URIs         *inspector_http.URIs
}

type ListHandlerVars struct {
	Records         []*embeddingsdb.Record
	Pagination      pagination.Results
	Models          []string
	Providers       []string
	CurrentModel    string
	CurrentProvider string
	EnableSearch    bool
	URIs            *inspector_http.URIs
}

func ListHandler(list_opts *ListHandlerOptions) (http.Handler, error) {

	t := list_opts.Templates.Lookup("list")

	if t == nil {
		return nil, fmt.Errorf("Failed to load 'list' template")
	}

	ctx := context.Background()
	pg_type, err := list_opts.Client.PaginationType(ctx)

	if err != nil {
		return nil, fmt.Errorf("Failed to determine pagination type, %w", err)
	}

	switch pg_type {
	case database.CountablePaginationType, database.CursorPaginationType:
		// ok
	default:
		return nil, fmt.Errorf("Unsupported pagination type, %T", pg_type)
	}

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		models, err := list_opts.Client.Models(ctx)

		if err != nil {
			logger.Error("Failed to retrieve models", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		providers, err := list_opts.Client.Providers(ctx)

		if err != nil {
			logger.Error("Failed to retrieve providers", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		var pg_opts pagination.Options

		switch pg_type {
		case database.CountablePaginationType:

			countable_opts, err := countable.NewCountableOptions()

			if err != nil {
				logger.Error("Failed to create pagination options", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			countable_opts.PerPage(int64(15))
			countable_opts.Pointer(int64(1))

			page, err := sanitize.GetInt64(req, "page")

			if err != nil {
				logger.Error("Failed to derive page query parameter", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			if page != 0 {
				countable_opts.Pointer(page)
			}

			pg_opts = countable_opts

		case database.CursorPaginationType:

			cursor_opts, err := cursor.NewCursorOptions()

			if err != nil {
				logger.Error("Failed to create pagination options", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			cursor_opts.PerPage(int64(15))
			cursor_opts.Pointer(int64(1))

			page, err := sanitize.GetString(req, "page")

			if err != nil {
				logger.Error("Failed to derive page query parameter", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			if page != "" {
				cursor_opts.Pointer(page)
			}

			pg_opts = cursor_opts

		default:
			logger.Error("Unsupported pagination type", "type", pg_type)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		opts := make([]options.Option, 0)

		model, err := sanitize.GetString(req, "model")

		if err != nil {
			logger.Error("Failed to derive model parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if model != "" {
			o := options.NewFilterOption("model", model)
			opts = append(opts, o)
		}

		provider, err := sanitize.GetString(req, "provider")

		if err != nil {
			logger.Error("Failed to derive provider parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if provider != "" {
			o := options.NewFilterOption("provider", provider)
			opts = append(opts, o)
		}

		records, pg_rsp, err := list_opts.Client.ListRecords(ctx, pg_opts, opts...)

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
			EnableSearch:    list_opts.EnableSearch,
			URIs:            list_opts.URIs,
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
