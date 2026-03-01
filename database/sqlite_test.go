package database

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/sfomuseum/go-embeddingsdb"
)

func TestSQLiteDatabase(t *testing.T) {

	ctx := context.Background()

	for _, compression := range sqlite_vec_compressions {

		if compression == sqlite_vec_quantize_compression {
			// continue
		}

		if compression == sqlite_vec_matroyshka_compression {
			continue
		}

		db_uri := fmt.Sprintf("sqlite3://?dsn=:memory:&dimensions=8&compression=%s", compression)

		db, err := NewSQLiteDatabase(ctx, db_uri)

		if err != nil {
			t.Fatalf("[%s] Failed to create database, %v", compression, err)
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
				0.0, 0.2344, 0.122873, 0.0007,
				0.987, 0.3244, 0.73, 0.0055664008,
			},
			Created: ts,
			Attributes: map[string]string{
				"hello": "world",
			},
		}

		err = db.AddRecord(ctx, rec)

		if err != nil {
			t.Fatalf("[%s] Failed to add record, %v", compression, err)
		}

		_, err = db.LastUpdate(ctx)

		if err != nil {
			t.Fatalf("[%s] Failed to determine last update value, %v", compression, err)
		}

		req := &embeddingsdb.GetRecordRequest{
			Provider:    "provider",
			Model:       "model",
			DepictionId: "1234",
		}

		get_rec, err := db.GetRecord(ctx, req)

		if err != nil {
			t.Fatalf("[%s] Failed to get record, %v", compression, err)
		}

		if get_rec.Key() != rec.Key() {
			t.Fatalf("[%s] Unexpected record key. Got '%s' but expected '%s'", compression, get_rec.Key(), rec.Key())
		}

		enc := json.NewEncoder(os.Stderr)
		err = enc.Encode(get_rec)

		if err != nil {
			t.Fatalf("[%s] Failed to encode record, %v", compression, err)
		}

		rec2 := &embeddingsdb.Record{
			Provider:    "provider",
			Model:       "model",
			DepictionId: "abc",
			SubjectId:   "def",
			Embeddings: []float32{
				0.00002, 0.3644, 0.52873, 0.0673,
				0.799, 0.3874, 0.8003, 0.264008,
			},
			Created: ts,
			Attributes: map[string]string{
				"foo": "bar",
			},
		}

		err = db.AddRecord(ctx, rec2)

		if err != nil {
			t.Fatalf("[%s] Failed to add record 2, %v", compression, err)
		}

		models, err := db.Models(ctx)

		if err != nil {
			t.Fatalf("[%s] Failed to derive models", compression)
		}

		if len(models) != 1 {
			t.Fatalf("[%s] Unexpected models length %d", compression, len(models))
		}

		providers, err := db.Providers(ctx)

		if err != nil {
			t.Fatalf("[%s] Failed to derive providers", compression)
		}

		if len(providers) != 1 {
			t.Fatalf("[%s] Unexpected providers length %d", compression, len(providers))
		}

		continue

		max_results := int32(10)

		similar_req := &embeddingsdb.SimilarRecordsRequest{
			SimilarProvider: &rec2.Provider,
			Model:           rec2.Model,
			Embeddings:      rec2.Embeddings,
			MaxResults:      &max_results,
		}

		similar_rsp, err := db.SimilarRecords(ctx, similar_req)

		if err != nil {
			t.Fatalf("Failed to determine similar records for rec 2, %v", err)
		}

		fmt.Println(len(similar_rsp))
	}
}
