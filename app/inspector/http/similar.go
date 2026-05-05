package http

import (
	"fmt"
	net_http "net/http"

	"github.com/aaronland/go-http/v4/sanitize"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

func GetSimilarRecordsFromRequest(req *net_http.Request, cl client.Client) ([]*embeddingsdb.SimilarRecord, error) {

	ctx := req.Context()

	record, err := GetRecordFromRequest(req, cl)

	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve record, %w", err)
	}

	// We assume that we have validated model above

	model, _ := sanitize.GetString(req, "model")

	max_dist, err := sanitize.PostFloat64(req, "max-distance")

	if err != nil {
		return nil, fmt.Errorf("Failed to derive max distance from query, %w", err)
	}

	similar_req := &embeddingsdb.SimilarRecordsRequest{
		Embeddings: record.Embeddings,
		Model:      model,
		Exclude: []string{
			record.Key(),
		},
	}

	custom_opts := make([]options.Option, 0)

	if max_dist > 0.0 {
		max32 := float32(max_dist)
		custom_opts = append(custom_opts, options.NewMaxDistanceOption(max32))
	}

	similar, err := cl.SimilarRecords(ctx, similar_req, custom_opts...)

	if err != nil {
		return nil, fmt.Errorf("Failed to get similar records, %w", err)
	}

	return similar, nil
}
