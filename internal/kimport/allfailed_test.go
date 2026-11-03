package kimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllReadsFailed(t *testing.T) {
	t.Parallel()
	ok := Result{TFType: "kion_ou", Status: "ok"}
	empty := Result{TFType: "kion_label", Status: "empty"}
	failed := Result{TFType: "kion_project", Status: "error", Reason: "401 Unauthorized"}
	unsupported := Result{TFType: "kion_aws_iam_policy", Status: "unsupported"}

	cases := []struct {
		name    string
		results []Result
		want    bool
	}{
		{"every read errored", []Result{failed, failed}, true},
		{"errors beside unsupported types", []Result{failed, unsupported}, true},
		{"one read succeeded", []Result{failed, ok}, false},
		{"an empty read is a successful read", []Result{failed, empty}, false},
		// Nothing was read, so nothing failed: an --include of only aliases.
		{"only unsupported types", []Result{unsupported}, false},
		{"no results", nil, false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, AllReadsFailed(tc.results), tc.name)
	}
}
