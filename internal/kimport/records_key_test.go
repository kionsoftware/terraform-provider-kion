package kimport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"terraform-provider-kion/internal/kgen/importmanifest"
)

// The OU cloud access role listing carries its exemptions beside two other
// arrays, so the generic unwrapping cannot pick them; ListUnder takes the key.
const ouRolesBody = `{"status":200,"data":{
	"to_keep":[{"id":40,"name":"role"}],
	"to_remove":[{"id":41}],
	"project_exemptions":null,
	"ou_exemptions":[{"id":7,"ou_id":266,"ou_cloud_access_role_id":40,"reason":"why"}]}}`

func TestListUnderNamedKey(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, ouRolesBody)
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL, "k", false, "").ListUnder(context.Background(), "/v1/ou/266/ou-cloud-access-role", "ou_exemptions")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, float64(7), got[0]["id"])
	assert.Equal(t, "why", got[0]["reason"])
}

func TestListUnderNullKeyIsEmpty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `{"status":200,"data":{"to_keep":[{"id":1}],"ou_exemptions":null}}`)
	}))
	defer srv.Close()

	got, err := NewClient(srv.URL, "k", false, "").ListUnder(context.Background(), "/v1/ou/1/ou-cloud-access-role", "ou_exemptions")
	require.NoError(t, err)
	assert.Empty(t, got, "a null exemption list is none, never the sibling roles")
}

// TestEnumerateReadsChildRecordsKey: with ChildRecordsKey set, the enumerator
// reads the named list and not the parent's cloud access roles beside it.
func TestEnumerateReadsChildRecordsKey(t *testing.T) {
	l := &routeLister{routes: map[string]any{
		"/v3/ou": []map[string]any{{"id": float64(266)}},
		"/v1/ou/266/ou-cloud-access-role": map[string]any{
			"to_keep":       []any{map[string]any{"id": float64(40)}},
			"ou_exemptions": []any{map[string]any{"id": float64(7), "ou_id": float64(266), "ou_cloud_access_role_id": float64(40)}},
		},
	}}
	r := importmanifest.Resource{
		TFType: "kion_ou_cloud_access_role_exemption", Kind: "ou_cloud_access_role_exemption",
		Archetype: "no_read", ReadShape: importmanifest.ShapeParentList, Readable: true,
		Parent: &importmanifest.Parent{
			Kind: "ou", ListPath: "/v3/ou", ChildPath: "/v1/ou/{parent_id}/ou-cloud-access-role",
			ParentIDField: "ou_id", ChildRecordsKey: "ou_exemptions",
		},
		ImportID: importmanifest.ImportID{Format: importmanifest.FormatParentSlashKey, KeyField: "id"},
	}

	res := Enumerate(context.Background(), l, r)
	require.Equal(t, "ok", res.Status, res.Reason)
	require.Len(t, res.Records, 1)
	assert.Equal(t, "266/7", res.Records[0].ID)
}
