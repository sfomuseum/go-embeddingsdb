package http

import (
	"fmt"
	net_http "net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

func GetSimilarRecordsFromRequest(req *net_http.Request, db database.Database) ([]*embeddingsdb.SimilarRecord, error) {

	ctx := req.Context()

	record, err := GetRecordFromRequest(req, db)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve record, %w", err)
	}

	model, _ := sanitize.GetString(req, "model")

	similar_req := &embeddingsdb.SimilarRecordsRequest{
		Embeddings: record.Embeddings,
		Model:      model,
	}

	similar, err := db.SimilarRecords(ctx, similar_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to get similar records, %w", err)
	}

	return similar, nil
}
