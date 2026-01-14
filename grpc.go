package embeddingsdb

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/url"
	"strconv"

	embeddingsdb_grpc "github.com/sfomuseum/go-embeddingsdb/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// GrpcEmbeddingsDBClient implements the [EmbeddingsDBClient] interface for a gRPC-based embeddings database.
type GrpcEmbeddingsDBClient struct {
	conn   *grpc.ClientConn
	client embeddingsdb_grpc.EmbeddingsDBServiceClient
}

func init() {
	ctx := context.Background()
	RegisterEmbeddingsDBClient(ctx, "grpc", NewGrpcEmbeddingsDBClient)
}

// NewGrpcEmbeddingsDBClient will return a new [GrpcEmbeddingsDBClient] instance implementing the [EmbeddingsDBClient] interface
// derived from 'uri' which is expected to take the port of:
//
//	grpc://{HOST}:{PORT}?{PARAMETERS}

// Where {PARAMETERS} may be one or more of the following:
func NewGrpcEmbeddingsDBClient(ctx context.Context, uri string) (EmbeddingsDBClient, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	q_tls_cert := q.Get("tls-certificate")
	q_tls_key := q.Get("tls-key")
	q_tls_ca := q.Get("tls-ca-certificate")
	q_tls_insecure := q.Get("tls-insecure")

	addr := u.Host

	opts := make([]grpc.DialOption, 0)

	if q_tls_cert != "" && q_tls_key != "" {

		cert, err := tls.LoadX509KeyPair(q_tls_cert, q_tls_key)

		if err != nil {
			return nil, fmt.Errorf("Failed to load TLS pair, %w", err)
		}

		tls_config := &tls.Config{
			Certificates: []tls.Certificate{cert},
		}

		if q_tls_ca != "" {

			ca_cert, err := ioutil.ReadFile(q_tls_ca)

			if err != nil {
				return nil, fmt.Errorf("Failed to create CA certificate, %w", err)
			}

			cert_pool := x509.NewCertPool()

			ok := cert_pool.AppendCertsFromPEM(ca_cert)

			if !ok {
				return nil, fmt.Errorf("Failed to append CA certificate, %w", err)
			}

			tls_config.RootCAs = cert_pool

		} else if q_tls_insecure != "" {

			v, err := strconv.ParseBool(q_tls_insecure)

			if err != nil {
				return nil, fmt.Errorf("Failed to parse ?tls-insecure= parameter, %w", err)
			}

			tls_config.InsecureSkipVerify = v
		}

		tls_credentials := credentials.NewTLS(tls_config)
		opts = append(opts, grpc.WithTransportCredentials(tls_credentials))

	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	conn, err := grpc.NewClient(addr, opts...)

	if err != nil {
		return nil, fmt.Errorf("Failed to dial '%s', %w", addr, err)
	}

	client := embeddingsdb_grpc.NewEmbeddingsDBServiceClient(conn)

	e := &GrpcEmbeddingsDBClient{
		conn:   conn,
		client: client,
	}

	return e, nil
}

// AddRecord adds 'record' to a gRPC-backed embeddings database.
func (e *GrpcEmbeddingsDBClient) AddRecord(ctx context.Context, record *Record) error {

	db_record := e.recordToEmbeddingsDBRecord(record)

	req := &embeddingsdb_grpc.AddRecordRequest{
		Record: db_record,
	}

	_, err := e.client.AddRecord(ctx, req)

	if err != nil {
		return fmt.Errorf("Failed to add record, %w", err)
	}

	return nil
}

// GetRecord retrieves the record matching 'provider', 'depiction_id' and 'model' from a gRPC-backed embeddings database.
func (e *GrpcEmbeddingsDBClient) GetRecord(ctx context.Context, provider string, depiction_id string, model string) (*Record, error) {

	req := &embeddingsdb_grpc.GetRecordRequest{
		Provider:    provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	rsp, err := e.client.GetRecord(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive embeddings, %w", err)
	}

	return e.embeddingsRecordToRecord(rsp.Record), nil
}

// SimilarRecords retrieves records with embeddings similar to those defined in 'req' from a gRPC-backed embeddings database.
func (e *GrpcEmbeddingsDBClient) SimilarRecords(ctx context.Context, req *SimilarRequest) ([]*SimilarResult, error) {

	grpc_req := &embeddingsdb_grpc.SimilarRecordsRequest{
		Model:      req.Model,
		Embeddings: req.Embeddings,
	}

	if req.Provider != nil {
		grpc_req.Provider = req.Provider
	}

	rsp, err := e.client.SimilarRecords(ctx, grpc_req)

	if err != nil {
		return nil, fmt.Errorf("Failed to query records, %w", err)
	}

	return e.embeddingsSimilarResultsToSimilarResults(rsp.Records), nil
}

// SimilarRecordsById retrieves records with embeddings similar to those for the record matching 'provider', 'depiction_id' and 'model' from a gRPC-backed embeddings database.
func (e *GrpcEmbeddingsDBClient) SimilarRecordsById(ctx context.Context, provider string, depiction_id string, model string) ([]*SimilarResult, error) {

	req := &embeddingsdb_grpc.SimilarRecordsByIdRequest{
		Provider:    &provider,
		DepictionId: depiction_id,
		Model:       model,
	}

	rsp, err := e.client.SimilarRecordsById(ctx, req)

	if err != nil {
		return nil, fmt.Errorf("Failed to query records, %w", err)
	}

	return e.embeddingsSimilarResultsToSimilarResults(rsp.Records), nil
}

func (e *GrpcEmbeddingsDBClient) embeddingsSimilarResultsToSimilarResults(records []*embeddingsdb_grpc.SimilarRecord) []*SimilarResult {

	count := len(records)
	results := make([]*SimilarResult, count)

	for idx, rec := range records {

		qr := &SimilarResult{
			Provider:    rec.Provider,
			DepictionId: rec.DepictionId,
			SubjectId:   rec.SubjectId,
			Similarity:  rec.Similarity,
		}

		results[idx] = qr
	}

	return results
}

func (e *GrpcEmbeddingsDBClient) recordToEmbeddingsDBRecord(record *Record) *embeddingsdb_grpc.EmbeddingsDBRecord {

	db_rec := &embeddingsdb_grpc.EmbeddingsDBRecord{}

	return db_rec
}

func (e *GrpcEmbeddingsDBClient) embeddingsRecordToRecord(db_record *embeddingsdb_grpc.EmbeddingsDBRecord) *Record {

	rec := &Record{
		Provider:    db_record.Provider,
		DepictionId: db_record.DepictionId,
		SubjectId:   db_record.SubjectId,
		Model:       db_record.Model,
		Embeddings:  db_record.Embeddings,
		Attributes:  db_record.Attributes,
		Created:     db_record.Created,
	}

	return rec
}
