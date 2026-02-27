package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	sfom_sql "github.com/sfomuseum/go-database/sql"
	"github.com/sfomuseum/go-embeddingsdb"
)

type SQLiteDatabase struct {
	Database
	db *sql.DB
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "sqlite3", NewSQLiteDatabase)

	if err != nil {
		panic(err)
	}
}

func NewSQLiteDatabase(ctx context.Context, uri string) (Database, error) {

	sqlite_vec.Auto()

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
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

	dsn := q.Get("dsn")

	vec_db, err := sql.Open("sqlite3", dsn)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	setup_opts := &setupSQLiteDatabaseOptions{
		Dimensions: dimensions,
	}

	if u.Path != "" {

		abs_path, err := filepath.Abs(u.Path)

		if err != nil {
			return nil, fmt.Errorf("Failed to derive absolute path for database, %w", err)
		}

		setup_opts.DatabasePath = abs_path
	}

	err = setupSQLiteDatabase(ctx, vec_db, setup_opts)

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

	db := &SQLiteDatabase{
		db_uri:       uri,
		vec_db:       vec_db,
		dimensions:   dimensions,
		max_distance: max_distance,
		max_results:  max_results,
	}

	return db, nil
}

func (db *SQLiteDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *SQLiteDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {
	return nil
}

func (db *SQLiteDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {
	return nil, fmt.Errorf("Not found")
}

func (db *SQLiteDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *SQLiteDatabase) LastUpdate(ctx context.Context) (int64, error) {
	return 0, nil
}

func (db *SQLiteDatabase) URI() string {
	return db.db_uri
}

func (db *SQLiteDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {
	models := make([]string, 0)
	return models, nil
}

func (db *SQLiteDatabase) Providers(ctx context.Context) ([]string, error) {
	providers := make([]string, 0)
	return providers, nil
}

func (db *SQLiteDatabase) Close(ctx context.Context) error {
	return db.vec_db.Close()
}

type setupSQLiteDatabaseOptions struct {
	Dimensions int
}

func setupSQLiteDatabase(ctx context.Context, db *sql.DB, opts *setupSQLiteDatabaseOptions) error {

	t1 := time.Now()

	defer func() {
		slog.Debug("Finished setting up database", "time", time.Since(t1))
	}()

	vec_t, err := NewSQLiteTable(ctx, "")

	if err != nil {
		return err
	}

	return sfom_sql.CreateTableIfNecessary(ctx, db, vec_t)
}
