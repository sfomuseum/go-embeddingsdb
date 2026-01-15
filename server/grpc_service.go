package server

import (
	"context"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/grpc"
)

type grpcService struct {
	grpc.EmbeddingsDBServiceServer
	db database.Database
}

func (s *grpcService) AddRecord(ctx context.Context, req *grpc.AddRecordRequest) (*grpc.AddRecordResponse, error) {

	record := embeddingsdb.GrpcEmbeddingsRecordToEmbeddingsDBRecord(req.Record)

	err := s.db.AddRecord(ctx, record)

	if err != nil {
		return nil, err
	}

	rsp := &grpc.AddRecordResponse{}
	return rsp, nil
}

func (s *grpcService) GetRecord(ctx context.Context, req *grpc.GetRecordRequest) (*grpc.GetRecordResponse, error) {

	provider := req.Provider
	depiction_id := req.DepictionId
	model := req.Model

	record, err := s.db.GetRecord(ctx, provider, depiction_id, model)

	if err != nil {
		return nil, err
	}

	grpc_record := embeddingsdb.EmbeddingsDBRecordToGrpcEmbeddingsDBRecord(record)

	rsp := &grpc.GetRecordResponse{
		Record: grpc_record,
	}

	return rsp, nil
}

func (s *grpcService) SimilarRecords(ctx context.Context, req *grpc.SimilarRecordsRequest) (*grpc.SimilarRecordsResponse, error) {

	db_req := &embeddingsdb.SimilarRequest{
		Model:           req.Model,
		Embeddings:      req.Embeddings,
		Exclude:         req.Exclude,
		SimilarProvider: req.SimilarProvider,
	}

	records, err := s.db.SimilarRecords(ctx, db_req)

	if err != nil {
		return nil, err
	}

	grpc_records := embeddingsdb.EmbeddingsDBSimilarResultsToGrpcSimilarRecords(records)

	rsp := &grpc.SimilarRecordsResponse{
		Records: grpc_records,
	}

	return rsp, nil
}

func (s *grpcService) SimilarRecordsById(ctx context.Context, req *grpc.SimilarRecordsByIdRequest) (*grpc.SimilarRecordsResponse, error) {

	provider := req.Provider
	depiction_id := req.DepictionId
	model := req.Model

	record, err := s.db.GetRecord(ctx, provider, depiction_id, model)

	if err != nil {
		return nil, err
	}

	similar_req := &grpc.SimilarRecordsRequest{
		Model:           record.Model,
		Embeddings:      record.Embeddings,
		Exclude:         []string{depiction_id},
		SimilarProvider: req.SimilarProvider,
	}

	return s.SimilarRecords(ctx, similar_req)
}
