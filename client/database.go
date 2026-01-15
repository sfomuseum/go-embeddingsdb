package client

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

type DatabaseEmbeddingsDBClient struct {
	EmbeddingsDBClient
	db database.Database
}

func init() {
	ctx := context.Background()
	RegisterEmbeddingsDBClient(ctx, "database", NewDatabaseEmbeddingsDBClient)
}

// DatabaseEmbeddingsDBClient will return a new [DatabaseEmbeddingsDBClient] instance implementing the [EmbeddingsDBClient] interface
// derived from 'uri' which is expected to take the port of:
//
//	database://?{PARAMETERS}

// Where {PARAMETERS} may be one or more of the following:
func NewDatabaseEmbeddingsDBClient(ctx context.Context, uri string) (EmbeddingsDBClient, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	if !q.Has("database-uri") {
		return nil, fmt.Errorf("Missing database URI, %w", err)
	}

	db_uri := q.Get("database-uri")

	db, err := database.NewDatabase(ctx, db_uri)

	if err != nil {
		return nil, err
	}

	cl := &DatabaseEmbeddingsDBClient{
		db: db,
	}

	return cl, nil
}

func (cl *DatabaseEmbeddingsDBClient) AddRecord(ctx context.Context, record *embeddingsdb.Record) error {
	return cl.db.AddRecord(ctx, record)
}

func (cl *DatabaseEmbeddingsDBClient) GetRecord(ctx context.Context, provider string, depiction_id string, model string) (*embeddingsdb.Record, error) {
	return cl.db.GetRecord(ctx, provider, depiction_id, model)
}

func (cl *DatabaseEmbeddingsDBClient) SimilarRecords(ctx context.Context, req *embeddingsdb.SimilarRequest) ([]*embeddingsdb.SimilarResult, error) {
	return cl.db.SimilarRecords(ctx, req)
}

func (cl *DatabaseEmbeddingsDBClient) SimilarRecordsById(ctx context.Context, provider string, depiction_id string, model string) ([]*embeddingsdb.SimilarResult, error) {

	rec, err := cl.GetRecord(ctx, provider, depiction_id, model)

	if err != nil {
		return nil, err
	}

	similar_req := &embeddingsdb.SimilarRequest{
		Model:      rec.Model,
		Embeddings: rec.Embeddings,
		Exclude:    []string{rec.DepictionId},
	}

	return cl.SimilarRecords(ctx, similar_req)
}
