package database

import (
	"context"
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"net/url"

	"github.com/aaronland/go-aws/v3/auth"
	"github.com/aaronland/go-pagination"
	"github.com/aaronland/go-pagination/countable"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors"
	"github.com/aws/aws-sdk-go-v2/service/s3vectors/types"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/options"
)

type S3VectorsDatabase struct {
	Database
	uri        string
	dimensions int
	bucket     string
	client     *s3vectors.Client
}

func init() {

	ctx := context.Background()
	err := RegisterDatabase(ctx, "null", NewS3VectorsDatabase)

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

	region := q.Get("region")
	creds := q.Get("credetials")
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
	return nil, fmt.Errorf("Not found")
}

// Remove a record from an EmbeddingsDB instance.
func (db *S3VectorsDatabase) RemoveRecord(ctx context.Context, req *embeddingsdb.RemoveRecordRequest, opts ...options.Option) error {
	return nil
}

// Find similar records for a given model and record instance.
func (db *S3VectorsDatabase) SimilarRecords(ctx context.Context, rec *embeddingsdb.SimilarRecordsRequest, opts ...options.Option) ([]*embeddingsdb.SimilarRecord, error) {
	results := make([]*embeddingsdb.SimilarRecord, 0)
	return results, nil
}

// ListRecords returns a paginated list of records stored in the database.
func (db *S3VectorsDatabase) ListRecords(ctx context.Context, pg_opts pagination.Options, opts ...options.Option) ([]*embeddingsdb.Record, pagination.Results, error) {

	records := make([]*embeddingsdb.Record, 0)

	pg, err := countable.NewResultsFromCountWithOptions(pg_opts, 0)

	if err != nil {
		return nil, nil, err
	}

	return records, pg, nil
}

// IterateRecords returns an [iter.Seq2[*embeddingsdb.Record, error]] for each record stored in the database.
func (db *S3VectorsDatabase) IterateRecords(ctx context.Context, opts ...options.Option) iter.Seq2[*embeddingsdb.Record, error] {
	return func(yield func(*embeddingsdb.Record, error) bool) {}
}

// Return the Unix timestamp of the last update to the Database instance.
func (db *S3VectorsDatabase) LastUpdate(ctx context.Context, opts ...options.Option) (int64, error) {
	return 0, nil
}

// Return the list of dimensions supported by this Database  implementation.
func (db *S3VectorsDatabase) Dimensions(ctx context.Context, opts ...options.Option) []int {

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

// Close performs and terminating functions required by the database.
func (db *S3VectorsDatabase) Close(ctx context.Context) error {
	return nil
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

			logger.Debug("Index not found, creating")

			_, err = cl.CreateIndex(ctx, &s3vectors.CreateIndexInput{
				DataType:         types.DataTypeFloat32,
				VectorBucketName: aws.String(bucket),
				IndexName:        aws.String(index),
				Dimension:        aws.Int32(int32(dimensions)),
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
