package crypto

// gpg --full-generate-key

import (
	"context"
	"fmt"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
)

// LoadKey retrieves a PGP key from the supplied 'key_uri' and, if the
// key is encrypted, unlocks it using the password found at 'pass_uri'.
// Both the 'key_uri' and 'pass_uri' variables are expected to be registed
// [gocloud.dev/runtimevar] URIs.
//
// 'key_uri' is expected to resolve an ASCII‑armored key block.
// 'pass_uri' is expected to resolve to the key's password but is only processed
// if the key is locked. As such an it may be an empty string.
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

// NewKey generates a new PGP key pair with the given user ID (name
// and email).  The resulting key is returned in unlocked form.  If
// pswd is non‑nil, the key is immediately locked with that
// passphrase.
//
// The function uses the [ProtonMail/gopenpgp/v3] package to create the key.
// It does not persist the key. The caller is responsible for storing the
// key elsewhere if desired.
func NewKey(name string, email string, pswd []byte) (*pgp_crypto.Key, error) {

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
