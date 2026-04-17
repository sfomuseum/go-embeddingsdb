package database

// https://github.com/blevesearch/bleve/blob/master/docs/vectors.md

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/blevesearch/bleve/v2"
	index_api "github.com/blevesearch/bleve_index_api"
	"github.com/sfomuseum/go-embeddingsdb"
)

type BleveDatabase struct {
	Database
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
	kw_mapping.Store = true
	kw_mapping.Index = true
	kw_mapping.DocValues = true

	num_mapping := bleve.NewNumericFieldMapping()

	vec_mapping := bleve.NewVectorFieldMapping()
	vec_mapping.Dims = dimensions
	vec_mapping.Similarity = "l2_norm"

	idx_mapping := bleve.NewIndexMapping()
	idx_mapping.DefaultMapping.AddFieldMappingsAt("embeddings", vec_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("created", num_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("model", kw_mapping)
	idx_mapping.DefaultMapping.AddFieldMappingsAt("provider", kw_mapping)

	var index bleve.Index

	switch path_index {
	case "":
		index, err = bleve.NewMemOnly(idx_mapping)
	default:

		_, err = os.Stat(path_index)

		if err != nil {
			index, err = bleve.New(path_index, idx_mapping)
		} else {
			index, err = bleve.Open(path_index)
		}
	}

	if err != nil {
		return nil, err
	}

	db := &BleveDatabase{
		index: index,
	}

	return db, nil
}

func (db *BleveDatabase) Export(ctx context.Context, uri string) error {
	return nil
}

func (db *BleveDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record) error {

	return db.index.Index(rec.Key(), rec)
}

func (db *BleveDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest) (*embeddingsdb.Record, error) {

	id := fmt.Sprintf("%s-%s-%s", req.Provider, req.DepictionId, req.Model)

	q := bleve.NewDocIDQuery([]string{id})

	bl_req := bleve.NewSearchRequestOptions(q, 1, 0, false)
	rsp, err := db.index.Search(bl_req)

	if err != nil {
		return nil, err
	}

	if rsp.Total != 1 {
		return nil, fmt.Errorf("Not found")
	}

	first := rsp.Hits[0]

	d, err := db.index.Document(first.ID)

	if err != nil {
		return nil, err
	}

	rec := new(embeddingsdb.Record)

	d.VisitFields(func(f index_api.Field) {
		slog.Info("F", "f", f.Name(), "v", string(f.Value()))

		switch {
		case strings.HasPrefix(f.Name(), "attributes."):

			attrs := rec.Attributes

			if attrs == nil {
				attrs = make(map[string]string)
			}

			k := strings.Replace(f.Name(), "attributes.", "", 1)
			attrs[k] = string(f.Value())
			rec.Attributes = attrs

		default:
			switch f.Name() {
			case "subject_id":
				rec.SubjectId = string(f.Value())
			case "depiction_id":
				rec.DepictionId = string(f.Value())
			case "model":
				rec.Model = string(f.Value())
			case "provider":
				rec.Provider = string(f.Value())
			case "created":

				str_v := string(f.Value())
				str_v = strings.TrimSpace(str_v)
				v, err := strconv.ParseInt(str_v, 10, 64)

				if err != nil {
					slog.Error("Failed to parse created time", "v", str_v)
				} else {
					rec.Created = v
				}
			case "embeddings":
				// Y NO APPEAR EMBEDDINGS ???
			default:
				slog.Warn("Unrecognized field", "field", f.Name())
			}
		}
	})

	return rec, nil
}

func (db *BleveDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest) error {
	return nil
}

func (db *BleveDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

func (db *BleveDatabase) ListRecords(ctx context.Context, opts pagination.Options, filters ...*ListRecordsFilter) ([]*embeddingsdb.Record, pagination.Results, error) {

	records := make([]*embeddingsdb.Record, 0)

	pg, err := countable.NewResultsFromCountWithOptions(opts, 0)

	if err != nil {
		return nil, nil, err
	}

	return records, pg, nil
}

func (db *BleveDatabase) IterateRecords(ctx context.Context) iter.Seq2[*embeddingsdb.Record, error] {
	return func(yield func(*embeddingsdb.Record, error) bool) {}
}

func (db *BleveDatabase) LastUpdate(ctx context.Context) (int64, error) {
	return 0, nil
}

func (db *BleveDatabase) URI() string {
	return "bleve://"
}

func (db *BleveDatabase) Models(ctx context.Context, providers ...string) ([]string, error) {

	facet := bleve.NewFacetRequest("model", 1)
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
		fmt.Printf("model = %s (count=%d)\n", term.Term, term.Count)
		models[idx] = term.Term
	}

	return models, nil
}

func (db *BleveDatabase) Providers(ctx context.Context) ([]string, error) {

	facet := bleve.NewFacetRequest("provider", 1)
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
		fmt.Printf("provider = %s (count=%d)\n", term.Term, term.Count)
		providers[idx] = term.Term
	}

	return providers, nil
}

func (db *BleveDatabase) Close(ctx context.Context) error {
	return nil
}
