package www

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aaronland/go-http/v4/slog"
	inspector_http "github.com/sfomuseum/go-embeddingsdb/app/inspector/http"
	"github.com/sfomuseum/go-embeddingsdb/client"
)

type UploadHandlerOptions struct {
	Client        client.Client
	Templates     *template.Template
	EnableUploads bool
	URIs          *inspector_http.URIs
}

type UploadHandlerFormVars struct {
	Models        []string
	Providers     []string
	EnableUploads bool
	URIs          *inspector_http.URIs
}

func UploadHandler(opts *UploadHandlerOptions) (http.Handler, error) {

	t := opts.Templates.Lookup("upload")

	if t == nil {
		return nil, fmt.Errorf("Failed to load 'upload_form' template")
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

		vars := UploadHandlerFormVars{
			Models:        models,
			Providers:     providers,
			EnableUploads: opts.EnableUploads,
			URIs:          opts.URIs,
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
