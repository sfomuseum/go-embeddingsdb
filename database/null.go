package database

import (
	"context"
	"fmt"

	"github.com/sfomuseum/go-embeddingsdb"
)

type NullDatabase struct {
	Database
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "null", NewNullDatabase)

	if err != nil {
		panic(err)
	}
}

func NewNullDatabase(ctx context.Context, uri string) (Database, error) {
	db := &NullDatabase{}
	return db, nil
}

func (db *NullDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *NullDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {
	return nil
}

func (db *NullDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *NullDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *NullDatabase) LastUpdate(ctx context.Context) (int64, error) {
	return 0, nil
}

func (db *NullDatabase) URI() string {
	return "null://"
}

func (db *NullDatabase) Close(ctx context.Context) error {
	return nil
}
