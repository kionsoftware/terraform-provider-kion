package kimport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListPagesPastPaddingRecords pins the shape that lost real records.
//
// /v4/compliance/program/{id}/family pads every page out to `total` instead of
// to the page size: with count=100 against 110 records, page 1 is 100 real plus
// 10 zero-valued fillers and page 2 is 10 real plus 100 fillers. Counting the
// padding made page 1 look complete, so paging stopped and the 10 real records
// on page 2 were never fetched -- reported as "10 record(s) skipped: no id",
// which reads like junk was discarded rather than records going missing.
func TestListPagesPastPaddingRecords(t *testing.T) {
	t.Parallel()

	const total, pageSize = 110, 100
	blank := map[string]any{"id": 0, "name": "", "description": "", "compliance_program_id": 0}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		items := make([]map[string]any, 0, total)
		switch page {
		case "1":
			for i := 1; i <= pageSize; i++ {
				items = append(items, map[string]any{"id": i, "name": fmt.Sprintf("fam-%d", i)})
			}
			for len(items) < total {
				items = append(items, blank)
			}
		case "2":
			for i := pageSize + 1; i <= total; i++ {
				items = append(items, map[string]any{"id": i, "name": fmt.Sprintf("fam-%d", i)})
			}
			for len(items) < total {
				items = append(items, blank)
			}
		default:
			for len(items) < total {
				items = append(items, blank)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"status": 200,
			"data":   map[string]any{"items": items, "total": total},
		}))
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL, "k", false, "").List(context.Background(), "/v4/compliance/program/5/family")
	require.NoError(t, err)

	assert.Len(t, got, total, "every real record should be fetched, including page 2's")

	ids := make(map[string]bool, len(got))
	for _, rec := range got {
		id := stringify(rec["id"])
		assert.NotEqual(t, "0", id, "padding records must not survive")
		ids[id] = true
	}
	assert.True(t, ids["101"], "the first record only reachable on page 2 is missing")
	assert.True(t, ids["110"], "the last record only reachable on page 2 is missing")
}

func TestIsBlankRecord(t *testing.T) {
	t.Parallel()

	blank := []map[string]any{
		{},
		{"id": float64(0), "name": ""},
		{"id": nil, "enabled": false, "tags": []any{}},
		{"outer": map[string]any{"id": float64(0)}},
	}
	for i, rec := range blank {
		assert.True(t, isBlankRecord(rec), "case %d should be blank: %v", i, rec)
	}

	populated := []map[string]any{
		{"id": float64(1)},
		{"id": float64(0), "name": "still a record"},
		{"enabled": true},
		{"outer": map[string]any{"id": float64(7)}},
	}
	for i, rec := range populated {
		assert.False(t, isBlankRecord(rec), "case %d should not be blank: %v", i, rec)
	}
}
