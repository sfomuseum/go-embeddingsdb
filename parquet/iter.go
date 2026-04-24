package parquet

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

// To do: Add filtering mechanism

func Iterate(ctx context.Context, uri string) iter.Seq2[*embeddingsdb.Record, error] {

	return func(yield func(*embeddingsdb.Record, error) bool) {

		switch {
		case strings.HasPrefix(uri, "http"):

			u, err := url.Parse(uri)

			if err != nil {
				yield(nil, err)
				return
			}

			for row, err := range IterateRemote(ctx, u) {
				if !yield(row, err) {
					return
				}
			}

		default:

			abs_path, err := filepath.Abs(uri)

			if err != nil {
				yield(nil, err)
				return
			}

			r, err := os.Open(abs_path)

			if err != nil {
				yield(nil, err)
				return
			}

			defer r.Close()

			for row, err := range IterateReader(ctx, r) {

				if !yield(row, err) {
					return
				}
			}
		}
	}
}

func IterateReader(ctx context.Context, r io.ReaderAt) iter.Seq2[*embeddingsdb.Record, error] {

	logger := slog.Default()

	batch_size := 100

	return func(yield func(*embeddingsdb.Record, error) bool) {

		parquet_r := parquet_go.NewGenericReader[*embeddingsdb.Record](r)
		buf := make([]*embeddingsdb.Record, batch_size)

		for {

			select {
			case <-ctx.Done():
				logger.Debug("Context signaled done, exiting")
				return
			default:

				n, err := parquet_r.Read(buf)

				switch {
				case err == io.EOF:
					return
				case err != nil:
					yield(nil, err)
					return
				}

				for _, rec := range buf[:n] {
					if !yield(rec, nil) {
						return
					}
				}
			}
		}
	}
}

func IterateRemote(ctx context.Context, uri *url.URL) iter.Seq2[*embeddingsdb.Record, error] {

	logger := slog.Default()

	return func(yield func(*embeddingsdb.Record, error) bool) {

		db, err := sql.Open("duckdb", "")

		if err != nil {
			yield(nil, err)
			return
		}

		defer db.Close()

		q := fmt.Sprintf(`SELECT provider, depiction_id, subject_id, model, embeddings, created, CAST(TO_JSON(attributes) AS VARCHAR) AS attributes FROM read_parquet('%s')`, uri.String())

		rows, err := db.QueryContext(ctx, q)

		if err != nil {
			yield(nil, err)
			return
		}

		defer rows.Close()

		for rows.Next() {

			select {
			case <- ctx.Done():
				logger.Debug("Context signaled done, exiting")
				return
			default:
				row, err := database.InflateDuckDBRecord(ctx, rows)
				
				if !yield(row, err) {
					return
				}
			}
		}

		err = rows.Close()

		if err != nil {
			yield(nil, err)
			return
		}

		err = rows.Err()

		if err != nil {
			yield(nil, err)
		}
	}

}
