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

To create a Signer, you provide a URI. The scheme of the URI determines which implementation is used:

* `pgp://` Uses OpenPGP keys.
* `tls://` Uses X.509 certificates and private keys.

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

### TLS Signing

To use TLS, your URI must include `certificate-uri` and `key-uri` query paramters. For example:

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "tls://?certificate-uri=file:///path/to/cert.pem&key-uri=file:///path/to/key.pem"

signer, _ := signatures.NewSigner(ctx, uri)

data := []byte("message to sign")
sig, _ := signer.Sign(ctx, data)
```

_Error handling omitted for the sake of brevity._

## Verifier

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

### TLS Verification

To verify using a TLS certificate, your URI must include a `certificate-uri` query parameter. For example:

```
import(
	"context"
	
	"github.com/sfomuseum/go-embeddingsdb/signatures"
)

ctx := context.Background()
uri := "tls://?certificate-uri=file:///path/to/cert.pem"

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
