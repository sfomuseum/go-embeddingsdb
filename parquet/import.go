package parquet

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/client"
)

func Import(ctx context.Context, cl client.Client, r io.ReaderAt) error {

	logger := slog.Default()
	parquet_r := parquet_go.NewGenericReader[*embeddingsdb.Record](r)

	rows := make([]*embeddingsdb.Record, 0, 10000)

	ticker := time.NewTicker(60 * time.Second)
	done_ch := make(chan bool)

	count := 0

	defer func() {
		ticker.Stop()
		done_ch <- true
	}()

	go func() {

		select {
		case <-done_ch:
			logger.Debug("Records imported", "count", count)
			return
		case <-ticker.C:
			logger.Debug("Records imported", "count", count)
		}

	}()

	for {

		n, err := parquet_r.Read(rows[:cap(rows)])

		if err != nil {

			if err == io.EOF {
				break
			}

			return fmt.Errorf("Failed to read record, %w", err)
		}

		rows = rows[:n]

		for _, row := range rows {

			err := cl.AddRecord(ctx, row)

			if err != nil {
				return fmt.Errorf("Failed to add record '%s', %w", row.Key(), err)
			}

			count += 1
		}
	}

	return nil

}
