package server

import (
	"context"
	"fmt"
)

type GrpcServer struct {
	Server
}

func NewGrpcServer(ctx context.Context, uri string) (Server, error) {

	s := &GrpcServer{}

	return s, nil
}

func (s *GrpcServer) ListenAndServe(ctx context.Context) error {
	return fmt.Errorf("Not implemented")
}
