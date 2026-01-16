package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"strings"

	"github.com/sfomuseum/go-embeddingsdb/server"
	"github.com/sfomuseum/go-flags/flagset"
)

const database_placeholder string = "{database}"
const token_placeholder string = "{token}"

func main() {

	var server_uri string
	var database_uri string
	var token_uri string
	var verbose bool

	server_uri_default := fmt.Sprintf("grpc://localhost:8081?database-uri=%s&token-uri={token}", database_placeholder, token_placeholder)

	database_uri_desc := fmt.Sprintf("An optional value which be used to replace the '%s' placeholder, if present, in the -server-uri flag. This is expected to be a registered sfomuseum/go-embeddingsdb/database.Database URI", database_placeholder)

	token_uri_desc := fmt.Sprintf("An optional value which be used to replace the '%s' placeholder, if present, in the -server-uri flag. This is expected to be a registered gocloud.dev/runtimevar URI that resolves to a shared authentication token.", token_placeholder)

	fs := flagset.NewFlagSet("server")

	fs.StringVar(&server_uri, "server-uri", server_uri_default, "A registered sfomuseum/go-embeddingsdb/server.EmbeddingsDBServer URI.")
	fs.StringVar(&database_uri, "database-uri", "", database_uri_desc)
	fs.StringVar(&token_uri, "token-uri", "", token_uri_desc)
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	if strings.Contains(server_uri, database_placeholder) {

		server_u, err := url.Parse(server_uri)

		if err != nil {
			log.Fatalf("Failed to parse server URI, %w", err)
		}

		server_q := server_u.Query()

		server_q.Del("database-uri")
		server_q.Set("database-uri", database_uri)

		server_u.RawQuery = server_q.Encode()
		server_uri = server_u.String()
	}

	svr, err := server.NewServer(ctx, server_uri)

	if err != nil {
		log.Fatalf("Failed to create server, %v", err)
	}

	err = svr.ListenAndServe(ctx)

	if err != nil {
		log.Fatalf("Failed to start server, %v", err)
	}
}
