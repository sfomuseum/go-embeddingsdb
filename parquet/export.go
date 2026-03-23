package parquet

import (
	"context"
	"io"

	parquet_go "github.com/parquet-go/parquet-go"
	"github.com/sfomuseum/go-embeddingsdb"
	"github.com/sfomuseum/go-embeddingsdb/database"
)

func Export(ctx context.Context, db database.Database, wr io.Writer) error {

	p_wr := parquet_go.NewGenericWriter[*embeddingsdb.Record](wr)

	for row, err := range db.Range() {

		if err != nil {
			return err
		}

		_, err = p_wr.Write([]*embeddingsdb.Record{
			row,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
