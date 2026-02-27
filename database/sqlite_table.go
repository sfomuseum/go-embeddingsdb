package database

import (
	"context"
	"database/sql"
	"fmt"

	sfom_sql "github.com/sfomuseum/go-database/sql"
)

type SQLiteTable struct {
	sfom_sql.Table
}

func NewSQLiteTable(ctx context.Context, uri string) (sfom_sql.Table, error) {

	t := &SQLiteTable{
		name: "vec",
	}

	return t, nil
}

func (t *SQLiteTable) Name() string {
	return t.name
}

func (t *SQLiteTable) Schema(*sql.DB) (string, error) {
	return "", nil
}

func (t *SQLiteTable) InitializeTable(context.Context, *sql.DB) error {
	return nil
}

func (t *SQLiteTable) IndexRecord(context.Context, *sql.DB, *sql.Tx, interface{}) error {
	return fmt.Errorf("Not implemented")
}
