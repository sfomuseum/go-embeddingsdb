package tls

import (
	"context"
	"crypto"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	"github.com/aaronland/gocloud/runtimevar"
)

const CERTIFICATE string = "CERTIFICATE"
const SIGNATURE string = "SIGNATURE"

func LoadCertFromURI(ctx context.Context, uri string) (*x509.Certificate, error) {

	cert_body, err := runtimevar.StringVar(ctx, uri)

	if err != nil {
		return nil, err
	}

	cert_body = strings.TrimSpace(cert_body)
	return LoadCertFromPEM(ctx, []byte(cert_body))
}

func LoadKeyFromURI(ctx context.Context, uri string) (crypto.PrivateKey, error) {

	key_body, err := runtimevar.StringVar(ctx, uri)

	if err != nil {
		return nil, err
	}

	key_body = strings.TrimSpace(key_body)
	return LoadKeyFromPEM(ctx, []byte(key_body))
}

func LoadCertFromPEM(ctx context.Context, data []byte) (*x509.Certificate, error) {

	block, _ := pem.Decode(data)

	if block == nil || block.Type != CERTIFICATE {
		return nil, fmt.Errorf("Failed to decode PEM certificate")
	}

	return x509.ParseCertificate(block.Bytes)
}

func LoadKeyFromPEM(ctx context.Context, data []byte) (crypto.PrivateKey, error) {

	block, _ := pem.Decode(data)

	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM key")
	}

	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)

	if err == nil {
		return priv, nil
	}

	priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)

	if err == nil {
		return priv, nil
	}

	priv, err = x509.ParseECPrivateKey(block.Bytes)

	if err == nil {
		return priv, nil
	}

	return nil, err
}

// IsArmored checks if the input byte slice contains at least one valid PEM block.
func IsArmored(data []byte) bool {
	block, _ := pem.Decode(data)
	return block != nil
}

func EncodeSignature(data []byte) []byte {

	pem_block := &pem.Block{
		Type:  SIGNATURE,
		Bytes: data,
	}

	return pem.EncodeToMemory(pem_block)
}

func DecodeSignature(data []byte) ([]byte, error) {

	if !IsArmored(data) {
		return data, nil
	}

	block, _ := pem.Decode(data)

	if block == nil {
		return nil, fmt.Errorf("Failed to parse PEM")
	}

	if block.Type != SIGNATURE {
		return nil, fmt.Errorf("Invalid block type %s", block.Type)
	}

	return block.Bytes, nil
}
