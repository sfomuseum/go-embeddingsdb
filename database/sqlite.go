package database

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
	"github.com/bwmarrin/snowflake"
	sfom_sql "github.com/sfomuseum/go-database/sql"
	"github.com/sfomuseum/go-embeddingsdb"
)

//go:embed sqlite_*_schema.txt
var sqlite_schema_fs embed.FS

var snowflake_node *snowflake.Node

type SQLiteDatabase struct {
	Database
	db_uri        string
	vec_db        *sql.DB
	vec_table     sfom_sql.Table
	records_table sfom_sql.Table
	dimensions    int
	max_distance  float32
	max_results   int32
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "sqlite", NewSQLiteDatabase)

	if err != nil {
		panic(err)
	}

	n, err := snowflake.NewNode(1)

	if err != nil {
		panic(err)
	}

	snowflake_node = n
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

	// vec_db, err := sfom_sql.OpenWithURI(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to open database connection, %w", err)
	}

	t_query := url.Values{}
	t_query.Set("dimensions", strconv.Itoa(dimensions))

	t_uri := url.URL{}
	t_uri.Scheme = "sqlite"
	t_uri.RawQuery = t_query.Encode()

	vec_t, err := NewSQLiteVec0Table(ctx, t_uri.String())

	if err != nil {
		return nil, err
	}

	err = sfom_sql.CreateTableIfNecessary(ctx, vec_db, vec_t)

	if err != nil {
		return nil, fmt.Errorf("Failed to setup vec table, %w", err)
	}

	records_t, err := NewSQLiteRecordsTable(ctx, "sqlite://")

	if err != nil {
		return nil, err
	}

	err = sfom_sql.CreateTableIfNecessary(ctx, vec_db, records_t)

	if err != nil {
		return nil, fmt.Errorf("Failed to setup records table, %w", err)
	}

	if q.Has("max-conns") {

		v, err := strconv.Atoi(q.Get("max-conns"))

		if err != nil {
			return nil, err
		}

		vec_db.SetMaxOpenConns(v)
	}

	db := &SQLiteDatabase{
		db_uri:        uri,
		vec_db:        vec_db,
		vec_table:     vec_t,
		records_table: records_t,
		dimensions:    dimensions,
		max_distance:  max_distance,
		max_results:   max_results,
	}

	return db, nil
}

func (db *SQLiteDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *SQLiteDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	id, err := db.uidForRecord(ctx, rec.Provider, rec.DepictionId, rec.Model)

	if err != nil {
		return err
	}

	enc_e, err := sqlite_vec.SerializeFloat32(rec.Embeddings)

	if err != nil {
		return err
	}

	vec_q := fmt.Sprintf("INSERT INTO %s (rowid, embedding) VALUES (?, ?)", db.vec_table.Name())

	_, err = db.vec_db.ExecContext(ctx, vec_q, id, enc_e)

	if err != nil {
		return err
	}

	enc_attrs, err := json.Marshal(rec.Attributes)

	if err != nil {
		return err
	}

	now := time.Now()
	lastmod := now.Unix()

	records_q := fmt.Sprintf("INSERT INTO %s (id, provider, depiction_id, subject_id, model, attributes, created, lastmodified) VALUES (?, ?, ?, ?, ?, ?, ?, ?)", db.records_table.Name())

	_, err = db.vec_db.ExecContext(ctx, records_q, id, rec.Provider, rec.DepictionId, rec.SubjectId, rec.Model, string(enc_attrs), rec.Created, lastmod)

	if err != nil {
		return err
	}

	return err
}

func (db *SQLiteDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {

	id, err := db.uidForRecord(ctx, req.Provider, req.DepictionId, req.Model)

	if err != nil {
		return nil, err
	}

	records_q := fmt.Sprintf("SELECT v.embedding, r.provider, r.depiction_id, r.subject_id, r.model, r.attributes, r.created FROM %s r, %s v  WHERE r.id = v.rowid AND r.id = ?", db.records_table.Name(), db.vec_table.Name())

	row := db.vec_db.QueryRowContext(ctx, records_q, id)

	var enc_embeddings []byte
	var provider string
	var depiction_id string
	var subject_id string
	var model string
	var str_attrs string
	var created int64

	err = row.Scan(&enc_embeddings, &provider, &depiction_id, &subject_id, &model, &str_attrs, &created)

	if err != nil {
		return nil, err
	}

	e32, err := DeserializeFloat32(enc_embeddings)

	if err != nil {
		slog.Error("Failed to deserialize embeddings", "error", err)
	}

	var attrs map[string]string

	err = json.Unmarshal([]byte(str_attrs), &attrs)

	if err != nil {
		return nil, err
	}

	r := &embeddingsdb.Record{
		Provider:    provider,
		Embeddings:  e32,
		DepictionId: depiction_id,
		SubjectId:   subject_id,
		Model:       model,
		Attributes:  attrs,
		Created:     created,
	}

	return r, nil
}

func (db *SQLiteDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *SQLiteDatabase) LastUpdate(ctx context.Context) (int64, error) {

	q := fmt.Sprintf("SELECT lastmodified FROM %s ORDER BY lastmodified DESC LIMIT 1", db.records_table.Name())

	row := db.vec_db.QueryRowContext(ctx, q)

	var lastmod int64
	err := row.Scan(&lastmod)

	switch {
	case err == sql.ErrNoRows:
		new_id := snowflake_node.Generate()
		return new_id.Int64(), nil
	case err != nil:
		return 0, err
	default:
		return lastmod, nil
	}
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

func (db *SQLiteDatabase) uidForRecord(ctx context.Context, provider string, depiction_id string, model string) (int64, error) {

	q := fmt.Sprintf("SELECT id FROM %s WHERE provider= ? AND depiction_id = ? AND model = ?", db.records_table.Name())

	row := db.vec_db.QueryRowContext(ctx, q, provider, depiction_id, model)

	var id int64
	err := row.Scan(&id)

	switch {
	case err == sql.ErrNoRows:
		new_id := snowflake_node.Generate()
		return new_id.Int64(), nil
	case err != nil:
		return 0, err
	default:
		return id, nil
	}
}

// Compliment method to SerializeFloat32
// https://github.com/asg017/sqlite-vec-go-bindings/blob/main/cgo/lib.go#L33

func DeserializeFloat32(b []byte) ([]float32, error) {

	if len(b)%4 != 0 {
		return nil, fmt.Errorf("byte slice length %d is not a multiple of 4", len(b))
	}

	n := len(b) / 4           // number of float32 values
	vec := make([]float32, n) // allocate destination slice

	buf := bytes.NewReader(b)

	// binary.Read will read n float32 values into vec
	if err := binary.Read(buf, binary.LittleEndian, vec); err != nil {
		return nil, err
	}
	return vec, nil
}
