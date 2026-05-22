package crypto

import (
	"context"
	"encoding/json"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb"
)

func NewSigner(ctx context.Context, key *pgp_crypto.Key) (pgp_crypto.PGPSign, error) {

	pgp := pgp_crypto.PGP()

	signer, err := pgp.Sign().SigningKey(key).New()

	if err != nil {
		return nil, err
	}

	return signer, nil
}

func SignRecord(ctx context.Context, key *pgp_crypto.Key, rec *embeddingsdb.Record) ([]byte, error) {

	signer, err := NewSigner(ctx, key)

	if err != nil {
		return nil, err
	}

	return SignRecordWithSigner(ctx, signer, rec)
}

func SignRecordWithSigner(ctx context.Context, signer pgp_crypto.PGPSign, rec *embeddingsdb.Record) ([]byte, error) {

	enc, err := json.Marshal(rec)

	if err != nil {
		return nil, err
	}

	return signer.SignCleartext(enc) //, crypto.Armor)
}
