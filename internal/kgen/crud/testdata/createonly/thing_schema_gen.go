package testdata

// ThingModel is testdata/stringid's model plus secret, which only the create
// response carries.
type ThingModel struct {
	Description types.String `tfsdk:"description"`
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Secret      types.String `tfsdk:"secret"`
}
