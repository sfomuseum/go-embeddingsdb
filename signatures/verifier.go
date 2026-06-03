package signatures

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// Verifier defines the interface for validating digital signatures against 
// raw data payloads.
type Verifier interface {
	// Verify checks if the provided signature is valid for the given data.
	Verify(context.Context, []byte, []byte) (bool, error)
}

// VerifierInitializationFunc is a function defined by individual verifier package and used to create
// an instance of that verifier
type VerifierInitializationFunc func(ctx context.Context, uri string) (Verifier, error)

var verifier_roster roster.Roster

// RegisterVerifier registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Verifier` instances by the `NewVerifier` method.
func RegisterVerifier(ctx context.Context, scheme string, init_func VerifierInitializationFunc) error {

	err := ensureVerifierRoster()

	if err != nil {
		return err
	}

	return verifier_roster.Register(ctx, scheme, init_func)
}

func ensureVerifierRoster() error {

	if verifier_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		verifier_roster = r
	}

	return nil
}

// NewVerifier returns a new `Verifier` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `VerifierInitializationFunc`
// function used to instantiate the new `Verifier`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterVerifier` method.
func NewVerifier(ctx context.Context, uri string) (Verifier, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := verifier_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	init_func := i.(VerifierInitializationFunc)
	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func VerifierSchemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureVerifierRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range verifier_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
