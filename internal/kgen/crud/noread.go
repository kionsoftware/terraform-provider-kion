package crud

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed noreadtest.gtpl
var noReadTestTmpl string

// noReadKind is the crud_archetypes.yaml kind for a resource with no
// single-record GET endpoint (create → id, no-op read keeping state, delete).
const noReadKind = "no_read"

// resolveNoRead assembles a ResourceModel for a no-read resource: create,
// (optional) update, and delete ops, with Read left empty so buildEntityData
// renders the no-read template.
func resolveNoRead(name string, ops resOps, idx sdkIndex, model []ModelField) (ResourceModel, error) {
	pascal := pascalCase(name)
	rm := ResourceModel{Name: name, Pascal: pascal, Model: pascal + "Model"}

	create, err := resolveOp("create", ops.Create, idx)
	if err != nil {
		return rm, err
	}
	if create == nil {
		return rm, fmt.Errorf("%s: no-read archetype requires a create op", name)
	}
	rm.Create = *create

	if rm.Update, err = resolveOp("update", ops.Update, idx); err != nil {
		return rm, err
	}
	if rm.Delete, err = resolveOp("delete", ops.Delete, idx); err != nil {
		return rm, err
	}

	for _, mf := range model {
		if mf.TFSDK == "id" {
			rm.IDField = mf
			continue
		}
		rm.Fields = append(rm.Fields, mf)
	}
	if rm.IDField.TFSDK == "" {
		return rm, fmt.Errorf("%s: generated model %s has no tfsdk:%q field; add one via codegen/schema_overrides.yaml for no-read resources", name, rm.Model, "id")
	}
	return rm, nil
}

// generateNoRead writes the files for a no-read resource: the resource, a stub
// sweeper, and a resource-only service_package.go (no data source, since there
// is no read endpoint to back one).
func (g *generator) generateNoRead(dir, name string, ops resOps, idx sdkIndex, model []ModelField, tvPath string, gated, force bool) (int, error) {
	rm, err := resolveNoRead(name, ops, idx, model)
	if err != nil {
		return 0, err
	}
	rm.Gated = gated

	// A no_read resource may still be readable over a private collection; see
	// parentread.go for why a resource without one imports as an empty shell.
	if pe, ok := g.privEnds[name]; ok && pe.ParentRead != nil {
		byTF := map[string]ModelField{}
		for _, mf := range model {
			byTF[mf.TFSDK] = mf
		}
		if rm.ParentRead, err = buildParentRead(name, *pe.ParentRead, byTF, rm.IDField.GoName); err != nil {
			return 0, err
		}
	}

	resourceGo, err := renderEntity(rm)
	if err != nil {
		return 0, err
	}
	sweepGo, sweepReason, err := renderSweep(rm)
	if err != nil {
		return 0, err
	}
	if sweepReason != "" {
		g.unswept = append(g.unswept, downgrade{Resource: name, Reason: sweepReason})
	}
	// DataSourceCtor must be supplied even when it resolves to "": the template
	// branches on it, and Go's text/template errors on a field that is absent
	// from the data struct rather than treating it as empty. Omitting it here
	// made every no_read resource fail template execution outright, not merely
	// lose its data source, so `kgen crud` skipped the resource entirely and
	// still exited 0. Keep this in step with the assoc, blended and raw_http
	// call sites, which pass the same three fields.
	dsCtor := dataSourceCompanionCtor(name, rm.Pascal)
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newServicePackageData(name, rm.Pascal, dsCtor), "service_package.go")
	if err != nil {
		return 0, err
	}

	files := []genFile{
		{filepath.Join(dir, name+".go"), resourceGo},
		{filepath.Join(dir, "sweep.go"), sweepGo},
		{filepath.Join(dir, "service_package.go"), pkgGo},
	}
	// Only worth saying when nothing was registered, and say the actual reason:
	// this branch is reached for resources that DO declare a read path but whose
	// read could not be resolved to a single-item op, so "no read endpoint" was
	// misleading. A companion template is what supplies a data source here.
	if dsCtor == "" {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no-read archetype; no data source emitted (no companion template registered in bespoke.go)\n", name)
	}
	for _, f := range files {
		if err := g.writeFile(f.path, f.data, force); err != nil {
			return 0, err
		}
	}
	tv, hasTV, err := loadTestValues(tvPath, name)
	if err != nil {
		return 0, err
	}
	if hasTV {
		listPath := ""
		if ops.Read != nil {
			listPath = ops.Read.Path
		}
		var pr *parentRead
		if pe, ok := g.privEnds[name]; ok {
			pr = pe.ParentRead
		}
		scan := noReadScanFor(name, pr, listPath, model)
		if scan.ListPath == "" {
			return 0, fmt.Errorf("%s: no-read archetype has test values but no read path to scan; record one in codegen/config_overrides.yaml", name)
		}
		if strings.Contains(scan.ListPath, "{id}") && scan.ParentIDTF == "" {
			return 0, fmt.Errorf("%s: collection path %q is parent-scoped but no *_id attribute matches", name, scan.ListPath)
		}
		td, err := buildNoReadTestData(rm, "kion_"+name, scan, tv)
		if err != nil {
			return 0, err
		}
		test, err := execGoTemplate("noreadtest", noReadTestTmpl, td, name+"_test.go")
		if err != nil {
			return 0, err
		}
		if err := g.writeFile(filepath.Join(dir, name+"_test.go"), test, force); err != nil {
			return 0, err
		}
	} else {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no test_values entry; skipping acceptance tests\n", name)
	}
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	if err := g.pruneUnwritten(dir, name); err != nil {
		return 0, err
	}
	return 1, nil
}

// noReadScan is the collection a no_read acceptance test scans for the record.
type noReadScan struct {
	ListPath   string // "{id}" stands for the parent id
	ParentIDTF string
	RecordsKey string // key inside data holding the records; "" when data is the list
}

// noReadScanFor picks the collection the existence check reads: the one the
// resource's own Read uses when it has a parent_read, else the spec read path.
func noReadScanFor(name string, pr *parentRead, readPath string, model []ModelField) noReadScan {
	if pr != nil {
		return noReadScan{
			ListPath:   strings.ReplaceAll(pr.Path, "{parent_id}", "{id}"),
			ParentIDTF: pr.ParentTF,
			RecordsKey: pr.Records,
		}
	}
	s := noReadScan{ListPath: readPath}
	// A {id} in the collection path is the PARENT's id.
	if strings.Contains(readPath, "{id}") {
		s.ParentIDTF = parentAttrFor(name, model)
	}
	return s
}

// parentAttrFor picks the model attribute naming the parent a parent-scoped
// collection path is keyed by, matching the leading path segment (an
// /v3/ou/{id}/... path is keyed by ou_id).
func parentAttrFor(name string, model []ModelField) string {
	for _, mf := range model {
		if !strings.HasSuffix(mf.TFSDK, "_id") {
			continue
		}
		if strings.HasPrefix(name, strings.TrimSuffix(mf.TFSDK, "_id")+"_") {
			return mf.TFSDK
		}
	}
	return ""
}

// dataSourceCompanionCtor returns the data-source constructor for a resource
// whose archetype emits no read, but which nonetheless ships a hand-authored
// data source as a companion file (see companionsByName).
//
// Without this the noread service package always emitted `return nil`, so those
// data sources were generated, compiled, tested, and never registered, leaving
// them unreachable from a practitioner's config. Registering them by hand did
// not survive, since the next `kgen crud` regenerated the file.
func dataSourceCompanionCtor(name, pascal string) string {
	for _, f := range companionsByName[name] {
		if strings.HasSuffix(f.outName, "_data_source.go") {
			return "New" + pascal + "DataSource"
		}
	}
	return ""
}
