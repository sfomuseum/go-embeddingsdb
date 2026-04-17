package database

// https://github.com/blevesearch/bleve/blob/master/docs/vectors.md

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"os"
	"slices"
	"strconv"

	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
	"github.com/sfomuseum/go-embeddingsdb"
)

type BleveDatabase struct {
	Database
	uri   string
	index bleve.Index
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "bleve", NewBleveDatabase)

	if err != nil {
		panic(err)
	}
}

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

	kw_mapping := bleve.NewTextFieldMapping()
	kw_mapping.Analyzer = "keyword"
	kw_mapping.Store = false
	kw_mapping.Index = true
	kw_mapping.DocValues = true

	vec_mapping := bleve.NewVectorFieldMapping()
	vec_mapping.Dims = dimensions
	vec_mapping.Similarity = "l2_norm"
	vec_mapping.Store = true
	vec_mapping.Index = true
	vec_mapping.DocValues = true

	idx_mapping := bleve.NewIndexMapping()
	idx_mapping.DefaultMapping.AddFieldMappingsAt("embeddings", vec_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("model", kw_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("provider", kw_mapping)

	var index bleve.Index

	bleve_opts := map[string]interface{}{
		"forceSegments":   true,
		"persistInterval": 2000,
	}

	switch path_index {
	case "":
		index, err = bleve.NewMemOnly(idx_mapping)
	default:

		_, err = os.Stat(path_index)

		if err != nil {
			index, err = bleve.NewUsing(path_index, idx_mapping, "scorch", "scorch", bleve_opts)
		} else {
			index, err = bleve.OpenUsing(path_index, bleve_opts)
		}
	}

	if err != nil {
		return nil, err
	}

	db := &BleveDatabase{
		uri:   uri,
		index: index,
	}

	return db, nil
}

func (db *BleveDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *BleveDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	id := rec.Key()

	err := db.index.Index(id, rec)

	if err != nil {
		return err
	}

	raw, err := json.Marshal(rec)

	if err != nil {
		return err
	}

	return db.index.SetInternal([]byte(id), raw)
}

func (db *BleveDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {

	id := fmt.Sprintf("%s-%s-%s", req.Provider, req.DepictionId, req.Model)

	q := bleve.NewDocIDQuery([]string{id})

	bl_req := bleve.NewSearchRequestOptions(q, 1, 0, false)
	bl_req.Fields = []string{"*"}

	rsp, err := db.index.Search(bl_req)

	if err != nil {
		return nil, err
	}

	if rsp.Total != 1 {
		return nil, fmt.Errorf("Not found")
	}

	first := rsp.Hits[0]

	return db.getInternal(first.ID)
}

func (db *BleveDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest) error {

	id := fmt.Sprintf("%s-%s-%s", req.Provider, req.DepictionId, req.Model)
	return db.index.Delete(id)
}

func (db *BleveDatabase) SimilarRecords(ctx context.Context, req *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {

	results := make([]*embeddingsdb.SimilarRecord, 0)

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
	search_req.Size = k

	rsp, err := db.index.Search(search_req)

	if err != nil {
		return nil, err
	}

	for _, hit := range rsp.Hits {

		dist := float32(hit.Score)

		if req.MaxDistance != nil && dist > *req.MaxDistance {
			continue
		}

		rec, err := db.getInternal(hit.ID)

		if err != nil {
			return nil, err
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

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, nil, err
	}

	for _, h := range rsp.Hits {
		rec, err := db.getInternal(h.ID)

		if err != nil {
			return nil, nil, err
		}

		records = append(records, rec)
	}

	pg_rsp, err := countable.NewResultsFromCountWithOptions(pg_opts, int64(rsp.Total))

	if err != nil {
		return nil, nil, err
	}

	return records, pg_rsp, nil
}

func (db *BleveDatabase) IterateRecords(ctx context.Context) iter.Seq2[*embeddingsdb.Record, error] {

	return func(yield func(*embeddingsdb.Record, error) bool) {

		idx, err := db.index.Advanced()

		if err != nil {
			yield(nil, err)
			return
		}

		idx_r, err := idx.Reader()

		if err != nil {
			yield(nil, err)
			return
		}

		defer idx_r.Close()

		id_r, err := idx_r.DocIDReaderAll()

		if err != nil {
			yield(nil, err)
			return
		}

		defer id_r.Close()

		for {

			id, err := id_r.Next()

			if err != nil {
				yield(nil, err)
				return
			}

			rec, err := db.getInternal(string(id))

			if !yield(rec, err) {
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

	rsp, err := db.index.Search(req)

	if err != nil {
		return 0, err
	}

	if rsp.Total == 0 {
		return 0, fmt.Errorf("no records found")
	}

	rec, err := db.getInternal(rsp.Hits[0].ID)

	if err != nil {
		return 0, err
	}

	return rec.Created, nil
}

func (db *BleveDatabase) URI() string {
	return db.uri
}

func (db *BleveDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {

	facet := bleve.NewFacetRequest("model", 1000000)
	query := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequest(query)
	req.AddFacet("model", facet)

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, err
	}

	terms := rsp.Facets["model"].Terms.Terms()
	models := make([]string, len(terms))

	for idx, term := range terms {
		models[idx] = term.Term
	}

	return models, nil
}

func (db *BleveDatabase) Providers(ctx context.Context) ([]string, error) {

	facet := bleve.NewFacetRequest("provider", 1000000)
	query := bleve.NewMatchAllQuery()
	req := bleve.NewSearchRequest(query)
	req.AddFacet("provider", facet)

	rsp, err := db.index.Search(req)

	if err != nil {
		return nil, err
	}

	terms := rsp.Facets["provider"].Terms.Terms()
	providers := make([]string, len(terms))

	for idx, term := range terms {
		providers[idx] = term.Term
	}

	return providers, nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {
	return nil
}

func (db *BleveDatabase) getInternal(id string) (*embeddingsdb.Record, error) {

	raw, err := db.index.GetInternal([]byte(id))

	if err != nil {
		return nil, err
	}

	if raw == nil {
		return nil, fmt.Errorf("Internal record missing")
	}

	var rec *embeddingsdb.Record

	err = json.Unmarshal(raw, &rec)

	if err != nil {
		return nil, err
	}

	return rec, nil
}
