package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"

	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/grpc"
	core "google.golang.org/grpc"
)

type GrpcEmbeddingsDBServer struct {
	EmbeddingsDBServer
	host    string
	port    string
	service grpc.EmbeddingsDBServiceServer
}

func init() {

	ctx := context.Background()
	err := RegisterEmbeddingsDBServer(ctx, "grpc", NewGrpcEmbeddingsDBServer)

	if err != nil {
		panic(err)
	}
}

func NewGrpcEmbeddingsDBServer(ctx context.Context, uri string) (EmbeddingsDBServer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	host := u.Hostname()
	port := u.Port()

	q := u.Query()

	if !q.Has("database-uri") {
		return nil, fmt.Errorf("Missing database URI, %w", err)
	}

	db_uri := q.Get("database-uri")

	db, err := database.NewDatabase(ctx, db_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create database, %w", err)
	}

	svc := &grpcService{
		db: db,
	}

	s := &GrpcEmbeddingsDBServer{
		host:    host,
		port:    port,
		service: svc,
	}

	return s, nil
}

func (s *GrpcEmbeddingsDBServer) ListenAndServe(ctx context.Context) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.host, s.port))

	if err != nil {
		return err
	}

	svr := core.NewServer()

	grpc.RegisterEmbeddingsDBServiceServer(svr, s.service)

	slog.Info("Server listening", "address", lis.Addr())
	err = svr.Serve(lis)

	if err != nil {
		return err
	}

	return nil
}

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

func (s *grpcService) SimilarRecords(context.Context, *grpc.SimilarRecordsRequest) (*grpc.SimilarRecordsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *grpcService) SimilarRecordsById(context.Context, *grpc.SimilarRecordsByIdRequest) (*grpc.SimilarRecordsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}
