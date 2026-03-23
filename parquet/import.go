package parquet

import (
	"context"
	"io"
	_ "log/slog"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

func Import(ctx context.Context, db database.Database, r io.ReaderAt) error {

	parquet_r := parquet_go.NewGenericReader[*embeddingsdb.Record](r)

	rows := make([]*embeddingsdb.Record, 0, 10000)

	for {

		n, err := parquet_r.Read(rows[:cap(rows)])

		if err != nil {

			if err == io.EOF {
				break
			}

			return err
		}

		rows = rows[:n]

		for _, row := range rows {

			err := db.AddRecord(ctx, row)

			if err != nil {
				return err
			}

			// slog.Info("Added record", "depiction_id", row.DepictionId)
		}
	}

	return nil

}
