package http

import (
	"fmt"
	net_http "net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
)

func GetSimilarRecordsFromRequest(req *net_http.Request, cl client.Client) ([]*embeddingsdb.SimilarRecord, error) {

	ctx := req.Context()

	record, err := GetRecordFromRequest(req, cl)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve record, %w", err)
	}

	model, _ := sanitize.GetString(req, "model")

	similar_req := &embeddingsdb.SimilarRecordsRequest{
		Embeddings: record.Embeddings,
		Model:      model,
	}

	similar, err := cl.SimilarRecords(ctx, similar_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to get similar records, %w", err)
	}

	return similar, nil
}
