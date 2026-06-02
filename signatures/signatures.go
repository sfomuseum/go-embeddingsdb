package signatures

import (
	"context"
)

type Signer interface {
	Sign(context.Context, []byte) ([]byte, error)
	Type() string
}

type Verifier interface {
	Verify(context.Context, []byte, []byte) (bool, error)
}
