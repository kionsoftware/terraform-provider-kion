package testdata

type WidgetModel struct {
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Scopes types.List   `tfsdk:"scopes"`
}
