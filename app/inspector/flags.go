package inspector

import (
	"flag"

	"github.com/sfomuseum/go-flags/flagset"
)

var server_uri string
var database_uri string

var enable_uploads bool
var embeddings_client_uri string
var max_upload_size int64

var max_results int

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("inspect")

	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "A registered aaronland/go-http/v4/server.Server URI.")
	fs.StringVar(&database_uri, "database-uri", "", "A registered sfomuseum/go-embeddingsdb/database.Database URI.")
	fs.IntVar(&max_results, "max-results", 20, "The maximum number of similar results to return.")

	fs.BoolVar(&enable_uploads, "enable-uploads", false, "Enable search by upload functionality.")
	fs.StringVar(&embeddings_client_uri, "embeddings-client-uri", "", "A registered go-embeddings.Client URI. This is required if the -enable-uploads flag is true.")

	// https://github.com/gangleri/humanbytes/blob/master/humanbytes.go
	fs.Int64Var(&max_upload_size, "max-upload-size", 10*1024*1024, "...")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	return fs
}
