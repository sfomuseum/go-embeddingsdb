package oembeddings

import (
	_ "embed"
	"encoding/json"
	"io"
	"sync"

	"github.com/google/jsonschema-go/jsonschema"
)

//go:embed oembeddings.json
var OEmbeddingsJSONSchema []byte

var loadSchema = sync.OnceValues(func() (*jsonschema.Resolved, error) {

	var s jsonschema.Schema

	err := json.Unmarshal(OEmbeddingsJSONSchema, &s)

	if err != nil {
		return nil, err
	}

	resolved, err := s.Resolve(nil)

	if err != nil {
		return nil, err
	}

	return resolved, nil
})

func ValidateWithOEmbeddings(oe *OEmbeddings) (bool, error) {

	body, err := json.Marshal(oe)

	if err != nil {
		return false, err
	}

	return Validate(body)
}

func ValidateWithReader(r io.Reader) (bool, error) {

	body, err := io.ReadAll(r)

	if err != nil {
		return false, err
	}

	return Validate(body)
}

func Validate(body []byte) (bool, error) {

	schema, err := loadSchema()

	if err != nil {
		return false, err
	}

	var oe any

	err = json.Unmarshal(body, &oe)

	if err != nil {
		return false, err
	}

	err = schema.Validate(oe)

	if err != nil {
		return false, err
	}

	return true, nil
}
