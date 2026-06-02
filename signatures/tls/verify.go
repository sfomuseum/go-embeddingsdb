package tls

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"errors"
)

// VerifyDetachedSignature checks that `sig` is a valid signature of `data`
// using the public key that is stored in `cert`.  It returns nil on success
// or an error describing why the verification failed.
func VerifyDetachedSignature(ctx context.Context, cert *x509.Certificate, data []byte, sig []byte) error {

	// Hash the data – the same hash algorithm that was used when signing
	hash := sha256.Sum256(data)

	switch pub := cert.PublicKey.(type) {

	case *rsa.PublicKey:
		return rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sig)

	case *ecdsa.PublicKey:
		ok := ecdsa.VerifyASN1(pub, hash[:], sig)

		if !ok {
			return errors.New("Invalid")
		}

		return nil

	case ed25519.PublicKey:

		if !ed25519.Verify(pub, data, sig) {
			return errors.New("ed25519 signature verification failed")
		}
		return nil

	default:
		return errors.New("unsupported public key type in certificate")
	}
}
