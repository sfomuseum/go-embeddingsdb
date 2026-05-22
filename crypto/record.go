package crypto

import (
	"context"
	"encoding/json"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb"
)

func SignRecord(ctx context.Context, key *pgp_crypto.Key, rec *embeddingsdb.Record) ([]byte, error) {

	enc, err := json.Marshal(rec)

	if err != nil {
		return nil, err
	}

	pgp := pgp_crypto.PGP()

	signer, err := pgp.Sign().SigningKey(key).New()

	if err != nil {
		return nil, err
	}

	return signer.SignCleartext(enc) //, crypto.Armor)
}
