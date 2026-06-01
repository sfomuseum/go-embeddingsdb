package crypto

import (
	"context"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
)

func LoadVerificationHandler(ctx context.Context, key_uri string) (pgp_crypto.PGPVerify, error) {

	k, err := LoadKey(ctx, key_uri)

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

	return LoadVerificationHandlerWithKey(ctx, pub_k)
}

func LoadVerificationHandlerWithKey(ctx context.Context, pub_k *pgp_crypto.Key) (pgp_crypto.PGPVerify, error) {

	builder := pgp.Verify()
	builder = builder.VerificationKey(pub_k)

	return builder.New()
}
