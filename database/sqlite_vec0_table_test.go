package database

import (
	"context"
	"fmt"
	"testing"
)

func TestSQLiteTable(t *testing.T) {

	ctx := context.Background()

	for _, compression := range sqlite_vec_compressions {

		tb_uri := fmt.Sprintf("sqlite://?dimensions=512&compression=%s", compression)

		tb, err := NewSQLiteVec0Table(ctx, tb_uri)

		if err != nil {
			t.Fatalf("[%s] Failed to create new SQLite table, %v", compression, err)
		}

		_, err = tb.Schema(nil)

		if err != nil {
			t.Fatalf("[%s] Failed to derive SQLite table schema, %v", compression, err)
		}
	}
}
