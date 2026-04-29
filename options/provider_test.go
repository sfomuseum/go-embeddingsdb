package options

import (
	"testing"
)

func TestProviderOptions(t *testing.T) {

	p := "sfomuseum"
	o := NewProviderOption(p)

	if o.Type() != ProviderOptionType {
		t.Fatalf("Invalid type, %v", o.Type())
	}

	switch o.(type) {
	case *ProviderOption:
		// pass
	default:
		t.Fatalf("Invalid type")
	}

	v := o.(*ProviderOption).Provider()

	if v != p {
		t.Fatalf("Invalid provider, %s. Expected %s", v, p)
	}
}
