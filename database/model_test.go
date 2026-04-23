package database

import (
	"context"
	"testing"
)

func TestDeriveModelDimensions(t *testing.T) {

	ctx := context.Background()

	tests_ok := map[string]int{
		"apple/mobileclip_s0":            512,
		"google/siglip-base-patch16-224": 1024,
	}

	tests_fail := map[string]int{
		"test/debug": 0,
	}

	for m, expected := range tests_ok {

		i, err := DeriveModelDimensions(ctx, m)

		if err != nil {
			t.Fatalf("Expected %s model to return answer", m)
		}

		if i != expected {
			t.Fatalf("Unexpected value (%d) for model %s. Expected %d", i, m, expected)
		}
	}

	for m, _ := range tests_fail {

		_, err := DeriveModelDimensions(ctx, m)

		if err == nil {
			t.Fatalf("Model %s was not supposed to return a value", m)
		}
	}
}
