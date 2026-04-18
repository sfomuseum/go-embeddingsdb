//go:build bleve

package database

// To do:
// Store/retrieve embeddings to/from an external source (sqlite? duckdb since it's already loaded?)
// Filter by provider for list view

import (
	"context"
	"database/sql"
	"embed"
	"encoding/binary"
	"fmt"
	"iter"
	"log/slog"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"

	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/blevesearch/bleve/v2/search"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/blevesearch/bleve_index_api"
	sfom_sql "github.com/sfomuseum/go-database/sql"
	"github.com/sfomuseum/go-embeddingsdb"
)

//go:embed bleve_*_schema.txt
var bleve_schema_fs embed.FS

type BleveDatabase struct {
	Database
	uri           string
	dimensions    int
	index         bleve.Index
	embeddings_db *sql.DB
	embeddings_t  sfom_sql.Table
	batch         *bleve.Batch
	batch_size    int
	mu            *sync.RWMutex
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "bleve", NewBleveDatabase)

	if err != nil {
		panic(err)
	}
}

// Create a new [BleveDatabase] instance for managing embeddings using the Bleve document store derived from 'uri' which is expected to take the form of:
//
//	bleve://{PATH}?{QUERY_PARAMETERS}
//
// If {PATH} is omitted an in-memory database will be created.
//
// Valid query parameters are:
// * `dimensions` – The number of dimensions for the embeddings being stored. Default is 512.
func NewBleveDatabase(ctx context.Context, uri string) (Database, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	path_index := u.Path

	q := u.Query()

	dimensions := 512

	if q.Has("dimensions") {

		v, err := strconv.Atoi(q.Get("dimensions"))

		if err != nil {
			return nil, fmt.Errorf("Invalid ?dimensions= parameter, %w", err)
		}

		dimensions = v
		slog.Debug("Reassign dimensions", "value", dimensions)
	}

	var index bleve.Index
	var path_embeddings string

	switch path_index {
	case "":
		idx_mapping := bleveMappings(dimensions)
		index, err = bleve.NewMemOnly(idx_mapping)
	default:

		_, err = os.Stat(path_index)

		if err != nil {

			idx_config := map[string]interface{}{
				"forceSegmentType":    "zap",
				"forceSegmentVersion": 16, // 16 required for vector search
				"initialMmapSize":     512 * 1024 * 1024,
			}

			idx_mapping := bleveMappings(dimensions)
			index, err = bleve.NewUsing(path_index, idx_mapping, "scorch", "scorch", idx_config)
		} else {
			index, err = bleve.Open(path_index)
		}

		path_embeddings = filepath.Join(path_index, "embeddingsdb")
	}

	if err != nil {
		return nil, fmt.Errorf("Failed to create Bleve index, %w", err)
	}

	// START OF

	emb_db, err := sql.Open("duckdb", path_embeddings)

	if err != nil {
		return nil, fmt.Errorf("Failed to create (static) embeddings index, %w", err)
	}

	emb_t, err := NewBleveEmbeddingsTable(ctx, "bleve://")

	if err != nil {
		return nil, fmt.Errorf("Failed to create (static) embeddings table, %w", err)
	}

	err = sfom_sql.CreateTableIfNecessary(ctx, emb_db, emb_t)

	if err != nil {
		return nil, fmt.Errorf("Failed to setup (static) embeddings table, %w", err)
	}

	// END OF

	batch := index.NewBatch()
	mu := new(sync.RWMutex)

	db := &BleveDatabase{
		uri:           uri,
		index:         index,
		dimensions:    dimensions,
		embeddings_db: emb_db,
		batch:         batch,
		batch_size:    200,
		mu:            mu,
	}

	return db, nil
}

func (db *BleveDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *BleveDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) (bool, error) {

	err := db.batchRecord(ctx, rec)

	if err != nil {
		return false, err
	}

	if db.batch.Size() >= db.batch_size {

		err := db.AddBatchedRecords(ctx)

		if err != nil {
			return false, err
		}
	}

	return true, nil
}

func (db *BleveDatabase) BatchedRecordsCount(ctx context.Context) int {

	db.mu.Lock()
	defer db.mu.Unlock()

	return db.batch.Size()
}

func (db *BleveDatabase) AddBatchedRecords(ctx context.Context) error {

	db.mu.Lock()
	defer db.mu.Unlock()

	select {
	case <-ctx.Done():
		return nil
	default:
		// pass
	}

	if db.batch.Size() == 0 {
		return nil
	}

	logger := slog.Default()
	logger = logger.With("size", db.batch.Size())
	logger.Debug("Flush batch")

	t1 := time.Now()

	defer func() {
		slog.Debug("Time to flush batch", "time", time.Since(t1))
	}()

	err := db.index.Batch(db.batch)

	if err != nil {
		return err
	}

	db.batch = db.index.NewBatch()
	return nil
}

func (db *BleveDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {

	id := req.Key()

	q := bleve.NewDocIDQuery([]string{id})

	bl_req := bleve.NewSearchRequestOptions(q, 1, 0, false)
	bl_req.Fields = []string{"*"}

	rsp, err := db.index.Search(bl_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute search to get record, %w", err)
	}

	if rsp.Total != 1 {
		return nil, fmt.Errorf("Not found")
	}

	first := rsp.Hits[0]

	return db.inflateRecordWithMatch(ctx, first)
}

func (db *BleveDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest) error {

	id := req.Key()
	return db.index.Delete(id)
}

func (db *BleveDatabase) SimilarRecords(ctx context.Context, req *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {

	results := make([]*embeddingsdb.SimilarRecord, 0)

	if len(req.Embeddings) != db.dimensions {
		logger := slog.Default()
		logger.Warn("Invalid embeddings", "dimensions", len(req.Embeddings))
		return results, nil
	}

	k := 10

	if req.MaxResults != nil {
		k = int(*req.MaxResults)
	}

	var filters []query.Query

	modelQ := bleve.NewTermQuery(req.Model)
	modelQ.SetField("model")

	filters = append(filters, modelQ)

	if req.SimilarProvider != nil && *req.SimilarProvider != "" {
		provQ := bleve.NewTermQuery(*req.SimilarProvider)
		provQ.SetField("provider")
		filters = append(filters, provQ)
	}

	if len(req.Exclude) > 0 {
		notQuery := bleve.NewDocIDQuery(req.Exclude)
		boolQuery := bleve.NewBooleanQuery()
		boolQuery.AddMustNot(notQuery)
		filters = append(filters, boolQuery)
	}

	filterQuery := bleve.NewConjunctionQuery(filters...)

	search_req := bleve.NewSearchRequest(filterQuery)
	search_req.AddKNN("embeddings", req.Embeddings, int64(k), 1.0)
	search_req.SortBy([]string{"_score"})
	search_req.Size = k
	search_req.Fields = []string{"*"}

	rsp, err := db.index.Search(search_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute search request to find similar, %w", err)
	}

	for _, hit := range rsp.Hits {

		dist := float32(hit.Score)

		if req.MaxDistance != nil && dist > *req.MaxDistance {
			continue
		}

		rec, err := db.inflateRecordWithMatch(ctx, hit)

		if err != nil {
			return nil, fmt.Errorf("Failed to get internal record for '%s', %w", hit.ID, err)
		}

		// Why isn't this filter being applied above?

		if len(req.Exclude) > 0 {

			if slices.Contains(req.Exclude, rec.Key()) {
				continue
			}
		}

		results = append(results, &embeddingsdb.SimilarRecord{
			Provider:    rec.Provider,
			DepictionId: rec.DepictionId,
			SubjectId:   rec.SubjectId,
			Attributes:  rec.Attributes,
			Distance:    dist,
		})
	}

	return results, nil
}

func (db *BleveDatabase) ListRecords(ctx context.Context, pg_opts pagination.Options, filters ...*ListRecordsFilter) ([]*embeddingsdb.Record, pagination.Results, error) {

	records := make([]*embeddingsdb.Record, 0)

	per_page := int(pg_opts.PerPage())
	from := int(countable.PageFromOptions(pg_opts))

	query := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequestOptions(query, per_page, from, false)
	req.Fields = []string{"*"}

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to execute search, %w", err)
	}

	for _, h := range rsp.Hits {

		rec, err := db.inflateRecordWithMatch(ctx, h)

		if err != nil {
			return nil, nil, fmt.Errorf("Failed to retrieve internal record '%s', %w", h.ID, err)
		}

		records = append(records, rec)
	}

	pg_rsp, err := countable.NewResultsFromCountWithOptions(pg_opts, int64(rsp.Total))

	if err != nil {
		return nil, nil, fmt.Errorf("Failed to derive pagination results, %w", err)
	}

	return records, pg_rsp, nil
}

func (db *BleveDatabase) IterateRecords(ctx context.Context) iter.Seq2[*embeddingsdb.Record, error] {

	return func(yield func(*embeddingsdb.Record, error) bool) {

		idx, err := db.index.Advanced()

		if err != nil {
			yield(nil, fmt.Errorf("Failed to retrieve internal index, %w", err))
			return
		}

		idx_r, err := idx.Reader()

		if err != nil {
			yield(nil, fmt.Errorf("Failed to retrieve internal reader, %w", err))
			return
		}

		defer idx_r.Close()

		id_r, err := idx_r.DocIDReaderAll()

		if err != nil {
			yield(nil, fmt.Errorf("Failed to retrieve document reader, %w", err))
			return
		}

		defer id_r.Close()

		for {

			id, err := id_r.Next()

			if err != nil {
				yield(nil, fmt.Errorf("Document reader did not yield next, %w", err))
				return
			}

			doc, err := idx_r.Document(string(id))

			if err != nil {
				yield(nil, err)
				return
			}

			rec, err := db.inflateRecordWithDocument(ctx, doc)

			if !yield(rec, fmt.Errorf("Failed to retrieve internal record for '%s', %w", string(id), err)) {
				return
			}
		}
	}
}

func (db *BleveDatabase) LastUpdate(ctx context.Context) (int64, error) {

	q := bleve.NewMatchAllQuery()

	req := bleve.NewSearchRequestOptions(q, 1, 0, false)

	req.Size = 1
	req.SortBy([]string{"-created"})
	req.Fields = []string{"created"}

	rsp, err := db.index.Search(req)

	if err != nil {
		return 0, fmt.Errorf("Failed to execute search request, %w", err)
	}

	if rsp.Total == 0 {
		return 0, fmt.Errorf("no records found")
	}

	rec, err := db.inflateRecordWithMatch(ctx, rsp.Hits[0])

	if err != nil {
		return 0, fmt.Errorf("Failed to retrieve internal record for '%s', %w", rsp.Hits[0].ID, err)
	}

	return rec.Created, nil
}

func (db *BleveDatabase) URI() string {
	return db.uri
}

func (db *BleveDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {

	count, err := db.index.DocCount()

	if err != nil {
		return nil, err
	}

	facet := bleve.NewFacetRequest("model", int(count))
	query := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequest(query)
	req.AddFacet("model", facet)

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute search request, %w", err)
	}

	terms := rsp.Facets["model"].Terms.Terms()
	models := make([]string, len(terms))

	for idx, term := range terms {
		models[idx] = term.Term
	}

	return models, nil
}

func (db *BleveDatabase) Providers(ctx context.Context) ([]string, error) {

	count, err := db.index.DocCount()

	if err != nil {
		return nil, err
	}

	facet := bleve.NewFacetRequest("provider", int(count))
	query := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequest(query)
	req.AddFacet("provider", facet)

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, fmt.Errorf("Failed to execute search request, %w", err)
	}

	terms := rsp.Facets["provider"].Terms.Terms()
	providers := make([]string, len(terms))

	for idx, term := range terms {
		providers[idx] = term.Term
	}

	return providers, nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {

	logger := slog.Default()
	logger.Debug("Close database")

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to close database", "time", time.Since(t1))
	}()

	if db.batch.Size() >= 0 {

		err := db.AddBatchedRecords(ctx)

		if err != nil {
			return err
		}
	}

	t2 := time.Now()

	defer func() {
		logger.Debug("Time to close index", "time", time.Since(t2))
	}()

	return db.index.Close()
}

func (db *BleveDatabase) batchRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	db.mu.Lock()
	defer db.mu.Unlock()

	id := rec.Key()

	db.batch.Index(id, rec)

	return db.storeEmbeddings(ctx, rec)
}

func (db *BleveDatabase) inflateRecordWithMatch(ctx context.Context, doc *search.DocumentMatch) (*embeddingsdb.Record, error) {

	rec := new(embeddingsdb.Record)

	for k, v := range doc.Fields {

		switch {
		case strings.HasPrefix(k, "attributes."):

			k = strings.Replace(k, "attributes.", "", 1)

			attrs := rec.Attributes

			if attrs == nil {
				attrs = make(map[string]string)
			}

			attrs[k] = v.(string)
			rec.Attributes = attrs

		default:

			switch k {
			case "provider":
				rec.Provider = v.(string)
			case "model":
				rec.Model = v.(string)
			case "depiction_id":
				rec.DepictionId = v.(string)
			case "subject_id":
				rec.SubjectId = v.(string)
			case "created":
				rec.Created = int64(v.(float64))
			default:
				slog.Warn("Unsupported key", "k", k, "v", v)
			}
		}
	}

	err := db.assignEmbeddings(ctx, rec)

	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (db *BleveDatabase) inflateRecordWithDocument(ctx context.Context, doc index.Document) (*embeddingsdb.Record, error) {

	rec := new(embeddingsdb.Record)

	doc.VisitFields(func(f index.Field) {

		k := f.Name()
		v := f.Value()

		switch {
		case strings.HasPrefix(k, "attributes."):

			k = strings.Replace(k, "attributes.", "", 1)

			attrs := rec.Attributes

			if attrs == nil {
				attrs = make(map[string]string)
			}

			attrs[k] = string(v)
			rec.Attributes = attrs

		default:

			switch k {
			case "provider":
				rec.Provider = string(v)
			case "model":
				rec.Model = string(v)
			case "depiction_id":
				rec.DepictionId = string(v)
			case "subject_id":
				rec.SubjectId = string(v)
			case "created":
				bits := binary.BigEndian.Uint64(v)
				f64 := math.Float64frombits(bits)
				rec.Created = int64(f64)
			default:
				slog.Warn("Unsupported key", "k", k, "v", v)
			}
		}
	})

	err := db.assignEmbeddings(ctx, rec)

	if err != nil {
		return nil, err
	}

	return rec, nil
}

func (db *BleveDatabase) storeEmbeddings(ctx context.Context, rec *embeddingsdb.Record) error {

	id := rec.Key()

	q := fmt.Sprintf(`INSERT OR REPLACE INTO %s(id, embeddings) VALUES(?, ?)`, db.embeddings_t.Name())

	_, err := db.embeddings_db.ExecContext(ctx, q, id, rec.Embeddings)

	return err

	/*
		buf := make([]byte, len(data)*4)

		for i, f := range data {
			bits := math.Float32bits(f)
			binary.BigEndian.PutUint32(buf[i*4:], bits)
		}

		db.batch.SetInternal([]byte(id), buf)
		return nil
	*/
}

func (db *BleveDatabase) assignEmbeddings(ctx context.Context, rec *embeddingsdb.Record) error {

	id := rec.Key()

	q := fmt.Sprintf(`SELECT embeddings FROM %s WHERE id = ?`, db.embeddings_t.Name())

	row := db.embeddings_db.QueryRowContext(ctx, q, id)

	var embeddings []float32

	err := row.Scan(&embeddings)

	if err != nil {
		return err
	}

	rec.Embeddings = embeddings
	return nil

	/*
		buf, err := db.index.GetInternal([]byte(id))

		if err != nil || buf == nil {
			return nil, err
		}

		count := len(buf) / 4
		result := make([]float32, count)

		for i := 0; i < count; i++ {
			bits := binary.BigEndian.Uint32(buf[i*4 : (i+1)*4])
			result[i] = math.Float32frombits(bits)
		}

		return result, nil
	*/
}

func bleveMappings(dimensions int) *mapping.IndexMappingImpl {

	kw_mapping := bleve.NewTextFieldMapping()
	kw_mapping.Analyzer = "keyword"
	kw_mapping.Store = true
	kw_mapping.Index = true
	kw_mapping.DocValues = true
	kw_mapping.IncludeTermVectors = false
	kw_mapping.IncludeInAll = false

	txt_mapping := bleve.NewTextFieldMapping()
	txt_mapping.Store = true
	txt_mapping.Index = true
	txt_mapping.DocValues = false
	txt_mapping.IncludeTermVectors = false
	txt_mapping.IncludeInAll = false

	num_mapping := bleve.NewNumericFieldMapping()
	num_mapping.Store = true
	num_mapping.Index = true
	num_mapping.DocValues = true
	num_mapping.IncludeInAll = false

	vec_mapping := bleve.NewVectorFieldMapping()
	vec_mapping.Dims = dimensions
	vec_mapping.Similarity = "l2_norm"
	vec_mapping.Store = false
	vec_mapping.Index = true
	vec_mapping.DocValues = false
	vec_mapping.IncludeInAll = false

	sf_mapping := bleve.NewTextFieldMapping()
	sf_mapping.Store = true

	map_mapping := bleve.NewDocumentMapping()
	map_mapping.Dynamic = true
	map_mapping.DefaultAnalyzer = "keyword"
	map_mapping.AddFieldMapping(sf_mapping)

	idx_mapping := bleve.NewIndexMapping()
	idx_mapping.DefaultMapping.Enabled = true
	idx_mapping.DefaultMapping.Dynamic = false

	idx_mapping.DefaultMapping.AddFieldMappingsAt("embeddings", vec_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("model", kw_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("provider", kw_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("subject_id", txt_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("depiction_id", txt_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("created", num_mapping)

	idx_mapping.DefaultMapping.AddSubDocumentMapping("attributes", map_mapping)

	return idx_mapping
}
