package kimport

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotPortable(t *testing.T) {
	t.Parallel()
	results := []Result{
		{TFType: "kion_funding_source", Status: "ok", Records: []Record{
			{ID: "31", Name: "Future FS", Raw: map[string]any{"id": 31.0}},
			{ID: "32", Name: "Zero", Raw: map[string]any{"id": 32.0, "ou_id": 0.0}},
			{ID: "50", Name: "IRAD", Raw: map[string]any{"id": 50.0, "ou_id": 1.0}},
		}},
		{TFType: "kion_permission_scheme", Status: "ok", Records: []Record{
			{ID: "1", Name: "Global Permissions Scheme", Raw: map[string]any{"system_managed": true}},
			{ID: "9", Name: "Custom", Raw: map[string]any{"system_managed": false}},
		}},
		{TFType: "kion_user_group", Status: "ok", Records: []Record{
			{ID: "1", Name: "Administrators", Raw: map[string]any{"name": "Administrators"}},
			{ID: "2", Name: "Admins", Raw: map[string]any{"name": "Admins"}},
		}},
		{TFType: "kion_custom_variable", Status: "ok", Records: []Record{
			{ID: "12", Name: "system-resourcetags", Raw: map[string]any{"name": "system-resourcetags"}},
			{ID: "13", Name: "Department", Raw: map[string]any{"name": "Department"}},
		}},
	}
	got := NotPortable(results)
	var keys []string
	for _, w := range got {
		keys = append(keys, w.TFType+"/"+w.ID)
	}
	assert.Equal(t, []string{
		"kion_custom_variable/12",
		"kion_funding_source/31",
		"kion_funding_source/32",
		"kion_permission_scheme/1",
		"kion_user_group/1",
	}, keys)
}

func TestNotPortableRendered(t *testing.T) {
	t.Parallel()
	results := []Result{{TFType: "kion_funding_source", Status: "ok", Records: []Record{
		{ID: "31", Name: "Future FS", Raw: map[string]any{"id": 31.0}},
	}}}

	out := RenderImports(results, "1.0.0")
	i := strings.Index(out, "# not portable:")
	require.GreaterOrEqual(t, i, 0, out)
	assert.Less(t, i, strings.Index(out, `id = "31"`), "the warning sits above its import block")

	rep := FormatReport(results)
	assert.Contains(t, rep, "Will not apply to another install as-is (1):")
	assert.Contains(t, rep, `kion_funding_source "Future FS" (31)`)
}
