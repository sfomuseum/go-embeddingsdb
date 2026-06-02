package pgp

import (
	"context"
	"encoding/json"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
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
func SignRecord(ctx context.Context, key *crypto.Key, rec *embeddingsdb.Record) ([]byte, error) {

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
func SignRecordWithSigner(ctx context.Context, signer crypto.PGPSign, rec *embeddingsdb.Record) ([]byte, error) {

	enc, err := json.Marshal(rec)

	if err != nil {
		return nil, err
	}

	return signer.Sign(enc, crypto.Armor)
}

func VerifyRecordSignature(ctx context.Context, key *crypto.Key, rec *embeddingsdb.Record, sig []byte) (bool, error) {

	verifier, err := LoadVerificationHandlerWithKey(ctx, key)

	if err != nil {
		return false, err
	}

	return VerifyRecordSignatureWithVerifier(ctx, verifier, rec, sig)
}

func VerifyRecordSignatureWithVerifier(ctx context.Context, verifier crypto.PGPVerify, rec *embeddingsdb.Record, sig []byte) (bool, error) {

	enc, err := json.Marshal(rec)

	if err != nil {
		return false, err
	}

	return VerifyRecordSignatureWithVerifierAndBody(ctx, verifier, enc, sig)
}

func VerifyRecordSignatureWithVerifierAndBody(ctx context.Context, verifier crypto.PGPVerify, enc []byte, sig []byte) (bool, error) {

	rsp, err := verifier.VerifyDetached(enc, sig, crypto.Armor)

	if err != nil {
		return false, err
	}

	err = rsp.SignatureError()

	if err != nil {
		return false, nil
	}

	return true, nil
}
