package parquet

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/sfomuseum/go-embeddingsdb"
)

type Statistics struct {
	Models    []string `json:"models"`
	Providers []string `json:"providers"`
	mu        *sync.RWMutex
}

func NewStatistics() *Statistics {

	models := make([]string, 0)
	providers := make([]string, 0)

	mu := new(sync.RWMutex)

	s := &Statistics{
		Models:    models,
		Providers: providers,
		mu:        mu,
	}

	return s
}

func (s *Statistics) AddRecord(rec *embeddingsdb.Record) error {

	s.mu.Lock()

	if !slices.Contains(s.Models, rec.Model) {
		s.Models = append(s.Models, rec.Model)
	}

	if !slices.Contains(s.Providers, rec.Provider) {
		s.Providers = append(s.Providers, rec.Provider)
	}

	s.mu.Unlock()
	return nil
}

func AppendStatistics(ctx context.Context, wr *ParquetWriter, uris ...string) error {

	stats := NewStatistics()

	for _, uri := range uris {

		for rec, err := range Iterate(ctx, uri) {

			if err != nil {
				return fmt.Errorf("Iterator yield an error, %w", err)
			}

			_, err = wr.Write([]*embeddingsdb.Record{rec})

			if err != nil {
				return err
			}

			stats.AddRecord(rec)
		}
	}

	p_wr := wr.ParquetWriter()

	p_wr.SetKeyValueMetadata("models", strings.Join(stats.Models, ";"))
	p_wr.SetKeyValueMetadata("providers", strings.Join(stats.Providers, ";"))

	return nil
}
