package testdata

import "context"

// Client is a miniature SDK client with an integer id whose create returns a
// pointer envelope. Parsed with go/parser only; never compiled.

// PostWidget invokes PostWidget operation.
//
// POST /v3/widget
func (c *Client) PostWidget(ctx context.Context, request OptWidgetCreateRequest) (PostWidgetRes, error) {
	return nil, nil
}

// GetWidget invokes GetWidget operation.
//
// GET /v3/widget/{id}
func (c *Client) GetWidget(ctx context.Context, params GetWidgetParams) (GetWidgetRes, error) {
	return nil, nil
}

// PutWidget invokes PutWidget operation.
//
// PUT /v3/widget/{id}
func (c *Client) PutWidget(ctx context.Context, request OptWidgetUpdateRequest, params PutWidgetParams) (PutWidgetRes, error) {
	return nil, nil
}

// DeleteWidget invokes DeleteWidget operation.
//
// DELETE /v3/widget/{id}
func (c *Client) DeleteWidget(ctx context.Context, params DeleteWidgetParams) (DeleteWidgetRes, error) {
	return nil, nil
}
