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
{{- end }}
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-kion/internal/framework"
)
{{ if .IncludeComments }}
// TIP: ==== FILE STRUCTURE ====
// All data sources should follow this basic outline.
//
// 1. Package declaration
// 2. Imports
// 3. Main data source struct with schema method
// 4. Read method
// 5. Other functions (flatteners, etc.)
{{- end }}

// Data source name constant.
const (
	DSName{{ .DataSource }} = "{{ .HumanDataSourceName }} Data Source"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &{{ .DataSourceLowerCamel }}DataSource{}
	_ datasource.DataSourceWithConfigure = &{{ .DataSourceLowerCamel }}DataSource{}
)

// New{{ .DataSource }}DataSource returns a new instance of the data source.
func New{{ .DataSource }}DataSource() datasource.DataSource {
	return &{{ .DataSourceLowerCamel }}DataSource{}
}

type {{ .DataSourceLowerCamel }}DataSource struct {
	framework.DataSourceWithConfigure
}

func (d *{{ .DataSourceLowerCamel }}DataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_{{ .DataSourceSnake }}"
}

{{ if .IncludeComments -}}
// TIP: ==== SCHEMA ====
// Data sources typically have filter criteria as arguments
// and the results as computed attributes.
{{- end }}
func (d *{{ .DataSourceLowerCamel }}DataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to access information about Kion {{ .HumanDataSourceName }}s.",
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Description: "The ID of the {{ .HumanDataSourceName }}.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the {{ .HumanDataSourceName }}.",
				Computed:    true,
			},
		},
	}
}

func (d *{{ .DataSourceLowerCamel }}DataSource) Read(_ context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	{{- if .IncludeComments }}
	// TIP: ==== DATA SOURCE READ ====
	// 1. conn := d.Meta().Client
	// 2. Get config: req.Config.Get(ctx, &data)
	// 3. Call SDK: out, err := conn.GetXxx(ctx, generated.GetXxxParams{ID: data.ID.ValueInt64()})
	// 4. If errs.IsNotFound(out): add error diagnostic
	// 5. Flatten response into model
	// 6. Set state: resp.State.Set(ctx, &data)
	{{- end }}
	resp.Diagnostics.AddError("not implemented", "Read for {{ .HumanDataSourceName }} data source is not yet implemented")
}

{{ if .IncludeComments -}}
// TIP: ==== FLATTEN ====
// flatten{{ .DataSource }}DataSource converts an SDK response into the Terraform model.
// See the resource template for examples of type-switching.
{{- end }}
func flatten{{ .DataSource }}DataSource(apiObject any, _ *{{ .DataSourceLowerCamel }}DataSourceModel) diag.Diagnostics {
	// TODO: implement, type-switch the response and map SDK fields to model fields
	_ = apiObject
	return nil
}

{{ if .IncludeComments -}}
// TIP: ==== DATA STRUCTURES ====
// These structs should match the schema definition exactly.
{{- end }}
type {{ .DataSourceLowerCamel }}DataSourceModel struct {
	ID   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	// TODO: Add fields matching your schema attributes here.
}
