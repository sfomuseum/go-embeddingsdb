package signatures

import (
	"context"
	"fmt"
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

// NewACMSigner creates a new Signer instance using an X.509 certificate stored in the AWS
// Certificate Manager service and key defined by 'uri'
// which is expected to take the form of:
//
//	acm://?{QUERY_PARAMETERS}
//
// Where valid query parameters are:
// * `region` – The AWS region to connect to. (require)
// * `credentials` – A recognized `aaronland/go-aws/v3/auth` credentials string to use with the AWS API. (required)
// * `arn` - The AWS Certificate Manager ARN for the certificate you want to use to sign records. (required)
func NewACMSigner(ctx context.Context, uri string) (Signer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to parse URI, %w", err)
	}

	q := u.Query()

	cl, err := acm.NewClient(ctx, uri)

	if err != nil {
		return nil, fmt.Errorf("Failed to create ACM client, %w", err)
	}

	arn := q.Get("arn")

	acm_cert, err := acm.ExportCertificateNoPassword(ctx, cl, arn)

	if err != nil {
		return nil, fmt.Errorf("Failed to export ACM certificate, %w", err)
	}

	cert, err := tls.LoadCertFromPEM(ctx, acm_cert.Certificate)

	if err != nil {
		return nil, fmt.Errorf("Failed to load certificate, %w", err)
	}

	key, err := tls.LoadKeyFromPEM(ctx, acm_cert.PrivateKey)
	
	if err != nil {
		return nil, fmt.Errorf("Failed to load private key, %w", err)
	}

	s := &X509Signer{
		cert: cert,
		key:  key,
	}

	return s, nil
}
