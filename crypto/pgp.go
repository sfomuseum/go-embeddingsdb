package crypto

import (
	"context"
	"fmt"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
)

var pgp = pgp_crypto.PGP()

func LoadKey(ctx context.Context, key_uri string) (*pgp_crypto.Key, error) {

	k_str, err := runtimevar.StringVar(ctx, key_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to resolve key URI, %w", err)
	}

	return LoadKeyFromArmor(ctx, k_str)
}

func LoadKeyFromArmor(ctx context.Context, k_str string) (*pgp_crypto.Key, error) {

	k, err := pgp_crypto.NewKeyFromArmored(k_str)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive key from armored, %w", err)
	}

	return k, nil
}
