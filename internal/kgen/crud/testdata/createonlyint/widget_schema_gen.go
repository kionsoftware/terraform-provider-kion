package testdata

// WidgetModel: count is a string attribute over an OptInt64 create field, and
// size an Int64 attribute; both exist to be refused as create_only.
type WidgetModel struct {
	Count  types.String `tfsdk:"count"`
	Id     types.String `tfsdk:"id"`
	Name   types.String `tfsdk:"name"`
	Note   types.String `tfsdk:"note"`
	Secret types.String `tfsdk:"secret"`
	Size   types.Int64  `tfsdk:"size"`
}
