package embeddingsdb

type Record struct {
	DepictionId int64     `json:"depiction_id"`
	SubjectId   int64     `json:"subject_id"`
	Model       string    `json:"model"`
	Dimensions  int       `json:"dimensions"`
	Embeddings  []float32 `json:"embeddings"`
	Created     int64     `json:"created"`
	URI         string    `json:"uri"`
}

type Record2 struct {
	DepictionId string            `json:"depiction_id"`
	SubjectId   string            `json:"subject_id"`
	Model       string            `json:"model"`
	Dimensions  int               `json:"dimensions"`
	Embeddings  []float32         `json:"embeddings"`
	Created     int64             `json:"created"`
	Provider    string            `json:"provider"`
	Meta        map[string]string `json:"meta"`
}

type QueryResult struct {
	DepictionId int64   `json:"depiction_id"`
	SubjectId   int64   `json:"subject_id"`
	Similarity  float32 `json:"similarity"`
}

type QueryResult2 struct {
	DepictionId string  `json:"depiction_id"`
	SubjectId   string  `json:"subject_id"`
	Provider    string  `json:"provider"`
	Similarity  float32 `json:"similarity"`
}
