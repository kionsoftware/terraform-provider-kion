package framework

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// AtLeastOneOfOnCreate requires at least one of attrs when the resource is
// being created. resourcevalidator.AtLeastOneOf cannot see prior state, so it
// also rejects an existing record Kion holds without any of them (built-in
// compliance checks, Kion-managed policies), which makes those unimportable.
//
// It is attached as a plan modifier on the first attribute; only set and list
// attributes are supported, and anything else panics at schema build.
func AtLeastOneOfOnCreate(s *schema.Schema, attrs ...string) {
	m := atLeastOneOfOnCreate{attrs: attrs}
	switch a := s.Attributes[attrs[0]].(type) {
	case schema.SetAttribute:
		a.PlanModifiers = append(a.PlanModifiers, m)
		s.Attributes[attrs[0]] = a
	case schema.ListAttribute:
		a.PlanModifiers = append(a.PlanModifiers, m)
		s.Attributes[attrs[0]] = a
	default:
		panic(fmt.Sprintf("AtLeastOneOfOnCreate: %s is %T, not a set or list attribute", attrs[0], a))
	}
}

type atLeastOneOfOnCreate struct{ attrs []string }

func (m atLeastOneOfOnCreate) Description(ctx context.Context) string {
	return m.MarkdownDescription(ctx)
}

func (m atLeastOneOfOnCreate) MarkdownDescription(_ context.Context) string {
	return fmt.Sprintf("at least one of %s is required on create", strings.Join(m.attrs, ", "))
}

func (m atLeastOneOfOnCreate) PlanModifySet(_ context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	resp.Diagnostics.Append(m.check(req.Config, req.State, req.Path)...)
}

func (m atLeastOneOfOnCreate) PlanModifyList(_ context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	resp.Diagnostics.Append(m.check(req.Config, req.State, req.Path)...)
}

// check reads the raw config so one implementation covers every attribute
// type. An unknown value counts as set: it is decided at apply.
func (m atLeastOneOfOnCreate) check(config tfsdk.Config, state tfsdk.State, at path.Path) diag.Diagnostics {
	if !state.Raw.IsNull() || config.Raw.IsNull() || !config.Raw.IsKnown() {
		return nil
	}
	var obj map[string]tftypes.Value
	if err := config.Raw.As(&obj); err != nil {
		return nil
	}
	for _, name := range m.attrs {
		if v, ok := obj[name]; ok && !v.IsNull() {
			return nil
		}
	}
	// Same summary and detail as resourcevalidator.AtLeastOneOf, which this
	// replaces, so existing expectations on the message still hold.
	var d diag.Diagnostics
	d.AddAttributeError(at, "Missing Attribute Configuration",
		fmt.Sprintf("At least one of these attributes must be configured: [%s]", strings.Join(m.attrs, ",")))
	return d
}
