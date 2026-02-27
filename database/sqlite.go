package database

import (
	"context"
	"fmt"
	"database/sql"
	
	_ "github.com/mattn/go-sqlite3"
	
	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/sfomuseum/go-embeddingsdb"
)

type SQLiteDatabase struct {
	Database
	db *sql.DB
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "null", NewSQLiteDatabase)

	if err != nil {
		panic(err)
	}
}

func NewSQLiteDatabase(ctx context.Context, uri string) (Database, error) {

	sqlite_vec.Auto()
	
	db := &SQLiteDatabase{

	}
	
	return db, nil
}

func (db *SQLiteDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *SQLiteDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {
	return nil
}

func (db *SQLiteDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *SQLiteDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *SQLiteDatabase) LastUpdate(ctx context.Context) (int64, error) {
	return 0, nil
}

func (db *SQLiteDatabase) URI() string {
	return "null://"
}

func (db *SQLiteDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {
	models := make([]string, 0)
	return models, nil
}

func (db *SQLiteDatabase) Providers(ctx context.Context) ([]string, error) {
	providers := make([]string, 0)
	return providers, nil
}

func (db *SQLiteDatabase) Close(ctx context.Context) error {
	return nil
}
