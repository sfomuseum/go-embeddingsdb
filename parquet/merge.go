package parquet

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
)

// Merge [*embeddingsdb.Record] records stored in a Parquet file and adds them to 'wr'
func Merge(ctx context.Context, wr *ParquetWriter, r io.ReaderAt) (int64, error) {

	logger := slog.Default()
	parquet_r := parquet_go.NewGenericReader[*embeddingsdb.Record](r)

	rows := make([]*embeddingsdb.Record, 0, 1000)

	ticker := time.NewTicker(60 * time.Second)
	done_ch := make(chan bool)

	count := int64(0)

	defer func() {
		ticker.Stop()
		done_ch <- true
	}()

	go func() {

		for {
			select {
			case <-done_ch:
				logger.Debug("Records imported", "count", count)
				return
			case <-ticker.C:
				logger.Debug("Records imported", "count", count)
			}
		}
	}()

	for {

		n, err := parquet_r.Read(rows[:cap(rows)])

		if err != nil {

			if err == io.EOF {
				break
			}

			return count, fmt.Errorf("Failed to read record, %w", err)
		}

		rows = rows[:n]
		wr.Write(rows)
	}

	return count, nil

}
