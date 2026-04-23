package database

import (
	"context"
	"fmt"
	"iter"
	"net/url"
	"slices"
	"strconv"
	"sync"

	"github.com/aaronland/go-pagination"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

const MultiDatabaseScheme string = "multi"

type MultiDatabase struct {
	Database
	uri       string
	databases *sync.Map
}

type execCallbackFunc func(context.Context, Database) error

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
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	db_uris := q["database-uri"]

	if len(db_uris) == 0 {
		return nil, fmt.Errorf("At least one ?database-uri parameter is required")
	}

	db_map := new(sync.Map)
	db_errors := make([]error, 0)

	wg := new(sync.WaitGroup)

	err_ch := make(chan error)
	done_ch := make(chan bool)

	go func() {

		for {
			select {
			case <-done_ch:
				return
			case err := <-err_ch:
				db_errors = append(db_errors, err)
			}
		}
	}()

	for _, target_uri := range db_uris {

		wg.Go(func() {

			db_u, err := url.Parse(target_uri)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to parse target URI '%s', %w", target_uri, err)
				return
			}

			db_q := db_u.Query()

			if !db_q.Has("dimensions") {
				err_ch <- fmt.Errorf("Target database '%s' missing ?dimensions= parameter", target_uri)
				return
			}

			d, err := strconv.Atoi(db_q.Get("dimensions"))

			if err != nil {
				err_ch <- fmt.Errorf("Target database '%s' has invalid ?dimensions parameter, %w", target_uri, err)
				return
			}

			_, registered := db_map.Load(d)

			if registered {
				err_ch <- fmt.Errorf("Database with %d dimensions already registered", d)
				return
			}

			target_db, err := NewDatabase(ctx, target_uri)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to create target database '%s', %w", target_uri, err)
				return
			}

			db_map.Store(d, target_db)
		})
	}

	wg.Wait()
	done_ch <- true

	if len(db_errors) > 0 {
		return nil, fmt.Errorf("...")
	}

	db := &MultiDatabase{
		uri:       uri,
		databases: db_map,
	}
	return db, nil
}

// Return the URI string used to instantiate the Database instance.
func (db *MultiDatabase) URI() string {
	return db.uri
}

// Export the contents of the database. This is current a no/op for this Database implementation.
func (db *MultiDatabase) Export(ctx context.Context, uri string, opts ...options.Option) error {
	return nil
}

// Add adds a [embeddingsdb.Record] instance to the underlying database implementation. Returns true or false if the addition was batched.
func (db *MultiDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record, opts ...options.Option) (bool, error) {

	target_db, err := db.loadDatabase(ctx, len(rec.Embeddings))

	if err != nil {
		return false, err
	}

	return target_db.AddRecord(ctx, rec)
}

// The number of batched records currently waiting to be added.
func (db *MultiDatabase) BatchedRecordsCount(ctx context.Context, opts ...options.Option) (int, error) {

	dims := GetAllDimensionsFromOptions(ctx, opts...)

	if len(dims) == 0 {
		dims = db.Dimensions()
	}

	total := 0

	mu := new(sync.RWMutex)

	cb := func(ctx context.Context, target_db Database) error {

		count, err := target_db.BatchedRecordsCount(ctx)

		if err != nil {
			return err
		}

		mu.Lock()
		total += count
		mu.Unlock()
		return nil
	}

	err := db.exec(ctx, cb, dims...)

	if err != nil {
		return int(total), err
	}

	return total, nil
}

// Add the pending batched records.
func (db *MultiDatabase) AddBatchedRecord(ctx context.Context, opts ...options.Option) error {

	dims := GetAllDimensionsFromOptions(ctx, opts...)

	if len(dims) == 0 {
		dims = db.Dimensions()
	}

	cb := func(ctx context.Context, target_db Database) error {
		return target_db.AddBatchedRecords(ctx)
	}

	return db.exec(ctx, cb, dims...)

}

// Return the EmbeddingsDB instance record matching 'provider', 'depiction_id' and 'model'.
// The `MultiDatabase` implementation requires a minimum of (1) [options.Option] implementing [option.DimensionsOption]
// in order to determine which underlying embeddings database to query.
func (db *MultiDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest, opts ...options.Option) (*embeddingsdb.Record, error) {

	d, err := DeriveModelDimensions(ctx, req.Model, opts...)

	if err != nil {
		return nil, err
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, err
	}

	return target_db.GetRecord(ctx, req)
}

// Remove a record from an EmbeddingsDB instance.
// The `MultiDatabase` implementation requires a minimum of (1) [options.Option] implementing [option.DimensionsOption]
//
//	in order to determine which underlying embeddings database to update.
func (db *MultiDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest, opts ...options.Option) error {

	d, err := DeriveModelDimensions(ctx, req.Model, opts...)

	if err != nil {
		return fmt.Errorf("Failed to derive model dimensions, %w", err)
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return err
	}

	return target_db.RemoveRecord(ctx, req)
}

// Find similar records for a given model and record instance.
func (db *MultiDatabase) SimilarRecords(ctx context.Context, req *embeddingsdb.SimilarRecordsRequest, opts ...options.Option) ([]*embeddingsdb.SimilarRecord, error) {

	d := len(req.Embeddings)

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, err
	}

	return target_db.SimilarRecords(ctx, req, opts...)
}

// ListRecords returns a paginated list of records stored in the database.
// The `MultiDatabase` implementation requires a minimum of (1) [options.Option] implementing [option.DimensionsOption]
//
//	in order to determine which underlying embeddings database to update.
func (db *MultiDatabase) ListRecords(ctx context.Context, pg_opts pagination.Options, opts ...options.Option) ([]*embeddingsdb.Record, pagination.Results, error) {

	d, err := GetDimensionFromOptions(ctx, opts...)

	if err != nil {
		return nil, nil, err
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, nil, err
	}

	return target_db.ListRecords(ctx, pg_opts, opts...)
}

// IterateRecords returns an [iter.Seq2[*embeddingsdb.Record, error]] for each record stored in the database.
func (db *MultiDatabase) IterateRecords(ctx context.Context, opts ...options.Option) iter.Seq2[*embeddingsdb.Record, error] {

	dims := GetAllDimensionsFromOptions(ctx, opts...)

	if len(dims) == 0 {
		dims = db.Dimensions()
	}

	return func(yield func(*embeddingsdb.Record, error) bool) {

		for _, d := range dims {

			target_db, err := db.loadDatabase(ctx, d)

			if err != nil {
				if !yield(nil, err) {
					continue
				}
			}

			for rec, err := range target_db.IterateRecords(ctx) {

				if !yield(rec, err) {
					return
				}
			}
		}
	}
}

// Return the Unix timestamp of the last update to the Database instance.
func (db *MultiDatabase) LastUpdate(ctx context.Context, opts ...options.Option) (int64, error) {

	dims := GetAllDimensionsFromOptions(ctx, opts...)

	if len(dims) == 0 {
		dims = db.Dimensions()
	}

	lastupdate := int64(0)
	mu := new(sync.RWMutex)

	cb := func(ctx context.Context, target_db Database) error {

		target_lastupdate, err := target_db.LastUpdate(ctx)

		if err != nil {
			return fmt.Errorf("Failed to determine last update for %s, %w", target_db.URI(), err)
		}

		mu.Lock()

		if target_lastupdate > lastupdate {
			lastupdate = target_lastupdate
		}

		mu.Unlock()
		return nil
	}

	err := db.exec(ctx, cb, dims...)

	if err != nil {
		return lastupdate, fmt.Errorf("Failed to issue last update commands, %w", err)
	}

	return lastupdate, nil
}

// Return the unique list of models, for zero (all) or more providers, across all the embeddings.
func (db *MultiDatabase) Models(ctx context.Context, opts ...options.Option) ([]string, error) {

	dims := GetAllDimensionsFromOptions(ctx, opts...)

	if len(dims) == 0 {
		dims = db.Dimensions()
	}

	models := make([]string, 0)
	mu := new(sync.RWMutex)

	cb := func(ctx context.Context, target_db Database) error {

		target_models, err := target_db.Models(ctx, opts...)

		if err != nil {
			return fmt.Errorf("Failed to determine models for %s, %w", target_db.URI(), err)
		}

		mu.Lock()

		for _, m := range target_models {
			if !slices.Contains(models, m) {
				models = append(models, m)
			}
		}

		mu.Unlock()
		return nil

	}

	err := db.exec(ctx, cb, dims...)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute models commands, %w", err)
	}

	return models, nil
}

// Return the unique list of providers across all the embeddings.
func (db *MultiDatabase) Providers(ctx context.Context, opts ...options.Option) ([]string, error) {

	dims := GetAllDimensionsFromOptions(ctx, opts...)

	if len(dims) == 0 {
		dims = db.Dimensions()
	}

	providers := make([]string, 0)
	mu := new(sync.RWMutex)

	cb := func(ctx context.Context, target_db Database) error {

		target_providers, err := target_db.Providers(ctx, opts...)

		if err != nil {
			return fmt.Errorf("Failed to determine providers for %s, %w", target_db.URI(), err)
		}

		mu.Lock()

		for _, p := range target_providers {

			if !slices.Contains(providers, p) {
				providers = append(providers, p)
			}
		}
		mu.Unlock()
		return nil

	}

	err := db.exec(ctx, cb, dims...)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute providers commands, %w", err)
	}

	return providers, nil
}

// Return the list of dimensions supported by this Database implementation.
func (db *MultiDatabase) Dimensions() []int {

	dimensions := make([]int, 0)

	db.databases.Range(func(k, v any) bool {
		dimensions = append(dimensions, v.(int))
		return true
	})

	return dimensions
}

// Close performs and terminating functions required by the database.
func (db *MultiDatabase) Close(ctx context.Context) error {

	dims := db.Dimensions()

	cb := func(ctx context.Context, target_db Database) error {
		return target_db.Close(ctx)
	}

	return db.exec(ctx, cb, dims...)
}

func (db *MultiDatabase) exec(ctx context.Context, cb execCallbackFunc, dimensions ...int) error {

	if len(dimensions) == 0 {
		return fmt.Errorf("At least one dimension needs to be specified")
	}

	wg := new(sync.WaitGroup)

	errors := make([]error, 0)

	err_ch := make(chan error)
	done_ch := make(chan bool)

	go func() {

		for {
			select {
			case <-done_ch:
				return
			case err := <-err_ch:
				errors = append(errors, err)
			}
		}
	}()

	for _, d := range dimensions {

		target_db, err := db.loadDatabase(ctx, d)

		if err != nil {
			errors = append(errors, fmt.Errorf("Failed to load database for %d dimensions, %w", d, err))
			continue
		}

		wg.Go(func() {

			err = cb(ctx, target_db)

			if err != nil {
				err_ch <- fmt.Errorf("Failed to execute callback for %s, %w", target_db.URI(), err)
			}
		})
	}

	wg.Wait()
	done_ch <- true

	if len(errors) > 0 {
		return fmt.Errorf("...")
	}

	return nil
}

func (db *MultiDatabase) loadDatabase(ctx context.Context, d int) (Database, error) {

	v, exists := db.databases.Load(d)

	if !exists {
		return nil, fmt.Errorf("Database not registered for %d embeddings", d)
	}

	return v.(Database), nil
}
