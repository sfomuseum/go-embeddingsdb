package inspector

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var server_uri string
var client_uri string

var enable_uploads bool
var embeddings_client_uri string
var max_upload_size int64

var max_results int

var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("inspect")

	fs.StringVar(&server_uri, "server-uri", "http://localhost:8080", "A registered aaronland/go-http/v4/server.Server URI.")
	fs.StringVar(&client_uri, "client-uri", "grpc://localhost:8080", "A validsfomuseum/go-embeddingsdb/client.Client URI.")
	fs.IntVar(&max_results, "max-results", 20, "The maximum number of similar results to return.")

	fs.BoolVar(&enable_uploads, "enable-uploads", false, "Enable search by upload functionality.")
	fs.StringVar(&embeddings_client_uri, "embeddings-client-uri", "", "A registered go-embeddings.Client URI. This is required if the -enable-uploads flag is true.")

	// https://github.com/gangleri/humanbytes/blob/master/humanbytes.go
	fs.Int64Var(&max_upload_size, "max-upload-size", 10*1024*1024, "The maximum size (in bytes) for uploads.")
	fs.BoolVar(&verbose, "verbose", false, "Enable verbose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "A minimalist web-interface for inspecting documents stored in a `embeddingsdb-server` instance.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
