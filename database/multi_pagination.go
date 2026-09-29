package database

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
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
	InternalPage  int64                  `json:"page"`
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
	perPage         int64
	total           int64
	nextPointer     any
	previousPointer any
	method          pagination.Method
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
	return m.nextPointer
}

func (m *MultiDatabasePaginationResults) Previous() any {
	return m.previousPointer
}
func (m *MultiDatabasePaginationResults) Method() pagination.Method {
	return m.method
}

func (m *MultiDatabasePaginationResults) NextURL(t *uritemplates.UriTemplate) (string, error) {

	if m.nextPointer == nil || t == nil {
		return "", nil
	}

	state, ok := m.nextPointer.(MultiDatabaseCursorState)

	if !ok {

		ptr, ok := m.nextPointer.(*MultiDatabaseCursorState)

		if ok {
			state = *ptr
		} else {
			return "", fmt.Errorf("invalid next pointer type")
		}
	}

	return t.Expand(map[string]interface{}{
		"pointer":  state.String(),
		"per_page": m.perPage,
	})
}

func (m *MultiDatabasePaginationResults) PreviousURL(t *uritemplates.UriTemplate) (string, error) {

	if m.previousPointer == nil || t == nil {
		return "", nil
	}

	state, ok := m.previousPointer.(MultiDatabaseCursorState)

	if !ok {

		ptr, ok := m.previousPointer.(*MultiDatabaseCursorState)

		if ok {
			state = *ptr
		} else {
			return "", fmt.Errorf("invalid previous pointer type")
		}
	}

	return t.Expand(map[string]interface{}{
		"pointer":  state.String(),
		"per_page": m.perPage,
	})
}
