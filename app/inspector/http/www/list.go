package www

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	Records               []*embeddingsdb.Record
	Pagination            pagination.Results
	PaginationType        string
	PaginationNextURL     string
	PaginationPreviousURL string
	Models                []string
	Providers             []string
	CurrentModel          string
	CurrentProvider       string
	EnableSearch          bool
	URIs                  *inspector_http.URIs
}

func ListHandler(opts *ListHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("list")

	if t == nil {
		return nil, fmt.Errorf("Failed to load 'list' template")
	}

	ctx := context.Background()
	pg_type, err := opts.Client.PaginationType(ctx)

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

		list_opts := make([]options.Option, 0)
		models_opts := make([]options.Option, 0)
		providers_opts := make([]options.Option, 0)

		model, err := sanitize.GetString(req, "model")

		if err != nil {
			logger.Error("Failed to derive model parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if model != "" {
			list_opts = append(list_opts, options.NewFilterOption("model", model))
			providers_opts = append(providers_opts, options.NewModelOption(model))
		}

		provider, err := sanitize.GetString(req, "provider")

		if err != nil {
			logger.Error("Failed to derive provider parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if provider != "" {
			list_opts = append(list_opts, options.NewFilterOption("provider", provider))
			models_opts = append(models_opts, options.NewProviderOption(provider))
		}

		models, err := opts.Client.Models(ctx, models_opts...)

		if err != nil {
			logger.Error("Failed to retrieve models", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		providers, err := opts.Client.Providers(ctx, providers_opts...)

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

			cursor, err := sanitize.GetString(req, "cursor")

			if err != nil {
				logger.Error("Failed to derive page query parameter", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			if cursor != "" {
				cursor_opts.Pointer(cursor)
			}

			pg_opts = cursor_opts

		default:
			logger.Error("Unsupported pagination type", "type", pg_type)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		records, pg_rsp, err := opts.Client.ListRecords(ctx, pg_opts, list_opts...)

		if err != nil {
			logger.Error("Failed to list records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		var pg_next string
		var pg_prev string

		list_root := opts.URIs.List

		switch pg_type {
		case database.CountablePaginationType:

			prev := pg_rsp.Previous().(int64)
			next := pg_rsp.Next().(int64)

			if prev != 0 {
				str_prev := strconv.FormatInt(prev, 10)
				pg_prev = paginationURL(list_root, "page", str_prev, provider, model)
			}

			if next != 0 {
				str_next := strconv.FormatInt(next, 10)
				pg_next = paginationURL(list_root, "page", str_next, provider, model)
			}

		case database.CursorPaginationType:

			prev := pg_rsp.Previous().(string)
			next := pg_rsp.Next().(string)

			if prev != "" {
				prev = strings.Replace(prev, "before-", "", 1)
				pg_prev = paginationURL(list_root, "cursor", prev, provider, model)
			}

			if next != "" {
				next = strings.Replace(next, "after-", "", 1)
				pg_next = paginationURL(list_root, "cursor", next, provider, model)
			}

		}

		vars := ListHandlerVars{
			Records:               records,
			Pagination:            pg_rsp,
			PaginationType:        pg_type.String(),
			PaginationPreviousURL: pg_prev,
			PaginationNextURL:     pg_next,
			Models:                models,
			CurrentModel:          model,
			CurrentProvider:       provider,
			Providers:             providers,
			EnableSearch:          opts.EnableSearch,
			URIs:                  opts.URIs,
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

func paginationURL(root string, param string, pointer string, provider string, model string) string {

	q := url.Values{}
	q.Set(param, pointer)

	if provider != "" {
		q.Set("provider", provider)
	}

	if model != "" {
		q.Set("model", model)
	}

	u, _ := url.Parse(root)
	u.RawQuery = q.Encode()

	return u.String()
}
