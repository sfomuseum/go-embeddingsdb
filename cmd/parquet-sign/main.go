package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb/crypto"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var key_uri string
	var pswd_uri string

	var verify bool
	var verbose bool

	fs := flagset.NewFlagSet("emit")

	fs.StringVar(&key_uri, "key-uri", "", "A registered gocloud.dev/runtimevar URI which is expected to resolve to an ASCII‑armored key block.")
	fs.StringVar(&pswd_uri, "password-uri", "", "A registered gocloud.dev/runtimevar URI which is expected to resolve to the key's password. This is only necessary if the key is locked and, as such, may be left empty.")

	fs.BoolVar(&verify, "verify", true, "Verify signature before recording.")
	fs.BoolVar(&verbose, "verbose", false, "Enable vebose (debug) logging.")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Generate \n")
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

	uris := fs.Args()

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

		fmt.Println(sig.Key(), sig.String())
	}
}
