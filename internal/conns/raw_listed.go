package conns

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
)

// RawIDListed reports whether the collection at path lists a record with id.
// It accepts both list envelopes Kion uses: {"data":[…]} and
// {"data":{"items":[…]}}.
//
// It exists for private by-id reads that answer a missing record with a 500
// rather than a 404 (GET /v1/app-role/{id}): the public list, which omits
// deleted records, is the only way to tell "gone" from "broken". Any failure
// here is an error, never "absent", so a broken list cannot drop a live record
// from state. A 404 on the list itself is the one exception: it means the list's
// owner no longer exists, so nothing is listed under it.
func (c *KionClient) RawIDListed(ctx context.Context, path string, id int64) (bool, error) {
	ids, err := c.RawListIDs(ctx, path)
	if err != nil {
		// The list's owner is gone (a deleted parent), so nothing is under it.
		if IsRawNotFound(err) {
			return false, nil
		}
		return false, err
	}
	for _, got := range ids {
		if got == id {
			return true, nil
		}
	}
	return false, nil
}

// RawKeyListed is RawIDListed for a record whose id is a string (an OAuth
// client's), compared as text so a numeric id matches too.
func (c *KionClient) RawKeyListed(ctx context.Context, path, key string) (bool, error) {
	body, err := c.RawGet(ctx, path)
	if err != nil {
		if IsRawNotFound(err) {
			return false, nil
		}
		return false, err
	}
	recs, err := listRecords[keyRecord](path, body)
	if err != nil {
		return false, err
	}
	return slices.ContainsFunc(recs, func(r keyRecord) bool {
		var s string
		if json.Unmarshal(r.ID, &s) == nil {
			return s == key
		}
		return string(r.ID) == key
	}), nil
}

type keyRecord struct {
	ID json.RawMessage `json:"id"`
}

// listRecords decodes either list envelope: {"data":[…]} or {"data":{"items":[…]}}.
func listRecords[T any](path string, body []byte) ([]T, error) {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("decoding %s: %w", path, err)
	}
	var recs []T
	if err := json.Unmarshal(env.Data, &recs); err != nil {
		var paged struct {
			Items []T `json:"items"`
		}
		if perr := json.Unmarshal(env.Data, &paged); perr != nil {
			return nil, fmt.Errorf("decoding %s: data is neither a list nor {items}", path)
		}
		recs = paged.Items
	}
	return recs, nil
}

// RawListIDs returns the ids of every record the collection at path lists,
// from either list envelope: {"data":[…]} or {"data":{"items":[…]}}.
func (c *KionClient) RawListIDs(ctx context.Context, path string) ([]int64, error) {
	body, err := c.RawGet(ctx, path)
	if err != nil {
		return nil, err
	}
	recs, err := listRecords[struct {
		ID int64 `json:"id"`
	}](path, body)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, len(recs))
	for i, r := range recs {
		ids[i] = r.ID
	}
	return ids, nil
}
