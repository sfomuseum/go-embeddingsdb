package signatures

import (
	"context"
	"net/url"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb/signatures/pgp"
)

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

func NewPGPVerifier(ctx context.Context, uri string) (Verifier, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	key_uri := q.Get("key-uri")

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
