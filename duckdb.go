//go:build duckdb

// This is all up for debate. Just testing things right now.

package embeddingsdb

// https://duckdb.org/2024/05/03/vector-similarity-search-vss.html
// https://duckdb.org/docs/api/go.html
// https://pkg.go.dev/github.com/marcboeker/go-duckdb
// https://github.com/marcboeker/go-duckdb/tree/main?tab=readme-ov-file#vendoring

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	_ "github.com/marcboeker/go-duckdb/v2"
)

type DuckDBDatabase struct {
	// The underlying SQLite database used to store and query embeddings.
	vec_db *sql.DB
	// The number of dimensions for embeddings
	dimensions int
	// The maximum number of results for queries
	max_results int
	// The compression type to use for embeddings. Valid options are: quantize, matroyshka, none (default)
	compression string
	// ..
	max_distance float32
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "duckdb", NewDuckDBDatabase)

	if err != nil {
		panic(err)
	}
}

func NewDuckDBDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	dimensions := 512
	max_distance := float32(1.0)
	max_results := 10

	if q.Has("dimensions") {

		v, err := strconv.Atoi(q.Get("dimensions"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?dimensions= parameter, %w", err)
		}

		dimensions = v
	}

	if q.Has("max-distance") {

		v, err := strconv.ParseFloat(q.Get("max-distance"), 64)

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-distance= parameter, %w", err)
		}

		max_distance = float32(v)
	}

	if q.Has("max-results") {

		v, err := strconv.Atoi(q.Get("max-results"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?max-results= parameter, %w", err)
		}

		max_results = v
	}

	vec_db, err := sql.Open("duckdb", "")

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	err = setupDuckDBDatabase(ctx, vec_db, u.Path, dimensions)

	if err != nil {
		return nil, fmt.Errorf("Failed to setup database, %w", err)
	}

	if q.Has("max-conns") {

		v, err := strconv.Atoi(q.Get("max-conns"))

		if err != nil {
			return nil, err
		}

		vec_db.SetMaxOpenConns(v)
	}

	db := &DuckDBDatabase{
		vec_db:       vec_db,
		dimensions:   dimensions,
		max_distance: max_distance,
		max_results:  max_results,
	}

	return db, nil
}

func (db *DuckDBDatabase) Export(ctx context.Context, uri string) error {

	// There does not appear to be any way to use query placeholders for this.
	// Note: Export directory does not need to exist before calling EXPORT...

	q := fmt.Sprintf("EXPORT DATABASE '%s'", uri)
	_, err := db.vec_db.ExecContext(ctx, q)
	return err
}

func (db *DuckDBDatabase) Add(ctx context.Context, rec *Record) error {

	depiction_id := rec.DepictionId
	subject_id := rec.SubjectId
	content := rec.URI
	model := rec.Model

	v, err := json.Marshal(rec.Embeddings)

	if err != nil {
		return fmt.Errorf("Failed to marshal embeddings for ID %s, %w", depiction_id, err)
	}

	q := "INSERT OR REPLACE INTO embeddings (depiction_id, subject_id, model, content, vec) VALUES (?, ?, ?, ?, ?)"

	_, err = db.vec_db.ExecContext(ctx, q, depiction_id, subject_id, model, content, string(v), subject_id, model, content, string(v))

	if err != nil {
		return fmt.Errorf("Failed to add embeddings for %s, %w", depiction_id, err)
	}

	return nil
}

func (db *DuckDBDatabase) Models(ctx context.Context) ([]string, error) {

	q := "SELECT DISTINCT(model) AS model FROM embeddings"

	rows, err := db.vec_db.QueryContext(ctx, q)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query (%s), %w", q, err)
	}

	models := make([]string, 0)

	for rows.Next() {

		var model string

		err = rows.Scan(&model)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		models = append(models, model)
	}

	return models, nil
}

func (db *DuckDBDatabase) Similar(ctx context.Context, rec *Record) ([]*QueryResult, error) {

	results := make([]*QueryResult, 0)

	v, err := json.Marshal(rec.Embeddings)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize query, %w", err)
	}

	q := fmt.Sprintf(`SELECT depiction_id, subject_id, content, array_distance(vec, ?::FLOAT[%d]) AS distance
			  FROM embeddings WHERE depiction_id != ? AND model == ? AND distance <= ? ORDER BY distance ASC LIMIT %d`,
		db.dimensions, db.max_results)

	slog.Debug(q)

	t1 := time.Now()

	rows, err := db.vec_db.QueryContext(ctx, q, string(v), rec.DepictionId, rec.Model, db.max_distance)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query (%s), %w", q, err)
	}

	slog.Debug("Query context", "time", time.Since(t1))

	for rows.Next() {

		var depiction_id string
		var subject_id string
		var content string
		var distance float64

		err = rows.Scan(&depiction_id, &subject_id, &content, &distance)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		r := &QueryResult{
			SubjectId:   subject_id,
			DepictionId: depiction_id,
			Content:     content,
			Similarity:  float32(distance),
		}

		slog.Debug("Result", "depiction id", depiction_id, "content", content, "distance", distance)

		results = append(results, r)
	}

	slog.Debug("Query rows", "time", time.Since(t1))

	return results, nil
}

func (db *DuckDBDatabase) Flush(ctx context.Context) error {
	return nil
}

func (db *DuckDBDatabase) Close(ctx context.Context) error {
	return db.vec_db.Close()
}

func setupDuckDBDatabase(ctx context.Context, db *sql.DB, path string, dimensions int) error {

	t1 := time.Now()

	defer func() {
		slog.Debug("Finished setting up database", "time", time.Since(t1))
	}()

	cmds := []string{
		"INSTALL vss",
		"LOAD vss",
	}

	if path != "" {
		cmds = append(cmds, fmt.Sprintf("IMPORT DATABASE '%s'", path))
	} else {
		cmds = append(cmds, fmt.Sprintf("CREATE TABLE embeddings(depiction_id INTEGER, subject_id INTEGER, model TEXT, content TEXT, vec FLOAT[%d])", dimensions))
		cmds = append(cmds, "CREATE UNIQUE INDEX id_model ON embeddings (depiction_id, model)")
		cmds = append(cmds, "CREATE INDEX idx ON embeddings USING HNSW (vec)")
	}

	for _, q := range cmds {

		slog.Debug(q)

		_, err := db.ExecContext(ctx, q)

		if err != nil {
			return fmt.Errorf("Failed to configure data - query failed, %w (%s)", err, q)
		}
	}

	return nil
}
