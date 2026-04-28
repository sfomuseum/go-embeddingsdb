package database

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"net/url"
	"strconv"
	"strings"

	"github.com/aaronland/go-aws/v3/auth"
	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/cursor"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors/document"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

type S3VectorsDatabase struct {
	Database
	uri    string
	bucket string
	index  string
	client *s3vectors.Client
}

const S3VectorsDatabaseScheme = "s3vectors"

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, S3VectorsDatabaseScheme, NewS3VectorsDatabase)

	if err != nil {
		panic(err)
	}
}

func NewS3VectorsDatabase(ctx context.Context, uri string) (Database, error) {

	dimensions := 512

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	bucket := u.Host

	if bucket == "" {
		return nil, fmt.Errorf("Missing bucket name as host")
	}

	required := []string{
		"region",
		"credentials",
		"index",
	}

	for _, k := range required {

		if !q.Has(k) {
			return nil, fmt.Errorf("Required ?%s= parameter missing", k)
		}

		if q.Get(k) == "" {
			return nil, fmt.Errorf("Required ?%s= parameter may not be empty", k)
		}
	}

	if q.Has("dimensions") {

		v, err := strconv.Atoi(q.Get("dimensions"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?dimension= parameter, %w", err)
		}

		dimensions = v
	}

	region := q.Get("region")
	creds := q.Get("credentials")
	index := q.Get("index")

	cfg_q := url.Values{}
	cfg_q.Set("region", region)
	cfg_q.Set("credentials", creds)

	cfg_u := url.URL{}
	cfg_u.Scheme = "aws"
	cfg_u.RawQuery = cfg_q.Encode()

	cfg_uri := cfg_u.String()

	cfg, err := auth.NewConfig(ctx, cfg_uri)

	if err != nil {
		return nil, err
	}

	cl := s3vectors.NewFromConfig(cfg)

	err = setupS3VectorsBucketAndIndex(ctx, cl, bucket, index, dimensions)

	if err != nil {
		return nil, err
	}

	db := &S3VectorsDatabase{
		client: cl,
		bucket: bucket,
		index:  index,
		uri:    uri,
	}

	return db, nil
}

// Export the contents of the database. Where and how a database is exported are left as details for specific implementations.
func (db *S3VectorsDatabase) Export(ctx context.Context, uri string, opts ...options.Option) error {
	return nil
}

// Add adds a [embeddingsdb.Record] instance to the underlying database implementation. Returns true or false if the addition was batched.
func (db *S3VectorsDatabase) AddRecord(ctx context.Context, rec *embeddingsdb.Record, opts ...options.Option) (bool, error) {

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors#Client.PutVectors

	attrs := rec.Attributes

	if attrs == nil {
		attrs = make(map[string]string)
	}

	attrs["x-model"] = rec.Model
	attrs["x-provider"] = rec.Provider
	attrs["x-subject-id"] = rec.SubjectId
	attrs["x-depiction-id"] = rec.DepictionId
	attrs["x-created"] = strconv.FormatInt(rec.Created, 10)

	meta := document.NewLazyDocument(attrs)

	vecs := []types.PutInputVector{
		types.PutInputVector{
			Data: &types.VectorDataMemberFloat32{
				Value: rec.Embeddings,
			},
			Key:      aws.String(rec.Key()),
			Metadata: meta,
		},
	}

	put_opts := &s3vectors.PutVectorsInput{
		VectorBucketName: aws.String(db.bucket),
		IndexName:        aws.String(db.index),
		Vectors:          vecs,
	}

	_, err := db.client.PutVectors(ctx, put_opts)

	if err != nil {
		return false, err
	}

	return false, nil
}

// The number of batched records currently waiting to be added.
func (db *S3VectorsDatabase) BatchedRecordsCount(ctx context.Context, opts ...options.Option) (int, error) {
	return 0, nil
}

// Add the pending batched records.
func (db *S3VectorsDatabase) AddBatchedRecord(ctx context.Context, opts ...options.Option) error {
	return nil
}

// Return the EmbeddingsDB instance record matching 'provider', 'depiction_id' and 'model'.
func (db *S3VectorsDatabase) GetRecord(ctx context.Context, req *embeddingsdb.GetRecordRequest, opts ...options.Option) (*embeddingsdb.Record, error) {

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors#Client.QueryVectors

	get_opts := &s3vectors.GetVectorsInput{
		Keys: []string{
			req.Key(),
		},
		IndexName:        aws.String(db.index),
		VectorBucketName: aws.String(db.bucket),
		ReturnMetadata:   true,
		ReturnData:       true,
	}

	rsp, err := db.client.GetVectors(ctx, get_opts)

	if err != nil {
		return nil, err
	}

	if len(rsp.Vectors) == 0 {
		return nil, RecordNotFound
	}

	vec := rsp.Vectors[0]

	return db.s3VectorToRecord(vec.Data, vec.Metadata)
}

// Remove a record from an EmbeddingsDB instance.
func (db *S3VectorsDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest, opts ...options.Option) error {

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors#Client.DeleteVectors

	del_opts := &s3vectors.DeleteVectorsInput{
		Keys: []string{
			req.Key(),
		},
		IndexName:        aws.String(db.index),
		VectorBucketName: aws.String(db.bucket),
	}

	_, err := db.client.DeleteVectors(ctx, del_opts)

	if err != nil {
		return fmt.Errorf("Failed to remove record, %w", err)
	}

	return nil
}

// Find similar records for a given model and record instance.
func (db *S3VectorsDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest, opts ...options.Option) ([]*embeddingsdb.SimilarRecord, error) {

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors#Client.QueryVectors
	// https://docs.aws.amazon.com/AmazonS3/latest/userguide/s3-vectors-metadata-filtering.html

	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

// ListRecords returns a paginated list of records stored in the database.
func (db *S3VectorsDatabase) ListRecords(ctx context.Context, pg_opts pagination.Options, opts ...options.Option) ([]*embeddingsdb.Record, pagination.Results, error) {

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors#NewListVectorsPaginator

	list_opts := &s3vectors.ListVectorsInput{
		VectorBucketName: aws.String(db.bucket),
		IndexName:        aws.String(db.index),
		ReturnData:       true,
		ReturnMetadata:   true,
	}

	per_page := pg_opts.PerPage()
	pointer := pg_opts.Pointer()

	var prev_cursor string
	var next_cursor string

	if per_page > 0 {
		list_opts.MaxResults = aws.Int32(int32(per_page))
	}

	if pointer != nil {

		if token, ok := pointer.(string); ok && token != "" {
			prev_cursor = token
			list_opts.NextToken = aws.String(token)
		}
	}

	rsp, err := db.client.ListVectors(ctx, list_opts)

	if err != nil {
		return nil, nil, err
	}

	if rsp.NextToken != nil {
		next_cursor = *rsp.NextToken
	}

	pg, err := cursor.NewPaginationFromCursors(prev_cursor, next_cursor)

	if err != nil {
		return nil, nil, err
	}

	count := len(rsp.Vectors)
	records := make([]*embeddingsdb.Record, count)

	for i, vec := range rsp.Vectors {

		rec, err := db.s3VectorToRecord(vec.Data, vec.Metadata)

		if err != nil {
			return nil, nil, err
		}

		records[i] = rec
	}

	return records, pg, nil
}

// IterateRecords returns an [iter.Seq2[*embeddingsdb.Record, error]] for each record stored in the database.
func (db *S3VectorsDatabase) IterateRecords(ctx context.Context, opts ...options.Option) iter.Seq2[*embeddingsdb.Record, error] {

	return func(yield func(*embeddingsdb.Record, error) bool) {

		list_opts := &s3vectors.ListVectorsInput{
			VectorBucketName: aws.String(db.bucket),
			IndexName:        aws.String(db.index),
			ReturnData:       true,
			ReturnMetadata:   true,
		}

		pg := s3vectors.NewListVectorsPaginator(db.client, list_opts)

		for pg.HasMorePages() {

			rsp, err := pg.NextPage(ctx)

			if err != nil {
				yield(nil, err)
				return
			}

			for _, vec := range rsp.Vectors {

				rec, err := db.s3VectorToRecord(vec.Data, vec.Metadata)

				if !yield(rec, err) {
					return
				}
			}
		}
	}
}

// Return the Unix timestamp of the last update to the Database instance.
func (db *S3VectorsDatabase) LastUpdate(ctx context.Context, opts ...options.Option) (int64, error) {
	return 0, nil
}

// Return the list of dimensions supported by this Database  implementation.
func (db *S3VectorsDatabase) Dimensions(ctx context.Context, opts ...options.Option) []int {

	// https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors#Client.ListVectors

	dims := make([]int, 0)

	for idx, err := range db.listIndexes(ctx) {

		if err != nil {
			continue // fix me
		}

		dims = append(dims, int(*idx.Dimension))
	}

	return dims
}

// Return the unique list of models, for zero (all) or more providers, across all the embeddings.
func (db *S3VectorsDatabase) Models(ctx context.Context, opts ...options.Option) ([]string, error) {
	models := make([]string, 0)
	return models, nil
}

// Return the unique list of providers across all the embeddings.
func (db *S3VectorsDatabase) Providers(ctx context.Context, opts ...options.Option) ([]string, error) {
	providers := make([]string, 0)
	return providers, nil
}

// Return the pagination type used by the database.
func (db *S3VectorsDatabase) PaginationType(ctx context.Context, opts ...options.Option) (PaginationType, error) {
	return CursorPaginationType, nil
}

// Close performs and terminating functions required by the database.
func (db *S3VectorsDatabase) Close(ctx context.Context) error {
	return nil
}

func (db *S3VectorsDatabase) s3VectorToRecord(vec_data types.VectorData, vec_meta document.Interface) (*embeddingsdb.Record, error) {

	meta := make(map[string]string)

	err := vec_meta.UnmarshalSmithyDocument(&meta)

	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal metadata")
	}

	var provider string
	var model string
	var depiction_id string
	var subject_id string
	var created int64

	attrs := make(map[string]string)

	for k, v := range meta {

		switch {
		case strings.HasPrefix(k, "x-"):

			switch k {
			case "x-provider":
				provider = v
			case "x-model":
				model = v
			case "x-depiction-id":
				depiction_id = v
			case "x-subject-id":
				subject_id = v
			case "x-created":

				created_v, err := strconv.ParseInt(v, 10, 64)

				if err != nil {
					slog.Warn("Failed to parse string created date", "date", v, "error", err)
				} else {
					created = created_v
				}

			default:
				slog.Debug("Unrecognized x- key", "k", k)
			}

		default:
			attrs[k] = v
		}
	}

	rec := &embeddingsdb.Record{
		Embeddings:  vec_data.(*types.VectorDataMemberFloat32).Value,
		Provider:    provider,
		Model:       model,
		DepictionId: depiction_id,
		SubjectId:   subject_id,
		Created:     created,
		Attributes:  attrs,
	}

	return rec, nil

}

func (db *S3VectorsDatabase) listIndexes(ctx context.Context) iter.Seq2[*types.Index, error] {

	// Maybe move in to aaronland/go-aws/s3vectors? TBD...

	return func(yield func(*types.Index, error) bool) {

		list_opts := &s3vectors.ListIndexesInput{
			VectorBucketName: aws.String(db.bucket),
		}

		rsp, err := db.client.ListIndexes(ctx, list_opts)

		if err != nil {
			if !yield(nil, err) {
				return
			}
		}

		for _, idx_meta := range rsp.Indexes {

			rsp, err := db.client.GetIndex(ctx, &s3vectors.GetIndexInput{
				VectorBucketName: aws.String(db.bucket),
				IndexName:        idx_meta.IndexName,
			})

			var idx *types.Index

			if err == nil {
				idx = rsp.Index
			}

			if !yield(idx, err) {
				return
			}
		}
	}
}

func setupS3VectorsBucketAndIndex(ctx context.Context, cl *s3vectors.Client, bucket string, index string, dimensions int) error {

	// Maybe move in to aaronland/go-aws/s3vectors? TBD...

	logger := slog.Default()
	logger = logger.With("bucket", bucket)
	logger = logger.With("index", index)

	_, err := cl.GetVectorBucket(ctx, &s3vectors.GetVectorBucketInput{
		VectorBucketName: aws.String(bucket),
	})

	if err != nil {

		var notFound *types.NotFoundException

		if errors.As(err, &notFound) {
			logger.Debug("Bucket not found, creating")

			_, err = cl.CreateVectorBucket(ctx, &s3vectors.CreateVectorBucketInput{
				VectorBucketName: aws.String(bucket),
			})

			if err != nil {
				logger.Error("Failed to create bucket", "error", err)
				return err
			}
		} else {
			logger.Error("Failed to check bucket", "error", err)
			return err
		}
	}

	_, err = cl.GetIndex(ctx, &s3vectors.GetIndexInput{
		VectorBucketName: aws.String(bucket),
		IndexName:        aws.String(index),
	})

	if err != nil {

		var notFound *types.NotFoundException

		if errors.As(err, &notFound) {

			logger.Debug("Index not found, creating", "dimensions", dimensions)

			_, err = cl.CreateIndex(ctx, &s3vectors.CreateIndexInput{
				DataType:         types.DataTypeFloat32,
				VectorBucketName: aws.String(bucket),
				IndexName:        aws.String(index),
				Dimension:        aws.Int32(int32(dimensions)),
				DistanceMetric:   types.DistanceMetricCosine, // https://pkg.go.dev/github.com/aws/aws-sdk-go-v2/service/s3vectors@v1.6.7/types#DistanceMetric
			})

			if err != nil {
				logger.Error("Failed to create index", "error", err)
				return err
			}
		} else {
			logger.Error("Failed to check index", "error", err)
			return err
		}
	}

	return nil
}
