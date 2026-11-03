package examples

import (
	"strings"
	"testing"

	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestResourceSchemaToTF_Basic(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"id": rsschema.Int64Attribute{
				Computed: true,
			},
			"name": rsschema.StringAttribute{
				Required: true,
			},
			"description": rsschema.StringAttribute{
				Optional: true,
			},
			"count": rsschema.Int64Attribute{
				Required: true,
			},
			"enabled": rsschema.BoolAttribute{
				Optional: true,
			},
			"computed_field": rsschema.StringAttribute{
				Computed: true,
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	// Check resource declaration.
	if !strings.Contains(result, `resource "kion_test" "example" {`) {
		t.Errorf("missing resource declaration, got:\n%s", result)
	}

	// Required section should have name and count.
	if !strings.Contains(result, "# Required") {
		t.Error("missing Required section")
	}
	if !strings.Contains(result, `name  = "example"`) {
		t.Errorf("missing required 'name' attribute, got:\n%s", result)
	}
	if !strings.Contains(result, "count = 1") {
		t.Errorf("missing required 'count' attribute, got:\n%s", result)
	}

	// Optional section should have description and enabled.
	if !strings.Contains(result, "# Optional") {
		t.Error("missing Optional section")
	}
	if !strings.Contains(result, `# description = "example"`) {
		t.Errorf("missing optional 'description' attribute, got:\n%s", result)
	}
	if !strings.Contains(result, "# enabled     = false") {
		t.Errorf("missing optional 'enabled' attribute, got:\n%s", result)
	}

	// Computed-only fields should not appear.
	if strings.Contains(result, "computed_field") {
		t.Errorf("computed-only field 'computed_field' should not appear, got:\n%s", result)
	}
	if strings.Contains(result, `"id"`) {
		t.Errorf("computed-only field 'id' should not appear in config, got:\n%s", result)
	}
}

func TestResourceSchemaToTF_WithBlocks(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"name": rsschema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]rsschema.Block{
			"settings": rsschema.SingleNestedBlock{
				Attributes: map[string]rsschema.Attribute{
					"mode": rsschema.StringAttribute{
						Optional: true,
					},
					"value": rsschema.Int64Attribute{
						Optional: true,
					},
				},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# settings {") {
		t.Errorf("missing block start, got:\n%s", result)
	}
	if !strings.Contains(result, `#   mode  = "example"`) {
		t.Errorf("missing block attribute 'mode', got:\n%s", result)
	}
	if !strings.Contains(result, "#   value = 1") {
		t.Errorf("missing block attribute 'value', got:\n%s", result)
	}
	if !strings.Contains(result, "# }") {
		t.Errorf("missing block end, got:\n%s", result)
	}
}

func TestResourceSchemaToTF_AllTypes(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"str_field": rsschema.StringAttribute{
				Required: true,
			},
			"int_field": rsschema.Int64Attribute{
				Required: true,
			},
			"bool_field": rsschema.BoolAttribute{
				Required: true,
			},
			"float_field": rsschema.Float64Attribute{
				Required: true,
			},
			"map_field": rsschema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"list_field": rsschema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
			},
			"set_field": rsschema.SetAttribute{
				Required:    true,
				ElementType: types.Int64Type,
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	expectations := map[string]string{
		"str_field":   `"example"`,
		"int_field":   "1",
		"bool_field":  "false",
		"float_field": "0.0",
		"map_field":   "{}",
		"list_field":  "[]",
		"set_field":   "[]",
	}

	for field, expected := range expectations {
		if !strings.Contains(result, field+" ") || !strings.Contains(result, expected) {
			t.Errorf("expected field %q with placeholder %q in output:\n%s", field, expected, result)
		}
	}
}

func TestResourceSchemaToTF_OptionalComputed(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"name": rsschema.StringAttribute{
				Required: true,
			},
			"email": rsschema.StringAttribute{
				Optional: true,
				Computed: true,
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	// Optional+Computed should appear in optional section.
	if !strings.Contains(result, `# email = "example"`) {
		t.Errorf("optional+computed field 'email' should appear as optional, got:\n%s", result)
	}
}

func TestDataSourceSchemaToTF(t *testing.T) {
	s := dsschema.Schema{
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"name": dsschema.StringAttribute{
				Computed: true,
			},
		},
	}

	result := dataSourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, `data "kion_test" "example" {`) {
		t.Errorf("missing data source declaration, got:\n%s", result)
	}

	// id is Optional+Computed, should appear.
	if !strings.Contains(result, "# id = 1") {
		t.Errorf("missing optional 'id' attribute, got:\n%s", result)
	}

	// name is Computed-only, should not appear.
	if strings.Contains(result, "name") {
		t.Errorf("computed-only 'name' should not appear, got:\n%s", result)
	}
}

func TestResourceSchemaToTF_ListNestedBlock(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"name": rsschema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]rsschema.Block{
			"items": rsschema.ListNestedBlock{
				NestedObject: rsschema.NestedBlockObject{
					Attributes: map[string]rsschema.Attribute{
						"key": rsschema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	if !strings.Contains(result, "# items {") {
		t.Errorf("missing ListNestedBlock, got:\n%s", result)
	}
	if !strings.Contains(result, `#   key = "example"`) {
		t.Errorf("missing block attribute, got:\n%s", result)
	}
}

func TestSortedKeys(t *testing.T) {
	m := map[string]int{"c": 3, "a": 1, "b": 2}
	keys := sortedKeys(m)

	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Errorf("expected [a b c], got %v", keys)
	}
}

func TestResourceSchemaToTF_ComputedOnlyBlocks(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"name": rsschema.StringAttribute{
				Required: true,
			},
		},
		Blocks: map[string]rsschema.Block{
			"info": rsschema.SingleNestedBlock{
				Attributes: map[string]rsschema.Attribute{
					"computed_only": rsschema.StringAttribute{
						Computed: true,
					},
					"editable": rsschema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	if strings.Contains(result, "computed_only") {
		t.Errorf("computed-only block attribute should not appear, got:\n%s", result)
	}
	if !strings.Contains(result, "editable") {
		t.Errorf("optional block attribute should appear, got:\n%s", result)
	}
}

// Nested attributes used to be dropped entirely: they fell through the
// required/optional predicates, so a resource whose shape is mostly a nested
// block got an example showing only its scalars. Shaped like the real
// billing-source connection blocks.
func TestResourceSchemaToTF_nestedAttributes(t *testing.T) {
	s := rsschema.Schema{
		Attributes: map[string]rsschema.Attribute{
			"name": rsschema.StringAttribute{Optional: true},
			"aws_connection": rsschema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]rsschema.Attribute{
					"account_number": rsschema.StringAttribute{Optional: true},
					"storage_bucket": rsschema.StringAttribute{Optional: true},
				},
			},
			"required_block": rsschema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]rsschema.Attribute{
					"tenant_id": rsschema.Int64Attribute{Required: true},
				},
			},
			"computed_block": rsschema.SingleNestedAttribute{
				Computed:   true,
				Attributes: map[string]rsschema.Attribute{"x": rsschema.StringAttribute{Computed: true}},
			},
		},
	}

	result := resourceSchemaToTF("kion_test", s)

	// An optional nested attribute renders as a commented multi-line object,
	// with its sub-attributes commented once, not twice.
	for _, want := range []string{
		"# aws_connection = {",
		`#   account_number = "example"`,
		`#   storage_bucket = "example"`,
		"# }",
	} {
		if !strings.Contains(result, want) {
			t.Errorf("optional nested attribute missing %q, got:\n%s", want, result)
		}
	}
	if strings.Contains(result, "#   # ") {
		t.Errorf("sub-attributes should not be double-commented, got:\n%s", result)
	}

	// A required nested attribute renders uncommented.
	for _, want := range []string{"required_block = {", "tenant_id = 1"} {
		if !strings.Contains(result, want) {
			t.Errorf("required nested attribute missing %q, got:\n%s", want, result)
		}
	}

	// A computed-only nested attribute is omitted, like a computed-only scalar.
	if strings.Contains(result, "computed_block") {
		t.Errorf("computed-only nested attribute should be omitted, got:\n%s", result)
	}
}
