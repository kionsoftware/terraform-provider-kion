package module

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// A placeholder that fails the attribute's own validator surfaces as a failing
// generated `terraform test`, not as anything useful at generation time. These
// cover the two shapes that broke the test:modules stage.

func TestSampleSatisfying_regexValidatedAttribute(t *testing.T) {
	// billing_start_date is YYYY-MM. The name heuristic only knew "_datecode",
	// so this previously fell through to "example" and failed the plan.
	attr := rsschema.StringAttribute{
		Required: true,
		Validators: []validator.String{
			stringvalidator.RegexMatches(regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])$`), ""),
		},
	}
	got := sampleSatisfying(attr)
	if got == "" {
		t.Fatal("expected a candidate satisfying the YYYY-MM regex, got none")
	}
	if !regexp.MustCompile(`^\d{4}-(?:0[1-9]|1[0-2])$`).MatchString(got) {
		t.Errorf("sample %q does not satisfy the attribute's own validator", got)
	}
}

func TestSampleSatisfying_noValidators(t *testing.T) {
	if got := sampleSatisfying(rsschema.StringAttribute{Required: true}); got != "" {
		t.Errorf("an attribute with no validators should yield no forced sample, got %q", got)
	}
	if got := sampleSatisfying(rsschema.Int64Attribute{Required: true}); got != "" {
		t.Errorf("a non-string attribute should yield no forced sample, got %q", got)
	}
}

func TestObjectSample_fillsRequiredFieldsOnly(t *testing.T) {
	// Shaped like gcp_billing_account_create: a required nested object holding a
	// required nested object, plus a mix of required and optional scalars.
	typ := "object({ big_query_export = object({ dataset_name = optional(string) }), " +
		"billing_start_date = string, gcp_id = string, is_reseller = optional(bool), " +
		"service_account_id = number })"
	got := sample(typ)

	if got == "null" {
		t.Fatal("a required object attribute must not sample to null; Terraform rejects it")
	}
	for _, want := range []string{
		"big_query_export = {}",          // nested object, all fields optional
		`billing_start_date = "2026-01"`, // name heuristic applies inside objects too
		`gcp_id = "example"`,
		"service_account_id = 1",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("object sample missing %q, got:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"is_reseller", "dataset_name"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("optional field %q should be omitted, got:\n%s", unwanted, got)
		}
	}
}

func TestSplitTopLevel_doesNotSplitInsideNestedTypes(t *testing.T) {
	parts := splitTopLevel("a = object({ x = string, y = number }), b = string")
	if len(parts) != 2 {
		t.Fatalf("expected 2 top-level fields, got %d: %#v", len(parts), parts)
	}
	if !strings.HasPrefix(parts[0], "a = object(") {
		t.Errorf("nested object was split apart: %#v", parts)
	}
	if parts[1] != "b = string" {
		t.Errorf("second field mis-parsed: %q", parts[1])
	}
}

func TestSampleFor_startDateNames(t *testing.T) {
	// Reachable inside object literals, where the attribute (and so its
	// validators) is not available.
	for _, n := range []string{"billing_start_date", "start_datecode"} {
		if got := sampleFor(n, "string"); got != `"2026-01"` {
			t.Errorf("sampleFor(%q) = %s, want \"2026-01\"", n, got)
		}
	}
}
