package testdata

import "context"

// Client is a miniature stand-in for the ogen-generated SDK client. Only the
// Label ops are present. The crud generator parses this file with go/parser
// (no type-checking), so unresolved identifiers in the signatures are fine;
// keeping it minimal avoids cross-fixture duplicate declarations and gofmt
// churn. Go tooling never compiles files under testdata/.

// PostLabel invokes PostLabel operation.
//
// Creates a new app label in the application.
//
// POST /v3/label
func (c *Client) PostLabel(ctx context.Context, request OptCreateLabel) (PostLabelRes, error) {
	return nil, nil
}

// GetLabel invokes GetLabel operation.
//
// GET /v3/label/{id}
func (c *Client) GetLabel(ctx context.Context, params GetLabelParams) (GetLabelRes, error) {
	return nil, nil
}

// PatchLabel invokes PatchLabel operation.
//
// PATCH /v3/label/{id}
func (c *Client) PatchLabel(ctx context.Context, request OptUpdateLabel, params PatchLabelParams) (PatchLabelRes, error) {
	return nil, nil
}

// DeleteLabel invokes DeleteLabel operation.
//
// DELETE /v3/label/{id}
func (c *Client) DeleteLabel(ctx context.Context, params DeleteLabelParams) (DeleteLabelRes, error) {
	return nil, nil
}

// GetLabelIndex invokes GetLabelIndex operation.
//
// GET /v3/label
func (c *Client) GetLabelIndex(ctx context.Context, params GetLabelIndexParams) (GetLabelIndexRes, error) {
	return nil, nil
}

// GetLabelFlat stands in for the many Kion index endpoints whose response puts
// the items slice directly in `data` and which take no pagination params at all
// (e.g. GET /v3/user -> UserListResponse{Data []User}).
//
// GET /v3/label-flat
func (c *Client) GetLabelFlat(ctx context.Context) (GetLabelFlatRes, error) {
	return nil, nil
}

// GetLabelChild stands in for a nested collection: the items are a real Label
// slice, but the op cannot be called without a parent id.
//
// GET /v3/label/{id}/child
func (c *Client) GetLabelChild(ctx context.Context, params GetLabelChildParams) (GetLabelChildRes, error) {
	return nil, nil
}

// GetLabelForeign stands in for an index whose element type is unrelated to the
// single read's payload.
//
// GET /v3/label-foreign
func (c *Client) GetLabelForeign(ctx context.Context) (GetLabelForeignRes, error) {
	return nil, nil
}
