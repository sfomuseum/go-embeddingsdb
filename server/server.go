package server

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/aaronland/go-roster"
)

// Server defines an interface for a network-based interface for interacting with an embeddings database.
type Server interface {
	// ListenAndServe starts a new server and listens for requests.
	ListenAndServe(context.Context) error
}

var server_roster roster.Roster

// ServerInitializationFunc is a function defined by individual server package and used to create
// an instance of that server
type ServerInitializationFunc func(ctx context.Context, uri string) (Server, error)

// RegisterServer registers 'scheme' as a key pointing to 'init_func' in an internal lookup table
// used to create new `Server` instances by the `NewServer` method.
func RegisterServer(ctx context.Context, scheme string, init_func ServerInitializationFunc) error {

	err := ensureServerRoster()

	if err != nil {
		return err
	}

	return server_roster.Register(ctx, scheme, init_func)
}

func ensureServerRoster() error {

	if server_roster == nil {

		r, err := roster.NewDefaultRoster()

		if err != nil {
			return err
		}

		server_roster = r
	}

	return nil
}

// NewServer returns a new `Server` instance configured by 'uri'. The value of 'uri' is parsed
// as a `url.URL` and its scheme is used as the key for a corresponding `ServerInitializationFunc`
// function used to instantiate the new `Server`. It is assumed that the scheme (and initialization
// function) have been registered by the `RegisterServer` method.
func NewServer(ctx context.Context, uri string) (Server, error) {

	u, err := url.Parse(uri)

	if err != nil {
		return nil, err
	}

	scheme := u.Scheme

	i, err := server_roster.Driver(ctx, scheme)

	if err != nil {
		return nil, err
	}

	if i == nil {
		return nil, fmt.Errorf("Unregistered server")
	}

	init_func := i.(ServerInitializationFunc)

	return init_func(ctx, uri)
}

// Schemes returns the list of schemes that have been registered.
func Schemes() []string {

	ctx := context.Background()
	schemes := []string{}

	err := ensureServerRoster()

	if err != nil {
		return schemes
	}

	for _, dr := range server_roster.Drivers(ctx) {
		scheme := fmt.Sprintf("%s://", strings.ToLower(dr))
		schemes = append(schemes, scheme)
	}

	sort.Strings(schemes)
	return schemes
}
