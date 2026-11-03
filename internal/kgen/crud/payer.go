package crud

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// payerData parameterizes payer.gtpl.
type payerData struct {
	Pkg, Pascal, Model, ResConst, TypeName string
	SchemaVersion                          int
	Gated                                  bool
	CreatePath, ReadPath, UpdatePath       string
	DeletePath                             string
	CreateSeed                             string // JSON object literal
	MinKionVersion                         string // "" means ungated
	WireStructGo, FlattenGo, MergeGo       string
}

// generatePayer emits a payer-kind resource from its private_endpoints.yaml
// entry: create_seed, read_shape and write_shape. There is no data source.
func (g *generator) generatePayer(dir, name, tvPath string, gated, force bool) (int, error) {
	pe, ok := g.privEnds[name]
	if !ok {
		return 0, fmt.Errorf("%s: payer kind needs a private_endpoints.yaml entry", name)
	}
	if pe.ReadShape == nil || len(pe.WriteShape) == 0 {
		return 0, fmt.Errorf("%s: payer kind needs read_shape and write_shape", name)
	}
	for verb, op := range map[string]rawOp{"create": pe.Create, "read": pe.Read, "update": pe.Update, "delete": pe.Delete} {
		if op.Path == "" {
			return 0, fmt.Errorf("%s: payer kind needs a %s path", name, verb)
		}
	}
	seed := pe.CreateSeed
	if seed == "" {
		seed = "{}"
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(seed), &obj); err != nil {
		return 0, fmt.Errorf("%s: create_seed is not a JSON object: %w", name, err)
	}

	pascal := pascalCase(name)
	schemaGen := filepath.Join(dir, name+"_schema_gen.go")
	model, err := g.src.ModelFields(schemaGen, pascal+"Model")
	if err != nil {
		return 0, err
	}
	byTF := map[string]ModelField{}
	for _, mf := range model {
		byTF[mf.TFSDK] = mf
	}
	if err := checkReadShapeKinds(*pe.ReadShape, byTF, nil); err != nil {
		return 0, fmt.Errorf("%s: %w", name, err)
	}
	wireGo, err := buildWireStruct(name, *pe.ReadShape)
	if err != nil {
		return 0, fmt.Errorf("%s read_shape wire: %w", name, err)
	}
	flattenGo, err := buildNestedFlatten(*pe.ReadShape, byTF)
	if err != nil {
		return 0, fmt.Errorf("%s read_shape flatten: %w", name, err)
	}
	mergeGo, err := buildMergeFields(pe.WriteShape, byTF)
	if err != nil {
		return 0, fmt.Errorf("%s write_shape: %w", name, err)
	}

	data := payerData{
		Pkg: name, Pascal: pascal, Model: pascal + "Model", ResConst: "ResName" + pascal,
		TypeName: "kion_" + name, Gated: gated,
		CreatePath: pe.Create.Path, ReadPath: pe.Read.Path, UpdatePath: pe.Update.Path, DeletePath: pe.Delete.Path,
		CreateSeed: seed, WireStructGo: wireGo, FlattenGo: flattenGo, MergeGo: mergeGo,
		MinKionVersion: pe.MinKionVersion,
	}
	if data.MinKionVersion != "" && data.Gated {
		return 0, fmt.Errorf("%s: min_kion_version and a version_support gate both apply; keep one", name)
	}
	if _, ok := g.upgrades["kion_"+name]; ok {
		data.SchemaVersion = 1
	}
	n, err := g.emitBespoke(dir, name, data, []bespokeFile{{payerResourceTmpl, name + ".go"}}, force)
	if err != nil {
		return n, err
	}
	if err := g.emitCompanions(dir, name, force); err != nil {
		return n, err
	}
	if err := g.emitPayerTest(dir, name, pascal, tvPath, pe, model, force); err != nil {
		return n, err
	}
	// No data source, like the other typed billing sources: drop the scaffold's
	// stub and register the resource alone.
	if err := g.pruneUnwritten(dir, name); err != nil {
		return n, err
	}
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newServicePackageData(name, pascal, ""), "service_package.go")
	if err != nil {
		return n, err
	}
	return n, g.writeFile(filepath.Join(dir, "service_package.go"), pkgGo, force)
}

// emitPayerTest derives the acceptance test from test_values.yaml through the
// blended test template: both read through a private route. read_shape's
// write-only attributes are never read back, so import cannot verify them.
func (g *generator) emitPayerTest(dir, name, pascal, tvPath string, pe rawResourceOps, model []ModelField, force bool) error {
	tv, hasTV, err := loadTestValues(tvPath, name)
	if err != nil {
		return err
	}
	if !hasTV {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no test_values entry; skipping acceptance tests\n", name)
		return nil
	}
	td := blendedTestData{
		acctestData: acctestData{Pkg: name, Pascal: pascal, SDKAlias: "generated", SDKImport: sdkImport()},
		TypeName:    "kion_" + name,
		RawRead:     &rawReadData{Method: "RawGet", Path: pe.Read.Path},
	}
	if err := fillConfig(&td.acctestData, tv, modelTypes(model), len(tv.Update) > 0); err != nil {
		return err
	}
	for _, sc := range pe.ReadShape.Scalars {
		if sc.WriteOnly {
			td.ImportVerifyIgnore = append(td.ImportVerifyIgnore, sc.TF)
		}
	}
	test, err := execGoTemplate("blendedtest", blendedTestTmpl, td, name+"_test.go")
	if err != nil {
		return err
	}
	return g.writeFile(filepath.Join(dir, name+"_test.go"), test, force)
}
