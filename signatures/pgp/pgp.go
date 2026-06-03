package pgp

import (
	"context"
	"fmt"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
)

// LoadArmored retrieves the armored string associated with the provided gocloud.dev/runtimevar URI.
func LoadArmored(ctx context.Context, key_uri string) (string, error) {

	return runtimevar.StringVar(ctx, key_uri)
}

// LoadKey retrieves a PGP key by first resolving the provided gocloud.dev/runtimevar URI to an 
// armored string and then parsing that string into a crypto.Key.
func LoadKey(ctx context.Context, key_uri string) (*crypto.Key, error) {

	armored, err := LoadArmored(ctx, key_uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to resolve key URI, %w", err)
	}

	return LoadKeyFromArmor(ctx, armored)
}

// LoadKeyFromArmor parses a raw armored PGP string into a crypto.Key.
func LoadKeyFromArmor(ctx context.Context, k_str string) (*crypto.Key, error) {

	k, err := crypto.NewKeyFromArmored(k_str)

	if err != nil {
		return nil, fmt.Errorf("Failed to derive key from armored, %w", err)
	}

	return k, nil
}
