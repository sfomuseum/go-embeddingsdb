package database

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

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

	now := time.Now()
	ts := now.Unix()

	rec := &embeddingsdb.Record{
		Provider:    "provider",
		Model:       "model",
		DepictionId: "1234",
		SubjectId:   "6789",
		Embeddings: []float32{
			0.0, 0.2344, 0.122873,
		},
		Created: ts,
		Attributes: map[string]string{
			"hello": "world",
		},
	}

	err = db.AddRecord(ctx, rec)

	if err != nil {
		t.Fatalf("Failed to add record, %v", err)
	}

	_, err = db.LastUpdate(ctx)

	if err != nil {
		t.Fatalf("Failed to determine last update value, %v", err)
	}

	req := &embeddingsdb.GetRecordRequest{
		Provider:    "provider",
		Model:       "model",
		DepictionId: "1234",
	}

	rec2, err := db.GetRecord(ctx, req)

	if err != nil {
		t.Fatalf("Failed to get record, %v", err)
	}

	if rec2.Key() != rec.Key() {
		t.Fatalf("Unexpected record key. Got '%s' but expected '%s'", rec2.Key(), rec.Key())
	}

	enc := json.NewEncoder(os.Stderr)
	err = enc.Encode(rec2)

	if err != nil {
		t.Fatalf("Failed to encode record, %v", err)
	}
}
