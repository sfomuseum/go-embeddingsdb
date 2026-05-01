package models

import (
	"testing"
)

func TestDeriveDimensions(t *testing.T) {

	lookup, err := KnownDimensions()

	if err != nil {
		t.Fatalf("Failed to derive dimensions, %v", err)
	}

	for model, expected_dims := range lookup {

		dims, exists := DeriveDimensionsFromModel(model)

		if !exists {
			t.Fatalf("Model '%s' not found", model)
		}

		if dims != expected_dims {
			t.Fatalf("Unexpected dimensions (%d) for model %s", dims, model)
		}
	}
}
