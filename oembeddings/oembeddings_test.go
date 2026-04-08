package oembeddings

import (
	"encoding/json"
	"os"
	"testing"
)

func TestOEmbeddings(t *testing.T) {

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

		var o *OEmbeddings

		dec := json.NewDecoder(r)
		err = dec.Decode(&o)

		if err != nil {
			t.Fatalf("Failed to decode %s, %v", path, err)
		}
	}
}
