package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/grpc"
	core "google.golang.org/grpc"
)

type grpcService struct {
	grpc.EmbeddingsDBServiceServer
	db database.Database
}

func (s *GrpcServer) AddRecord(context.Context, *grpc.AddRecordRequest) (*grpc.AddRecordResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *GrpcServer) GetRecord(context.Context, *grpc.GetRecordRequest) (*grpc.GetRecordResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *GrpcServer) SimilarRecords(context.Context, *grpc.SimilarRecordsRequest) (*grpc.SimilarRecordsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *GrpcServer) SimilarRecordsById(context.Context, *grpc.SimilarRecordsByIdRequest) (*grpc.SimilarRecordsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

type GrpcServer struct {
	Server
	host    string
	port    int
	service grpc.EmbeddingsDBServiceServer
}

func NewGrpcServer(ctx context.Context, uri string) (Server, error) {

	s := &GrpcServer{}

	return s, nil
}

func (s *GrpcServer) ListenAndServe(ctx context.Context) error {

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))

	if err != nil {
		return err
	}

	svr := core.NewServer()
	svc := grpcService{}

	grpc.RegisterEmbeddingsDBServiceServer(svr, svc)

	slog.Info("Server listening", "address", lis.Addr())
	err = svr.Serve(lis)

	if err != nil {
		return err
	}

	return nil
}
