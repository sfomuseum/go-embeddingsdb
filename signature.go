package embeddingsdb

import (
	"fmt"
)

type Signature struct {
	Provider string `json:"provider" parquet:"provider,dict,zstd" bleve:"store"`
	// DepictionId is the unique identifier for the depiction for which embeddings have been generated.
	DepictionId string `json:"depiction_id" parquet:"depiction_id,dict,zstd" bleve:"store"`
	// Model is the label for the model used to generate embeddings for DepictionId.
	Model string `json:"model" parquet:"model,dict,zstd" bleve:"store"`
	// Return the hex-encoded SHA256 hash of the record that was signed.
	Hash string `json:"hash" parquet:"signature,dict,zstd" bleve:"store"`
	// Return the detached signature associated with the signed record as an ASCII armor-encoded string
	Signature string `json:"signature" parquet:"signature,dict,zstd" bleve:"store"`
}

func (s *Signature) Key() string {
	return fmt.Sprintf("%s-%s-%s-%s", s.Provider, s.DepictionId, s.Model, s.Hash)
}

// Return the detached signature associated with the signed record as an ASCII armor-encoded string
func (s *Signature) String() string {
	return s.Signature
}
