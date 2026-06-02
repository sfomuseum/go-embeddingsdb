package tls

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"fmt"
	"net/url"

	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

// TLSSigner knows how to sign data with the private key that
// belongs to a TLS (X.509) certificate.
type TLSSigner struct {
	signatures.Signer
	cert *x509.Certificate
	key  crypto.PrivateKey
}

func NewTLSSigner(ctx context.Context, uri string) (signatures.Signer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cert_uri := q.Get("certificate-uri")
	key_uri := q.Get("key-uri")

	cert, err := LoadCertFromURI(ctx, cert_uri)

	if err != nil {
		return nil, err
	}

	key, err := LoadKeyFromURI(ctx, key_uri)

	if err != nil {
		return nil, err
	}

	s := &TLSSigner{
		cert: cert,
		key:  key,
	}

	return s, nil
}

// NewTLSSignerFromPEM loads a cert and its key from two PEM buffers.
func NewTLSSignerFromPEM(ctx context.Context, cert_pem []byte, key_pem []byte) (signatures.Signer, error) {

	cert, err := LoadCertFromPEM(ctx, cert_pem)

	if err != nil {
		return nil, err
	}

	key, err := LoadKeyFromPEM(ctx, key_pem)

	if err != nil {
		return nil, err
	}

	s := &TLSSigner{
		cert: cert,
		key:  key,
	}

	return s, nil
}

func (s *TLSSigner) Verifier(ctx context.Context) (signatures.Verifier, error) {
	return NewTLSVerifierWithCertificate(ctx, s.cert)
}

// Sign produces a raw detached signature of the supplied data.
// The signature algorithm is inferred from the key type.
func (s *TLSSigner) Sign(ctx context.Context, data []byte) ([]byte, error) {

	hash := sha256.Sum256(data)

	switch k := s.key.(type) {
	case *rsa.PrivateKey:
		return rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, hash[:])
	case *ecdsa.PrivateKey:
		return ecdsa.SignASN1(rand.Reader, k, hash[:])
	case ed25519.PrivateKey:
		return ed25519.Sign(k, data), nil
	default:
		return nil, fmt.Errorf("Unsupported private key type")
	}
}
