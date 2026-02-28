package database

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"text/template"

	sfom_sql "github.com/sfomuseum/go-database/sql"
)

//go:embed sqlite_schema.txt
var fs embed.FS

type SQLiteTable struct {
	sfom_sql.Table
	dimensions   int
	max_distance float32
	max_results  int32
	name         string
}

type SQLiteTableSchemaVars struct {
	Dimensions int
	TableName  string
}

func NewSQLiteTable(ctx context.Context, uri string) (sfom_sql.Table, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	dimensions := 512
	max_distance := float32(1.0)
	max_results := int32(10)

	if q.Has("dimensions") {

		v, err := strconv.Atoi(q.Get("dimensions"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?dimensions= parameter, %w", err)
		}

		dimensions = v
		slog.Debug("Reassign dimensions", "value", dimensions)
	}

	if q.Has("max-distance") {

		v, err := strconv.ParseFloat(q.Get("max-distance"), 64)

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-distance= parameter, %w", err)
		}

		max_distance = float32(v)
		slog.Debug("Reassign max distance", "value", max_distance)
	}

	if q.Has("max-results") {

		v, err := strconv.Atoi(q.Get("max-results"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-results= parameter, %w", err)
		}

		max_results = int32(v)
		slog.Debug("Reassign max results", "value", max_results)
	}

	t := &SQLiteTable{
		name:         "vec",
		dimensions:   dimensions,
		max_results:  max_results,
		max_distance: max_distance,
	}

	return t, nil
}

func (t *SQLiteTable) Name() string {
	return t.name
}

func (t *SQLiteTable) Schema(*sql.DB) (string, error) {

	tp, err := template.ParseFS(fs, "*_schema.txt")

	if err != nil {
		return "", err
	}

	tp = tp.Lookup("sqlite_schema")

	if tp == nil {
		return "", fmt.Errorf("Missing schema template")
	}

	vars := SQLiteTableSchemaVars{
		Dimensions: t.dimensions,
		TableName:  t.Name(),
	}

	var buf bytes.Buffer
	wr := bufio.NewWriter(&buf)

	err = tp.Execute(wr, vars)

	if err != nil {
		return "", err
	}

	wr.Flush()
	return buf.String(), nil
}

func (t *SQLiteTable) InitializeTable(context.Context, *sql.DB) error {
	return nil
}

func (t *SQLiteTable) IndexRecord(context.Context, *sql.DB, *sql.Tx, interface{}) error {
	return fmt.Errorf("Not implemented")
}
