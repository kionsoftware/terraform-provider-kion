package framework_test

import (
	"testing"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/framework"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceWithConfigure(t *testing.T) {
	t.Parallel()

	t.Run("stores valid client", func(t *testing.T) {
		t.Parallel()
		var r framework.ResourceWithConfigure
		client := &conns.KionClient{APIURL: "https://x"}
		resp := &resource.ConfigureResponse{}
		r.Configure(t.Context(), resource.ConfigureRequest{ProviderData: client}, resp)

		require.False(t, resp.Diagnostics.HasError())
		assert.Same(t, client, r.Meta())
	})

	t.Run("nil provider data is a no-op", func(t *testing.T) {
		t.Parallel()
		var r framework.ResourceWithConfigure
		resp := &resource.ConfigureResponse{}
		r.Configure(t.Context(), resource.ConfigureRequest{ProviderData: nil}, resp)

		require.False(t, resp.Diagnostics.HasError())
		assert.Nil(t, r.Meta())
	})

	t.Run("wrong type errors", func(t *testing.T) {
		t.Parallel()
		var r framework.ResourceWithConfigure
		resp := &resource.ConfigureResponse{}
		r.Configure(t.Context(), resource.ConfigureRequest{ProviderData: "not a client"}, resp)

		require.True(t, resp.Diagnostics.HasError())
	})
}

func TestDataSourceWithConfigure(t *testing.T) {
	t.Parallel()

	t.Run("stores valid client", func(t *testing.T) {
		t.Parallel()
		var d framework.DataSourceWithConfigure
		client := &conns.KionClient{APIURL: "https://x"}
		resp := &datasource.ConfigureResponse{}
		d.Configure(t.Context(), datasource.ConfigureRequest{ProviderData: client}, resp)

		require.False(t, resp.Diagnostics.HasError())
		assert.Same(t, client, d.Meta())
	})

	t.Run("wrong type errors", func(t *testing.T) {
		t.Parallel()
		var d framework.DataSourceWithConfigure
		resp := &datasource.ConfigureResponse{}
		d.Configure(t.Context(), datasource.ConfigureRequest{ProviderData: 123}, resp)

		require.True(t, resp.Diagnostics.HasError())
	})
}
