package acm

import (
	"context"

	"github.com/aaronland/go-string/random"
	"github.com/aws/aws-sdk-go-v2/aws"
	aws_acm "github.com/aws/aws-sdk-go-v2/service/acm"
)

func ExportCertificate(ctx context.Context, cl *aws_acm.Client, arn string, pswd string) (*Certificate, error) {

	opts := &aws_acm.ExportCertificateInput{
		CertificateArn: aws.String(arn),
		Passphrase:     []byte(pswd),
	}

	rsp, err := cl.ExportCertificate(ctx, opts)

	if err != nil {
		return nil, err
	}

	cert := &Certificate{
		Certificate:      []byte(*rsp.Certificate),
		CertificateChain: []byte(*rsp.CertificateChain),
		PrivateKey:       []byte(*rsp.PrivateKey),
	}

	return cert, nil
}

func ExportCertificateNoPassword(ctx context.Context, cl *aws_acm.Client, arn string) (*Certificate, error) {

	random_opts := &random.Options{
		Length:       32,
		AlphaNumeric: true,
	}

	pswd, err := random.String(random_opts)

	if err != nil {
		return nil, err
	}

	cert, err := ExportCertificate(ctx, cl, arn, pswd)

	if err != nil {
		return nil, err
	}

	nopass, err := cert.RemovePassword(pswd)

	if err != nil {
		return nil, err
	}

	cert.PrivateKey = nopass
	return cert, nil
}
