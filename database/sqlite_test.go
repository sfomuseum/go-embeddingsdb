package database

import (
	"context"
	"testing"

	"github.com/sfomuseum/go-embeddingsdb"
)

func TestSQLiteDatabase(t *testing.T) {

	ctx := context.Background()

	db_uri := "sqlite3://?dsn=:memory:&dimensions=3"

	db, err := NewSQLiteDatabase(ctx, db_uri)

	if err != nil {
		t.Fatalf("Failed to create database, %v", err)
	}

	defer db.Close(ctx)

	rec := &embeddingsdb.Record{
		Provider:    "provider",
		Model:       "model",
		DepictionId: "1234",
		Embeddings: []float32{
			0.0, 0.2344, 0.122873,
		},
	}

	err = db.AddRecord(ctx, rec)

	if err != nil {
		t.Fatalf("Failed to add record, %v", err)
	}
}
