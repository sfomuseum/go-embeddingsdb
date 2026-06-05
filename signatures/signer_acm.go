package signatures

import (
	"context"
	_ "fmt"
	"net/url"

	"github.com/sfomuseum/go-embeddingsdb/signatures/tls"
	"github.com/aaronland/go-aws/v3/acm"
)

func init() {

	err := RegisterSigner(context.Background(), "acm", NewACMSigner)

	if err != nil {
		panic(err)
	}
}

// NewTLSSigner creates a new Signer instance using an X.509 certificate and key defined by 'uri'
// which is expected to take the form of:
//
//	tls://?{QUERY_PARAMETERS}
//
// Where valid query parameters are:
// * `certificate-uri` – A URI pointing to a PEM-encoded x509 certificate. (required)
// * `key-uri` – A URI pointing to a PEM-encoded private key file. (required)
//
// In both cases URIs may be: A path on the local filesystem "cwd://{PATH}" which will look for
// {PATH} in the current directory; A valid gocloud.dev/runtimevar URI.
func NewACMSigner(ctx context.Context, uri string) (Signer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	q := u.Query()

	cl, err := acm.NewClient(ctx, uri)

	if err != nil {
		return nil, err
	}

	arn := q.Get("arn")

	acm_cert, err := acm.ExportCertificateNoPassword(ctx, cl, arn)

	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadCertFromPEM(ctx, acm_cert.Certificate)

	if err != nil {
		return nil, err
	}

	key, err := tls.LoadKeyFromPEM(ctx, acm_cert.PrivateKey)
	
	if err != nil {
		return nil, err
	}

	s := &TLSSigner{
		cert: cert,
		key:  key,
	}

	return s, nil
}
