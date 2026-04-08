package oembeddings

import (
	"os"
	"testing"
)

func TestOEmbeddingsJSONSchema(t *testing.T) {

	tests := []string{
		"../fixtures/oembeddings_image.json",
		"../fixtures/oembeddings_text.json",
		"../fixtures/oembeddings_extra.json",
	}

	for _, path := range tests {

		r, err := os.Open(path)

		if err != nil {
			t.Fatalf("Failed to open %s for reading, %v", path, err)
		}

		defer r.Close()

		_, err = ValidateWithReader(r)

		if err != nil {
			t.Fatalf("Failed to valiate %s, %v", path, err)
		}
	}
}
