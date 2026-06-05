The signatures package provides a unified abstraction for generating and verifying digital signatures across different cryptographic protocols, specifically OpenPGP and TLS (X.509). The package is built upon two primary interfaces:

The `Signer` interface is used to generate digital signatures for data payloads.

```
type Signer interface {
    // Sign takes a byte slice and returns the resulting signature.
    Sign(context.Context, []byte) ([]byte, error)
    
    // Verifier returns a Verifier for the underlying credentials.
    Verifier(context.Context) (Verifier, error)
    
    // PublicKey returns the public representation of the signing key.
    PublicKey(context.Context) ([]byte, error)
}
```

The `Verifier` interface is used to validate whether a signature is valid for a given piece of data.

```
type Verifier interface {
    // Verify checks if the provided signature is valid for the given data.
    Verify(context.Context, []byte, []byte) (bool, error)
}
```

## Signer

To create a `Signer`, you provide a URI. The scheme of the URI determines which implementation is used:

| Scheme | Description |
| --- | --- |
| `pgp://` | Uses OpenPGP keys. |
| `x509://` | Uses X.509 certificates and private keys. |
| `acm://` | Uses X.509 certificates stored in the AWS Certificate Manager service |

### PGP Signing

To use PGP, your URI must include `key-uri` parameter (the path to the private key) and, if the key is locked, a `key-password-uri` parameter. For example

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "pgp://?key-uri=file:///path/to/private.key"

signer, _ := signatures.NewSigner(ctx, uri)

data := []byte("message to sign")
sig, _ := signer.Sign(ctx, data)
```

_Error handling omitted for the sake of brevity._

### x509 Signing

To use x509 certificates/keys to sign records your URI must include `certificate-uri` and `key-uri` query paramters. For example:

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "x509://?certificate-uri=file:///path/to/cert.pem&key-uri=file:///path/to/key.pem"

signer, _ := signatures.NewSigner(ctx, uri)

data := []byte("message to sign")
sig, _ := signer.Sign(ctx, data)
```

_Error handling omitted for the sake of brevity._

### x509 Signing (with AWS Certificate Manager)

To use x509 certificates/keys, derived from an AWS Certificate Manager listing, to sign records your URI must include `region`, `credentials` and `arn` query paramters. For example:

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "acm://?region={AWS_REGION}&credentials={AWS_CREDENTIALS}&arn={AWS_CERTIFICATE_MANAGER_ARN}"

signer, _ := signatures.NewSigner(ctx, uri)

data := []byte("message to sign")
sig, _ := signer.Sign(ctx, data)
```

_Error handling omitted for the sake of brevity._

### Credentials

Under the hood this package uses the [aaronland/go-aws/v3/auth](https://github.com/aaronland/go-aws/tree/main/auth#credentials) credentials strings to authenticate with AWS services. Valid options are:

| Label | Description |
| --- | --- |
| `anon:` | Empty or anonymous credentials. |
| `env:` | Read credentials from AWS defined environment variables. |
| `iam:` | Assume AWS IAM credentials are in effect. |
| `iam:{REGION}:{ARN}` | Assume AWS IAM credentials are in effect after assuming the IAM Role defined by `{ARN}` (in `{REGION}`). |
| `sts:{ARN}` | Assume the role defined by `{ARN}` using STS credentials. |
| `{AWS_PROFILE_NAME}` | This this profile from the default AWS credentials location. |
| `{AWS_CREDENTIALS_PATH}:{AWS_PROFILE_NAME}` | This this profile from a user-defined AWS credentials location. |

## Verifier

To create a `Verifier`, you provide a URI. The scheme of the URI determines which implementation is used:

| Scheme | Description |
| --- | --- |
| `pgp://` | Uses OpenPGP keys. |
| `x509://` | Uses X.509 certificates and private keys. |

### PGP Verification

To verify using PGP, your URI must include a `public-key-uri` query paramter. For example:

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "pgp://?public-key-uri=file:///path/to/public.key"

verifier, _ := signatures.NewVerifier(ctx, uri)

data := []byte("data to validate")
sig := []byte("signature of data")

is_valid, _ := verifier.Verify(ctx, data, sig)
```

_Error handling omitted for the sake of brevity._

### x509 Verification

To verify using a x509 certificate, your URI must include a `certificate-uri` query parameter. For example:

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "x509://?certificate-uri=file:///path/to/cert.pem"

verifier, _ := signatures.NewVerifier(ctx, uri)

data := []byte("data to validate")
sig := []byte("signature of data")

is_valid, _ := verifier.Verify(ctx, data, sig)
```

_Error handling omitted for the sake of brevity._

## Query parameters

### URIs

URIs for query parameters referencing public or private key data (`key-uri`, `key-password-uri`, `public-key-uri`, `certificate-uri`) may take any of the following forms:

* A path on the local filesystem
* `cwd://{PATH}` which will look for "{PATH}" in the current working directory
* A registered [gocloud.dev/runtimevar](https://pkg.go.dev/gocloud.dev/runtimevar) URI. Under the hood this package is using the [aaronland/gocloud/runtimevar](https://pkg.go.dev/github.com/aaronland/gocloud/runtimevar#section-readme) wrapper library.
