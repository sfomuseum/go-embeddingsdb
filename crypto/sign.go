package crypto

import (
	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
)

func NewSigner(key *pgp_crypto.Key) (pgp_crypto.PGPSign, error) {

	pgp := pgp_crypto.PGP()

	signer, err := pgp.Sign().SigningKey(key).New()

	if err != nil {
		return nil, err
	}

	return signer, nil
}
