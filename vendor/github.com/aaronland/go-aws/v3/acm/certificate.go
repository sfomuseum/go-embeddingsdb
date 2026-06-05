package acm

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/youmark/pkcs8"
)

type Certificate struct {
	Certificate      []byte
	CertificateChain []byte
	PrivateKey       []byte
}

func (c *Certificate) RemovePassword(password string) ([]byte, error) {

	block, _ := pem.Decode(c.PrivateKey)

	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	if block.Type != "ENCRYPTED PRIVATE KEY" {
		return nil, fmt.Errorf("PEM block is not an ENCRYPTED PRIVATE KEY: %s", block.Type)
	}

	key, err := pkcs8.ParsePKCS8PrivateKey(block.Bytes, []byte(password))

	if err != nil {
		return nil, fmt.Errorf("failed to decrypt PKCS#8 key: %w", err)
	}

	var pem_type string
	var buf []byte

	switch k := key.(type) {
	case *ecdsa.PrivateKey:
		// SEC1 EC Private Key format (Required for ES256 / Ring engine compatibility)
		pem_type = "EC PRIVATE KEY"
		buf, err = x509.MarshalECPrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ECDSA key to SEC1: %w", err)
		}

	case *rsa.PrivateKey:
		// PKCS#1 RSA Private Key format (Required for PS256 engine compatibility)
		pem_type = "RSA PRIVATE KEY"
		buf = x509.MarshalPKCS1PrivateKey(k)

	case ed25519.PrivateKey:
		// Ed25519 doesn't have an alternative legacy RFC header block like EC/RSA.
		// It relies on standard unencrypted PKCS#8 formatting.
		pem_type = "PRIVATE KEY"
		buf, err = x509.MarshalPKCS8PrivateKey(k)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal Ed25519 key to PKCS#8: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported asymmetric key algorithm type: %T", key)
	}

	// 3. Wrap with the specific PEM label computed for the algorithm type
	nopass := &pem.Block{
		Type:  pem_type,
		Bytes: buf,
	}

	return pem.EncodeToMemory(nopass), nil
}
