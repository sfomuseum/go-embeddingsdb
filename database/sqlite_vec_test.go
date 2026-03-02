package database

import (
	"slices"
	"testing"

	sqlite_vec "github.com/asg017/sqlite-vec-go-bindings/cgo"
)

func TestDeserializeFloat32(t *testing.T) {

	f32 := []float32{
		0.0,
		0.1,
		0.002,
	}

	s32, err := sqlite_vec.SerializeFloat32(f32)

	if err != nil {
		t.Fatalf("Failed to serialize floats, %v", err)
	}

	new_f32, err := DeserializeFloat32(s32)

	if err != nil {
		t.Fatalf("Failed to deserialize floats, %v", err)
	}

	if !slices.Equal(f32, new_f32) {
		t.Fatalf("Deserialized floats to do match input")
	}
}
