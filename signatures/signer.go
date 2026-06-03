package signatures

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

type Signer interface {
	Sign(context.Context, []byte) ([]byte, error)
	Verifier(context.Context) (Verifier, error)
	PublicKey(context.Context) ([]byte, error)
}

// SignerInitializationFunc is a function defined by individual signer package and used to create
// an instance of that signer
type SignerInitializationFunc func(ctx context.Context, uri string) (Signer, error)

var signer_roster roster.Roster

// RegisterSigner registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Signer` instances by the `NewSigner` method.
func RegisterSigner(ctx context.Context, scheme string, init_func SignerInitializationFunc) error {

	err := ensureSignerRoster()

	if err != nil {
		return err
	}

	return signer_roster.Register(ctx, scheme, init_func)
}

func ensureSignerRoster() error {

	if signer_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		signer_roster = r
	}

	return nil
}

// NewSigner returns a new `Signer` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `SignerInitializationFunc`
// function used to instantiate the new `Signer`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterSigner` method.
func NewSigner(ctx context.Context, uri string) (Signer, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := signer_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(SignerInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func SignerSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureSignerRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range signer_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
