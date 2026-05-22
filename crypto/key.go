package crypto

import (
	"context"
	"fmt"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
)

func LoadKey(ctx context.Context, key_uri string, pass_uri string) (*pgp_crypto.Key, error) {

	k_str, err := runtimevar.StringVar(ctx, key_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to resolve key URI, %w", err)
	}

	k, err := pgp_crypto.NewKeyFromArmored(k_str)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive key from armored, %w", err)
	}

	is_locked, err := k.IsLocked()

	if err != nil {
		return nil, fmt.Errorf("Failed to determine if key is locked, %w", err)
	}

	if is_locked {

		p_str, err := runtimevar.StringVar(ctx, pass_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to resolve key password URI, %w", err)
		}

		k, err = k.Unlock([]byte(p_str))

		if err != nil {
			return nil, fmt.Errorf("Failed to unlock key, %w", err)
		}
	}

	return k, nil
}

func NewKey(name string, email string, pswd []byte) (*pgp_crypto.Key, error) {

	pgp := pgp_crypto.PGP()

	builder := pgp.KeyGeneration()
	builder = builder.AddUserId(name, email)
	// builder.Lifetime(int32)

	key_gen := builder.New()

	k, err := key_gen.GenerateKey()

	if err != nil {
		return nil, fmt.Errorf("Failed to generate new key, %w", err)
	}

	if pswd != nil {

		k, err = pgp.LockKey(k, pswd)

		if err != nil {
			return nil, fmt.Errorf("Failed to lock key, %w", err)
		}
	}

	return k, nil
}
