package parquet

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/sfomuseum/go-embeddingsdb"
)

type Statistics struct {
	Models    []string
	Providers []string
	Count     int
}

func AppendStatistics(ctx context.Context, wr *ParquetWriter, uris ...string) error {

	models := make([]string, 0)
	providers := make([]string, 0)
	count := 0

	for _, uri := range uris {

		for rec, err := range Iterate(ctx, uri) {

			if err != nil {
				return fmt.Errorf("Iterator yield an error, %w", err)
			}

			_, err = wr.Write([]*embeddingsdb.Record{rec})

			if err != nil {
				return err
			}

			if !slices.Contains(models, rec.Model) {
				models = append(models, rec.Model)
			}

			if !slices.Contains(providers, rec.Provider) {
				models = append(providers, rec.Provider)
			}

			count += 1

		}
	}

	stats := &Statistics{
		Models:    models,
		Providers: providers,
		Count:     count,
	}

	p_wr := wr.ParquetWriter()

	p_wr.SetKeyValueMetadata("models", strings.Join(stats.Models, ";"))
	p_wr.SetKeyValueMetadata("providers", strings.Join(stats.Providers, ";"))
	p_wr.SetKeyValueMetadata("count", strconv.Itoa(stats.Count))

	return nil
}
