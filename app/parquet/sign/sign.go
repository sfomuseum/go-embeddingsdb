package sign

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aaronland/gocloud/blob/bucket"
	"github.com/aaronland/gocloud/blob/writer"
	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/parquet"
	"github.com/sfomuseum/go-embeddingsdb/signatures"
	"github.com/sfomuseum/go-flags/flagset"
)

func Run(ctx context.Context) error {
	fs := DefaultFlagSet()
	return RunWithFlagSet(ctx, fs)
}

func RunWithFlagSet(ctx context.Context, fs *flag.FlagSet) error {

	flagset.Parse(fs)

	if verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Debug("Verbose logging enabled")
	}

	target_bucket, err := bucket.OpenBucket(ctx, target_bucket_uri)

	if err != nil {
		return fmt.Errorf("Failed to open target bucket, %v", err)
	}

	defer target_bucket.Close()

	signer, err := signatures.NewSigner(ctx, signer_uri)

	if err != nil {
		return fmt.Errorf("Failed to load new signer, %w", err)
	}

	pubkey, err := signer.PublicKey(ctx)

	if err != nil {
		return fmt.Errorf("Failed to derive public key from signer, %w", err)
	}

	var verifier signatures.Verifier

	if verify {

		v, err := signer.Verifier(ctx)

		if err != nil {
			return fmt.Errorf("Failed to create verification handler, %w", err)
		}

		verifier = v
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

	// START OF move this in to a "run" function (or equivalent)
	// to account for S3/Lambda trigger events

	uris := fs.Args()

	for _, uri := range uris {

		fname := filepath.Base(uri)
		ext := filepath.Ext(fname)

		fname_root := strings.Replace(fname, ext, "", 1)
		fname_sigs := fmt.Sprintf("%s-signatures.parquet", fname_root)
		fname_pub := fmt.Sprintf("%s-signatures.pub", fname_root)

		wr, err := writer.NewWriterWithACL(ctx, target_bucket, fname_sigs, "public-read")

		if err != nil {
			log.Fatal("Failed to create new writer for %s, %w", fname_sigs, err)
		}

		p_wr := parquet_go.NewGenericWriter[*embeddingsdb.Signature](wr)
		p_buf := make([]*embeddingsdb.Signature, 0)

		batch_size := 10000

		for rec, err := range parquet.Iterate(ctx, uri) {

			if err != nil {
				return fmt.Errorf("Iterator yield an error, %w", err)
			}

			logger := slog.Default()
			logger = logger.With("uri", uri)
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
					return fmt.Errorf("Failed to write Parquet buffer, %w", err)
				}

				p_buf = make([]*embeddingsdb.Signature, 0)
			}

			atomic.AddInt64(&completed, 1)
		}

		if len(p_buf) >= 0 {

			_, err = p_wr.Write(p_buf)

			if err != nil {
				return fmt.Errorf("Failed to write final Parquet buffer, %w", err)
			}
		}

		p_wr.Flush()

		// write pub key as metadata?

		err = p_wr.Close()

		if err != nil {
			return fmt.Errorf("Failed to close Parquet writer, %w", err)
		}

		err = wr.Close()

		if err != nil {
			return fmt.Errorf("Failed to close after writing, %w", err)
		}

		pk_wr, err := writer.NewWriterWithACL(ctx, target_bucket, fname_pub, "public-read")

		if err != nil {
			return fmt.Errorf("Failed to open writer for %s, %w", fname_pub, err)
		}

		_, err = pk_wr.Write(pubkey)

		if err != nil {
			return fmt.Errorf("Failed to write pubkey %s, %w", fname_pub, err)
		}

		err = pk_wr.Close()

		if err != nil {
			return fmt.Errorf("Failed to close pubkey %s after writing, %w", fname_pub, err)
		}
	}

	report_metrics("Completed")
	return nil
}
