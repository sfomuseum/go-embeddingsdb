package crypto

import (
	"context"
	"encoding/json"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb"
)

func SignRecord(ctx context.Context, key *pgp_crypto.Key, rec *embeddingsdb.Record) ([]byte, error) {

	signer, err := NewSigner(key)

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
