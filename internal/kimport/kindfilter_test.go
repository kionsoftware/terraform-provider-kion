package kimport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

// v4BillingSource is the shape GET /v4/billing-source returns: one polymorphic
// collection holding every payer type, discriminated by which nested key is
// present. Only custom_billing_source records are kion_billing_source.
var v4BillingSource = []map[string]any{
	{"id": 2, "aws_payer": map[string]any{"id": 1, "name": "Sandbox2"}},
	{"id": 5, "gcp_payer": map[string]any{"id": 1}},
	{"id": 8, "azure_payer": map[string]any{"id": 5, "name": "Azure Gov EA"}},
	{"id": 22, "custom_billing_source": map[string]any{"id": 5, "name": "FOCUS Databricks"}},
	{"id": 23, "custom_billing_source": map[string]any{"id": 6, "name": "Another custom"}},
	{"id": 30, "oci_payer": map[string]any{"id": 2}},
	{"id": 31, "anthropic_billing_source": map[string]any{"id": 1}},
}

func billingSourceServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"status": 200,
			"data":   map[string]any{"items": v4BillingSource, "total": len(v4BillingSource)},
		}))
	}))
}

// A flat collection that mixes kinds must be filtered, exactly as a
// parent-scoped one is. Before this, kion_billing_source claimed all seven:
// the by-id read then found no custom_billing_source key, flattened to zero,
// and the resource refreshed as id="0" -- which made every payer_id reference
// pointing at it resolve to 0.
func TestEnumerateFiltersFlatCollectionByKind(t *testing.T) {
	t.Parallel()
	srv := billingSourceServer(t)
	defer srv.Close()

	r := importmanifest.Resource{
		TFType:            "kion_billing_source",
		Kind:              "billing_source",
		ReadShape:         importmanifest.ShapeGeneric,
		Readable:          true,
		ListPath:          "/v4/billing-source",
		RequireValidField: "custom_billing_source",
		ImportID:          importmanifest.ImportID{Format: importmanifest.FormatID},
	}

	res := Enumerate(context.Background(), NewClient(srv.URL, "k", false, ""), r)

	require.Len(t, res.Records, 2, "only the custom records belong to this resource")
	ids := []string{res.Records[0].ID, res.Records[1].ID}
	assert.ElementsMatch(t, []string{"22", "23"}, ids)

	// The drop is reported, not silent: five records went somewhere else, and
	// an operator comparing counts against the UI needs to know why.
	assert.Contains(t, res.Reason, "5 record(s) in the collection")
	assert.Contains(t, res.Reason, "custom_billing_source")
}

// Without a discriminator every record is kept, so the filter cannot quietly
// shrink the collections that hold exactly one kind.
func TestEnumerateKeepsEverythingWithoutDiscriminator(t *testing.T) {
	t.Parallel()
	srv := billingSourceServer(t)
	defer srv.Close()

	r := importmanifest.Resource{
		TFType:    "kion_thing",
		Kind:      "thing",
		ReadShape: importmanifest.ShapeGeneric,
		Readable:  true,
		ListPath:  "/v4/billing-source",
		ImportID:  importmanifest.ImportID{Format: importmanifest.FormatID},
	}

	res := Enumerate(context.Background(), NewClient(srv.URL, "k", false, ""), r)

	assert.Len(t, res.Records, 7)
	assert.NotContains(t, res.Reason, "in the collection")
}

func TestFilterByKind(t *testing.T) {
	t.Parallel()

	kept, wrong := filterByKind(v4BillingSource,
		importmanifest.Resource{RequireValidField: "custom_billing_source"})
	assert.Len(t, kept, 2)
	assert.Equal(t, 5, wrong)

	// An empty object is not a record of this kind either.
	kept, wrong = filterByKind([]map[string]any{
		{"id": 1, "custom_billing_source": map[string]any{}},
		{"id": 2, "custom_billing_source": nil},
	}, importmanifest.Resource{RequireValidField: "custom_billing_source"})
	assert.Empty(t, kept)
	assert.Equal(t, 2, wrong)
}
