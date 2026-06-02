package pgp

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
)

var pgp = crypto.PGP()

func LoadArmored(ctx context.Context, key_uri string) (string, error) {

	return runtimevar.StringVar(ctx, key_uri)
}

func LoadKey(ctx context.Context, key_uri string) (*crypto.Key, error) {

	armored, err := LoadArmored(ctx, key_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to resolve key URI, %w", err)
	}

	return LoadKeyFromArmor(ctx, armored)
}

func LoadKeyFromArmor(ctx context.Context, k_str string) (*crypto.Key, error) {

	k, err := crypto.NewKeyFromArmored(k_str)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive key from armored, %w", err)
	}

	return k, nil
}
