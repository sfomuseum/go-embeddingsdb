package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/aaronland/gocloud/runtimevar"
	"github.com/sfomuseum/go-embeddingsdb/database"
	embeddings_grpc "github.com/sfomuseum/go-embeddingsdb/grpc"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	errMissingMetadata = status.Errorf(codes.InvalidArgument, "missing metadata")
	errInvalidToken    = status.Errorf(codes.Unauthenticated, "invalid token")
)

type GrpcServer struct {
	Server
	host   string
	port   string
	db_uri string
	token  *string
	cert   *tls.Certificate
}

func init() {

	ctx := context.Background()
	err := RegisterServer(ctx, "grpc", NewGrpcServer)

	if err != nil {
		panic(err)
	}
}

// * `database-uri`
// * `token-uri`
// * `tls-certificate`
// * `tls-key`
func NewGrpcServer(ctx context.Context, uri string) (Server, error) {

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

	s := &GrpcServer{
		host:   host,
		port:   port,
		db_uri: db_uri,
	}

	if q.Has("token-uri") {

		token, err := runtimevar.StringVar(ctx, q.Get("token-uri"))

		if err != nil {
			return nil, fmt.Errorf("Failed to derive token, %w", err)
		}

		slog.Debug("TOKEN", "t", token)
		s.token = &token
	}

	if q.Has("tls-certificate") && q.Has("tls-key") {

		cert_file := q.Get("tls-certificate")
		key_file := q.Get("tls-key")

		cert, err := tls.LoadX509KeyPair(cert_file, key_file)

		if err != nil {
			return nil, fmt.Errorf("Failed to load key pair, %w", err)
		}

		s.cert = &cert
	}

	return s, nil
}

func (s *GrpcServer) ListenAndServe(ctx context.Context) error {

	slog.Debug("Set up database")

	db_u, err := url.Parse(s.db_uri)

	if err != nil {
		return fmt.Errorf("Failed to parse database URI, %w", err)
	}

	db, err := database.NewDatabase(ctx, s.db_uri)

	if err != nil {
		return fmt.Errorf("Failed to create database, %w", err)
	}

	defer db.Close(ctx)

	db_path := db_u.Path

	if db_path != "" {

		slog.Debug("Set up database export timer", "path", db_path)

		export_db := func() {

			slog.Debug("Export database")
			err := db.Export(ctx, db_path)

			if err != nil {
				slog.Error("Failed to export database", "db_path", db_path, "error", err)
			}

		}

		interval := 60
		ticker := time.NewTicker(time.Duration(interval) * time.Second)

		defer func() {
			ticker.Stop()
			export_db()
		}()

		go func() {

			for {
				select {
				case t := <-ticker.C:

					last_update, err := db.LastUpdate(ctx)

					if err != nil {
						slog.Warn("Failed to determine last update from database", "error", err)
						break
					}

					now := t.Unix()
					diff := now - last_update

					if diff < int64(interval) {
						export_db()
					}
				}
			}
		}()
	}

	slog.Debug("Set up listener")

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", s.host, s.port))

	if err != nil {
		return err
	}

	slog.Debug("Set up server")

	svc := &grpcService{
		db: db,
	}

	opts := []grpc.ServerOption{}

	if s.token != nil {
		slog.Debug("Set up token interceptor")
		opts = append(opts, grpc.UnaryInterceptor(s.ensureValidToken))
	}

	if s.cert != nil {
		slog.Debug("Set up TLS")
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(s.cert)))
	}

	svr := grpc.NewServer(opts...)

	embeddings_grpc.RegisterEmbeddingsDBServiceServer(svr, svc)

	slog.Info("Server listening", "address", lis.Addr())
	err = svr.Serve(lis)

	if err != nil {
		return err
	}

	return nil
}

func (s *GrpcServer) ensureValidToken(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {

	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		return nil, errMissingMetadata
	}

	if !s.valid(md["authorization"]) {
		return nil, errInvalidToken
	}

	return handler(ctx, req)
}

func (s *GrpcServer) valid(authorization []string) bool {

	if len(authorization) < 1 {
		return false
	}

	token := strings.TrimPrefix(authorization[0], "Bearer ")

	if token != *s.token {
		return false
	}

	return true
}
