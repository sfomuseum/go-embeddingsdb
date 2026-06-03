package verify

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
)

var signature_files multi.MultiString
var verbose bool
var verifier_uri string
var workers int

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("verify")

	fs.Var(&signature_files, "signatures", "One or more Parquet files containing signature data (for example, as produced by the parquet-sign tool).")
	fs.StringVar(&verifier_uri, "verifier-uri", "", "...")
	fs.IntVar(&workers, "workers", runtime.NumCPU(), "The maximum number of concurrent worker to verify records with.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Verify the PGP/GPG signatures associated with one or more go-embeddingsdb Parquet files.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file(N) parquet_file(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
