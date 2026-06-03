package signatures

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"net/url"

	"github.com/sfomuseum/go-embeddingsdb/signatures/tls"
)

type TLSVerifier struct {
	Verifier
	cert *x509.Certificate
}

func NewTLSVerifier(ctx context.Context, uri string) (Verifier, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cert_uri := q.Get("certificate-uri")

	cert, err := tls.LoadCertFromURI(ctx, cert_uri)

	if err != nil {
		return nil, err
	}

	return NewTLSVerifierWithCertificate(ctx, cert)
}

func NewTLSVerifierWithCertificate(ctx context.Context, cert *x509.Certificate) (Verifier, error) {

	v := &TLSVerifier{
		cert: cert,
	}

	return v, nil
}

func (v *TLSVerifier) Verify(ctx context.Context, data []byte, sig []byte) (bool, error) {

	sig, err := tls.DecodeSignature(sig)

	if err != nil {
		return false, err
	}

	hash := sha256.Sum256(data)

	switch pub := v.cert.PublicKey.(type) {

	case *rsa.PublicKey:

		err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sig)

		if err != nil {
			return false, err
		}

		return true, nil

	case *ecdsa.PublicKey:

		ok := ecdsa.VerifyASN1(pub, hash[:], sig)

		if !ok {
			return false, fmt.Errorf("Signature did not validate")
		}

		return true, nil

	case ed25519.PublicKey:

		if !ed25519.Verify(pub, data, sig) {
			return false, fmt.Errorf("ed25519 signature verification failed")
		}

		return true, nil

	default:
		return false, fmt.Errorf("Unsupported public key type in certificate")
	}
}
