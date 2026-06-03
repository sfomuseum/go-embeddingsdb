package signatures

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

	"github.com/sfomuseum/go-embeddingsdb/signatures/tls"
)

// TLSSigner knows how to sign data with the private key that
// belongs to a TLS (X.509) certificate.
type TLSSigner struct {
	Signer
	cert *x509.Certificate
	key  crypto.PrivateKey
}

func NewTLSSigner(ctx context.Context, uri string) (Signer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cert_uri := q.Get("certificate-uri")
	key_uri := q.Get("key-uri")

	cert, err := tls.LoadCertFromURI(ctx, cert_uri)

	if err != nil {
		return nil, err
	}

	key, err := tls.LoadKeyFromURI(ctx, key_uri)

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
func NewTLSSignerFromPEM(ctx context.Context, cert_pem []byte, key_pem []byte) (Signer, error) {

	cert, err := tls.LoadCertFromPEM(ctx, cert_pem)

	if err != nil {
		return nil, err
	}

	key, err := tls.LoadKeyFromPEM(ctx, key_pem)

	if err != nil {
		return nil, err
	}

	s := &TLSSigner{
		cert: cert,
		key:  key,
	}

	return s, nil
}

func (s *TLSSigner) Verifier(ctx context.Context) (Verifier, error) {
	return NewTLSVerifierWithCertificate(ctx, s.cert)
}

// Sign produces a raw detached signature of the supplied data.
// The signature algorithm is inferred from the key type.
func (s *TLSSigner) Sign(ctx context.Context, data []byte) ([]byte, error) {

	hash := sha256.Sum256(data)

	var sig []byte
	var err error

	switch k := s.key.(type) {
	case *rsa.PrivateKey:
		sig, err = rsa.SignPKCS1v15(rand.Reader, k, crypto.SHA256, hash[:])
	case *ecdsa.PrivateKey:
		sig, err = ecdsa.SignASN1(rand.Reader, k, hash[:])
	case ed25519.PrivateKey:
		sig, err = ed25519.Sign(k, data), nil
	default:
		return nil, fmt.Errorf("Unsupported private key type")
	}

	if err != nil {
		return nil, err
	}

	return tls.EncodeSignature(sig), nil
}
