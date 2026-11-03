package crud

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const dsGateCall = `framework.RequireKionVersionInRange(d.Meta(), minKionVersion, maxKionVersion, "kion_label")`

// readGated reports whether the data source's Read checks the version gate
// before it reaches the API.
func readGated(t *testing.T, src []byte) bool {
	t.Helper()
	i := bytes.Index(src, []byte(") Read(ctx context.Context"))
	if i < 0 {
		t.Fatalf("no Read in\n%s", src)
	}
	body := src[i:]
	g, c := bytes.Index(body, []byte(dsGateCall)), bytes.Index(body, []byte("conn := d.Meta().Client"))
	return g >= 0 && c >= 0 && g < c
}

func TestRenderDataSource_gatedReadChecksVersion(t *testing.T) {
	for _, tc := range []struct {
		name   string
		noList bool
	}{{"list", false}, {"id-only", true}} {
		t.Run(tc.name, func(t *testing.T) {
			rm := labelResourceModel(t)
			if tc.noList {
				rm.List = nil
			}
			rm.DSGated = true
			got, _, _, err := renderDataSource(rm, FieldPolicy{})
			if err != nil {
				t.Fatal(err)
			}
			if !readGated(t, got) {
				t.Errorf("gated data source Read does not check the version first:\n%s", got)
			}

			rm.DSGated = false
			got, _, _, err = renderDataSource(rm, FieldPolicy{})
			if err != nil {
				t.Fatal(err)
			}
			if bytes.Contains(got, []byte("RequireKionVersionInRange")) {
				t.Errorf("ungated data source checks a version gate it has no vars for")
			}
		})
	}
}

func TestRenderCompoundDataSource_gatedReadChecksVersion(t *testing.T) {
	d := compoundData{Pkg: "label", Pascal: "Label", Model: "LabelModel", TypeName: "kion_label",
		SDKAlias: "generated", SDKImport: sdkImport(), DSConst: "DSNameLabel", DSName: "Label Data Source",
		IDGo: "Id", ParentIDGo: "ScopeId", ParentIDTF: "scope_id", ChildIDGo: "CriteriaId", ChildIDTF: "criteria_id",
		ReadMethod: "GetScope", ReadParams: "GetScopeParams", ReadParentArg: "ID", DSGated: true}
	got, err := renderCompoundDataSource(d)
	if err != nil {
		t.Fatal(err)
	}
	if !readGated(t, got) {
		t.Errorf("gated compound data source Read does not check the version first:\n%s", got)
	}
}

func writeVS(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "version_support.yaml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestVersionSupport_dataSourceGate(t *testing.T) {
	vs, err := loadVersionSupport(writeVS(t, `
resources:
  same: {min: "3.16.0"}
  differs: {min: "3.16.0"}
  resonly: {min: "3.15.0"}
data_sources:
  same: {min: "3.16.0"}
  differs: {min: "3.17.0"}
  dsonly: {min: "3.16.0"}
`))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		res, ds bool
		errSub  string
	}{
		{"same", true, true, ""},
		{"resonly", true, false, ""},
		{"dsonly", false, true, ""},
		{"none", false, false, ""},
		// One minKionVersion/maxKionVersion pair serves both: a differing DS
		// range would be enforced as the resource's, so it must fail.
		{"differs", true, false, "differs from the resource's"},
	} {
		if got := vs.resourceGated(tc.name); got != tc.res {
			t.Errorf("%s: resourceGated = %v, want %v", tc.name, got, tc.res)
		}
		ds, err := vs.dataSourceGated(tc.name)
		if tc.errSub != "" {
			if err == nil || !strings.Contains(err.Error(), tc.errSub) {
				t.Errorf("%s: err = %v, want containing %q", tc.name, err, tc.errSub)
			}
			continue
		}
		if err != nil || ds != tc.ds {
			t.Errorf("%s: dataSourceGated = %v, %v; want %v", tc.name, ds, err, tc.ds)
		}
	}
}
