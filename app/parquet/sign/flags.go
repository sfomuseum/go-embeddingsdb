package sign

import (
	"flag"
	"fmt"
	"os"

	"github.com/sfomuseum/go-flags/flagset"
)

var target_bucket_uri string
var signer_uri string
var embed_public_key bool

var verify bool
var verbose bool

func DefaultFlagSet() *flag.FlagSet {

	fs := flagset.NewFlagSet("sign")

	fs.StringVar(&signer_uri, "signer-uri", "", "A valid sfomuseum/go-embeddingsdb/signatures.Signer URI.")

	fs.StringVar(&target_bucket_uri, "target-bucket-uri", "cwd://", "The URI where signature files and public keys will be written. One of the following: A valid gocloud.dev/blob.Bucket URI; The path to a folder on the local filesystem; \"cwd://\" which will cause files to be written to the current directory.")

	fs.BoolVar(&embed_public_key, "embed-public-key", false, "If true the public key (certicate) used to verify signatures will be written to the signature Parquet file in the \"embeddingsdb:signatures:public_key\" metadata key.")

	fs.BoolVar(&verify, "verify", true, "Verify signature before recording.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Generate a corresponding Parquet \"signature\" file with PGP/GPG or TLS detached signatures for embeddingsdb.Record records in one or more Parquet files.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file(N) parquet_file(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	return fs
}
