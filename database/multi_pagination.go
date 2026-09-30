package database

import (
	"encoding/base64"
	"encoding/json"
	_ "fmt"
	"log/slog"

	"github.com/aaronland/go-pagination"
	"github.com/jtacoma/uritemplates"
)

type MultiDatabaseDirection string

const (
	DirectionNext     MultiDatabaseDirection = "next"
	DirectionPrevious MultiDatabaseDirection = "prev"
)

type MultiDatabaseCursorState struct {
	DatabaseIndex int                    `json:"db_idx"`
	Page          int64                  `json:"page"`
	Cursor        string                 `json:"cursor"`
	Direction     MultiDatabaseDirection `json:"dir"` // Keeps track of which way we are shifting
}

// String serializes the MultiDatabaseCursorState into a URL-safe Base64 encoded string.
func (c MultiDatabaseCursorState) String() string {

	b, err := json.Marshal(c)

	if err != nil {
		slog.Error("Failed to marshal cursor", "error", err)
		return ""
	}

	return base64.URLEncoding.EncodeToString(b)
}

// ParseCursorState decodes a Base64 string back into a MultiDatabaseCursorState.
func ParseCursorState(s string) (MultiDatabaseCursorState, error) {

	var state MultiDatabaseCursorState
	b, err := base64.URLEncoding.DecodeString(s)

	if err != nil {
		return state, err
	}

	err = json.Unmarshal(b, &state)
	return state, err
}

type MultiDatabasePaginationResults struct {
	pagination.Results
	perPage  int64
	total    int64
	next     *MultiDatabaseCursorState
	previous *MultiDatabaseCursorState
	method   pagination.Method
}

func (m *MultiDatabasePaginationResults) Total() int64 {
	return m.total
}

func (m *MultiDatabasePaginationResults) PerPage() int64 {
	return m.perPage
}

func (m *MultiDatabasePaginationResults) Page() int64 {
	return 0
}

func (m *MultiDatabasePaginationResults) Pages() int64 {

	if m.perPage == 0 {
		return 0
	}

	return (m.total + m.perPage - 1) / m.perPage
}

func (m *MultiDatabasePaginationResults) Next() any {
	return m.next
}

func (m *MultiDatabasePaginationResults) Previous() any {
	return m.previous
}

func (m *MultiDatabasePaginationResults) Method() pagination.Method {
	return m.method
}

func (m *MultiDatabasePaginationResults) NextURL(t *uritemplates.UriTemplate) (string, error) {

	if m.next == nil || t == nil {
		return "", nil
	}

	return t.Expand(map[string]interface{}{
		"cursor":   m.next.String(),
		"per_page": m.perPage,
	})
}

func (m *MultiDatabasePaginationResults) PreviousURL(t *uritemplates.UriTemplate) (string, error) {

	if m.previous == nil || t == nil {
		return "", nil
	}

	return t.Expand(map[string]interface{}{
		"cursor":   m.previous.String(),
		"per_page": m.perPage,
	})
}
