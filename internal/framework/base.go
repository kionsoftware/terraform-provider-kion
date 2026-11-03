// Package framework provides base types for Terraform resources and data sources.
package framework

import (
	"context"

	"terraform-provider-kion/internal/conns"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ResourceWithConfigure is an embeddable struct for resources that need
// access to the configured KionClient. Embed this in your resource struct
// and call Meta() to get the client.
type ResourceWithConfigure struct {
	meta *conns.KionClient
}

// Meta returns the provider's KionClient. Panics if Configure has not been called.
func (r *ResourceWithConfigure) Meta() *conns.KionClient {
	return r.meta
}

// Configure stores the provider-configured client for later use by CRUD methods.
func (r *ResourceWithConfigure) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*conns.KionClient)
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Unexpected Resource Configure Type",
			"Expected *conns.KionClient, got unexpected type. Please report this issue to the provider developers.",
		))
		return
	}

	r.meta = client
}

// DataSourceWithConfigure is an embeddable struct for data sources that need
// access to the configured KionClient.
type DataSourceWithConfigure struct {
	meta *conns.KionClient
}

// Meta returns the provider's KionClient.
func (d *DataSourceWithConfigure) Meta() *conns.KionClient {
	return d.meta
}

// Configure stores the provider-configured client for later use by Read.
func (d *DataSourceWithConfigure) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*conns.KionClient)
	if !ok {
		resp.Diagnostics.Append(diag.NewErrorDiagnostic(
			"Unexpected DataSource Configure Type",
			"Expected *conns.KionClient, got unexpected type. Please report this issue to the provider developers.",
		))
		return
	}

	d.meta = client
}
