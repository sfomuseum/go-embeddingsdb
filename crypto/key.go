package crypto

import (
	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
)

func NewKey(name string, email string, pswd []byte) (*pgp_crypto.Key, error) {

	pgp := pgp_crypto.PGP()

	builder := pgp.KeyGeneration()
	builder = builder.AddUserId(name, email)
	// builder.Lifetime(int32)

	key_gen := builder.New()

	k, err := key_gen.GenerateKey()

	if err != nil {
		return nil, err
	}

	if pswd != nil {

		k, err = pgp.LockKey(k, pswd)

		if err != nil {
			return nil, err
		}
	}

	return k, nil
}
