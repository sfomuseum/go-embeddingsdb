package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/crypto"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var output string
	var key_uri string
	var pswd_uri string

	var verify bool
	var verbose bool

	fs := flagset.NewFlagSet("emit")

	fs.StringVar(&key_uri, "key-uri", "", "A registered gocloud.dev/runtimevar URI which is expected to resolve to an ASCII‑armored key block.")
	fs.StringVar(&pswd_uri, "password-uri", "", "A registered gocloud.dev/runtimevar URI which is expected to resolve to the key's password. This is only necessary if the key is locked and, as such, may be left empty.")
	fs.StringVar(&output, "output", "", "The path where Parquet-encoded data should be written. If \"-\" then data will be written to STDOUT.")

	fs.BoolVar(&verify, "verify", true, "Verify signature before recording.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Generate a corresponding Parquet \"signature\" file with detached PGP/GPG signatures for embeddingsdb.Record records in one or more Parquet files.\n")
		fmt.Fprintf(os.Stderr, "Usage:\n\t%s [options] parquet_file(N) parquet_file(N)\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Valid options are:\n")
		fs.PrintDefaults()
	}

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	ctx := context.Background()

	key, err := crypto.LoadSigningKey(ctx, key_uri, pswd_uri)

	if err != nil {
		log.Fatalf("Failed to load key, %v", err)
	}

	signer, err := crypto.NewSigner(key)

	if err != nil {
		log.Fatalf("Failed to create signer, %v", err)
	}

	var verifier pgp_crypto.PGPVerify

	if verify {

		v, err := crypto.LoadVerificationHandlerWithKey(ctx, key)

		if err != nil {
			log.Fatalf("Failed to create verification handler, %v", err)
		}

		verifier = v
	}

	// START OF update to use gocloud.dev/blob so that we can automatically
	// generate signature (Parquet) files when a record file is added to an
	// S3 bucket (trigers, etc.)

	var wr io.WriteCloser

	switch output {
	case "-":
		wr = os.Stdout
	default:

		w, err := os.OpenFile(output, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Failed to open %s for writing, %v", output, err)
		}

		wr = w
	}

	// END OF update to use gocloud.dev/blob

	p_wr := parquet_go.NewGenericWriter[*embeddingsdb.Signature](wr)
	p_buf := make([]*embeddingsdb.Signature, 0)

	batch_size := 10000

	// START OF move this in to a "run" function (or equivalent)
	// to account for S3/Lambda trigger events

	uris := fs.Args()

	for rec, err := range parquet.Iterate(ctx, uris...) {

		if err != nil {
			log.Fatalf("Iterator yield an error, %v", err)
		}

		record_sig, err := crypto.SignRecordWithSigner(ctx, signer, rec)

		if err != nil {
			log.Fatalf("Failed to sign record %s, %v", rec.Key(), err)
		}

		if verify {

			ok, err := crypto.VerifyRecordSignatureWithVerifier(ctx, verifier, rec, record_sig)

			if err != nil {
				log.Fatalf("Failed to verify signature for record %s, %v", rec.Key(), err)
			}

			if !ok {
				log.Fatalf("Failed to verify signature for record %s, undefined error", rec.Key())
			}
		}

		sig, err := rec.Signature(record_sig)

		if err != nil {
			log.Fatalf("Failed to hash record, %v", err)
		}

		p_buf = append(p_buf, sig)

		if len(p_buf) >= batch_size {

			_, err = p_wr.Write(p_buf)

			if err != nil {
				log.Fatalf("Failed to write Parquet buffer, %v", err)
			}

			p_buf = make([]*embeddingsdb.Signature, 0)
		}

	}

	if len(p_buf) >= batch_size {

		_, err = p_wr.Write(p_buf)

		if err != nil {
			log.Fatalf("Failed to write final Parquet buffer, %v", err)
		}
	}

	p_wr.Flush()

	err = p_wr.Close()

	if err != nil {
		log.Fatalf("Failed to close Parquet writer, %v", err)
	}

	switch output {
	case "-":
		// pass
	default:

		err = wr.Close()

		if err != nil {
			log.Fatalf("Failed to close %s after writing, %v", err)
		}
	}

	// END OF move this in to a "run" function (or equivalent)
}
