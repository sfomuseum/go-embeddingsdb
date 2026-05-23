package crypto

import (
	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
)

// NewSigner creates a PGPSign detached signer that uses the provided key
// for signing.  The signer is created via the [ProtonMail/gopenpgp/v3]
// package's Signer API and is ready to sign data immediately.
func NewSigner(key *pgp_crypto.Key) (pgp_crypto.PGPSign, error) {

	return pgp.Sign().SigningKey(key).Detached().New()
}
