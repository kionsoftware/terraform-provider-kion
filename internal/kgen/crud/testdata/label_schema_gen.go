package testdata

// Miniature stand-in for a generated <name>_schema_gen.go model. Parsed as
// bytes for the tfsdk tags + field types; the `types` selector need not
// resolve (never compiled — under testdata/).

type LabelModel struct {
	Color types.String `tfsdk:"color"`
	Id    types.String `tfsdk:"id"`
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}
