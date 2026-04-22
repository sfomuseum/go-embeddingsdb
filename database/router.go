package database

import (
	"context"
	"fmt"
	"iter"
	"sync"
	"net/url"
	"strconv"
	
	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
)

type RouterDatabase struct {
	Database
	uri       string
	databases *sync.Map
}

type execCallbackFunc func(context.Context, Database) error

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "router", NewRouterDatabase)

	if err != nil {
		panic(err)
	}
}

func NewRouterDatabase(ctx context.Context, uri string) (Database, error) {

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

	db := &RouterDatabase{
		uri:       uri,
		databases: db_map,
	}
	return db, nil
}

func (db *RouterDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *RouterDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) (bool, error) {

	target_db, err := db.loadDatabase(ctx, len(rec.Embeddings))

	if err != nil {
		return false, err
	}

	return target_db.AddRecord(ctx, rec)
}

func (db *RouterDatabase) BatchedRecordsCount(ctx context.Context) (int, error) {

	d := db.getAllDimensionsFromOptions(ctx)
	count := len(d)

	switch {
	case count == 0:
		return 0, fmt.Errorf("Missing dimensions option")
	case count > 1:
		return 0, fmt.Errorf("Multiple dimensions specified")
	default:
		// pass
	}

	target_db, err := db.loadDatabase(ctx, d[0])

	if err != nil {
		return 0, err
	}

	return target_db.BatchedRecordsCount(ctx)
}

func (db *RouterDatabase) AddBatchedRecord(ctx context.Context) error {

	dims := db.getAllDimensionsFromOptions(ctx)

	if len(dims) == 0 {
	   	     dims = db.Dimensions()
        }

	// Do all...
	
	target_db, err := db.loadDatabase(ctx, dims[0])

	if err != nil {
		return nil, err
	}

	return target_db.AddBatchedRecords(ctx)
}

func (db *RouterDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {

	d, err := db.getDimensionFromOptions(ctx)

	if err != nil {
		return nil, err
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, err
	}

	return target_db.GetRecord(ctx, req)
}

func (db *RouterDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest) error {

	d, err := db.getDimensionFromOptions(ctx)

	if err != nil {
		return nil, err
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, err
	}

	return target_db.RemoveRecord(ctx, req)
}

func (db *RouterDatabase) SimilarRecords(ctx context.Context, req *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {

	d := len(req.Embeddings)

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, err
	}

	return target_db.SimilarRecords(ctx, req)
}

func (db *RouterDatabase) ListRecords(ctx context.Context, opts pagination.Options, filters ...*ListRecordsFilter) ([]*embeddingsdb.Record, pagination.Results, error) {

	d, err := db.getDimensionFromOptions(ctx)

	if err != nil {
		return nil, err
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
		return nil, err
	}

	return target_db.ListRecords(ctx, opts, filters...)
}

func (db *RouterDatabase) IterateRecords(ctx context.Context) iter.Seq2[*embeddingsdb.Record, error] {

	d := -1

	return func(yield func(*embeddingsdb.Record, error) bool) {

		target_db, err := db.LoadDatabase(ctx, d)

		if err != nil {
			yield(nil, err)
		}

		for rec, err := range target_db.IterateRecords(ctx) {

			if !yield(rec, err) {
				return
			}
		}
	}
}

func (db *RouterDatabase) LastUpdate(ctx context.Context) (int64, error) {

	d, err := db.getDimensionFromOptions(ctx)

	if err != nil {
		return 0, err
	}

	target_db, err := db.loadDatabase(ctx, d)

	if err != nil {
	   return 0, nil
	}
	
	return 0, nil
}

func (db *RouterDatabase) URI() string {
	return db.uri
}

func (db *RouterDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {
	models := make([]string, 0)
	return models, nil
}

func (db *RouterDatabase) Providers(ctx context.Context) ([]string, error) {
	providers := make([]string, 0)
	return providers, nil
}

func (db *RouterDatabase) Close(ctx context.Context) error {

     cb := func(ctx context.Context, target_db Database) error {
     	erturn target_db.Close(ctx)
     }

     return db.exec(ctx, db)
}

func (db *RouterDatabase) exec(ctx context.Context, cb execCallbackFunc, dimensions ...int) error {

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
			errors = append(errors, err)
			continue
		}

		wg.Go(func() {

			err = cb(ctx, target_db)

			if err != nil {
				err_ch <- err
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

func (db *RouterDatabase) loadDatabase(ctx context.Context, d int) (Database, error) {

	v, exists := db.databases.Load(d)

	if !exists {
		return nil, fmt.Errorf("Database not registered")
	}

	return v.(Database)
}

func (db *RouterDatabase) getDimensionFromOptions(ctx context.Context, opts ...Option) (int, error){

	dims := db.getAllDimensionsFromOptions(ctx)

	switch {
	case count == 0:
		return 0, fmt.Errorf("Missing dimensions option")
	case count > 1:
		return 0, fmt.Errorf("Multiple dimensions specified")
	default:
		return dims[0]
	}

}

func (db *RouterDatabase) getAllDimensionsFromOptions(ctx context.Context, opts ...Option) []int {

	dimensions := make([]int, 0)

	for _, o := range opts {

		if o.Type() == DimensionsOptions {

			if !slices.Contains(dimensions, o.Value()) {
				dimensions = append(dimensions, o.Value())
			}
		}
	}

	return dimensions
}
