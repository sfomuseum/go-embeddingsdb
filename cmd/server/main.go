package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/sfomuseum/go-embeddingsdb/database"
	"github.com/sfomuseum/go-embeddingsdb/server"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

const database_placeholder string = "{database}"
const token_placeholder string = "{token}"

func main() {

	var server_uri string
	var database_uris multi.MultiString
	var token_uri string
	var verbose bool

	server_uri_default := fmt.Sprintf("grpc://localhost:8081?database-uri=%s&token-uri=%s", database_placeholder, token_placeholder)

	database_uri_desc := fmt.Sprintf("Zero or more optional values which be used to replace the '%s' placeholder, if present, in the -server-uri flag. These are expected to be a registered sfomuseum/go-embeddingsdb/database.Database URI strings. If multiple then each database URI is required to include a '#{DIMENSIONS} fragment for routing requests by embeddings dimensions.", database_placeholder)

	token_uri_desc := fmt.Sprintf("An optional value which be used to replace the '%s' placeholder, if present, in the -server-uri flag. This is expected to be a registered gocloud.dev/runtimevar URI that resolves to a shared authentication token.", token_placeholder)

	fs := flagset.NewFlagSet("server")

	fs.StringVar(&server_uri, "server-uri", server_uri_default, "A registered sfomuseum/go-embeddingsdb/server.EmbeddingsDBServer URI.")

	fs.Var(&database_uris, "database-uri", database_uri_desc)

	fs.StringVar(&token_uri, "token-uri", "", token_uri_desc)
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Start a network-based server for managing embeddings.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	swap_database := strings.Contains(server_uri, database_placeholder)
	swap_token := strings.Contains(server_uri, token_placeholder)

	if swap_database || swap_token {

		server_u, err := url.Parse(server_uri)

		if err != nil {
			log.Fatalf("Failed to parse server URI, %v", err)
		}

		server_q := server_u.Query()

		if swap_database {

			var database_uri string

			switch len(database_uris) {
			case 0:
				log.Fatal("Missing database uri(s)")
			case 1:
				database_uri = database_uris[0]
			default:
				database_uri = database.NewMultiDatabaseURIFromURIs(database_uris...)
			}

			server_q.Del("database-uri")
			server_q.Set("database-uri", database_uri)
		}

		if swap_token {
			server_q.Del("token-uri")
			server_q.Set("token-uri", token_uri)
		}

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
