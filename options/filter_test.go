package options

import (
	"testing"
)

func TestFilterOptions(t *testing.T) {

	k := "provider"
	v := "sfomuseum"

	o := NewFilterOption(k, v)

	if o.Type() != FilterOptionType {
		t.Fatalf("Invalid type, %v", o.Type())
	}

	switch o.(type) {
	case *FilterOption:
		// pass
	default:
		t.Fatalf("Invalid type")
	}

	f := o.(*FilterOption)

	ok := f.Key()
	ov := f.Value()

	if ok != k {
		t.Fatalf("Invalid key, %s. Expected %s", ok, k)
	}

	if ov != v {
		t.Fatalf("Invalid value, %v. Expected %v", ov, v)
	}

}
