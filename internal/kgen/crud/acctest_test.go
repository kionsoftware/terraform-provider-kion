package crud

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func labelTestValues() testValues {
	return testValues{
		Create: map[string]string{"color": "#0088ff", "key": "test-acc-%[1]s", "value": "test-acc-%[1]s"},
		Update: map[string]string{"color": "#ff0000", "key": "test-acc-%[1]s-upd", "value": "test-acc-%[1]s-upd"},
	}
}

func TestRenderResourceTest_label(t *testing.T) {
	got, err := renderResourceTest(labelResourceModel(t), labelTestValues())
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_test.go", got, parser.ParseComments); err != nil {
		t.Fatalf("generated resource test does not parse: %v\n%s", err, got)
	}
	wants := []string{
		"package label_test",
		"func TestAccKionLabel_basic(",
		"func TestAccKionLabel_update(",
		"testAccCheckLabelDestroy(ctx)",
		"testAccCheckLabelExists(ctx, resourceName)",
		"conn.Client.GetLabel(ctx, generated.GetLabelParams{ID: id})",
		"ImportStateVerify: true",
		`resource "kion_label" "test"`,
		`key = "test-acc-%[1]s"`,
		"fmt.Sprintf(",
	}
	for _, w := range wants {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated resource test missing %q", w)
		}
	}
}

func TestRenderDataSourceTest_label(t *testing.T) {
	got, err := renderDataSourceTest(labelResourceModel(t), labelTestValues(), []string{"color", "key", "value"})
	if err != nil {
		t.Fatalf("renderDataSourceTest: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_data_source_test.go", got, parser.ParseComments); err != nil {
		t.Fatalf("generated data source test does not parse: %v\n%s", err, got)
	}
	wants := []string{
		"func TestAccKionLabelDataSource_basic(",
		`data.kion_label.test`,
		`data "kion_label" "test"`,
		"id = kion_label.test.id",
	}
	for _, w := range wants {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated data source test missing %q", w)
		}
	}
}

func TestRenderResourceTest_requireEnvAndKnownIssues(t *testing.T) {
	tv := labelTestValues()
	tv.RequireEnv = []envRequirement{{Name: "KION_ACC_AZURE_PAYER_ID", Reason: `an Azure billing source; the "create" answers 500 without one`}}
	tv.KnownIssues = []knownIssue{
		{Issue: 42, Note: "the read drops a field.\nLeft failing on purpose.", ExpectNonEmptyPlan: true},
		{Issue: 43, Note: "a second note"},
	}
	got, err := renderResourceTest(labelResourceModel(t), tv)
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_test.go", got, parser.ParseComments); err != nil {
		t.Fatalf("generated resource test does not parse: %v\n%s", err, got)
	}
	precheck := "PreCheck: func() {\n\t\t\tacctest.PreCheck(t)\n\t\t\t" +
		`acctest.RequireEnv(t, "KION_ACC_AZURE_PAYER_ID", "an Azure billing source; the \"create\" answers 500 without one")` +
		"\n\t\t},"
	if n := bytes.Count(got, []byte(precheck)); n != 2 {
		t.Errorf("want the RequireEnv PreCheck in both tests, found %d:\n%s", n, got)
	}
	if bytes.Contains(got, []byte(":= acctest.RequireEnv")) {
		t.Errorf("require_env value is unused by the config and must not be bound to a variable:\n%s", got)
	}
	for _, w := range []string{
		"// Known issue #42: the read drops a field.\n// Left failing on purpose.",
		"// Known issue #43: a second note",
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated resource test missing %q:\n%s", w, got)
		}
	}
	// basic: one apply step; update: two. Import steps take no plan expectation.
	if n := bytes.Count(got, []byte("ExpectNonEmptyPlan: true")); n != 3 {
		t.Errorf("want ExpectNonEmptyPlan on the 3 apply steps, found %d:\n%s", n, got)
	}
}

func TestRenderResourceTest_knownIssueWithoutPlanFlag(t *testing.T) {
	tv := labelTestValues()
	tv.KnownIssues = []knownIssue{{Issue: 7, Note: "cosmetic"}}
	got, err := renderResourceTest(labelResourceModel(t), tv)
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	if bytes.Contains(got, []byte("ExpectNonEmptyPlan")) {
		t.Errorf("no known issue asked for ExpectNonEmptyPlan:\n%s", got)
	}
	if !bytes.Contains(got, []byte("PreCheck:                 func() { acctest.PreCheck(t) },")) {
		t.Errorf("PreCheck changed with no require_env:\n%s", got)
	}
}

func TestRenderDataSourceTest_requireEnv(t *testing.T) {
	tv := labelTestValues()
	tv.RequireEnv = []envRequirement{{Name: "KION_ACC_X", Reason: "an x"}}
	tv.KnownIssues = []knownIssue{{Issue: 9, Note: "n", ExpectNonEmptyPlan: true}}
	got, err := renderDataSourceTest(labelResourceModel(t), tv, []string{"color", "key", "value"})
	if err != nil {
		t.Fatalf("renderDataSourceTest: %v", err)
	}
	for _, w := range []string{
		"acctest.PreCheck(t)\n\t\t\tacctest.RequireEnv(t, \"KION_ACC_X\", \"an x\")",
		"// Known issue #9: n",
		"ExpectNonEmptyPlan: true",
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated data source test missing %q:\n%s", w, got)
		}
	}
}

// Every archetype's test template must render require_env and known_issues;
// one that ignored them would drop the gate in silence.
func TestTestTemplatesRenderRequireEnvAndKnownIssues(t *testing.T) {
	tmpls := map[string]string{
		"restest": resourceTestTmpl, "dstest": dataSourceTestTmpl, "assoctest": assocTestTmpl,
		"blendedtest": blendedTestTmpl, "noreadtest": noReadTestTmpl, "parentlisttest": parentListTestTmpl,
	}
	for name, src := range tmpls {
		for _, w := range []string{"{{range .RequireEnv}}", "{{range .KnownIssues}}", "{{if $.ExpectNonEmptyPlan}}"} {
			if !strings.Contains(src, w) {
				t.Errorf("%s template does not render %q", name, w)
			}
		}
		if got, want := strings.Count(src, "{{if $.ExpectNonEmptyPlan}}"), strings.Count(src, "\t\t\t\tConfig: "); got != want {
			t.Errorf("%s: %d apply steps but %d ExpectNonEmptyPlan hooks", name, want, got)
		}
	}
}

func TestLoadTestValues_requireEnvAndKnownIssues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tv.yaml")
	raw := `resources:
  thing:
    require_env:
      - name: KION_ACC_X
        reason: an x
    known_issues:
      - issue: 12
        note: broken
        expect_non_empty_plan: true
    create:
      name: "%[1]s"
`
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	tv, ok, err := loadTestValues(path, "thing")
	if err != nil || !ok {
		t.Fatalf("loadTestValues: ok=%v err=%v", ok, err)
	}
	if len(tv.RequireEnv) != 1 || tv.RequireEnv[0] != (envRequirement{Name: "KION_ACC_X", Reason: "an x"}) {
		t.Errorf("require_env = %+v", tv.RequireEnv)
	}
	if len(tv.KnownIssues) != 1 || tv.KnownIssues[0] != (knownIssue{Issue: 12, Note: "broken", ExpectNonEmptyPlan: true}) {
		t.Errorf("known_issues = %+v", tv.KnownIssues)
	}
}

func TestLoadTestValues_rejectsIncompleteEntries(t *testing.T) {
	cases := map[string]string{
		"require_env without reason":       "require_env: [{name: KION_ACC_X}]",
		"require_env without name":         "require_env: [{reason: an x}]",
		"require_env duplicating env_args": "env_args: [KION_ACC_X]\n    require_env: [{name: KION_ACC_X, reason: an x}]",
		"known_issue without number":       "known_issues: [{note: broken}]",
		"known_issue without note":         "known_issues: [{issue: 3}]",
		"import_unrecoverable w/o reason":  "import_unrecoverable: [{attr: color}]",
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "tv.yaml")
			raw := "resources:\n  thing:\n    " + body + "\n"
			if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, _, err := loadTestValues(path, "thing"); err == nil {
				t.Errorf("loadTestValues accepted %s", name)
			}
		})
	}
}

// collectionResourceModel is the label model plus a list and a set attribute.
func collectionResourceModel(t *testing.T) ResourceModel {
	rm := labelResourceModel(t)
	rm.Fields = append(rm.Fields,
		ModelField{GoName: "GrantTypes", TFSDK: "grant_types", Type: "types.List"},
		ModelField{GoName: "Scopes", TFSDK: "scopes", Type: "types.List"},
		ModelField{GoName: "OwnerIds", TFSDK: "owner_ids", Type: "types.Set"},
		ModelField{GoName: "Tags", TFSDK: "tags", Type: "types.Set"},
	)
	return rm
}

// terraform-plugin-testing rejects TestCheckResourceAttrSet on a list or set
// ("must be checked by element count key"), so a collection is checked by .#.
func TestRenderResourceTest_collectionsCheckedByCount(t *testing.T) {
	tv := labelTestValues()
	tv.Create["grant_types"] = `hcl:["authorization_code", "refresh_token"]`
	tv.Create["scopes"] = "hcl:var.scopes"
	tv.Create["owner_ids"] = "hcl:[kion_user.a.id]"
	tv.Create["tags"] = `hcl:["a", "b", "c"]`
	tv.Update["grant_types"] = `hcl:[]`
	got, err := renderResourceTest(collectionResourceModel(t), tv)
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	for _, w := range []string{
		`resource.TestCheckResourceAttr(resourceName, "grant_types.#", "2")`,
		`resource.TestCheckResourceAttr(resourceName, "grant_types.#", "0")`,
		`resource.TestCheckResourceAttrSet(resourceName, "scopes.#")`,
		`resource.TestCheckResourceAttrSet(resourceName, "owner_ids.#")`,
		`resource.TestCheckResourceAttr(resourceName, "tags.#", "3")`,
	} {
		if !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated resource test missing %q:\n%s", w, got)
		}
	}
	for _, name := range []string{"grant_types", "scopes", "owner_ids", "tags"} {
		if bad := `TestCheckResourceAttrSet(resourceName, "` + name + `")`; bytes.Contains(got, []byte(bad)) {
			t.Errorf("collection %s checked with AttrSet on the attribute itself", name)
		}
	}
}

func TestHCLListLen(t *testing.T) {
	cases := map[string]int{
		`[]`:                         0,
		`["a"]`:                      1,
		`["a", "b,c"]`:               2,
		`[ "a" , "b" , ]`:            2,
		`[{ a = 1, b = 2 }, [1, 2]]`: 2,
		`[kion_user.a.id, 3]`:        2,
	}
	for in, want := range cases {
		if n, ok := hclListLen(in); !ok || n != want {
			t.Errorf("hclListLen(%q) = %d, %v; want %d", in, n, ok, want)
		}
	}
	for _, in := range []string{`var.x`, `[for s in var.x : s]`, `concat(["a"], ["b"])`, `["a"]...`} {
		if _, ok := hclListLen(in); ok {
			t.Errorf("hclListLen(%q) claimed a literal count", in)
		}
	}
}

// A data source test may assert only what the data source's schema has.
func TestRenderDataSourceTest_onlyDataSourceAttrs(t *testing.T) {
	tv := labelTestValues()
	got, err := renderDataSourceTest(labelResourceModel(t), tv, []string{"key"})
	if err != nil {
		t.Fatalf("renderDataSourceTest: %v", err)
	}
	if !bytes.Contains(got, []byte(`TestCheckResourceAttrSet(dataSourceName, "key")`)) {
		t.Errorf("data source test does not assert key:\n%s", got)
	}
	for _, name := range []string{"color", "value"} {
		if bytes.Contains(got, []byte(`(dataSourceName, "`+name+`")`)) {
			t.Errorf("data source test asserts %s, which its schema lacks:\n%s", name, got)
		}
	}
}

// A create_only attribute exists only in the create response; the test must
// see it captured, unless the config produces none.
func TestRenderResourceTest_createOnlyAssertedSet(t *testing.T) {
	rm := labelResourceModel(t)
	rm.CreateOnly = []createOnlyAttr{{TF: "secret", ModelGo: "Secret"}}
	got, err := renderResourceTest(rm, labelTestValues())
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	if n := bytes.Count(got, []byte(`resource.TestCheckResourceAttrSet(resourceName, "secret")`)); n != 2 {
		t.Errorf("want secret asserted set after each create (basic and update), found %d:\n%s", n, got)
	}
	if !bytes.Contains(got, []byte(`ImportStateVerifyIgnore: []string{"secret"}`)) {
		t.Errorf("import verify no longer ignores secret:\n%s", got)
	}

	tv := labelTestValues()
	tv.CreateOnlyUnset = []string{"secret"}
	got, err = renderResourceTest(rm, tv)
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	if bytes.Contains(got, []byte(`(resourceName, "secret")`)) {
		t.Errorf("create_only_unset secret is still asserted:\n%s", got)
	}

	tv.CreateOnlyUnset = []string{"nope"}
	if _, err := renderResourceTest(rm, tv); err == nil {
		t.Error("create_only_unset naming a non-create_only attribute was accepted")
	}
}

// A map or nested object is held in state as name.% plus its members, and
// AttrSet on the name itself is refused like a collection's.
func TestRenderResourceTest_objectsCheckedByCount(t *testing.T) {
	rm := labelResourceModel(t)
	rm.Fields = append(rm.Fields,
		ModelField{GoName: "AzurePolicy", TFSDK: "azure_policy", Type: "AzurePolicyValue"},
		ModelField{GoName: "Labels", TFSDK: "labels", Type: "types.Map"},
	)
	tv := labelTestValues()
	tv.Create["azure_policy"] = `hcl:{ name = "x" }`
	tv.Create["labels"] = `hcl:{ a = "b" }`
	got, err := renderResourceTest(rm, tv)
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	for _, name := range []string{"azure_policy", "labels"} {
		if w := `resource.TestCheckResourceAttrSet(resourceName, "` + name + `.%")`; !bytes.Contains(got, []byte(w)) {
			t.Errorf("generated resource test missing %q:\n%s", w, got)
		}
		if bad := `TestCheckResourceAttrSet(resourceName, "` + name + `")`; bytes.Contains(got, []byte(bad)) {
			t.Errorf("object %s checked with AttrSet on the attribute itself", name)
		}
	}
}

// A compound "<parent>/<id>" import id is built from state, and an attribute
// the read cannot restore is left out of the import's verification.
func TestRenderResourceTest_importParentAndUnrecoverable(t *testing.T) {
	rm := labelResourceModel(t)
	rm.ImportParentTF = "color"
	tv := labelTestValues()
	tv.ImportUnrecoverable = []importUnrecoverable{{Attr: "value", Reason: "the read never returns it"}}
	got, err := renderResourceTest(rm, tv)
	if err != nil {
		t.Fatalf("renderResourceTest: %v", err)
	}
	if _, err := parser.ParseFile(token.NewFileSet(), "label_test.go", got, parser.ParseComments); err != nil {
		t.Fatalf("generated resource test does not parse: %v\n%s", err, got)
	}
	for w, n := range map[string]int{
		"ImportStateIdFunc: testAccLabelImportID(resourceName),":           2,
		`return rs.Primary.Attributes["color"] + "/" + rs.Primary.ID, nil`: 1,
		"// value: the read never returns it":                              2,
		`ImportStateVerifyIgnore: []string{"value"},`:                      2,
	} {
		if c := bytes.Count(got, []byte(w)); c != n {
			t.Errorf("want %d of %q, found %d:\n%s", n, w, c, got)
		}
	}

	tv.ImportUnrecoverable = []importUnrecoverable{{Attr: "nope", Reason: "r"}}
	if _, err := renderResourceTest(rm, tv); err == nil {
		t.Error("import_unrecoverable naming no attribute was accepted")
	}
}
