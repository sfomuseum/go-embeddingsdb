package crypto

import (
	"context"
	"encoding/json"

	pgp_crypto "github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/sfomuseum/go-embeddingsdb"
)

// SignRecord signs an embeddingsdb.Record using the supplied key.
// It internally creates a signer from the key and forwards the
// signing to SignRecordWithSigner.
//
// The record is first marshalled to JSON; the resulting bytes are
// signed in cleartext mode (i.e. the output contains the cleartext
// data followed by an ASCII‑armored signature).  The returned
// slice contains the signed data.
func SignRecord(ctx context.Context, key *pgp_crypto.Key, rec *embeddingsdb.Record) ([]byte, error) {

	signer, err := NewSigner(key)

	if err != nil {
		return nil, err
	}

	return SignRecordWithSigner(ctx, signer, rec)
}

// SignRecordWithSigner signs an embeddingsdb.Record using a signer
// that has already been created.  This function is useful when the
// caller wishes to reuse the same signer for multiple records.
//
// The record is first marshalled to JSON, then signed in cleartext
// mode.  The returned slice contains the signed data.
func SignRecordWithSigner(ctx context.Context, signer pgp_crypto.PGPSign, rec *embeddingsdb.Record) ([]byte, error) {

	enc, err := json.Marshal(rec)

	if err != nil {
		return nil, err
	}

	return signer.SignCleartext(enc) //, crypto.Armor)
}
