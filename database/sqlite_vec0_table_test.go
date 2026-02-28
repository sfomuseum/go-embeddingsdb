package database

import (
	"context"
	"fmt"
	"testing"
)

func TestSQLiteTable(t *testing.T) {

	ctx := context.Background()

	tb_uri := "sqlite://?dimensions=384"

	tb, err := NewSQLiteVec0Table(ctx, tb_uri)

	if err != nil {
		t.Fatalf("Failed to create new SQLite table, %v", err)
	}

	schema, err := tb.Schema(nil)

	if err != nil {
		t.Fatalf("Failed to derive SQLite table schema, %v", err)
	}

	fmt.Println(schema)
}
