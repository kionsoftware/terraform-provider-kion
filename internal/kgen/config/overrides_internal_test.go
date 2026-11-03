package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestApplyOverrides checks that overrides fill missing verbs, clear the
// heuristic flag (marking Overridden), and add override-only entries. Uses
// unexported symbols, so it lives in the internal test package.
func TestApplyOverrides(t *testing.T) {
	configs := []ServiceConfig{
		{Name: "foo", Resource: &CRUD{Create: &Op{Method: "POST", Path: "/v3/foo"}}, ResourceHeuristic: true},
	}
	ov := overridesFile{
		Resources: map[string]ovEntry{
			"foo": {Read: &yOp{Path: "/v3/foo/{id}", Method: "GET"}},
		},
		DataSources: map[string]ovEntry{
			"bar": {Read: &yOp{Path: "/v3/bar", Method: "GET"}},
		},
	}

	got := applyOverrides(configs, ov)
	byName := map[string]ServiceConfig{}
	for _, c := range got {
		byName[c.Name] = c
	}

	foo := byName["foo"]
	require.NotNil(t, foo.Resource.Read)
	assert.Equal(t, "/v3/foo/{id}", foo.Resource.Read.Path)
	assert.Equal(t, "/v3/foo", foo.Resource.Create.Path, "existing verb preserved")
	assert.False(t, foo.ResourceHeuristic, "override clears heuristic")
	assert.True(t, foo.Overridden)

	bar := byName["bar"]
	assert.Nil(t, bar.Resource)
	require.NotNil(t, bar.DataSourceRead, "override-only data source added")
	assert.Equal(t, "/v3/bar", bar.DataSourceRead.Path)
}

func TestToPascal(t *testing.T) {
	cases := map[string]string{
		"label":           "Label",
		"ou_note":         "OUNote",
		"account_linkage": "AccountLinkage",
		"aws_account":     "AWSAccount",
		"gcp_iam_role":    "GCPIAMRole",
	}
	for in, want := range cases {
		if got := toPascal(in); got != want {
			t.Errorf("toPascal(%q) = %q, want %q", in, got, want)
		}
	}
}
