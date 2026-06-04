package verify

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/duckdb/duckdb-go/v2"

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

	db, err := sql.Open("duckdb", "")

	if err != nil {
		return fmt.Errorf("Failed to open DuckDB, %w", err)
	}

	defer db.Close()

	verifier, err := signatures.NewVerifier(ctx, verifier_uri)

	if err != nil {
		return fmt.Errorf("Failed to create new verifier, %w", err)
	}

	sigs := make([]string, len(signature_files))

	for i := 0; i < len(signature_files); i++ {
		sigs[i] = fmt.Sprintf("'%s'", signature_files[i])
	}

	str_sigs := strings.Join(sigs, ",")

	wg := new(sync.WaitGroup)

	throttle := make(chan bool, workers)

	for i := 0; i < workers; i++ {
		throttle <- true
	}

	count := int64(0)
	errors := int64(0)
	missing := int64(0)
	invalid := int64(0)
	valid := int64(0)

	done_ch := make(chan bool)

	report_metrics := func(msg string) {
		slog.Info(msg, "count", atomic.LoadInt64(&count), "valid", atomic.LoadInt64(&valid), "invalid", atomic.LoadInt64(&invalid), "missing", atomic.LoadInt64(&missing), "errors", atomic.LoadInt64(&errors))
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

	uris := fs.Args()

	for rec, err := range parquet.Iterate(ctx, uris...) {

		if err != nil {
			return fmt.Errorf("Iterator yield an error, %w", err)
		}

		<-throttle

		key := rec.Key()
		index := atomic.AddInt64(&count, 1)

		logger := slog.Default()
		logger = logger.With("index", index)
		logger = logger.With("record", key)

		// Note how we are doing all the things we might need
		// to do with 'rec' before invoking the Go routine lest
		// the iterating code get confused

		hash, err := rec.Hash()

		if err != nil {
			atomic.AddInt64(&errors, 1)
			logger.Error("Failed to hash record", "error", err)
			continue

		}

		enc, err := json.Marshal(rec)

		if err != nil {
			atomic.AddInt64(&errors, 1)
			logger.Error("Failed to marshal record", "error", err)
			continue
		}

		wg.Go(func() {

			defer func() {
				throttle <- true
			}()

			logger := slog.Default()
			logger = logger.With("record", key)

			q := fmt.Sprintf("SELECT record_signature FROM read_parquet(%s) WHERE record_hash = ?", str_sigs)

			row := db.QueryRowContext(ctx, q, hash)

			var sig []byte

			err = row.Scan(&sig)

			if err != nil {

				if err == sql.ErrNoRows {
					logger.Debug("No rows for record")
					atomic.AddInt64(&missing, 1)
				} else {
					logger.Error("Failed to scan signature", "error", err)
					atomic.AddInt64(&errors, 1)
				}

				return
			}

			ok, err := verifier.Verify(ctx, enc, sig)

			if err != nil {
				atomic.AddInt64(&errors, 1)
				logger.Error("Failed to verify signature", "error", err)
				return
			}

			if !ok {
				atomic.AddInt64(&invalid, 1)
				logger.Warn("Invalid signature")
				return
			}

			atomic.AddInt64(&valid, 1)
			logger.Debug("Verified")
		})
	}

	wg.Wait()
	done_ch <- true

	report_metrics("Completed")
	return nil
}
