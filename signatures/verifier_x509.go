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

// X509Verifier implements the Verifier interface using X.509 certificates.
type X509Verifier struct {
	Verifier
	cert *x509.Certificate
}

func init() {

	err := RegisterVerifier(context.Background(), "x509", NewX509Verifier)

	if err != nil {
		panic(err)
	}
}

// NewX509Verifier creates a new X509Verifier derived from 'uri' which is expected to
// take the form of:
//
//	tls://?{QUERY_PARAMETERS}
//
// Where valid query parameters are:
// * `certificate-uri` – A URI pointing to a PEM-encoded x509 certificate. (required)
//
// URIs may be: A path on the local filesystem "cwd://{PATH}" which will look for
// {PATH} in the current directory; A valid gocloud.dev/runtimevar URI.
func NewX509Verifier(ctx context.Context, uri string) (Verifier, error) {

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

	return NewX509VerifierWithCertificate(ctx, cert)
}

// NewX509VerifierWithCertificate creates a new X509Verifier using a pre-loaded x509.Certificate.
func NewX509VerifierWithCertificate(ctx context.Context, cert *x509.Certificate) (Verifier, error) {

	v := &X509Verifier{
		cert: cert,
	}

	return v, nil
}

// Verify checks the validity of a signature against the provided data using
// the public key found in the certificate (supporting RSA, ECDSA, and Ed25519).
func (v *X509Verifier) Verify(ctx context.Context, data []byte, sig []byte) (bool, error) {

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
