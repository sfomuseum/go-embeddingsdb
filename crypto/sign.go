package crypto

import (
	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
)

// NewSigner creates a PGPSign signer that uses the provided key
// for signing.  The signer is created via the [ProtonMail/gopenpgp/v3]
// package's Signer API and is ready to sign data immediately.
func NewSigner(key *pgp_crypto.Key) (pgp_crypto.PGPSign, error) {

	pgp := pgp_crypto.PGP()

	signer, err := pgp.Sign().SigningKey(key).New()

	if err != nil {
		return nil, err
	}

	return signer, nil
}
