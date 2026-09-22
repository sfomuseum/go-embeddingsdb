package database

import (
	"context"
	"net/url"
	"testing"
)

func TestMultiDatabase(t *testing.T) {

	ctx := context.Background()

	db_uris := []string{
		"null://#512",
		"null://#768",
		"null://#1024",
	}

	db_q := url.Values{}
	db_q["database"] = db_uris

	db_u := url.URL{}
	db_u.Scheme = "multi"
	db_u.RawQuery = db_q.Encode()

	_, err := NewMultiDatabase(ctx, db_u.String())

	if err != nil {
		t.Fatalf("Failed to create multi database from string, %v", err)
	}

	db, err := NewMultiDatabaseFromURIs(ctx, db_uris...)

	if err != nil {
		t.Fatalf("Failed to create multi database from URIs, %v", err)
	}

	dims, err := db.Dimensions(ctx)

	if err != nil {
		t.Fatalf("Failed to derive dimensions, %v", err)
	}

	if len(dims) != len(db_uris) {
		t.Fatalf("Unexpected dimensions count, %d", len(dims))
	}
}
