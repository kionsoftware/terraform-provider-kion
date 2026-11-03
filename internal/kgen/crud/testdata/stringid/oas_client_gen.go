package testdata

import "context"

// Client is a miniature SDK client whose record id is a string, shaped like
// the v3_17 OAuth client ops. Parsed with go/parser only; never compiled.

// PostThing invokes PostThing operation.
//
// POST /v3/thing
func (c *Client) PostThing(ctx context.Context, request OptThingCreateRequest) (PostThingRes, error) {
	return nil, nil
}

// GetThing invokes GetThing operation.
//
// GET /v3/thing/{id}
func (c *Client) GetThing(ctx context.Context, params GetThingParams) (GetThingRes, error) {
	return nil, nil
}

// PutThing invokes PutThing operation.
//
// PUT /v3/thing/{id}
func (c *Client) PutThing(ctx context.Context, request OptThingUpdateRequest, params PutThingParams) (PutThingRes, error) {
	return nil, nil
}

// DeleteThing invokes DeleteThing operation.
//
// DELETE /v3/thing/{id}
func (c *Client) DeleteThing(ctx context.Context, params DeleteThingParams) (DeleteThingRes, error) {
	return nil, nil
}

// GetThings invokes GetThings operation.
//
// GET /v3/thing
func (c *Client) GetThings(ctx context.Context) (GetThingsRes, error) {
	return nil, nil
}
