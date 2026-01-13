package embeddingsdb

type Record struct {
	Provider    string            `json:"provider"`
	DepictionId string            `json:"depiction_id"`
	SubjectId   string            `json:"subject_id"`
	Model       string            `json:"model"`
	Embeddings  []float32         `json:"embeddings"`
	Created     int64             `json:"created"`
	Attributes  map[string]string `json:"attributes"`
}

type SimilarRequest struct {
	Provider   *string   `json:"provider"`
	Model      string    `json:"model"`
	Embeddings []float32 `json:"embeddings"`
}

type SimilarResult struct {
	Provider    string  `json:"provider"`
	DepictionId string  `json:"depiction_id"`
	SubjectId   string  `json:"subject_id"`
	Similarity  float32 `json:"similarity"`
}
