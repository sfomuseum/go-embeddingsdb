package options

import (
	"testing"
)

func TestOptionDimensions(t *testing.T) {

	d := 512
	o := NewDimensionsOption(d)

	if o.Type() != DimensionsOptionType {
		t.Fatalf("Invalid type, %v", o.Type())
	}

	switch o.(type) {
	case *DimensionsOption:
		// pass
	default:
		t.Fatalf("Invalid type")
	}

	v := o.(*DimensionsOption).Dimensions()

	if v != d {
		t.Fatalf("Invalid dimensions, %d. Expected %d", v, d)
	}
}
