package database

import (
	"context"
	"testing"
)

func TestSQLiteRecordsTable(t *testing.T) {

	ctx := context.Background()
	tb_uri := "sqlite://"

	tb, err := NewSQLiteRecordsTable(ctx, tb_uri)

	if err != nil {
		t.Fatalf("Failed to create new SQLite records table, %v", err)
	}

	_, err = tb.Schema(nil)

	if err != nil {
		t.Fatalf("Failed to derive SQLite records table schema, %v", err)
	}
}
