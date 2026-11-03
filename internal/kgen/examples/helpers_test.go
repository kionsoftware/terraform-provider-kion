package examples

import (
	"strings"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDataSourceSchemaToTF_SkipsReadOnlyBlock(t *testing.T) {
	// A block whose attributes are all computed (the classic "list" output
	// block) must be skipped entirely.
	s := dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"query": dsschema.StringAttribute{Optional: true},
		},
		Blocks: map[string]dsschema.Block{
			"list": dsschema.ListNestedBlock{
				NestedObject: dsschema.NestedBlockObject{
					Attributes: map[string]dsschema.Attribute{
						"id":   dsschema.Int64Attribute{Computed: true},
						"name": dsschema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}

	result := dataSourceSchemaToTF("kion_test", s)

	if strings.Contains(result, "# list {") {
		t.Errorf("read-only 'list' block should be skipped, got:\n%s", result)
	}
	if !strings.Contains(result, "# query = \"example\"") {
		t.Errorf("optional 'query' attribute should appear, got:\n%s", result)
	}
}

func TestDataSourceSchemaToTF_KeepsConfigurableBlock(t *testing.T) {
	// A block with at least one configurable (non-computed) attribute must be
	// rendered.
	s := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"filter": dsschema.ListNestedBlock{
				NestedObject: dsschema.NestedBlockObject{
					Attributes: map[string]dsschema.Attribute{
						"field": dsschema.StringAttribute{Required: true},
						"out":   dsschema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}

	result := dataSourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# filter {") {
		t.Errorf("configurable 'filter' block should be rendered, got:\n%s", result)
	}
	if !strings.Contains(result, "#   field = \"example\"") {
		t.Errorf("configurable block attribute 'field' should appear, got:\n%s", result)
	}
	if strings.Contains(result, "out") {
		t.Errorf("computed-only block attribute 'out' should not appear, got:\n%s", result)
	}
}

func TestDataSourceSchemaToTF_NestedConfigurableBlock(t *testing.T) {
	// A block whose direct attributes are all computed but which contains a
	// nested block with configurable input must be kept.
	s := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"outer": dsschema.SingleNestedBlock{
				Attributes: map[string]dsschema.Attribute{
					"computed_out": dsschema.StringAttribute{Computed: true},
				},
				Blocks: map[string]dsschema.Block{
					"inner": dsschema.SingleNestedBlock{
						Attributes: map[string]dsschema.Attribute{
							"needle": dsschema.StringAttribute{Optional: true},
						},
					},
				},
			},
		},
	}

	result := dataSourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# outer {") {
		t.Errorf("outer block with configurable nested input should render, got:\n%s", result)
	}
	if !strings.Contains(result, "# inner {") {
		t.Errorf("inner configurable block should render, got:\n%s", result)
	}
	if !strings.Contains(result, "needle") {
		t.Errorf("nested optional 'needle' should appear, got:\n%s", result)
	}
}

func TestDataSourceSchemaToTF_SetNestedBlock(t *testing.T) {
	s := dsschema.Schema{
		Blocks: map[string]dsschema.Block{
			"tags": dsschema.SetNestedBlock{
				NestedObject: dsschema.NestedBlockObject{
					Attributes: map[string]dsschema.Attribute{
						"key": dsschema.StringAttribute{Optional: true},
					},
				},
			},
		},
	}

	result := dataSourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# tags {") {
		t.Errorf("SetNestedBlock should render, got:\n%s", result)
	}
	if !strings.Contains(result, "#   key = \"example\"") {
		t.Errorf("SetNestedBlock attribute should render, got:\n%s", result)
	}
}

func TestResourceSchemaToTF_SetNestedBlock(t *testing.T) {
	s := rsschema.Schema{
		Blocks: map[string]rsschema.Block{
			"rules": rsschema.SetNestedBlock{
				NestedObject: rsschema.NestedBlockObject{
					Attributes: map[string]rsschema.Attribute{
						"action": rsschema.StringAttribute{Optional: true},
					},
				},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# rules {") {
		t.Errorf("resource SetNestedBlock should render, got:\n%s", result)
	}
	if !strings.Contains(result, "#   action = \"example\"") {
		t.Errorf("resource SetNestedBlock attribute should render, got:\n%s", result)
	}
}

func TestResourceSchemaToTF_NestedBlockRecursion(t *testing.T) {
	s := rsschema.Schema{
		Blocks: map[string]rsschema.Block{
			"outer": rsschema.SingleNestedBlock{
				Attributes: map[string]rsschema.Attribute{
					"flag": rsschema.BoolAttribute{Optional: true},
				},
				Blocks: map[string]rsschema.Block{
					"inner": rsschema.ListNestedBlock{
						NestedObject: rsschema.NestedBlockObject{
							Attributes: map[string]rsschema.Attribute{
								"depth": rsschema.Int64Attribute{Optional: true},
							},
						},
					},
				},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# outer {") {
		t.Errorf("outer block should render, got:\n%s", result)
	}
	if !strings.Contains(result, "# inner {") {
		t.Errorf("nested inner block should render, got:\n%s", result)
	}
	if !strings.Contains(result, "depth") {
		t.Errorf("deeply nested attribute should render, got:\n%s", result)
	}
}

func TestResourceSchemaToTF_RequiredOptionalAndBlocksSpacing(t *testing.T) {
	// Exercises the branch that inserts a blank line between required, optional,
	// and block sections all present together.
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"name": rsschema.StringAttribute{Required: true},
			"desc": rsschema.StringAttribute{Optional: true},
		},
		Blocks: map[string]rsschema.Block{
			"cfg": rsschema.SingleNestedBlock{
				Attributes: map[string]rsschema.Attribute{
					"mode": rsschema.StringAttribute{Optional: true},
				},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	for _, want := range []string{"# Required", "# Optional", "# cfg {"} {
		if !strings.Contains(result, want) {
			t.Errorf("expected %q in output, got:\n%s", want, result)
		}
	}
}

func TestResourcePlaceholder_Default(t *testing.T) {
	// An attribute type not in the switch falls through to the string default.
	// ObjectAttribute is a resource attribute type not handled explicitly.
	got := resourcePlaceholder(rsschema.ObjectAttribute{})
	if got != `"example"` {
		t.Errorf("expected default placeholder, got %q", got)
	}
}

func TestDataSourcePlaceholder_Default(t *testing.T) {
	got := dataSourcePlaceholder(dsschema.ObjectAttribute{})
	if got != `"example"` {
		t.Errorf("expected default placeholder, got %q", got)
	}
}

func TestIsDataSourceComputedOnly_NestedAttributeTypes(t *testing.T) {
	cases := map[string]dsschema.Attribute{
		"object":      dsschema.ObjectAttribute{Computed: true},
		"single":      dsschema.SingleNestedAttribute{Computed: true},
		"list_nested": dsschema.ListNestedAttribute{Computed: true},
		"set_nested":  dsschema.SetNestedAttribute{Computed: true},
		"map_nested":  dsschema.MapNestedAttribute{Computed: true},
	}
	for name, attr := range cases {
		if !isDataSourceComputedOnly(attr) {
			t.Errorf("%s: expected computed-only=true", name)
		}
	}
}

func TestIsResourceRequiredOptional_UnhandledType(t *testing.T) {
	// ObjectAttribute is not handled by the resource required/optional switches,
	// so it must report neither required nor optional (and thus be omitted).
	if isResourceRequired(rsschema.ObjectAttribute{Required: true}) {
		t.Error("unhandled resource attribute type should not report required")
	}
	if isResourceOptional(rsschema.ObjectAttribute{Optional: true}) {
		t.Error("unhandled resource attribute type should not report optional")
	}
	if isResourceComputedOnly(rsschema.ObjectAttribute{Computed: true}) {
		t.Error("unhandled resource attribute type should not report computed-only")
	}
}

func TestIsDataSourceRequiredOptional_AllScalarTypes(t *testing.T) {
	req := []dsschema.Attribute{
		dsschema.StringAttribute{Required: true},
		dsschema.Int64Attribute{Required: true},
		dsschema.BoolAttribute{Required: true},
		dsschema.Float64Attribute{Required: true},
		dsschema.MapAttribute{Required: true, ElementType: types.StringType},
		dsschema.ListAttribute{Required: true, ElementType: types.StringType},
		dsschema.SetAttribute{Required: true, ElementType: types.StringType},
	}
	for i, a := range req {
		if !isDataSourceRequired(a) {
			t.Errorf("case %d: expected required=true", i)
		}
	}

	opt := []dsschema.Attribute{
		dsschema.StringAttribute{Optional: true},
		dsschema.Int64Attribute{Optional: true},
		dsschema.BoolAttribute{Optional: true},
		dsschema.Float64Attribute{Optional: true},
		dsschema.MapAttribute{Optional: true, ElementType: types.StringType},
		dsschema.ListAttribute{Optional: true, ElementType: types.StringType},
		dsschema.SetAttribute{Optional: true, ElementType: types.StringType},
	}
	for i, a := range opt {
		if !isDataSourceOptional(a) {
			t.Errorf("case %d: expected optional=true", i)
		}
	}
}
