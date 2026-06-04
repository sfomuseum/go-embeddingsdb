package signatures

import (
	"context"
	"net/url"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb/signatures/pgp"
)

// PGPVerifier implements the Verifier interface using OpenPGP keys.
type PGPVerifier struct {
	Verifier
	key      *crypto.Key
	verifier crypto.PGPVerify
}

func init() {

	err := RegisterVerifier(context.Background(), "pgp", NewPGPVerifier)

	if err != nil {
		panic(err)
	}
}

// NewPGPVerifier creates a new PGPVerifier derived from 'uri' which
// is expected to take the form of:
// Where valid query parameters are:
// * `public-key-uri` – A URI pointing to a PEM-encoded PGP public key. (required)
//
// URIs may be: A path on the local filesystem "cwd://{PATH}" which will look for
// {PATH} in the current directory; A valid gocloud.dev/runtimevar URI.
func NewPGPVerifier(ctx context.Context, uri string) (Verifier, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	key_uri := q.Get("public-key-uri")

	k, err := pgp.LoadKey(ctx, key_uri)

	if err != nil {
		return nil, err
	}

	pub_k := k

	if k.IsPrivate() {

		pbk, err := k.ToPublic()

		if err != nil {
			return nil, err
		}

		pub_k = pbk
	}

	return NewPGPVerifierWithKey(ctx, pub_k)
}

// NewPGPVerifierWithKey creates a new PGPVerifier using a pre-loaded crypto.Key.
func NewPGPVerifierWithKey(ctx context.Context, key *crypto.Key) (Verifier, error) {

	pgp_ctx := crypto.PGP()

	builder := pgp_ctx.Verify()
	builder = builder.VerificationKey(key)

	verifier, err := builder.New()

	if err != nil {
		return nil, err
	}

	v := &PGPVerifier{
		key:      key,
		verifier: verifier,
	}

	return v, nil
}

// Verify checks the validity of a PGP signature against the provided data.
func (v *PGPVerifier) Verify(ctx context.Context, data []byte, sig []byte) (bool, error) {

	rsp, err := v.verifier.VerifyDetached(data, sig, crypto.Armor)

	if err != nil {
		return false, err
	}

	err = rsp.SignatureError()

	if err != nil {
		return false, nil
	}

	return true, nil
}
