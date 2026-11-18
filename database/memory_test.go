package database

import (
	"context"
	"testing"

	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

func TestMemoryDatabase(t *testing.T) {

	ctx := context.Background()

	db, err := NewDatabase(ctx, "memory://?dimensions=2&max-distance=5")

	if err != nil {
		t.Fatal(err)
	}

	defer db.Close(ctx)

	vectors := map[string][]float32{
		"a": {0, 0},
		"b": {1, 0},
		"c": {3, 0},
		"d": {10, 0},
	}

	for id, v := range vectors {

		_, err := db.AddRecord(ctx, &embeddingsdb.Record{
			Provider:    "test",
			DepictionId: id,
			SubjectId:   "s-" + id,
			Model:       "m",
			Embeddings:  v,
		})

		if err != nil {
			t.Fatal(err)
		}
	}

	_, err = db.AddRecord(ctx, &embeddingsdb.Record{Provider: "test", DepictionId: "x", Model: "m", Embeddings: []float32{1}})

	if err == nil {
		t.Fatal("Expected dimension error")
	}

	count, _ := db.CountRecords(ctx)

	if count != 4 {
		t.Fatalf("Unexpected count %d", count)
	}

	// "a" (distance 0) and "d" (beyond max distance) are omitted.
	similar, err := db.SimilarRecords(ctx, &embeddingsdb.SimilarRecordsRequest{
		Model:      "m",
		Embeddings: []float32{0, 0},
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(similar) != 2 || similar[0].DepictionId != "b" || similar[1].DepictionId != "c" {
		t.Fatalf("Unexpected similar results: %v", similar)
	}

	similar, err = db.SimilarRecords(ctx, &embeddingsdb.SimilarRecordsRequest{
		Model:      "m",
		Embeddings: []float32{0, 0},
		Exclude:    []string{"b"},
	}, options.NewMaxResultsOption(1))

	if err != nil {
		t.Fatal(err)
	}

	if len(similar) != 1 || similar[0].DepictionId != "c" {
		t.Fatalf("Unexpected excluded results: %v", similar)
	}

	pg_opts, _ := countable.NewCountableOptions()
	pg_opts.PerPage(3)
	pg_opts.Pointer(int64(2))

	records, pg, err := db.ListRecords(ctx, pg_opts)

	if err != nil {
		t.Fatal(err)
	}

	if len(records) != 1 || records[0].DepictionId != "d" || pg.Total() != 4 {
		t.Fatalf("Unexpected page: %d records", len(records))
	}

	err = db.RemoveRecord(ctx, &embeddingsdb.RemoveRecordRequest{Provider: "test", DepictionId: "a", Model: "m"})

	if err != nil {
		t.Fatal(err)
	}

	_, err = db.GetRecord(ctx, &embeddingsdb.GetRecordRequest{Provider: "test", DepictionId: "a", Model: "m"})

	if err != RecordNotFound {
		t.Fatalf("Expected RecordNotFound, got %v", err)
	}

	models, _ := db.Models(ctx)

	if len(models) != 1 || models[0] != "m" {
		t.Fatalf("Unexpected models %v", models)
	}
}
