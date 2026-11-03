package config_test

import (
	"sort"
	"testing"

	"terraform-provider-kion/internal/kgen/config"
	"terraform-provider-kion/internal/kgen/config/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDerive checks the core mapping against a mocked Source: implemented
// resources come from the code's ops, stubs fall back to the spec naming
// convention (flagged heuristic), and resources with no resolvable ops drop out.
func TestDerive(t *testing.T) {
	m := mocks.NewMockSource(t)

	ops := map[string]config.Op{
		"PostLabel":     {Method: "POST", Path: "/v3/label"},
		"GetLabel":      {Method: "GET", Path: "/v3/label/{id}"},
		"PatchLabel":    {Method: "PATCH", Path: "/v3/label/{id}"},
		"DeleteLabel":   {Method: "DELETE", Path: "/v3/label/{id}"},
		"PostWidget":    {Method: "POST", Path: "/v3/widget"},
		"GetWidget":     {Method: "GET", Path: "/v3/widget/{id}"},
		"GetThingIndex": {Method: "GET", Path: "/v3/thing"},
	}
	svcs := []config.ServiceOps{
		{Name: "label", Create: []string{"PostLabel"}, Read: []string{"GetLabel"}, Update: []string{"PatchLabel"}, Delete: []string{"DeleteLabel"}},
		{Name: "widget"}, // stub -> heuristic (Post/GetWidget exist)
		{Name: "orphan"}, // stub, no matching ops -> dropped
		{Name: "thing", DataSourceRead: []string{"GetThingIndex"}},
	}
	m.EXPECT().Operations("spec.json").Return(ops, nil)
	m.EXPECT().ServiceOps("svc").Return(svcs, nil)

	got, err := config.Derive(m, config.Options{Spec: "spec.json", ServiceRoot: "svc"})
	require.NoError(t, err)

	byName := map[string]config.ServiceConfig{}
	for _, c := range got {
		byName[c.Name] = c
	}

	// label: authoritative from code, not heuristic.
	lbl := byName["label"]
	require.NotNil(t, lbl.Resource)
	require.NotNil(t, lbl.Resource.Create)
	assert.False(t, lbl.ResourceHeuristic)
	assert.Equal(t, "/v3/label", lbl.Resource.Create.Path)
	assert.Equal(t, "/v3/label/{id}", lbl.Resource.Read.Path)
	assert.Equal(t, "DELETE", lbl.Resource.Delete.Method)

	// widget: heuristic from spec naming (create + read resolved, no patch/delete op).
	w := byName["widget"]
	require.NotNil(t, w.Resource)
	assert.True(t, w.ResourceHeuristic)
	assert.Equal(t, "/v3/widget", w.Resource.Create.Path)
	assert.Equal(t, "/v3/widget/{id}", w.Resource.Read.Path)
	assert.Nil(t, w.Resource.Delete)

	// orphan: nothing resolvable -> absent.
	_, ok := byName["orphan"]
	assert.False(t, ok)

	// thing: data source read only.
	th := byName["thing"]
	assert.Nil(t, th.Resource)
	require.NotNil(t, th.DataSourceRead)
	assert.Equal(t, "/v3/thing", th.DataSourceRead.Path)

	assert.True(t, sort.SliceIsSorted(got, func(i, j int) bool { return got[i].Name < got[j].Name }), "output must be sorted by name")
}
