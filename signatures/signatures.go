package signatures

import (
	"context"
)

type Signer interface {
	Sign(context.Context, []byte) ([]byte, error)
	Verifier(context.Context) (Verifier, error)
}

type Verifier interface {
	Verify(context.Context, []byte, []byte) (bool, error)
}
