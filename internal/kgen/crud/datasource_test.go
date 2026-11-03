package crud

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"
)

func TestRenderDataSource_labelDualMode(t *testing.T) {
	got, downgrade, _, err := renderDataSource(labelResourceModel(t), FieldPolicy{})
	if err != nil {
		t.Fatalf("renderDataSource: %v", err)
	}
	if downgrade != "" {
		t.Fatalf("label must not be downgraded: %s", downgrade)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_data_source.go", got, parser.ParseComments); err != nil {
		t.Fatalf("does not parse: %v\n%s", err, got)
	}
	for _, w := range []string{
		"func (d *labelDataSource) readByID(",
		"func (d *labelDataSource) readByFilter(",
		"func fetchAllLabel(ctx context.Context, conn *generated.Client)",
		"filter.Schema()",
		"filter.Match(ctx, data.Filter, labelToRow(lbl))",
		"conn.GetLabelIndex(ctx, generated.GetLabelIndexParams{",
		"generated.NewOptInt64(page)",
		"resp.Data.Value.Items",
		"resp.Data.Value.Total",
		"data.Id = flex.OptUint64ToFramework(lbl.ID)",
		"Filter []filter.Model `tfsdk:\"filter\"`",
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("dual-mode DS missing %q", w)
		}
	}
}

// A shape-B, non-paginated collection must still produce the dual-mode data
// source: one call, `resp.Data` read as the items slice, no page loop, and no
// unused indexPageSize constant.
func TestRenderDataSource_shapeBNotPaginated(t *testing.T) {
	rm := labelResourceModel(t)
	lm, err := resolveList(&opRef{Method: "GET", Path: "/v3/label-flat"}, fixtureIndex(t), rm.Read)
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}
	rm.List = lm

	got, downgrade, _, err := renderDataSource(rm, FieldPolicy{})
	if err != nil {
		t.Fatalf("renderDataSource: %v", err)
	}
	if downgrade != "" {
		t.Fatalf("shape-B list must not downgrade: %s", downgrade)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_data_source.go", got, parser.ParseComments); err != nil {
		t.Fatalf("does not parse: %v\n%s", err, got)
	}
	for _, w := range []string{
		"filter.Schema()",
		"conn.GetLabelFlat(ctx)",
		"items := resp.Data",
		"all = append(all, items...)",
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("shape-B DS missing %q\n%s", w, got)
		}
	}
	for _, unwanted := range []string{"indexPageSize", "page++", "resp.Data.Value", "items.Set"} {
		if bytes.Contains(got, []byte(unwanted)) {
			t.Errorf("non-paginated shape-B DS must not emit %q\n%s", unwanted, got)
		}
	}
}

// The sweeper shares the envelope walk, so it must degrade the same way.
func TestRenderSweep_shapeBNotPaginated(t *testing.T) {
	rm := labelResourceModel(t)
	lm, err := resolveList(&opRef{Method: "GET", Path: "/v3/label-flat"}, fixtureIndex(t), rm.Read)
	if err != nil {
		t.Fatalf("resolveList: %v", err)
	}
	rm.List = lm

	got, note, err := renderSweep(rm)
	if err != nil {
		t.Fatalf("renderSweep: %v", err)
	}
	if note != "" {
		t.Fatalf("shape-B sweeper must not fall back: %s", note)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "sweep.go", got, parser.ParseComments); err != nil {
		t.Fatalf("does not parse: %v\n%s", err, got)
	}
	for _, w := range []string{"conn.Client.GetLabelFlat(ctx)", "for _, item := range items {"} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("shape-B sweeper missing %q\n%s", w, got)
		}
	}
	if bytes.Contains(got, []byte("sweepPageSize")) {
		t.Errorf("non-paginated sweeper must not declare sweepPageSize\n%s", got)
	}
}

func TestRenderDataSource_idOnlyFallback(t *testing.T) {
	rm := labelResourceModel(t)
	rm.List = nil // simulate a resource with no list endpoint
	got, _, _, err := renderDataSource(rm, FieldPolicy{})
	if err != nil {
		t.Fatalf("renderDataSource: %v", err)
	}
	if bytes.Contains(got, []byte("filter.Schema()")) {
		t.Error("id-only fallback must not reference filter")
	}
	if !bytes.Contains(got, []byte("func (d *labelDataSource) Read(")) {
		t.Error("id-only fallback missing Read")
	}
}

func TestRenderIDOnlyDataSource_label(t *testing.T) {
	got, err := renderIDOnlyDataSource(labelResourceModel(t))
	if err != nil {
		t.Fatalf("renderIDOnlyDataSource: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_data_source.go", got, parser.ParseComments); err != nil {
		t.Fatalf("generated data source does not parse: %v\n%s", err, got)
	}

	wants := []string{
		"func (d *labelDataSource) Read(",
		"conn.GetLabel(ctx, generated.GetLabelParams{ID: data.Id.ValueInt64()})",
		"data.Id = flex.OptUint64ToFramework(v.ID)",
		"data.Key = flex.StringToFramework(v.Key)",
		"type labelDataSourceModel struct",
		"`tfsdk:\"id\"`",
		"schema.Int64Attribute{",
	}
	for _, w := range wants {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated data source missing %q", w)
		}
	}

	// Documented reduction: the id-only DS must NOT reference the legacy filter
	// block or the list/index pagination path.
	for _, unwanted := range []string{"filter.", "GetLabelIndex", "ListNestedAttribute"} {
		if bytes.Contains(got, []byte(unwanted)) {
			t.Errorf("id-only data source unexpectedly references %q", unwanted)
		}
	}
}

// kion_ami.expires_at and kion_gcp_iam_role.gcp_role_launch_stage were dropped
// from their data sources because these SDK types had no list projection.
func TestListValueExprs_nullTimeAndEnum(t *testing.T) {
	cases := []struct {
		sdk, model, attrType, obj, row string
	}{
		{
			"OptNullTime", "types.String", "types.StringType",
			"types.StringValue(flex.OptNullTimeToFramework(x.F).ValueString())",
			"flex.OptNullTimeToFramework(x.F).ValueString()",
		},
		{
			"OptGCPRoleLaunchStage", "types.Int64", "types.Int64Type",
			"types.Int64Value(int64(x.F.Or(0)))",
			"int64(x.F.Or(0))",
		},
	}
	for _, c := range cases {
		attrType, obj, row, ok := listValueExprs("x.F", c.sdk, c.model)
		if !ok {
			t.Fatalf("%s: no list projection", c.sdk)
		}
		if attrType != c.attrType || obj != c.obj || row != c.row {
			t.Errorf("%s: got (%s, %s, %s)", c.sdk, attrType, obj, row)
		}
	}
}
