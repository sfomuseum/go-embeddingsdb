package database

import (
	"context"
	"testing"
)

func TestMultiDatabase(t *testing.T) {

	ctx := context.Background()

	db_uris := []string{
		"null://?dimensions=512",
		"null://?dimensions=768",
		"null://?dimensions=1024",
	}

	db_u := NewMultiDatabaseURIFromURIs(db_uris...)

	_, err := NewMultiDatabase(ctx, db_u)

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
