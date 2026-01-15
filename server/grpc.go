package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"time"

	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/grpc"
	core "google.golang.org/grpc"
)

type GrpcServer struct {
	Server
	host   string
	port   string
	db_uri string
}

func init() {

	ctx := context.Background()
	err := RegisterServer(ctx, "grpc", NewGrpcServer)

	if err != nil {
		panic(err)
	}
}

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

	svr := core.NewServer()
	grpc.RegisterEmbeddingsDBServiceServer(svr, svc)

	slog.Info("Server listening", "address", lis.Addr())
	err = svr.Serve(lis)

	if err != nil {
		return err
	}

	return nil
}
