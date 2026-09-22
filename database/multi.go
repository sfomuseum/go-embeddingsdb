package database

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"slices"
	"strconv"
	"sync"

	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

const MultiDatabaseScheme string = "multi"

type MultiDatabase struct {
	Database
	registry    map[int]Database
	model_cache *sync.Map
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, MultiDatabaseScheme, NewMultiDatabase)

	if err != nil {
		panic(err)
	}
}

func NewMultiDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	db_uris, ok := q["database"]

	if !ok {
		return nil, fmt.Errorf("URI missing one or more ?database= parameters")
	}

	return NewMultiDatabaseFromURIs(ctx, db_uris...)
}

func NewMultiDatabaseFromURIs(ctx context.Context, db_uris ...string) (Database, error) {

	registry := make(map[int]Database)

	for _, uri := range db_uris {

		db_u, err := url.Parse(uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse database URI '%s', %w", uri, err)
		}

		str_dims := db_u.Fragment

		if str_dims == "" {
			return nil, fmt.Errorf("Database URI '%s' missing #{DIMENSIONS} fragment", uri)
		}

		dims, err := strconv.Atoi(str_dims)

		if err != nil {
			return nil, fmt.Errorf("Failed to parse dimensions for database URI '%s', %w", uri, err)
		}

		_, exists := registry[dims]

		if exists {
			return nil, fmt.Errorf("Database already registered for dimensions '%d'", dims)
		}

		db_u.Fragment = ""

		other_db, err := NewDatabase(ctx, db_u.String())

		if err != nil {
			return nil, fmt.Errorf("Failed to create new database for '%s', %w", uri, err)
		}

		registry[dims] = other_db
	}

	return NewMultiDatabaseFromRegistry(ctx, registry)
}

func NewMultiDatabaseFromRegistry(ctx context.Context, registry map[int]Database) (Database, error) {

	model_cache := new(sync.Map)

	db := &MultiDatabase{
		registry:    registry,
		model_cache: model_cache,
	}

	return db, nil
}

// Return the URI string used to instantiate the Database instance.
func (db *MultiDatabase) URI() string {

	database_uris := make([]string, 0)

	for dims, target_db := range db.registry {

		target_uri := target_db.URI()
		target_u, err := url.Parse(target_uri)

		if err != nil {
			slog.Error("Failed to parse target URI", "uri", target_uri, "dims", dims, "error", err)
			continue
		}

		target_u.Fragment = strconv.Itoa(dims)
		database_uris = append(database_uris, target_u.String())
	}

	q := url.Values{}
	q["database"] = database_uris

	u := url.URL{}
	u.Scheme = "multi"
	u.RawQuery = q.Encode()

	return u.String()
}

// Export the contents of the database. Where and how a database is exported are left as details for specific implementations.
func (db *MultiDatabase) Export(ctx context.Context, uri string, opts ...options.Option) error {

	for _, target_db := range db.registry {

		err := target_db.Export(ctx, uri, opts...)

		if err != nil {
			return err
		}
	}

	return nil
}

// Add adds a [embeddingsdb.Record] instance to the underlying database implementation. Returns true or false if the addition was batched.
func (db *MultiDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record, opts ...options.Option) (bool, error) {

	dims := len(rec.Embeddings)

	target_db, ok := db.registry[dims]

	if !ok {
		return false, fmt.Errorf("Unregistered database for %d dimensions", dims)
	}

	return target_db.AddRecord(ctx, rec, opts...)
}

// The number of batched records currently waiting to be added.
func (db *MultiDatabase) BatchedRecordsCount(ctx context.Context, opts ...options.Option) (int, error) {

	total := 0

	for _, target_db := range db.registry {

		count, err := target_db.BatchedRecordsCount(ctx, opts...)

		if err != nil {
			return total, err
		}

		total += count
	}

	return total, nil
}

// Add the pending batched records.
func (db *MultiDatabase) AddBatchedRecord(ctx context.Context, opts ...options.Option) error {

	for _, target_db := range db.registry {

		err := target_db.AddBatchedRecords(ctx, opts...)

		if err != nil {
			return err
		}
	}

	return nil
}

// Return the EmbeddingsDB instance record matching 'provider', 'depiction_id' and 'model'.
func (db *MultiDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest, opts ...options.Option) (*embeddingsdb.Record, error) {

	target_db, err := db.databaseForModel(ctx, req.Model, opts...)

	if err != nil {
		return nil, err

	}

	return target_db.GetRecord(ctx, req, opts...)
}

// Remove a record from an EmbeddingsDB instance.
func (db *MultiDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest, opts ...options.Option) error {

	target_db, err := db.databaseForModel(ctx, req.Model, opts...)

	if err != nil {
		return err
	}

	return target_db.RemoveRecord(ctx, req, opts...)
}

// Find similar records for a given model and record instance.
func (db *MultiDatabase) SimilarRecords(ctx context.Context, req *embeddingsdb.SimilarRecordsRequest, opts ...options.Option) ([]*embeddingsdb.SimilarRecord, error) {

	target_db, err := db.databaseForModel(ctx, req.Model, opts...)

	if err != nil {
		return nil, err
	}

	return target_db.SimilarRecords(ctx, req, opts...)
}

// ListRecords returns a paginated list of records stored in the database.
func (db *MultiDatabase) ListRecords(ctx context.Context, pg_opts pagination.Options, opts ...options.Option) ([]*embeddingsdb.Record, pagination.Results, error) {

	records := make([]*embeddingsdb.Record, 0)

	pg, err := countable.NewResultsFromCountWithOptions(pg_opts, 0)

	if err != nil {
		return nil, nil, err
	}

	return records, pg, nil
}

// IterateRecords returns an [iter.Seq2[*embeddingsdb.Record, error]] for each record stored in the database.
func (db *MultiDatabase) IterateRecords(ctx context.Context, opts ...options.Option) iter.Seq2[*embeddingsdb.Record, error] {

	return func(yield func(*embeddingsdb.Record, error) bool) {

		keep_iterating := true

		for _, target_db := range db.registry {

			for rec, err := range target_db.IterateRecords(ctx, opts...) {

				if !yield(rec, err) {
					keep_iterating = false
					break
				}
			}

			if !keep_iterating {
				break
			}
		}
	}
}

// Return the Unix timestamp of the last update to the Database instance.
func (db *MultiDatabase) LastUpdate(ctx context.Context, opts ...options.Option) (int64, error) {

	lastupdate := int64(0)

	for _, target_db := range db.registry {

		u, err := target_db.LastUpdate(ctx, opts...)

		if err != nil {
			return 0, err
		}

		if u < lastupdate {
			continue
		}

		lastupdate = u

	}

	return lastupdate, nil
}

// Return the list of dimensions supported by this Database  implementation.
func (db *MultiDatabase) Dimensions(ctx context.Context, opts ...options.Option) ([]int, error) {

	dims := make([]int, 0)

	for d, _ := range db.registry {
		dims = append(dims, d)
	}

	return dims, nil
}

// Return the unique list of models, for zero (all) or more providers, across all the embeddings.
func (db *MultiDatabase) Models(ctx context.Context, opts ...options.Option) ([]string, error) {

	models := make([]string, 0)

	for _, target_db := range db.registry {

		target_models, err := target_db.Models(ctx, opts...)

		if err != nil {
			return nil, err
		}

		for _, m := range target_models {

			if !slices.Contains(models, m) {
				models = append(models, m)
			}
		}
	}

	return models, nil
}

// Return the unique list of providers across all the embeddings.
func (db *MultiDatabase) Providers(ctx context.Context, opts ...options.Option) ([]string, error) {

	providers := make([]string, 0)

	for _, target_db := range db.registry {

		target_providers, err := target_db.Providers(ctx, opts...)

		if err != nil {
			return nil, err
		}

		for _, p := range target_providers {

			if !slices.Contains(providers, p) {
				providers = append(providers, p)
			}
		}
	}

	return providers, nil
}

// Return the pagination type used by the database.
func (db *MultiDatabase) PaginationType(ctx context.Context, opts ...options.Option) (PaginationType, error) {
	return NullPaginationType, nil
}

// Close performs and terminating functions required by the database.
func (db *MultiDatabase) Close(ctx context.Context) error {

	for _, target_db := range db.registry {

		err := target_db.Close(ctx)

		if err != nil {
			return err
		}
	}

	return nil
}

func (db *MultiDatabase) databaseForModel(ctx context.Context, model string, opts ...options.Option) (Database, error) {

	v, ok := db.model_cache.Load(model)

	if ok {

		switch v.(type) {
		case Database:
			return v.(Database), nil
		default:
			return nil, fmt.Errorf("Model not found")
		}
	}

	var target_db Database

	for _, test_db := range db.registry {

		test_models, err := test_db.Models(ctx, opts...)

		if err != nil {
			return nil, err
		}

		if slices.Contains(test_models, model) {
			target_db = test_db
			break
		}
	}

	db.model_cache.Store(model, target_db)

	if target_db == nil {
		return nil, fmt.Errorf("No matching database for model")
	}

	return target_db, nil
}
