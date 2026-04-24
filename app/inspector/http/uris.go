package http

import (
	"net/url"
	"strings"
)

type URIs struct {
	CSS                   string
	JavaScript            string
	List                  string
	Record                string
	RecordWithVars        string
	Search                string
	APIEmbeddings         string
	APIEmbeddingsWithVars string
	APISearch             string
}

func DefaultURIs(prefix string) (*URIs, error) {

	u := &URIs{
		CSS:                   "/css/",
		JavaScript:            "/javascript/",
		List:                  "/",
		RecordWithVars:        "/record/{provider}/{depiction_id}/",
		Record:                "/record/",
		Search:                "/search/",
		APIEmbeddings:         "/api/embeddings/",
		APIEmbeddingsWithVars: "/api/embeddings/{provider}/{depiction_id}/",
		APISearch:             "/api/search/",
	}

	if prefix != "" {

		// The following hoop-jumping is necessary because
		// url.JoinPath will happily (and correctly) escape the
		// {foo} wilcard variables in URIs...

		prefix = strings.TrimLeft(prefix, "/")
		prefix, err := url.JoinPath("/", prefix)

		if err != nil {
			return nil, err
		}

		prefix = strings.TrimRight(prefix, "/")

		fields := []*string{
			&u.CSS,
			&u.JavaScript,
			&u.List,
			&u.Record,
			&u.RecordWithVars,
			&u.Search,
			&u.APIEmbeddings,
			&u.APIEmbeddingsWithVars,
			&u.APISearch,
		}

		for _, ptr := range fields {
			*ptr = prefix + *ptr
		}
	}

	return u, nil
}
