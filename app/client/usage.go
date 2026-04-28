package client

import (
	"flag"
	"fmt"
	"os"
)

func usage() {

	fmt.Fprintf(os.Stderr, "Command-line tool for interacting with a gRPC EmbeddingsDB \"service\". Results are written as a JSON-encoded string to STDOUT.\n")
	fmt.Fprintf(os.Stderr, "Usage:\n\t%s [command] [options]\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Valid commands are:\n")
	fmt.Fprintf(os.Stderr, "* record [options]\n")
	fmt.Fprintf(os.Stderr, "* remove [options]\n")
	fmt.Fprintf(os.Stderr, "* similar-by-id [options]\n")
	fmt.Fprintf(os.Stderr, "* list [options]\n")
	fmt.Fprintf(os.Stderr, "* models [options]\n")
	fmt.Fprintf(os.Stderr, "* providers [options]\n")
	flag.PrintDefaults()

	os.Exit(1)
}
