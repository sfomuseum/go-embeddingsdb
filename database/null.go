package database

import (
	"context"
	"fmt"
	"iter"

	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
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

func (db *NullDatabase) Export(ctx context.Context, uri string, opts ...options.Option) error {
	return nil
}

func (db *NullDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) (bool, error) {
	return false, nil
}

func (db *NullDatabase) BatchedRecordsCount(ctx context.Context, opts ...options.Option) (int, error) {
	return 0, nil
}

func (db *NullDatabase) AddBatchedRecord(ctx context.Context) error {
	return nil
}

func (db *NullDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest, opts ...options.Option) (*embeddingsdb.Record, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *NullDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest, opts ...options.Option) error {
	return nil
}

func (db *NullDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest, opts ...options.Option) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *NullDatabase) ListRecords(ctx context.Context, pg_opts pagination.Options, opts ...options.Option) ([]*embeddingsdb.Record, pagination.Results, error) {

	records := make([]*embeddingsdb.Record, 0)

	pg, err := countable.NewResultsFromCountWithOptions(pg_opts, 0)

	if err != nil {
		return nil, nil, err
	}

	return records, pg, nil
}

func (db *NullDatabase) IterateRecords(ctx context.Context, opts ...options.Option) iter.Seq2[*embeddingsdb.Record, error] {
	return func(yield func(*embeddingsdb.Record, error) bool) {}
}

func (db *NullDatabase) LastUpdate(ctx context.Context, opts ...options.Option) (int64, error) {
	return 0, nil
}

func (db *NullDatabase) Dimensions() []int {
	return []int{0}
}

func (db *NullDatabase) URI() string {
	return "null://"
}

func (db *NullDatabase) Models(ctx context.Context, opts ...options.Option) ([]string, error) {
	models := make([]string, 0)
	return models, nil
}

func (db *NullDatabase) Providers(ctx context.Context, opts ...options.Option) ([]string, error) {
	providers := make([]string, 0)
	return providers, nil
}

func (db *NullDatabase) Close(ctx context.Context) error {
	return nil
}
