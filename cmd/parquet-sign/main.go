package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"os"
	"sync/atomic"
	"time"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-embeddingsdb/signatures"
	"github.com/sfomuseum/go-flags/flagset"
)

func main() {

	var output string

	var signer_uri string

	var verify bool
	var verbose bool

	fs := flagset.NewFlagSet("emit")

	fs.StringVar(&signer_uri, "signer-uri", "", "...")

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

	signer, err := signatures.NewSigner(ctx, signer_uri)

	if err != nil {
		log.Fatalf("Failed to load new signer, %v", err)
	}

	var verifier signatures.Verifier

	if verify {

		v, err := signer.Verifier(ctx)

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

	count := int64(0)
	errors := int64(0)
	signed := int64(0)
	verified := int64(0)
	completed := int64(0)

	done_ch := make(chan bool)

	report_metrics := func(msg string) {
		slog.Info(msg, "count", atomic.LoadInt64(&count), "signed", atomic.LoadInt64(&signed), "verified", atomic.LoadInt64(&verified), "completed", atomic.LoadInt64(&completed), "errors", atomic.LoadInt64(&errors))
	}

	go func() {

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-done_ch:
				return
			case <-ticker.C:
				report_metrics("Processing")
			}
		}
	}()

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

		logger := slog.Default()
		logger = logger.With("key", rec.Key())

		atomic.AddInt64(&count, 1)

		data, err := json.Marshal(rec)

		if err != nil {
			atomic.AddInt64(&errors, 1)
			logger.Error("Failed to marshal record", "error", err)
			continue
		}

		record_sig, err := signer.Sign(ctx, data)

		if err != nil {
			atomic.AddInt64(&errors, 1)
			logger.Error("Failed to sign record", "error", err)
			continue
		}

		atomic.AddInt64(&signed, 1)

		if verify {

			ok, err := verifier.Verify(ctx, data, record_sig)

			if err != nil {
				atomic.AddInt64(&errors, 1)
				logger.Error("Failed to verify record", "error", err)
				continue
			}

			if !ok {
				atomic.AddInt64(&errors, 1)
				logger.Error("Record failed verification", "error", err)
				continue
			}

			atomic.AddInt64(&verified, 1)
		}

		sig, err := rec.Signature(record_sig)

		if err != nil {
			atomic.AddInt64(&errors, 1)
			logger.Error("Failed to create record signature", "error", err)
			continue
		}

		p_buf = append(p_buf, sig)

		if len(p_buf) >= batch_size {

			_, err = p_wr.Write(p_buf)

			if err != nil {
				log.Fatalf("Failed to write Parquet buffer, %v", err)
			}

			p_buf = make([]*embeddingsdb.Signature, 0)
		}

		atomic.AddInt64(&completed, 1)
	}

	report_metrics("Completed")

	if len(p_buf) >= 0 {

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

	// Something something something write signer.PublicKey() somewhere
	
	// END OF move this in to a "run" function (or equivalent)
}
