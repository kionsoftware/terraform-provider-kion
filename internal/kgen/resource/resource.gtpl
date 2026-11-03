// Package {{ .ServicePackage }} implements the Terraform resource and data source for {{ .ResourceSnake }}.
package {{ .ServicePackage }}

{{ if .IncludeComments -}}
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the kgen tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the Kion API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.
{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
{{- end }}
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-kion/internal/framework"
)
{{ if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All resources should follow this basic outline. Improve this resource's
// maintainability by sticking to it.
//
// 1. Package declaration
// 2. Imports
// 3. Main resource struct with schema method
// 4. Create, read, update, delete methods (in that order)
// 5. Other functions (flatteners, expanders, finders, etc.)

{{ end -}}

// Resource name constant.
const (
	ResName{{ .Resource }} = "{{ .HumanResourceName }}"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &{{ .ResourceLowerCamel }}Resource{}
	_ resource.ResourceWithConfigure   = &{{ .ResourceLowerCamel }}Resource{}
	_ resource.ResourceWithImportState = &{{ .ResourceLowerCamel }}Resource{}
)

// New{{ .Resource }}Resource returns a new instance of the resource.
func New{{ .Resource }}Resource() resource.Resource {
	return &{{ .ResourceLowerCamel }}Resource{}
}

type {{ .ResourceLowerCamel }}Resource struct {
	framework.ResourceWithConfigure
}

func (r *{{ .ResourceLowerCamel }}Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_{{ .ResourceSnake }}"
}

{{ if .IncludeComments -}}
// TIP: ==== SCHEMA ====
// In the schema, add each of the attributes in snake case (e.g.,
// delete_automated_backups).
//
// Formatting rules:
// * Alphabetize attributes to make them easier to find.
// * Do not add a blank line between attributes.
//
// Attribute basics:
// Required: true     the user must provide a value
// Optional: true     the user can configure or omit a value
// Computed: true     read-only, the provider sets the value
// Optional + Computed. Provider or user can provide a value
//
// For more about schema options, visit
// https://developer.hashicorp.com/terraform/plugin/framework/handling-data/schemas?page=schemas
{{- end }}
func (r *{{ .ResourceLowerCamel }}Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Kion {{ .HumanResourceName }}.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the {{ .HumanResourceName }}.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			{{- if .IncludeComments }}
			// TIP: Add your resource-specific attributes here.
			// Look up the SDK types for this resource in the generated package
			// to determine the correct fields.
			{{- end }}
			"name": schema.StringAttribute{
				Description: "The name of the {{ .HumanResourceName }}.",
				Required:    true,
			},
		},
	}
}

func (r *{{ .ResourceLowerCamel }}Resource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE CREATE ====
	// 1. conn := r.Meta().Client
	// 2. Get plan: req.Plan.Get(ctx, &plan)
	// 3. Build SDK input (e.g., generated.OptCreateXxx{Value: ..., Set: true})
	// 4. Call SDK: out, err := conn.PostXxx(ctx, input)
	// 5. Extract ID: id, diags := errs.CreatedID(out)
	// 6. Read back: readOut, err := conn.GetXxx(ctx, generated.GetXxxParams{ID: id})
	// 7. Flatten response into model
	// 8. Set state: resp.State.Set(ctx, &plan)
	{{- end }}
	resp.Diagnostics.AddError("not implemented", "Create for {{ .HumanResourceName }} is not yet implemented")
}

func (r *{{ .ResourceLowerCamel }}Resource) Read(_ context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE READ ====
	// 1. conn := r.Meta().Client
	// 2. Get state: req.State.Get(ctx, &state)
	// 3. Call SDK: out, err := conn.GetXxx(ctx, generated.GetXxxParams{ID: state.ID.ValueInt64()})
	// 4. If errs.IsNotFound(out): resp.State.RemoveResource(ctx); return
	// 5. Flatten response into model
	// 6. Set state: resp.State.Set(ctx, &state)
	{{- end }}
	resp.Diagnostics.AddError("not implemented", "Read for {{ .HumanResourceName }} is not yet implemented")
}

func (r *{{ .ResourceLowerCamel }}Resource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE UPDATE ====
	// 1. conn := r.Meta().Client
	// 2. Get plan: req.Plan.Get(ctx, &plan)
	// 3. Build SDK update input (Opt* wrappers for each field)
	// 4. Call SDK: out, err := conn.PatchXxx(ctx, input, generated.PatchXxxParams{ID: plan.ID.ValueInt64()})
	// 5. Check: errs.ResponseDiagnostics("updating ...", out)
	// 6. Read back and flatten
	// 7. Set state: resp.State.Set(ctx, &plan)
	{{- end }}
	resp.Diagnostics.AddError("not implemented", "Update for {{ .HumanResourceName }} is not yet implemented")
}

func (r *{{ .ResourceLowerCamel }}Resource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	{{- if .IncludeComments }}
	// TIP: ==== RESOURCE DELETE ====
	// 1. conn := r.Meta().Client
	// 2. Get state: req.State.Get(ctx, &state)
	// 3. Call SDK: out, err := conn.DeleteXxx(ctx, generated.DeleteXxxParams{ID: state.ID.ValueInt64()})
	// 4. If errs.IsNotFound(out): return (already deleted)
	// 5. Check: errs.ResponseDiagnostics("deleting ...", out)
	{{- end }}
	resp.Diagnostics.AddError("not implemented", "Delete for {{ .HumanResourceName }} is not yet implemented")
}

func (r *{{ .ResourceLowerCamel }}Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id, diags := framework.StringToInt64(req.ID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

{{ if .IncludeComments -}}
// TIP: ==== FLATTEN ====
// flatten{{ .Resource }} converts an SDK response into the Terraform model.
// Type-switch on the response interface to extract the concrete data.
//
// Example:
//   switch v := apiObject.(type) {
//   case *generated.XxxResponse:
//       if v.Data.Set {
//           model.ID = flex.OptUint64ToFramework(v.Data.Value.ID)
//           model.Name = flex.StringToFramework(v.Data.Value.Name)
//       }
//   default:
//       return errs.ResponseDiagnostics("reading Xxx", apiObject)
//   }
{{- end }}
func flatten{{ .Resource }}(apiObject any, _ *{{ .ResourceLowerCamel }}ResourceModel) diag.Diagnostics {
	// TODO: implement, type-switch the response and map SDK fields to model fields
	_ = apiObject
	return nil
}

{{ if .IncludeComments -}}
// TIP: ==== DATA STRUCTURES ====
// These structs should match the schema definition exactly, and the `tfsdk`
// tag value should match the attribute name.
{{- end }}
type {{ .ResourceLowerCamel }}ResourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// TODO: Add fields matching your schema attributes here.
}
