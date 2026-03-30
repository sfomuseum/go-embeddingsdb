package http

import (
	"fmt"
	"net/url"
)

type URIs struct {
	CSS                   string
	JavaScript            string
	List                  string
	Record                string
	RecordWithVars        string
	Upload                string
	APIEmbeddings         string
	APIEmbeddingsWithVars string
	APIUpload             string
}

func DefaultURIs(prefix string) (*URIs, error) {

	u := &URIs{
		CSS:                   "/css/",
		JavaScript:            "/javascript/",
		List:                  "/",
		RecordWithVars:        "/record/{provider}/{depiction_id}/",
		Record:                "/record/",
		Upload:                "/upload/",
		APIEmbeddings:         "/api/embeddings/",
		APIEmbeddingsWithVars: "/api/embeddings/{provider}/{depiction_id}/",
		APIUpload:             "/api/upload/",
	}

	if prefix != "" {

		fields := map[string]*string{
			"CSS":                   &u.CSS,
			"JavaScript":            &u.JavaScript,
			"List":                  &u.List,
			"Record":                &u.Record,
			"RecordWithVars":        &u.RecordWithVars,
			"Upload":                &u.Upload,
			"APIEmbeddings":         &u.APIEmbeddings,
			"APIEmbeddingsWithVars": &u.APIEmbeddingsWithVars,
			"APIUpload":             &u.APIUpload,
		}

		for name, ptr := range fields {

			new_uri, err := url.JoinPath(prefix, *ptr)

			if err != nil {
				return nil, fmt.Errorf("failed to append prefix to %s (%s): %w", *ptr, name, err)
			}

			*ptr = new_uri
		}
	}

	return u, nil
}
