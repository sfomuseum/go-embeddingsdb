package embeddingsdb

// Record defines a struct containing properties associated with individual records stored in an embeddings database.
type Record struct {
	// Provider is the name (or context) of the provider responsible for DepictionId.
	Provider string `json:"provider"`
	// DepictionId is the unique identifier for the depiction for which embeddings have been generated.
	DepictionId string `json:"depiction_id"`
	// SubjectId is the unique identifier associated with the record that DepictionId depicts.
	SubjectId string `json:"subject_id"`
	// Model is the label for the model used to generate embeddings for DepictionId.
	Model string `json:"model"`
	// Embeddings are the embeddings generated for DepictionId using Model.
	Embeddings []float32 `json:"embeddings"`
	// Created is the Unix timestamp when Embeddings were generated.
	Created int64 `json:"created"`
	// Attributes is an arbitrary map of key-value properties associated with the embeddings.
	Attributes map[string]string `json:"attributes"`
}

// SimilarRequest is a struct containing properties for retrieving records from an embeddings database.
type SimilarRequest struct {
	// The name of the provider to limit similar record queries to. If empty then all the records for the model chosen will be queried.
	SimilarProvider *string `json:"similar_provider,omitempty"`
	// Model is the name of the model to specify when querying for similar embeddings.
	Model string `json:"model"`
	// Embeddings are the embeddings to use for querying similar records.
	Embeddings []float32 `json:"embeddings"`
	// Zero or more depiction IDs to explicitly exclude from results.
	Exclude []string `json:"exclude,omitempty"`
	// ...
	MaxDistance float32 `json:"max_distance,omitempty"`
	// ...
	MaxResults int `json:"max_results,omitempty"`
}

// SimilarResult is a struct containing properties for similar records returned from an embeddings database.
type SimilarResult struct {
	// Provider is the name (or context) of the provider responsible for DepictionId.
	Provider string `json:"provider"`
	// DepictionID is the unique identifier	for the	depiction associated with the result.
	DepictionId string `json:"depiction_id"`
	// DepictionID is the unique identifier	for the	subject associated with the result depiction.
	SubjectId string `json:"subject_id"`
	// Similarlity is the distance between the input embeddings and the result embeddings.
	Similarity float32 `json:"similarity"`
	// Attributes is an arbitrary map of key-value properties associated with the embeddings.
	Attributes map[string]string `json:"attributes"`
}
