package www

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"

	// goslog "log/slog"

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
	case database.CountablePaginationType, database.CursorPaginationType, database.MultiPaginationType:
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

		per_page := int64(15)

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

		// Next cursor is the standard pagination cursor signal; as in "use this
		// cursor to get the next set of results".

		next_cursor, err := sanitize.GetString(req, "cursor")

		if err != nil {
			logger.Error("Failed to derive cursor query parameter", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Previous cursor is a hack to account for the lack of "backwards" pagination
		// tokens in token-based symptoms. Specifically if the pagination response instance
		// does not return a "previous" value then we assign the value of next_cursor
		// to a ?previous-cursor parameter assigned to the _next_ pagination URL (below)
		// as a way to enable backwards pagination. This will probably not work for all
		// database implementations but it does work for s3vector buckets.

		previous_cursor, err := sanitize.GetString(req, "previous-cursor")

		if err != nil {
			logger.Error("Failed to derive previous cursor query parameter", "error", err)
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

			countable_opts.PerPage(per_page)
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

		case database.CursorPaginationType, database.MultiPaginationType:

			cursor_opts, err := cursor.NewCursorOptions()

			if err != nil {
				logger.Error("Failed to create pagination options", "error", err)
				http.Error(rsp, "Internal server error", http.StatusInternalServerError)
				return
			}

			cursor_opts.PerPage(per_page)

			if next_cursor != "" {
				cursor_opts.Pointer(next_cursor)
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
				pg_prev = paginationURL(list_root, "page", str_prev, provider, model, "")
			}

			if next != 0 {
				str_next := strconv.FormatInt(next, 10)
				pg_next = paginationURL(list_root, "page", str_next, provider, model, "")
			}

		case database.CursorPaginationType, database.MultiPaginationType:

			prev := pg_rsp.Previous().(string)
			next := pg_rsp.Next().(string)

			if prev == "" {
				// See notes above
				prev = previous_cursor
			}

			if prev != "" {
				pg_prev = paginationURL(list_root, "cursor", prev, provider, model, "")
			}

			if next != "" {
				pg_next = paginationURL(list_root, "cursor", next, provider, model, next_cursor)
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

func paginationURL(root string, param string, pointer string, provider string, model string, previous_cursor string) string {

	q := url.Values{}
	q.Set(param, pointer)

	if provider != "" {
		q.Set("provider", provider)
	}

	if model != "" {
		q.Set("model", model)
	}

	// See notes above where next_cursor and previous_cursor
	// query parameters are derived.

	if previous_cursor != "" {
		q.Set("previous-cursor", previous_cursor)
	}

	u, _ := url.Parse(root)
	u.RawQuery = q.Encode()

	return u.String()
}
