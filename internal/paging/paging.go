// Package paging walks a Kion list endpoint to exhaustion. Endpoints serve ten
// records unless asked otherwise, so one unpaged request returns a truncated
// collection that looks complete.
//
// Not every collection caps an unpaged request, and the path does not say which.
// A handler on the backend's CreatePaginationQueryWithDefaults caps at ten
// (/v1/automation-policy); one on CreatePaginationQuery leaves the limit unset
// and returns everything (/v2/funding-source/{id}/funding-source-note). Check
// which before assuming either.
//
// Callers supply a fetch returning one decoded page; this decides how many to
// ask for. Decoding stays with the caller.
package paging

import (
	"context"
	"fmt"
)

const (
	// DefaultPageSize is what list endpoints are asked for. It is not the
	// server's default, which is ten.
	DefaultPageSize = 100

	// MaxPages bounds the walk when a reported total never matches what the
	// server serves. No real collection should reach it.
	MaxPages = 1000
)

// Page is one decoded page of a list response.
type Page[T any] struct {
	Items []T

	// Total is the collection size the server reports. A NEGATIVE Total means
	// the response was not a paginated envelope at all, so the first page is
	// the entire collection and no further requests are made.
	Total int
}

// Fetch returns one decoded page. Pages are 1-based.
type Fetch[T any] func(ctx context.Context, page int) (Page[T], error)

// All walks every page and returns the whole collection. label prefixes the
// page-cap error; errors from fetch pass through verbatim. A failure after the
// first page returns the records gathered so far alongside the error.
func All[T any](ctx context.Context, label string, fetch Fetch[T]) ([]T, error) {
	first, err := fetch(ctx, 1)
	if err != nil {
		return nil, err
	}
	records := first.Items
	if first.Total < 0 {
		return records, nil
	}

	for page := 2; len(records) < first.Total; page++ {
		if page > MaxPages {
			return records, fmt.Errorf("%s: paging exceeded max pages (%d)", label, MaxPages)
		}
		next, err := fetch(ctx, page)
		if err != nil {
			return records, err
		}
		// An empty page ends the walk even when total disagrees, so a wrong
		// total cannot spin this to the cap.
		if len(next.Items) == 0 {
			break
		}
		records = append(records, next.Items...)
	}
	return records, nil
}
