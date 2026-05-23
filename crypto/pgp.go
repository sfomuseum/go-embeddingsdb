package crypto

import (
	"context"
	"fmt"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
)

var pgp = pgp_crypto.PGP()

func loadKey(ctx context.Context, key_uri string) (*pgp_crypto.Key, error) {

	k_str, err := runtimevar.StringVar(ctx, key_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to resolve key URI, %w", err)
	}

	k, err := pgp_crypto.NewKeyFromArmored(k_str)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive key from armored, %w", err)
	}

	return k, nil
}
