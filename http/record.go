package http

import (
	"fmt"
	net_http "net/http"

	"github.com/aaronland/go-http/v4/slog"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

func GetRecordFromRequest(req *net_http.Request, db database.Database) (*embeddingsdb.Record, error) {

	ctx := req.Context()

	provider := req.PathValue("provider")
	logger := slog.LoggerWithRequest(req, nil)
	
	if provider == "" {
		return nil, fmt.Errorf("Missing or invalid provider")
	}

	model := req.PathValue("model")

	if model == "" {
		return nil, fmt.Errorf("Missing or invalid model")
	}

	depiction_id := req.PathValue("depiction_id")

	if depiction_id == "" {
		return nil, fmt.Errorf("Missing or invalid depiction ID")
	}

	logger.Debug("Fetch record", "provider", provider, "model", model, "depiction_id", depiction_id)
	
	record_req := &embeddingsdb.GetRecordRequest{
		Provider:    provider,
		Model:       model,
		DepictionId: depiction_id,
	}

	record, err := db.GetRecord(ctx, record_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to get record, %w", err)
	}

	return record, nil
}
