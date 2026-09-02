package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/grpc"
	"github.com/sfomuseum/go-embeddingsdb/options"
	"google.golang.org/grpc/peer"
)

// grpcService implements the grpc.EmbeddingsDBServiceServer interface.
// It forwards each RPC request to the underlying database.Database instance
// that was provided when the service was constructed.
type grpcService struct {
	grpc.EmbeddingsDBServiceServer
	db database.Database
}

// AddRecord adds a new embeddings record to the database.
// The request must contain a valid grpc.EmbeddingsDBRecord.  The
// method logs the time taken to insert the record and returns
// an empty AddRecordResponse on success.
func (s *grpcService) AddRecord(ctx context.Context, req *grpc.AddRecordRequest) (*grpc.AddRecordResponse, error) {

	logger := s.Logger(ctx)
	t1 := time.Now()

	defer func() {
		logger.Debug("Time to add record", "time", time.Since(t1))
	}()

	record := embeddingsdb.GrpcEmbeddingsRecordToEmbeddingsDBRecord(req.Record)

	logger = logger.With("provider", record.Provider)
	logger = logger.With("depiction_id", record.DepictionId)
	logger = logger.With("model", record.Model)

	_, err := s.db.AddRecord(ctx, record)

	if err != nil {
		logger.Error("Failed to add record", "error", err)
		return nil, err
	}

	rsp := &grpc.AddRecordResponse{}
	return rsp, nil
}

// GetRecord retrieves a single record identified by provider, depiction_id and model.
// It logs the time taken to perform the lookup and returns the
// corresponding grpc.EmbeddingsDBRecord wrapped in a GetRecordResponse.
func (s *grpcService) GetRecord(ctx context.Context, req *grpc.GetRecordRequest) (*grpc.GetRecordResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("provider", req.Provider)
	logger = logger.With("depiction_id", req.DepictionId)
	logger = logger.With("model", req.Model)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to get record", "time", time.Since(t1))
	}()

	db_req := &embeddingsdb.GetRecordRequest{
		Provider:    req.Provider,
		DepictionId: req.DepictionId,
		Model:       req.Model,
	}

	record, err := s.db.GetRecord(ctx, db_req)

	if err != nil {
		logger.Error("Failed to get record", "error", err)
		return nil, err
	}

	grpc_record := embeddingsdb.EmbeddingsDBRecordToGrpcEmbeddingsDBRecord(record)

	rsp := &grpc.GetRecordResponse{
		Record: grpc_record,
	}

	return rsp, nil
}

// RemoveRecord deletes the specified record from the database.
// The method logs the time taken to perform the deletion and
// returns an empty RemoveRecordResponse on success.
func (s *grpcService) RemoveRecord(ctx context.Context, req *grpc.RemoveRecordRequest) (*grpc.RemoveRecordResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("provider", req.Provider)
	logger = logger.With("depiction_id", req.DepictionId)
	logger = logger.With("model", req.Model)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to remove record", "time", time.Since(t1))
	}()

	db_req := &embeddingsdb.RemoveRecordRequest{
		Provider:    req.Provider,
		DepictionId: req.DepictionId,
		Model:       req.Model,
	}

	err := s.db.RemoveRecord(ctx, db_req)

	if err != nil {
		logger.Error("Failed to record record", "error", err)
		return nil, err
	}

	rsp := &grpc.RemoveRecordResponse{}
	return rsp, nil
}

// ListRecords returns a paginated list of records that match the supplied
// filters and pagination options.  The response includes the pagination
// metadata and a slice of grpc.EmbeddingsDBRecord objects.
func (s *grpcService) ListRecords(ctx context.Context, req *grpc.ListRecordsRequest) (*grpc.ListRecordsResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to list records", "time", time.Since(t1))
	}()

	pg_opts, err := countable.NewCountableOptions()

	if err != nil {
		logger.Error("Failed to create new countable options", "error", err)
		return nil, err
	}

	pg_opts.PerPage(req.Pagination.PerPage)
	pg_opts.Pointer(req.Pagination.Page)

	opts := make([]options.Option, len(req.Filters))

	for i, f := range req.Filters {
		opts[i] = options.NewFilterOption(f.Column, f.Value)
	}

	db_records, pg_rsp, err := s.db.ListRecords(ctx, pg_opts, opts...)

	if err != nil {
		logger.Error("Failed to list records", "error", err)
		return nil, err
	}

	grpc_records := make([]*grpc.EmbeddingsDBRecord, len(db_records))

	for i, r := range db_records {
		grpc_records[i] = embeddingsdb.EmbeddingsDBRecordToGrpcEmbeddingsDBRecord(r)
	}

	rsp := &grpc.ListRecordsResponse{
		Pagination: &grpc.PaginationResults{
			Total:   pg_rsp.Total(),
			Page:    pg_rsp.Page(),
			Pages:   pg_rsp.Pages(),
			PerPage: pg_rsp.PerPage(),
		},
		Records: grpc_records,
	}

	return rsp, nil
}

// SimilarRecords returns a list of records that are similar to the supplied
// embeddings vector.  The request can optionally restrict the search to a
// specific provider, set a maximum distance threshold, or limit the number
// of results.  The method logs the time taken to perform the query and
// returns the matched records in a SimilarRecordsResponse.
func (s *grpcService) SimilarRecords(ctx context.Context, req *grpc.SimilarRecordsRequest) (*grpc.SimilarRecordsResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("model", req.Model)

	if req.SimilarProvider != nil {
		logger = logger.With("provider", *req.SimilarProvider)
	}

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to retrieve similar records", "time", time.Since(t1))
	}()

	db_req := &embeddingsdb.SimilarRecordsRequest{
		Model:      req.Model,
		Embeddings: req.Embeddings,
		Exclude:    req.Exclude,
	}

	opts := make([]options.Option, 0)

	if req.SimilarProvider != nil {
		o := options.NewSimilarProviderOption(*req.SimilarProvider)
		opts = append(opts, o)
	}

	if req.MaxDistance != nil {
		o := options.NewMaxDistanceOption(*req.MaxDistance)
		opts = append(opts, o)
	}

	if req.MaxResults != nil {
		o := options.NewMaxResultsOption(*req.MaxResults)
		opts = append(opts, o)
	}

	records, err := s.db.SimilarRecords(ctx, db_req, opts...)

	if err != nil {
		logger.Error("Failed to retrieve similar records", "error", err)
		return nil, err
	}

	logger = logger.With("count", len(records))

	grpc_records := embeddingsdb.EmbeddingsDBSimilarRecordsToGrpcSimilarRecords(records)

	rsp := &grpc.SimilarRecordsResponse{
		Records: grpc_records,
	}

	return rsp, nil
}

// SimilarRecordsById first fetches the record identified by the supplied
// provider, depiction_id and model, then performs a SimilarRecords query
// using that record’s embeddings vector.
func (s *grpcService) SimilarRecordsById(ctx context.Context, req *grpc.SimilarRecordsByIdRequest) (*grpc.SimilarRecordsResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("provider", req.Provider)
	logger = logger.With("depiction_id", req.DepictionId)
	logger = logger.With("model", req.Model)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to retrieve similar records by ID", "time", time.Since(t1))
	}()

	record_req := &embeddingsdb.GetRecordRequest{
		Provider:    req.Provider,
		DepictionId: req.DepictionId,
		Model:       req.Model,
	}

	record, err := s.db.GetRecord(ctx, record_req)

	if err != nil {
		logger.Error("Failed to get record", "error", err)
		return nil, err
	}

	similar_req := &grpc.SimilarRecordsRequest{
		Model:      record.Model,
		Embeddings: record.Embeddings,
		Exclude: []string{
			record.DepictionId,
		},
		SimilarProvider: req.SimilarProvider,
		MaxDistance:     req.MaxDistance,
		MaxResults:      req.MaxResults,
	}

	return s.SimilarRecords(ctx, similar_req)
}

// GetModels retrieves the list of distinct model names available in the database.
// The request may filter by one or more providers.
func (s *grpcService) GetModels(ctx context.Context, req *grpc.GetModelsRequest) (*grpc.GetModelsResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to list models", "time", time.Since(t1))
	}()

	opts := make([]options.Option, len(req.Provider))

	for i, p := range req.Provider {
		opts[i] = options.NewProviderOption(p)
	}

	models, err := s.db.Models(ctx, opts...)

	if err != nil {
		logger.Error("Failed to list models", "error", err)
		return nil, err
	}

	rsp := &grpc.GetModelsResponse{
		Model: models,
	}

	return rsp, nil
}

// GetProviders returns a list of distinct provider names stored in the database.
func (s *grpcService) GetProviders(ctx context.Context, req *grpc.GetProvidersRequest) (*grpc.GetProvidersResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to list providers", "time", time.Since(t1))
	}()

	providers, err := s.db.Providers(ctx)

	if err != nil {
		logger.Error("Failed to list providers", "error", err)
		return nil, err
	}

	rsp := &grpc.GetProvidersResponse{
		Provider: providers,
	}

	return rsp, nil
}

// GetDimensions retrieves the dimensionality of the embeddings vectors stored in the database.
func (s *grpcService) GetDimensions(ctx context.Context, req *grpc.GetDimensionsRequest) (*grpc.GetDimensionsResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to list dimensions", "time", time.Since(t1))
	}()

	// options...

	dimensions, err := s.db.Dimensions(ctx)

	if err != nil {
		logger.Error("Failed to derive dimensions", "error", err)
		return nil, err
	}

	d32 := make([]int32, len(dimensions))

	for i, d := range dimensions {
		d32[i] = int32(d)
	}

	rsp := &grpc.GetDimensionsResponse{
		Dimensions: d32,
	}

	return rsp, nil
}

// GetPaginationType returns the pagination strategy used by the database
// implementation (e.g., cursor, offset).
func (s *grpcService) GetPaginationType(ctx context.Context, req *grpc.GetPaginationTypeRequest) (*grpc.GetPaginationTypeResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()

	defer func() {
		logger.Debug("Time to get pagination type", "time", time.Since(t1))
	}()

	// options...

	pg_type, err := s.db.PaginationType(ctx)

	if err != nil {
		return nil, err
	}

	rsp := &grpc.GetPaginationTypeResponse{
		PaginationType: pg_type.String(),
	}

	return rsp, nil
}

// Logger returns a slog.Logger instance that includes the remote address
// of the client (if available) as a structured log field.  The logger
// is used by all RPC methods to provide consistent logging output.
func (s *grpcService) Logger(ctx context.Context) *slog.Logger {

	logger := slog.Default()

	p, ok := peer.FromContext(ctx)

	if ok {
		logger = logger.With("remote address", p.Addr.String())
	}

	return logger
}
