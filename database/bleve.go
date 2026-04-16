package database

// https://github.com/blevesearch/bleve/blob/master/docs/vectors.md

import (
	"context"
	"fmt"
	"iter"

	_ "github.com/blevesearch/bleve/v2"

	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
)

type BleveDatabase struct {
	Database
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "bleve", NewBleveDatabase)

	if err != nil {
		panic(err)
	}
}

func NewBleveDatabase(ctx context.Context, uri string) (Database, error) {
	db := &BleveDatabase{}
	return db, nil
}

func (db *BleveDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *BleveDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {
	return nil
}

func (db *BleveDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *BleveDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest) error {
	return nil
}

func (db *BleveDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *BleveDatabase) ListRecords(ctx context.Context, opts pagination.Options, filters ...*ListRecordsFilter) ([]*embeddingsdb.Record, pagination.Results, error) {

	records := make([]*embeddingsdb.Record, 0)

	pg, err := countable.NewResultsFromCountWithOptions(opts, 0)

	if err != nil {
		return nil, nil, err
	}

	return records, pg, nil
}

func (db *BleveDatabase) IterateRecords(ctx context.Context) iter.Seq2[*embeddingsdb.Record, error] {
	return func(yield func(*embeddingsdb.Record, error) bool) {}
}

func (db *BleveDatabase) LastUpdate(ctx context.Context) (int64, error) {
	return 0, nil
}

func (db *BleveDatabase) URI() string {
	return "bleve://"
}

func (db *BleveDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {
	models := make([]string, 0)
	return models, nil
}

func (db *BleveDatabase) Providers(ctx context.Context) ([]string, error) {
	providers := make([]string, 0)
	return providers, nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {
	return nil
}
