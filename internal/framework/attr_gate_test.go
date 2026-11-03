package framework_test

import (
	"testing"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/framework"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// planWith builds a raw plan holding the given attributes; a nil value is an
// attribute the practitioner left unset.
func planWith(attrs map[string]*string) tfsdk.Plan {
	types := map[string]tftypes.Type{}
	vals := map[string]tftypes.Value{}
	for name, v := range attrs {
		types[name] = tftypes.String
		if v == nil {
			vals[name] = tftypes.NewValue(tftypes.String, nil)
			continue
		}
		vals[name] = tftypes.NewValue(tftypes.String, *v)
	}
	obj := tftypes.Object{AttributeTypes: types}
	return tfsdk.Plan{Raw: tftypes.NewValue(obj, vals)}
}

func meta(version string) *conns.KionClient {
	return &conns.KionClient{
		Version:         conns.MustParseKionVersion(version),
		VersionDetected: true,
	}
}

func TestRequireAttrKionVersions_errorsOnTooNewAttribute(t *testing.T) {
	set := "policy-1"
	diags := framework.RequireAttrKionVersions(
		meta("3.13.0"),
		planWith(map[string]*string{"automation_policy_ids": &set}),
		map[string]framework.AttrVersions{"automation_policy_ids": {{Min: conns.MustParseKionVersion("3.16.0")}}},
		"kion_cloud_rule",
	)
	require.True(t, diags.HasError(), "a 3.16 attribute on a 3.13 instance must error")
	assert.Contains(t, diags[0].Detail(), "kion_cloud_rule.automation_policy_ids")
	assert.Contains(t, diags[0].Detail(), "3.16.0")
	assert.Contains(t, diags[0].Detail(), "3.13.0")
}

// The silent-drop only happens for values actually sent, so an unset attribute
// must not error. Otherwise every 3.13 user is blocked from the resource.
func TestRequireAttrKionVersions_ignoresUnsetAttribute(t *testing.T) {
	diags := framework.RequireAttrKionVersions(
		meta("3.13.0"),
		planWith(map[string]*string{"automation_policy_ids": nil}),
		map[string]framework.AttrVersions{"automation_policy_ids": {{Min: conns.MustParseKionVersion("3.16.0")}}},
		"kion_cloud_rule",
	)
	assert.False(t, diags.HasError(), "an unset attribute is never sent, so nothing can be dropped")
}

func TestRequireAttrKionVersions_allowsNewEnoughInstance(t *testing.T) {
	set := "policy-1"
	diags := framework.RequireAttrKionVersions(
		meta("3.16.0"),
		planWith(map[string]*string{"automation_policy_ids": &set}),
		map[string]framework.AttrVersions{"automation_policy_ids": {{Min: conns.MustParseKionVersion("3.16.0")}}},
		"kion_cloud_rule",
	)
	assert.False(t, diags.HasError())
}

func TestRequireAttrKionVersions_quietWhenVersionUndetected(t *testing.T) {
	// The resource gate already warns once; repeating it per attribute would
	// bury the warning that matters.
	set := "policy-1"
	diags := framework.RequireAttrKionVersions(
		&conns.KionClient{VersionDetected: false},
		planWith(map[string]*string{"automation_policy_ids": &set}),
		map[string]framework.AttrVersions{"automation_policy_ids": {{Min: conns.MustParseKionVersion("3.16.0")}}},
		"kion_cloud_rule",
	)
	assert.Empty(t, diags)
}

func TestRequireAttrKionVersions_noMapIsNoop(t *testing.T) {
	set := "x"
	diags := framework.RequireAttrKionVersions(
		meta("3.12.0"), planWith(map[string]*string{"anything": &set}), nil, "kion_label")
	assert.Empty(t, diags)
}

// carWindow is the real availability of cloud_access_role_type_id: 3.15.13 on
// the 3.15 line, 3.16.5 on the 3.16 line, absent from 3.17.
func carWindow() map[string]framework.AttrVersions {
	return map[string]framework.AttrVersions{
		"cloud_access_role_type_id": {
			{Min: conns.MustParseKionVersion("3.15.13"), Before: conns.MustParseKionVersion("3.16.0")},
			{Min: conns.MustParseKionVersion("3.16.5"), Before: conns.MustParseKionVersion("3.17.0")},
		},
	}
}

func TestRequireAttrKionVersions_perLineFloors(t *testing.T) {
	set := "2"
	cases := []struct {
		version string
		accept  bool
	}{
		{"3.15.12", false}, // one patch below the 3.15 floor
		{"3.15.13", true},
		{"3.15.20", true},
		{"3.16.0", false}, // inside 3.16 but below its own floor
		{"3.16.4", false},
		{"3.16.5", true},
		{"3.17.0", false}, // dropped on the newer line
	}
	for _, c := range cases {
		diags := framework.RequireAttrKionVersions(
			meta(c.version),
			planWith(map[string]*string{"cloud_access_role_type_id": &set}),
			carWindow(),
			"kion_ou_cloud_access_role",
		)
		assert.Equal(t, !c.accept, diags.HasError(), "Kion %s", c.version)
	}
}

// A single range spanning 3.15.13 to 3.17.0 would accept 3.16.0, which is the
// case per-line floors exist to catch.
func TestRequireAttrKionVersions_diagnosticNamesEveryRange(t *testing.T) {
	set := "2"
	diags := framework.RequireAttrKionVersions(
		meta("3.16.0"),
		planWith(map[string]*string{"cloud_access_role_type_id": &set}),
		carWindow(),
		"kion_ou_cloud_access_role",
	)
	require.True(t, diags.HasError())
	d := diags[0].Detail()
	assert.Contains(t, d, "kion_ou_cloud_access_role.cloud_access_role_type_id")
	assert.Contains(t, d, "3.15.13")
	assert.Contains(t, d, "3.16.5", "both floors must appear, or the message is misleading")
	assert.Contains(t, d, "3.16.0", "must name the instance's own version")
	assert.Contains(t, d, "silently ignored")
}

func TestRequireAttrKionVersions_ignoresUnsetAttributeOutsideWindow(t *testing.T) {
	diags := framework.RequireAttrKionVersions(
		meta("3.17.0"),
		planWith(map[string]*string{"cloud_access_role_type_id": nil}),
		carWindow(),
		"kion_ou_cloud_access_role",
	)
	assert.False(t, diags.HasError(), "an attribute never sent cannot be dropped")
}

func TestAttrVersions_emptyAcceptsEverything(t *testing.T) {
	assert.True(t, framework.AttrVersions{}.Accepts(conns.MustParseKionVersion("3.17.0")))
}

func TestVersionRange_boundsAreInclusiveMinExclusiveBefore(t *testing.T) {
	r := framework.VersionRange{
		Min:    conns.MustParseKionVersion("3.16.5"),
		Before: conns.MustParseKionVersion("3.17.0"),
	}
	assert.True(t, r.Contains(conns.MustParseKionVersion("3.16.5")), "Min is inclusive")
	assert.True(t, r.Contains(conns.MustParseKionVersion("3.16.99")))
	assert.False(t, r.Contains(conns.MustParseKionVersion("3.17.0")), "Before is exclusive")
	assert.False(t, r.Contains(conns.MustParseKionVersion("3.16.4")))
}
