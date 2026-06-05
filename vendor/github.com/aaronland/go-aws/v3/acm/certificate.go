package acm

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	aws_acm "github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/youmark/pkcs8"
)

func ExportCertificate(ctx context.Context, cl *aws_acm.Client, arn string, pswd string) (string, string, error) {

	opts := &aws_acm.ExportCertificateInput{
		CertificateArn: aws.String(arn),
		Passphrase:     []byte(pswd),
	}

	rsp, err := cl.ExportCertificate(ctx, opts)

	if err != nil {
		return "", "", err
	}

	return *rsp.Certificate, *rsp.PrivateKey, nil
}

func RemovePassword(data string, password string) (string, error) {

	block, _ := pem.Decode([]byte(data))

	if block == nil {
		return "", fmt.Errorf("failed to parse PEM block")
	}

	if block.Type != "ENCRYPTED PRIVATE KEY" {
		return "", fmt.Errorf("PEM block is not an ENCRYPTED PRIVATE KEY: %s", block.Type)
	}

	key, err := pkcs8.ParsePKCS8PrivateKey(block.Bytes, []byte(password))

	if err != nil {
		return "", fmt.Errorf("failed to decrypt PKCS#8 key: %w", err)
	}

	key_nopass, err := x509.MarshalPKCS8PrivateKey(key)

	if err != nil {
		return "", fmt.Errorf("failed to marshal private key to PKCS#8: %w", err)
	}

	nopass := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: key_nopass,
	}

	return string(pem.EncodeToMemory(nopass)), nil
}
