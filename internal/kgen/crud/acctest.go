package crud

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

//go:embed restest.gtpl
var resourceTestTmpl string

//go:embed dstest.gtpl
var dataSourceTestTmpl string

// testValues are the create/update sample attribute values for one resource.
type testValues struct {
	// EnvArgs names environment variables whose values the config needs; each
	// becomes an extra parameter, readable as %[2]s, %[3]s, ... in order.
	EnvArgs []string `yaml:"env_args"`
	// Prereqs is HCL emitted ahead of the resource under test, for resources
	// that cannot exist without a parent. It is rendered through the same
	// Sprintf as the rest of the config, so %[1]s is the random name.
	Prereqs string `yaml:"prereqs"`
	// Fixtures name shared HCL builders in internal/acctest, called with rName
	// and prepended to the configuration.
	Fixtures []string `yaml:"fixtures"`
	// ExtraArgs are Go expressions appended to the config Sprintf after rName
	// and the env args, for a value the HCL cannot state literally.
	ExtraArgs []string `yaml:"extra_args"`
	// RequireEnv are variables the test cannot run without but whose values the
	// config does not use; each is checked in PreCheck and skips when unset.
	RequireEnv []envRequirement `yaml:"require_env"`
	// KnownIssues are open defects the test is expected to surface.
	KnownIssues []knownIssue `yaml:"known_issues"`
	// CreateOnlyUnset are create_only attributes this config does not produce,
	// so the test does not assert them set after create.
	CreateOnlyUnset []string `yaml:"create_only_unset"`
	// ImportUnrecoverable are attributes the read cannot restore, so an import
	// leaves them null; each is left out of the import's verification.
	ImportUnrecoverable []importUnrecoverable `yaml:"import_unrecoverable"`
	Create              map[string]string     `yaml:"create"`
	Update              map[string]string     `yaml:"update"`
}

// envRequirement is a skip guard: Reason completes "<Name> must be set to ...".
type envRequirement struct {
	Name   string `yaml:"name"`
	Reason string `yaml:"reason"`
}

// importUnrecoverable is one attribute import cannot restore, and why.
type importUnrecoverable struct {
	Attr   string `yaml:"attr"`
	Reason string `yaml:"reason"`
}

// knownIssue names an open defect in a comment atop the generated test file;
// ExpectNonEmptyPlan tolerates the plan drift it causes on every apply step.
type knownIssue struct {
	Issue              int    `yaml:"issue"`
	Note               string `yaml:"note"`
	ExpectNonEmptyPlan bool   `yaml:"expect_non_empty_plan"`
}

// validate rejects an entry whose guard or issue record is incomplete, so a
// typo cannot turn into a gate or a comment that says nothing.
func (tv testValues) validate() error {
	for _, r := range tv.RequireEnv {
		if r.Name == "" || r.Reason == "" {
			return fmt.Errorf("require_env entry %+v needs both name and reason", r)
		}
		if slices.Contains(tv.EnvArgs, r.Name) {
			return fmt.Errorf("require_env %s is already in env_args, which gates it", r.Name)
		}
	}
	for _, u := range tv.ImportUnrecoverable {
		if u.Attr == "" || strings.TrimSpace(u.Reason) == "" {
			return fmt.Errorf("import_unrecoverable entry %+v needs both attr and reason", u)
		}
	}
	for _, k := range tv.KnownIssues {
		if k.Issue <= 0 || strings.TrimSpace(k.Note) == "" {
			return fmt.Errorf("known_issues entry %+v needs an issue number and a note", k)
		}
	}
	return nil
}

// acctestKnownIssue is a knownIssue split into comment lines for the template.
type acctestKnownIssue struct {
	Issue int
	Lines []string
}

// fillGuards copies the skip guards and known issues into the payload.
func fillGuards(d *acctestData, tv testValues) {
	d.RequireEnv = tv.RequireEnv
	for _, k := range tv.KnownIssues {
		d.KnownIssues = append(d.KnownIssues, acctestKnownIssue{
			Issue: k.Issue, Lines: strings.Split(strings.TrimSpace(k.Note), "\n"),
		})
		d.ExpectNonEmptyPlan = d.ExpectNonEmptyPlan || k.ExpectNonEmptyPlan
	}
}

// hclPrefix marks a value as an HCL expression rather than a literal: it is
// emitted unquoted and asserted only for presence, since a reference like
// kion_project.test.id has no value until apply.
const hclPrefix = "hcl:"

type testValuesFile struct {
	Resources map[string]testValues `yaml:"resources"`
}

// loadTestValues reads codegen/test_values.yaml and returns the entry for one
// resource. A missing file or entry yields ok=false (caller skips test gen).
func loadTestValues(path, resource string) (testValues, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return testValues{}, false, nil
		}
		return testValues{}, false, fmt.Errorf("reading test values %s: %w", path, err)
	}
	var tvf testValuesFile
	if err := yaml.Unmarshal(raw, &tvf); err != nil {
		return testValues{}, false, fmt.Errorf("parsing test values %s: %w", path, err)
	}
	tv, ok := tvf.Resources[resource]
	if err := tv.validate(); err != nil {
		return testValues{}, false, fmt.Errorf("test values for %s: %w", resource, err)
	}
	return tv, ok, nil
}

type acctestAttr struct {
	Name  string
	Value string
	// Quoted is false for an attribute the schema types as a number or bool.
	// Every test value was a string until env_args arrived, so the template
	// quoted unconditionally and a numeric value came out as "3" -- which
	// Terraform coerces but `make ci-acctest-config` rejects.
	Quoted bool
	// CheckExpr is the Go expression for the value this attribute must hold in
	// state: a literal, or a Sprintf matching the one that built the config.
	// Terraform stores every primitive as a string, so numbers and bools are
	// compared as their unquoted text. Empty for an IsHCL attribute.
	CheckExpr string
	// IsHCL marks a value that is an expression to emit verbatim.
	IsHCL bool
	// Collection marks a list or set attribute, which state holds as name.# plus
	// one key per element; Set is a set, whose duplicate elements collapse.
	Collection, Set bool
	// Object marks a map or nested object, which state holds as name.% plus one
	// key per member; AttrSet on the name itself is refused.
	Object bool
}

// Check is the assertion a generated test makes for the attribute after apply.
// terraform-plugin-testing refuses AttrSet on a list or set itself, so a
// collection is checked by its element count, exactly when the value says it.
func (a acctestAttr) Check() string {
	switch {
	case a.Collection:
		if n, ok := a.literalLen(); ok {
			return fmt.Sprintf("resource.TestCheckResourceAttr(resourceName, %q, %q)", a.Name+".#", strconv.Itoa(n))
		}
		return fmt.Sprintf("resource.TestCheckResourceAttrSet(resourceName, %q)", a.Name+".#")
	case a.Object:
		return fmt.Sprintf("resource.TestCheckResourceAttrSet(resourceName, %q)", a.Name+".%")
	case a.IsHCL:
		return fmt.Sprintf("resource.TestCheckResourceAttrSet(resourceName, %q)", a.Name)
	default:
		return fmt.Sprintf("resource.TestCheckResourceAttr(resourceName, %q, %s)", a.Name, a.CheckExpr)
	}
}

// literalLen is the element count a collection's value fixes. A set counts only
// distinct constants: two expressions may evaluate to one element.
func (a acctestAttr) literalLen() (int, bool) {
	n, ok := hclListLen(a.Value)
	if !ok || !a.Set {
		return n, ok
	}
	elems, _ := hclListElems(a.Value)
	seen := map[string]bool{}
	for _, e := range elems {
		if !hclConstant.MatchString(e) {
			return 0, false
		}
		seen[e] = true
	}
	return len(seen), true
}

var hclConstant = regexp.MustCompile(`^("([^"\\]|\\.)*"|-?[0-9.]+|true|false)$`)

// hclListLen counts the elements of a plain HCL tuple literal ([a, b]). A for
// expression, function call, reference or expansion has no count until apply.
func hclListLen(v string) (int, bool) {
	elems, ok := hclListElems(v)
	return len(elems), ok
}

func hclListElems(v string) ([]string, bool) {
	v = strings.TrimSpace(v)
	if len(v) < 2 || v[0] != '[' || v[len(v)-1] != ']' {
		return nil, false
	}
	inner := v[1 : len(v)-1]
	var elems []string
	depth, start, inStr := 0, 0, false
	for i := 0; i < len(inner); i++ {
		c := inner[i]
		switch {
		case inStr && c == '\\':
			i++
		case c == '"':
			inStr = !inStr
		case inStr:
		case c == '[' || c == '{' || c == '(':
			depth++
		case c == ']' || c == '}' || c == ')':
			depth--
			if depth < 0 {
				return nil, false // the outer brackets do not enclose one tuple
			}
		case c == ',' && depth == 0:
			elems = append(elems, strings.TrimSpace(inner[start:i]))
			start = i + 1
		}
	}
	if inStr || depth != 0 {
		return nil, false
	}
	if last := strings.TrimSpace(inner[start:]); last != "" {
		elems = append(elems, last)
	}
	for _, e := range elems {
		if e == "" || strings.HasPrefix(e, "for ") || strings.HasSuffix(e, "...") {
			return nil, false
		}
	}
	return elems, true
}

type acctestData struct {
	Pkg, Pascal, ResourceType string
	SDKAlias                  string
	SDKImport                 string
	ReadMethod, ReadParams    string
	ReadIDParam               string
	IDParamType               string // "int64" | "uint64" | "string"
	// ConfirmListed/ConfirmListedTF: see ResourceModel.ConfirmListed.
	ConfirmListed, ConfirmListedTF string
	IDFormat                       string // "%d" | "%s"
	AttrNames                      []string
	CreateAttrs                    []acctestAttr
	UpdateAttrs                    []acctestAttr
	HasUpdate                      bool
	// BasicUsesRName is whether the inline HCL has a verb to interpolate, so it
	// must go through Sprintf; NeedsRName is whether the function takes rName at
	// all, which a fixture call needs even with nothing to interpolate.
	BasicUsesRName   bool
	UpdateUsesRName  bool
	BasicNeedsRName  bool
	UpdateNeedsRName bool
	// Prereqs is HCL for the parent resources the resource under test needs,
	// emitted ahead of it in both the basic and update configurations.
	Prereqs   string
	Fixtures  []string
	ExtraArgs []string
	// Parent-list fields: these resources have no by-id GET, so Exists and
	// Destroy list the parent's children and look for the record.
	ParentParam, ParentCast, ParentIDTF string
	ResponseType, RecordIDGo            string
	RecordIDOpt                         bool
	// EnvArgs are environment variables whose VALUES the config interpolates,
	// in order, as %[2]s, %[3]s, ... A value the API requires but the schema
	// marks optional (kion_category's payer_id) is install-specific, so it
	// cannot be a literal in test_values.yaml.
	EnvArgs []acctestEnvArg
	// RequireEnv are PreCheck skip guards whose values the config ignores.
	RequireEnv         []envRequirement
	KnownIssues        []acctestKnownIssue
	ExpectNonEmptyPlan bool
	// ImportVerifyIgnore are the create_only attributes: an import has no
	// create response to capture them from, so they are null after import.
	ImportVerifyIgnore []string
	// ImportUnrecoverable are test_values' import_unrecoverable entries, also
	// in ImportVerifyIgnore, kept whole for the comment naming each reason.
	ImportUnrecoverable []importUnrecoverable
	// MissingViaList: a 500 from the read is "gone" when this list omits the id.
	MissingViaList string
	// IDIsParent: the route answers a deleted record with an all-zero one.
	IDIsParent bool
	// ImportParentTF is the parent attribute a compound "<parent>/<id>" import
	// id carries; the test builds that id from state.
	ImportParentTF string
	// CreateOnlySet are the create_only attributes asserted set after create:
	// all of them, less those test_values says this config does not produce.
	CreateOnlySet []string
}

func buildAcctestData(rm ResourceModel, tv testValues) (acctestData, error) {
	readIDParam, readIDType, err := idParamName(rm.Read.Params)
	if err != nil {
		return acctestData{}, fmt.Errorf("%s acctest: %w", rm.Name, err)
	}
	d := acctestData{
		Pkg:          rm.Name,
		Pascal:       rm.Pascal,
		ResourceType: "kion_" + rm.Name,
		SDKAlias:     "generated",
		SDKImport:    sdkImport(),
		ReadMethod:   rm.Read.Method.Name,
		ReadParams:   rm.Read.Method.ParamsType,
		ReadIDParam:  readIDParam,
		IDParamType:  readIDType,
		IDFormat:     idFormat(readIDType),

		ConfirmListed:   rm.ConfirmListed,
		ConfirmListedTF: rm.ConfirmListedTF,
		ImportParentTF:  rm.ImportParentTF,
		MissingViaList:  rm.MissingViaList,
		IDIsParent:      rm.IDIsParent,
	}
	d.CreateAttrs = sortAttrs(tv.Create, rm)
	d.UpdateAttrs = sortAttrs(tv.Update, rm)
	for _, a := range d.CreateAttrs {
		d.AttrNames = append(d.AttrNames, a.Name)
	}
	d.HasUpdate = rm.Update != nil && len(tv.Update) > 0
	d.EnvArgs = envArgsFor(tv.EnvArgs)
	fillGuards(&d, tv)
	unset := map[string]bool{}
	for _, n := range tv.CreateOnlyUnset {
		unset[n] = true
	}
	for _, a := range rm.CreateOnly {
		d.ImportVerifyIgnore = append(d.ImportVerifyIgnore, a.TF)
		if !unset[a.TF] {
			d.CreateOnlySet = append(d.CreateOnlySet, a.TF)
		}
		delete(unset, a.TF)
	}
	for _, u := range tv.ImportUnrecoverable {
		if !slices.ContainsFunc(rm.Fields, func(f ModelField) bool { return f.TFSDK == u.Attr }) {
			return acctestData{}, fmt.Errorf("%s test_values: import_unrecoverable names %q, which is not an attribute", rm.Name, u.Attr)
		}
		d.ImportVerifyIgnore = append(d.ImportVerifyIgnore, u.Attr)
	}
	d.ImportUnrecoverable = tv.ImportUnrecoverable
	if len(unset) > 0 {
		return acctestData{}, fmt.Errorf("%s test_values: create_only_unset names %v, which are not create_only attributes", rm.Name, sortedKeys(unset))
	}
	d.Fixtures = tv.Fixtures
	d.ExtraArgs = tv.ExtraArgs
	setCheckExprs(d.CreateAttrs, d.EnvArgs, d.ExtraArgs)
	setCheckExprs(d.UpdateAttrs, d.EnvArgs, d.ExtraArgs)
	d.Prereqs = strings.TrimRight(tv.Prereqs, "\n")
	prereqVerb := strings.Contains(d.Prereqs, "%")
	d.BasicUsesRName = usesFormatVerb(d.CreateAttrs) || prereqVerb
	d.UpdateUsesRName = usesFormatVerb(d.UpdateAttrs) || prereqVerb
	d.BasicNeedsRName = d.BasicUsesRName || len(d.Fixtures) > 0
	d.UpdateNeedsRName = d.UpdateUsesRName || len(d.Fixtures) > 0
	return d, nil
}

// fillConfig populates the configuration half of the payload, shared by every
// archetype's test builder; the probe half differs per archetype. attrTypes
// maps each attribute to its model type.
func fillConfig(d *acctestData, tv testValues, attrTypes map[string]string, hasUpdate bool) error {
	if len(tv.CreateOnlyUnset) > 0 {
		return fmt.Errorf("%s test_values: create_only_unset, but this archetype has no create_only attributes", d.Pkg)
	}
	d.CreateAttrs = sortAttrsByType(tv.Create, attrTypes)
	d.UpdateAttrs = sortAttrsByType(tv.Update, attrTypes)
	for _, a := range d.CreateAttrs {
		d.AttrNames = append(d.AttrNames, a.Name)
	}
	d.HasUpdate = hasUpdate && len(tv.Update) > 0
	d.EnvArgs = envArgsFor(tv.EnvArgs)
	fillGuards(d, tv)
	d.Fixtures, d.ExtraArgs = tv.Fixtures, tv.ExtraArgs
	setCheckExprs(d.CreateAttrs, d.EnvArgs, d.ExtraArgs)
	setCheckExprs(d.UpdateAttrs, d.EnvArgs, d.ExtraArgs)
	d.Prereqs = strings.TrimRight(tv.Prereqs, "\n")
	verb := strings.Contains(d.Prereqs, "%")
	d.BasicUsesRName = usesFormatVerb(d.CreateAttrs) || verb
	d.UpdateUsesRName = usesFormatVerb(d.UpdateAttrs) || verb
	d.BasicNeedsRName = d.BasicUsesRName || len(d.Fixtures) > 0
	d.UpdateNeedsRName = d.UpdateUsesRName || len(d.Fixtures) > 0
	return nil
}

// modelTypes maps each model attribute to its type.
func modelTypes(model []ModelField) map[string]string {
	out := make(map[string]string, len(model))
	for _, mf := range model {
		out[mf.TFSDK] = mf.Type
	}
	return out
}

// blendedTestData is the blended payload: the shared configuration fields plus
// the private route the resource reads through.
type blendedTestData struct {
	acctestData
	TypeName string
	RawRead  *rawReadData
}

func buildBlendedTestData(d entityData, model []ModelField, tv testValues) (blendedTestData, error) {
	out := blendedTestData{
		acctestData: acctestData{Pkg: d.Pkg, Pascal: d.Pascal, SDKAlias: d.SDKAlias, SDKImport: d.SDKImport},
		TypeName:    d.TypeName, RawRead: d.RawRead,
	}
	return out, fillConfig(&out.acctestData, tv, modelTypes(model), d.HasUpdate)
}

// assocTestData is the association payload: shared configuration fields plus
// the membership list the row lives in.
type assocTestData struct {
	acctestData
	TypeName, ReadMethod, ReadParams string
	HasParent                        bool
	ParentArg, ParentTF, KeyTF       string
	RespType, DataGo, RecordKeyGo    string
}

func buildAssocTestData(a assocData, tv testValues) (assocTestData, error) {
	out := assocTestData{
		acctestData: acctestData{Pkg: a.Pkg, Pascal: a.Pascal, SDKAlias: a.SDKAlias, SDKImport: a.SDKImport},
		TypeName:    a.TypeName, ReadMethod: a.ReadMethod, ReadParams: a.ReadParams,
		HasParent: a.HasParent, ParentArg: a.ParentArg, ParentTF: a.ParentTF, KeyTF: a.KeyTF,
		RespType: a.RespType, DataGo: a.DataGo, RecordKeyGo: a.RecordKeyGo,
	}
	return out, fillConfig(&out.acctestData, tv, a.AttrTypes, a.HasUpdate)
}

// buildRawTestData reuses the blended payload: a raw_http resource reads
// through the same kind of private route.
func buildRawTestData(r rawData, model []ModelField, tv testValues) (blendedTestData, error) {
	out := blendedTestData{
		acctestData: acctestData{Pkg: r.Pkg, Pascal: r.Pascal, SDKAlias: "generated", SDKImport: sdkImport()},
		TypeName:    r.TypeName,
		RawRead:     &rawReadData{Method: r.ReadMethod, Path: r.ReadPath},
	}
	return out, fillConfig(&out.acctestData, tv, modelTypes(model), r.HasUpdate)
}

// noReadTestData is the no_read payload: the shared configuration fields plus
// the collection path a test must scan, since there is no single-record GET.
type noReadTestData struct {
	acctestData
	TypeName, ListPath, ParentIDTF, RecordsKey string
}

func buildNoReadTestData(rm ResourceModel, typeName string, scan noReadScan, tv testValues) (noReadTestData, error) {
	out := noReadTestData{
		acctestData: acctestData{Pkg: rm.Name, Pascal: rm.Pascal, SDKAlias: "generated", SDKImport: sdkImport()},
		TypeName:    typeName, ListPath: scan.ListPath, ParentIDTF: scan.ParentIDTF, RecordsKey: scan.RecordsKey,
	}
	return out, fillConfig(&out.acctestData, tv, modelTypes(rm.Fields), rm.Update != nil)
}

// buildParentListTestData fills the same payload from a parent_list resolution,
// whose read is a list under the parent rather than a get by id.
func buildParentListTestData(pl parentListData, tv testValues) (acctestData, error) {
	d := acctestData{
		Pkg: pl.Pkg, Pascal: pl.Pascal, ResourceType: pl.TypeName, SDKAlias: pl.SDKAlias, SDKImport: pl.SDKImport,
		ReadMethod: pl.ReadMethod, ReadParams: pl.ReadParams,
		ParentParam: pl.ParentParam, ParentCast: pl.ParentCast, ParentIDTF: pl.ParentIDTF,
		ResponseType: pl.ResponseType, RecordIDGo: pl.RecordIDGo, RecordIDOpt: pl.RecordIDOpt,
	}
	return d, fillConfig(&d, tv, pl.AttrTypes, pl.HasUpdate)
}

// sortAttrsByType is sortAttrs against a map of model types, for archetypes
// that carry no ResourceModel. An attribute it does not know is unquoted.
func sortAttrsByType(m map[string]string, attrTypes map[string]string) []acctestAttr {
	out := make([]acctestAttr, 0, len(m))
	for k, v := range m {
		out = append(out, newAcctestAttr(k, v, attrTypes[k], attrTypes[k] == "types.String"))
	}
	slices.SortFunc(out, func(a, b acctestAttr) int { return cmp.Compare(a.Name, b.Name) })
	return out
}

func sortAttrs(m map[string]string, rm ResourceModel) []acctestAttr {
	types := modelTypes(rm.Fields)
	out := make([]acctestAttr, 0, len(m))
	for k, v := range m {
		out = append(out, newAcctestAttr(k, v, types[k], attrIsString(rm, k)))
	}
	slices.SortFunc(out, func(a, b acctestAttr) int { return cmp.Compare(a.Name, b.Name) })
	return out
}

func newAcctestAttr(name, value, modelType string, quoted bool) acctestAttr {
	a := acctestAttr{Name: name, Value: value, Quoted: quoted}
	if expr, ok := strings.CutPrefix(value, hclPrefix); ok {
		a = acctestAttr{Name: name, Value: expr, IsHCL: true}
	}
	a.Collection = modelType == "types.List" || modelType == "types.Set"
	a.Set = modelType == "types.Set"
	a.Object = modelType == "types.Map" || modelType == "basetypes.ObjectValue" ||
		(!strings.Contains(modelType, ".") && strings.HasSuffix(modelType, "Value"))
	return a
}

// setCheckExprs fills in each attribute's CheckExpr. A value with no verb is
// its own literal; one with a verb is rebuilt by the same Sprintf the config
// used, so the assertion and the configuration cannot drift apart.
func setCheckExprs(attrs []acctestAttr, envArgs []acctestEnvArg, extra []string) {
	for i := range attrs {
		if attrs[i].IsHCL {
			continue
		}
		attrs[i].CheckExpr = checkExpr(attrs[i].Value, envArgs, extra)
	}
}

func checkExpr(value string, envArgs []acctestEnvArg, extra []string) string {
	if !strings.Contains(value, "%") {
		return strconv.Quote(value)
	}
	args := []string{"rName"}
	for _, e := range envArgs {
		args = append(args, e.Param)
	}
	args = append(args, extra...)
	// Passing more arguments than the verbs consume makes Sprintf emit
	// %!(EXTRA ...) and trips vet, so the list stops at the highest index used.
	if n := maxFormatIndex(value); n > 0 && n < len(args) {
		args = args[:n]
	}
	return fmt.Sprintf("fmt.Sprintf(%s, %s)", strconv.Quote(value), strings.Join(args, ", "))
}

// maxFormatIndex returns the highest N across the value's `%[N]` verbs, or 0
// when it uses none (unindexed verbs consume arguments in order instead).
func maxFormatIndex(value string) int {
	highest := 0
	for i := 0; i+1 < len(value); i++ {
		if value[i] != '%' || value[i+1] != '[' {
			continue
		}
		end := strings.IndexByte(value[i+2:], ']')
		if end < 0 {
			continue
		}
		n, err := strconv.Atoi(value[i+2 : i+2+end])
		if err == nil && n > highest {
			highest = n
		}
	}
	return highest
}

// usesFormatVerb reports whether any value contains a `%` verb, so the config
// function must fmt.Sprintf(..., rName) rather than return a raw literal.
func usesFormatVerb(attrs []acctestAttr) bool {
	for _, a := range attrs {
		if strings.Contains(a.Value, "%") {
			return true
		}
	}
	return false
}

func renderResourceTest(rm ResourceModel, tv testValues) ([]byte, error) {
	return renderTest("resourcetest", resourceTestTmpl, rm, tv)
}

// renderDataSourceTest renders the data source test; dsAttrs are the
// attributes the data source's schema has, the only ones it may assert.
func renderDataSourceTest(rm ResourceModel, tv testValues, dsAttrs []string) ([]byte, error) {
	return renderTest("datasourcetest", dataSourceTestTmpl, rm, tv, func(d *acctestData) {
		d.AttrNames = slices.DeleteFunc(d.AttrNames, func(n string) bool { return !slices.Contains(dsAttrs, n) })
	})
}

func renderTest(name, tmpl string, rm ResourceModel, tv testValues, adjust ...func(*acctestData)) ([]byte, error) {
	data, err := buildAcctestData(rm, tv)
	if err != nil {
		return nil, err
	}
	for _, f := range adjust {
		f(&data)
	}
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse %s template: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute %s template for %s: %w", name, rm.Name, err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated %s for %s: %w\n%s", name, rm.Name, err, buf.Bytes())
	}
	return src, nil
}

// acctestEnvArg pairs an environment variable with the Go parameter name the
// generated config function receives it as.
type acctestEnvArg struct {
	Env   string // e.g. KION_ACC_BILLING_SOURCE_ID
	Param string // e.g. billingSourceID
	// Describes is the phrase RequireEnv prints when the variable is unset.
	Describes string // e.g. "the billing source id"
}

// envArgsFor derives parameter names from variable names. Computed here rather
// than in the template: the templates are rendered with no FuncMap.
func envArgsFor(envs []string) []acctestEnvArg {
	out := make([]acctestEnvArg, 0, len(envs))
	for _, e := range envs {
		parts := strings.Split(strings.ToLower(strings.TrimPrefix(e, "KION_ACC_")), "_")
		for i, p := range parts {
			switch {
			case p == "":
			case i == 0:
			case p == "id":
				parts[i] = "ID"
			default:
				parts[i] = strings.ToUpper(p[:1]) + p[1:]
			}
		}
		words := strings.ToLower(strings.TrimPrefix(e, "KION_ACC_"))
		out = append(out, acctestEnvArg{
			Env:       e,
			Param:     strings.Join(parts, ""),
			Describes: "the " + strings.ReplaceAll(words, "_", " "),
		})
	}
	return out
}

// attrIsString reports whether the model types an attribute as a string, which
// is what decides whether the generated HCL quotes its value.
func attrIsString(rm ResourceModel, tfsdk string) bool {
	for _, f := range rm.Fields {
		if f.TFSDK == tfsdk {
			return f.Type == "types.String"
		}
	}
	// Unknown attributes keep the historical behavior.
	return true
}
