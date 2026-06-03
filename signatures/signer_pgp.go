package signatures

import (
	"context"
	"fmt"
	"net/url"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/aaronland/gocloud/runtimevar"
	"github.com/sfomuseum/go-embeddingsdb/signatures/pgp"
)

type PGPSigner struct {
	Signer
	key    *crypto.Key
	signer crypto.PGPSign
}

func init() {

	err := RegisterSigner(context.Background(), "pgp", NewPGPSigner)

	if err != nil {
		panic(err)
	}
}

func NewPGPSigner(ctx context.Context, uri string) (Signer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	key_uri := q.Get("key-uri")
	pswd_uri := q.Get("key-password-uri")

	key, err := pgp.LoadKey(ctx, key_uri)

	if err != nil {
		return nil, err
	}

	is_locked, err := key.IsLocked()

	if err != nil {
		return nil, fmt.Errorf("Failed to determine if key is locked, %w", err)
	}

	if is_locked {

		p_str, err := runtimevar.StringVar(ctx, pswd_uri)

		if err != nil {
			return nil, fmt.Errorf("Failed to resolve key password URI, %w", err)
		}

		key, err = key.Unlock([]byte(p_str))

		if err != nil {
			return nil, fmt.Errorf("Failed to unlock key, %w", err)
		}
	}

	return NewPGPSignerWithKey(ctx, key)
}

func NewPGPSignerWithKey(ctx context.Context, key *crypto.Key) (Signer, error) {

	pgp_ctx := crypto.PGP()

	signer, err := pgp_ctx.Sign().SigningKey(key).Detached().New()

	if err != nil {
		return nil, err
	}

	s := &PGPSigner{
		key:    key,
		signer: signer,
	}

	return s, nil
}

func (s *PGPSigner) Sign(ctx context.Context, data []byte) ([]byte, error) {
	return s.signer.Sign(data, crypto.Armor)
}

func (s *PGPSigner) Verifier(ctx context.Context) (Verifier, error) {
	return NewPGPVerifierWithKey(ctx, s.key)
}

func (s *PGPSigner) PublicKey(ctx context.Context) ([]byte, error) {

	pub_key, err := s.key.ToPublic()

	if err != nil {
		return nil, err
	}

	armor, err := pub_key.Armor()

	if err != nil {
		return nil, err
	}

	return []byte(armor), nil
}
