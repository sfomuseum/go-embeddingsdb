package server

import (
	"context"
	"log/slog"
	"time"

	"github.com/aaronland/go-pagination/countable"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/grpc"
	"google.golang.org/grpc/peer"
)

type grpcService struct {
	grpc.EmbeddingsDBServiceServer
	db database.Database
}

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

	err := s.db.AddRecord(ctx, record)

	if err != nil {
		logger.Error("Failed to add record", "error", err)
		return nil, err
	}

	rsp := &grpc.AddRecordResponse{}
	return rsp, nil
}

func (s *grpcService) GetRecord(ctx context.Context, req *grpc.GetRecordRequest) (*grpc.GetRecordResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("provider", req.Provider)
	logger = logger.With("depiction_id", req.DepictionId)
	logger = logger.With("model", req.Model)

	t1 := time.Now()
	defer logger.Debug("Time to get record", "time", time.Since(t1))

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

func (s *grpcService) RemoveRecord(ctx context.Context, req *grpc.RemoveRecordRequest) (*grpc.RemoveRecordResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("provider", req.Provider)
	logger = logger.With("depiction_id", req.DepictionId)
	logger = logger.With("model", req.Model)

	t1 := time.Now()
	defer logger.Debug("Time to remove record", "time", time.Since(t1))

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

	filters := make([]*database.ListRecordsFilter, len(req.Filters))

	for i, f := range req.Filters {

		filters[i] = &database.ListRecordsFilter{
			Column: f.Column,
			Value:  f.Value,
		}
	}

	db_records, pg_rsp, err := s.db.ListRecords(ctx, pg_opts, filters...)

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
		Model:           req.Model,
		Embeddings:      req.Embeddings,
		Exclude:         req.Exclude,
		SimilarProvider: req.SimilarProvider,
		MaxDistance:     req.MaxDistance,
		MaxResults:      req.MaxResults,
	}

	records, err := s.db.SimilarRecords(ctx, db_req)

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

func (s *grpcService) SimilarRecordsById(ctx context.Context, req *grpc.SimilarRecordsByIdRequest) (*grpc.SimilarRecordsResponse, error) {

	logger := s.Logger(ctx)
	logger = logger.With("provider", req.Provider)
	logger = logger.With("depiction_id", req.DepictionId)
	logger = logger.With("model", req.Model)

	t1 := time.Now()
	defer logger.Debug("Time to retrieve similar records by ID", "time", time.Since(t1))

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

func (s *grpcService) GetModels(ctx context.Context, req *grpc.GetModelsRequest) (*grpc.GetModelsResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()
	defer logger.Debug("Time to list models", "time", time.Since(t1))

	models, err := s.db.Models(ctx, req.Provider...)

	if err != nil {
		logger.Error("Failed to list models", "error", err)
		return nil, err
	}

	rsp := &grpc.GetModelsResponse{
		Model: models,
	}

	return rsp, nil
}

func (s *grpcService) GetProviders(ctx context.Context, req *grpc.GetProvidersRequest) (*grpc.GetProvidersResponse, error) {

	logger := s.Logger(ctx)

	t1 := time.Now()
	defer logger.Debug("Time to list providers", "time", time.Since(t1))

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

func (s *grpcService) Logger(ctx context.Context) *slog.Logger {

	logger := slog.Default()

	p, ok := peer.FromContext(ctx)

	if ok {
		logger = logger.With("remote address", p.Addr.String())
	}

	return logger
}
