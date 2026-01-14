//go:build duckdb

// This is all up for debate. Just testing things right now.

package database

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
	"strings"
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

func (db *DuckDBDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	provider := rec.Provider
	depiction_id := rec.DepictionId
	subject_id := rec.SubjectId
	model := rec.Model
	created := rec.Created

	now := time.Now()
	lastmod := now.Unix()

	embeddings, err := json.Marshal(rec.Embeddings)

	if err != nil {
		return fmt.Errorf("Failed to marshal embeddings for ID %s, %w", depiction_id, err)
	}

	attributes, err := json.Marshal(rec.Attributes)

	if err != nil {
		return fmt.Errorf("Failed to marshal attributes for ID %s, %w", depiction_id, err)
	}

	q := "INSERT OR REPLACE INTO embeddings (provider, depiction_id, subject_id, model, attributes, vec, created, lastmodified) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"

	_, err = db.vec_db.ExecContext(ctx, q, provider, depiction_id, subject_id, model, string(attributes), string(embeddings), created, lastmod)

	if err != nil {
		return fmt.Errorf("Failed to add embeddings for %s, %w", depiction_id, err)
	}

	return nil
}

func (db *DuckDBDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRequest) ([]*embeddingsdb.SimilarResult, error) {

	results := make([]*embeddingsdb.SimilarResult, 0)

	embeddings, err := json.Marshal(rec.Embeddings)

	if err != nil {
		return nil, fmt.Errorf("Failed to serialize query, %w", err)
	}

	conditions := make([]string, 0)

	args := []any{
		string(embeddings),
	}

	if rec.SimilarProvider != nil {
		conditions = append(conditions, "provider = ?")
		args = append(args, &rec.SimilarProvider)
	}

	conditions = append(conditions, "model == ?")
	args = append(args, rec.Model)

	conditions = append(conditions, "distance <= ?")
	args = append(args, db.max_distance)

	str_conditions := strings.Join(conditions, " AND ")

	q := fmt.Sprintf(`SELECT provider, depiction_id, subject_id, array_distance(vec, ?::FLOAT[%d]) AS distance
			  FROM embeddings WHERE %s ORDER BY distance ASC LIMIT %d`,
		str_conditions, db.dimensions, db.max_results)

	slog.Debug(q)

	t1 := time.Now()

	rows, err := db.vec_db.QueryContext(ctx, q, args...)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute query (%s), %w", q, err)
	}

	slog.Debug("Query context", "time", time.Since(t1))

	for rows.Next() {

		var provider string
		var depiction_id string
		var subject_id string
		var distance float64

		err = rows.Scan(&provider, &depiction_id, &subject_id, &distance)

		if err != nil {
			return nil, fmt.Errorf("Failed to scan row, %w", err)
		}

		r := &SimilarResult{
			Provider:    provider,
			SubjectId:   subject_id,
			DepictionId: depiction_id,
			Similarity:  float32(distance),
		}

		results = append(results, r)
	}

	slog.Debug("Query rows", "time", time.Since(t1))

	return results, nil
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
		cmds = append(cmds, fmt.Sprintf("CREATE TABLE embeddings(provider TEXT, depiction_id TEXT, subject_id TEXT, model TEXT, attributes JSON, vec FLOAT[%d], created BIGINT, lastmodified BIGINT", dimensions))
		cmds = append(cmds, "CREATE UNIQUE INDEX id_model ON embeddings (provider, depiction_id, model)")
		cmds = append(cmds, "CREATE INDEX by_provider ON embeddings (provider, model, created)")
		cmds = append(cmds, "CREATE INDEX by_model ON embeddings (model, provider, created)")
		cmds = append(cmds, "CREATE INDEX by_lastmod ON embeddings (lastmodified)")
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
