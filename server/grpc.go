package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strconv"
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

// GrpcServer implements a gRPC‐based server for managing embeddings
// operations. The server is configured via a URI that specifies the
// host, port, and various optional query parameters (see the
// documentation of NewGrpcServer for details).  The server
// interacts with a database implementation that satisfies the
// `sfomuseum/go-embeddingsdb/database.Database` interface.
type GrpcServer struct {
	Server
	host   string
	port   string
	db_uri string
	token  *string
	cert   *tls.Certificate
	// maxRecvMsgSize sets the max message size in bytes the gRPC server can receive. If this is not set, gRPC uses the default 4MB.
	max_msg_size int
}

func init() {

	ctx := context.Background()
	err := RegisterServer(ctx, "grpc", NewGrpcServer)

	if err != nil {
		panic(err)
	}
}

// Create a gRPC-based server for managing embeddings-related operations derived from 'uri' which is expected to take the form of:
//
//	grpc://{HOST}:{ADDRESS}?{QUERY_PARAMETERS}
//
// Where {QUERY_PARAMETERS} may be:
// * `database-uri` – A registered `sfomuseum/go-embeddingsdb/database.Database` URI for the underlying database implementation to use. (required)
// * `token-uri` – A registered `gocloud.dev/runtimevar` URI used to stored a shared authentication to require with client requests.
// * `tls-certificate` – The path to a valid TLS certificate to use for encrypted connections.
// * `tls-key` – The path to a valid TLS key file to use for encrypted connections.
// * `max-msg-size` - The maximum message size in bytes the gRPC server can receive. Default is 4MB.
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

		token = strings.TrimSpace(token)
		s.token = &token
	}

	if q.Has("tls-certificate") && q.Has("tls-key") {

		slog.Debug("Configure TLS")

		cert_file := q.Get("tls-certificate")
		key_file := q.Get("tls-key")

		cert, err := tls.LoadX509KeyPair(cert_file, key_file)

		if err != nil {
			return nil, fmt.Errorf("Failed to load key pair, %w", err)
		}

		s.cert = &cert
	}

	if q.Has("max-msg-size") {

		v, err := strconv.Atoi(q.Get("max-msg-size"))

		if err != nil {
			return nil, fmt.Errorf("Failed to parse ?max-msg-size= parameter, %w", err)
		}

		s.max_msg_size = v
	}

	return s, nil
}

// ListenAndServe starts the gRPC server and blocks until the context is
// cancelled or an error occurs.  It creates a database instance from the
// URI supplied to NewGrpcServer, sets up optional periodic database
// export and batching logic, and then listens for incoming gRPC
// connections on the configured host/port.  If TLS credentials were
// supplied to NewGrpcServer, the server will require TLS; otherwise
// it accepts insecure connections.
func (s *GrpcServer) ListenAndServe(ctx context.Context) error {

	logger := slog.Default()

	logger.Debug("Set up database")

	db_u, err := url.Parse(s.db_uri)

	if err != nil {
		logger.Error("Failed to parse database URI", "error", err)
		return fmt.Errorf("Failed to parse database URI, %w", err)
	}

	// ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	// defer cancel()

	db, err := database.NewDatabase(ctx, s.db_uri)

	if err != nil {
		logger.Error("Failed to create database", "error", err)
		return fmt.Errorf("Failed to create database, %w", err)
	}

	db_closefunc := func() {

		logger.Debug("Received signal handler, shutting down database")

		ctx := context.Background()
		err := db.Close(ctx)

		if err != nil {
			logger.Error("Failed to close database", "error", err)
		}
	}

	defer db_closefunc()

	db_path := db_u.Path

	if db_path != "" {

		interval := 60
		logger.Debug("Set up database export timer", "path", db_path, "interval", interval)

		export_db := func() {

			count, err := db.BatchedRecordsCount(ctx)

			switch {
			case err != nil:
				logger.Warn("Failed to derive batched records count", "error", err)
			case count > 0:

				logger.Debug("Batched record count", "count", count)
				err := db.AddBatchedRecords(ctx)

				if err != nil {
					logger.Error("Failed to add batched records", "error", err)
				}
			}

			logger.Debug("Export database")
			err = db.Export(ctx, db_path)

			if err != nil {
				slog.Error("Failed to export database", "db_path", db_path, "error", err)
			}

		}

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

					if err != nil && err != sql.ErrNoRows {
						slog.Warn("Failed to determine last update from database", "error", err)
						break
					}

					do_export := false

					if last_update == 0 {

						do_export = true

					} else {

						now := t.Unix()
						diff := now - last_update

						if diff < int64(interval) {
							do_export = true
						}
					}

					if do_export {
						logger.Debug("Export database")
						export_db()
					}

				}
			}
		}()
	}

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	logger.Debug("Set up listener")

	lis, err := net.Listen("tcp", addr)

	if err != nil {
		return err
	}

	logger.Debug("Set up server")

	svc := &grpcService{
		db: db,
	}

	opts := []grpc.ServerOption{}

	if s.token != nil {
		logger.Debug("Set up token interceptor")
		opts = append(opts, grpc.UnaryInterceptor(s.ensureValidToken))
	}

	if s.cert != nil {
		logger.Debug("Set up TLS")
		opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(s.cert)))
	} else {
		logger.Debug("Allow insecure connections")
		// opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		// opts = append(opts, grpc.WithInsecure())
	}

	if s.max_msg_size > 0 {
		opts = append(opts, grpc.MaxRecvMsgSize(s.max_msg_size))
	}

	svr := grpc.NewServer(opts...)

	embeddings_grpc.RegisterEmbeddingsDBServiceServer(svr, svc)

	logger.Info("Server listening", "address", addr)

	err = svr.Serve(lis)

	if err != nil {
		logger.Error("Failed to serve requests", "error", err)
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
