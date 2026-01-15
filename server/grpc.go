package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"

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

	svc := &grpcService{}

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

func (s *grpcService) AddRecord(context.Context, *grpc.AddRecordRequest) (*grpc.AddRecordResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *grpcService) GetRecord(context.Context, *grpc.GetRecordRequest) (*grpc.GetRecordResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *grpcService) SimilarRecords(context.Context, *grpc.SimilarRecordsRequest) (*grpc.SimilarRecordsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *grpcService) SimilarRecordsById(context.Context, *grpc.SimilarRecordsByIdRequest) (*grpc.SimilarRecordsResponse, error) {
	return nil, fmt.Errorf("Not implemented")
}

