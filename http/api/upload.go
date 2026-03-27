package api

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddings"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

type UploadHandlerOptions struct {
	Database         database.Database
	EmbeddingsClient embeddings.Embedder[float32]
	MaxUploadSize    int64
	MaxResults       int32
}

func UploadHandler(opts *UploadHandlerOptions) (http.Handler, error) {

	fn := func(rsp http.ResponseWriter, req *http.Request) {

		ctx := req.Context()
		logger := slog.LoggerWithRequest(req, nil)

		if req.Method != http.MethodPost {
			http.Error(rsp, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		models, err := opts.Database.Models(ctx)

		if err != nil {
			logger.Error("Failed to retrieve models", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		providers, err := opts.Database.Providers(ctx)

		if err != nil {
			logger.Error("Failed to retrieve providers", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		//

		req.Body = http.MaxBytesReader(rsp, req.Body, opts.MaxUploadSize)

		err = req.ParseMultipartForm(opts.MaxUploadSize)

		if err != nil {
			logger.Error("Failed to parse form", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		model, err := sanitize.PostString(req, "model")

		if err != nil {
			logger.Error("Failed to derive model from query", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if !slices.Contains(models, model) {
			logger.Error("Unsupported model parameter", "model", model, "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		// Hack...
		model = strings.Replace(model, "apple/mobileclip_", "", 1)

		// Now the file

		r, _, err := req.FormFile("upload")

		if err != nil {
			logger.Error("Failed to read upload", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		defer r.Close()

		im_body, err := io.ReadAll(r)

		if err != nil {
			logger.Error("Failed to read upload body", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		im_r := bytes.NewReader(im_body)

		_, _, err = image.Decode(im_r)

		if err != nil {
			logger.Error("Failed to parse upload as image", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		emb_req := &embeddings.EmbeddingsRequest{
			Body:  im_body,
			Model: model,
		}

		emb_rsp, err := opts.EmbeddingsClient.ImageEmbeddings(ctx, emb_req)

		if err != nil {
			logger.Error("Failed to derive embeddings for upload", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		//

		similar_req := &embeddingsdb.SimilarRecordsRequest{
			Embeddings: emb_rsp.Embeddings(),
			Model:      model,
			MaxResults: &opts.MaxResults,
		}

		similar_provider, err := sanitize.PostString(req, "similar-provider")

		if err != nil {
			logger.Error("Failed to derive similar-provider parameter", "error", err)
			http.Error(rsp, "Bad request", http.StatusBadRequest)
			return
		}

		if similar_provider != "" {

			if !slices.Contains(providers, similar_provider) {
				logger.Error("Unsupported similar-provider parameter", "provider", similar_provider, "error", err)
				http.Error(rsp, "Bad request", http.StatusBadRequest)
				return
			}

			similar_req.SimilarProvider = &similar_provider
		}

		similar, err := opts.Database.SimilarRecords(ctx, similar_req)

		if err != nil {
			logger.Error("Failed to retrieve similar records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		logger.Debug("Similar results", "count", len(similar))

		rsp.Header().Set("Content-type", "application/json")

		enc := json.NewEncoder(rsp)
		err = enc.Encode(similar)

		if err != nil {
			logger.Error("Failed to encode similar records", "error", err)
			http.Error(rsp, "Internal server error", http.StatusInternalServerError)
			return
		}

		return
	}

	return http.HandlerFunc(fn), nil
}
