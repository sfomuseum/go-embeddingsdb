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
		"../fixtures/oembeddings_no_depiction_url.json",
		"../fixtures/oembeddings_x_urn.json",
		"../fixtures/oembeddings_empty.json",
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
