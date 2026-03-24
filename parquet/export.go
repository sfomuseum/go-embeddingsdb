package parquet

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

func Export(ctx context.Context, db database.Database, wr io.Writer) error {

	logger := slog.Default()
	p_wr := parquet_go.NewGenericWriter[*embeddingsdb.Record](wr)

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
			logger.Debug("Records exported", "count", count)
			return
		case <-ticker.C:
			logger.Debug("Records exported", "count", count)
		}

	}()

	for row, err := range db.Iterate(ctx) {

		if err != nil {
			return fmt.Errorf("Iterator yielded an error, %w", err)
		}

		_, err = p_wr.Write([]*embeddingsdb.Record{
			row,
		})

		if err != nil {
			return fmt.Errorf("Failed to export %s, %w", row.Key(), err)
		}

		count += 1
	}

	return p_wr.Close()
}
