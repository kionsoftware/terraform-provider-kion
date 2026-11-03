package crud

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed assoctest.gtpl
var assocTestTmpl string

//go:embed association.gtpl
var associationTmpl string

// assocKind is the crud_archetypes.yaml kind for a single list-member row keyed
// by a field, scoped by an optional parent, over a bulk read + replace-list API
// (the permission-mapping resources).
const assocKind = "association"

// assocMember is one repeated-membership field (a types.Set of ids on the model,
// an OptNilUint64Array on the SDK record).
type assocMember struct {
	ModelGo  string // "UserIds"
	TF       string // "user_ids"
	RecordGo string // "UserIds"
}

// assocData is the association template payload.
type assocData struct {
	Pkg, Pascal, Model, SDKAlias string
	SDKImport                    string
	ResConst, ResName            string
	Gated                        bool
	TypeName                     string

	IDGo         string // composite-id model field ("Id")
	KeyGo, KeyTF string // "AppRoleId","app_role_id"
	RecordKeyGo  string // "AppRoleID" (NilUint64 field on the record)

	HasParent          bool
	ParentGo, ParentTF string // "FundingSourceId","funding_source_id"
	ParentArg          string // parent path-param GoName in the ops (e.g. "ID")

	RespType   string // "GlobalUserMappingListResponse"
	DataGo     string // "Data"
	RecordType string // "UserMapping"

	ReadMethod   string
	ReadParams   string
	CreateMethod string
	CreateParams string
	HasUpdate    bool
	UpdateMethod string
	UpdateParams string
	HasDeleteOp  bool
	DeleteMethod string
	DeleteParams string

	Members []assocMember

	// AttrTypes is each attribute's model type, for a generated test.
	AttrTypes map[string]string
}

// resolveAssoc assembles an assocData from the op-set, SDK index, archetype
// declaration, and model. Refuses (error) any shape it can't map.
func resolveAssoc(name string, ops resOps, idx sdkIndex, arch archetype, model []ModelField) (assocData, error) {
	pascal := pascalCase(name)
	d := assocData{
		Pkg:       name,
		Pascal:    pascal,
		Model:     pascal + "Model",
		SDKAlias:  "generated",
		SDKImport: sdkImport(),
		ResConst:  "ResName" + pascal,
		ResName:   pascal + " Resource",
		TypeName:  "kion_" + name,
	}

	byTF := map[string]ModelField{}
	d.AttrTypes = map[string]string{}
	for _, mf := range model {
		byTF[mf.TFSDK] = mf
		d.AttrTypes[mf.TFSDK] = mf.Type
		switch mf.TFSDK {
		case "id":
			d.IDGo = mf.GoName
		case arch.KeyField:
			d.KeyGo, d.KeyTF = mf.GoName, mf.TFSDK
		case arch.ParentField:
			d.ParentGo, d.ParentTF, d.HasParent = mf.GoName, mf.TFSDK, true
		}
	}
	if d.IDGo == "" || d.KeyGo == "" {
		return d, fmt.Errorf("%s: model missing id/%s", name, arch.KeyField)
	}
	if arch.ParentField != "" && !d.HasParent {
		return d, fmt.Errorf("%s: parent_field %q not in model", name, arch.ParentField)
	}

	// Read op → the list response + its []record Data field + record type.
	read, err := resolveOp("read", ops.Read, idx)
	if err != nil {
		return d, err
	}
	if read == nil {
		return d, fmt.Errorf("%s: association requires a read op", name)
	}
	d.ReadMethod = read.Method.Name
	d.ReadParams = read.Method.ParamsType
	marker := lowerFirst(read.Method.ResultType)
	var respType string
	for _, t := range idx.markerImpls[marker] {
		if !sharedResponse[t] {
			respType = t
		}
	}
	if respType == "" {
		return d, fmt.Errorf("%s: no list response impl for %s", name, marker)
	}
	d.RespType = respType
	var recType string
	for _, f := range idx.structs[respType].Fields {
		if f.GoName == "Data" || f.JSONName == "data" {
			d.DataGo = f.GoName
			if elem, ok := sliceElem(f.Type); ok {
				recType = elem
			}
		}
	}
	if d.DataGo == "" || recType == "" {
		return d, fmt.Errorf("%s: list response %s has no []record Data field", name, respType)
	}
	d.RecordType = recType

	rec, ok := idx.structs[recType]
	if !ok {
		return d, fmt.Errorf("%s: record type %s not found", name, recType)
	}
	recByJSON := map[string]Field{}
	for _, f := range rec.Fields {
		recByJSON[f.JSONName] = f
	}
	kf, ok := recByJSON[arch.KeyField]
	if !ok {
		return d, fmt.Errorf("%s: record %s has no %s field", name, recType, arch.KeyField)
	}
	d.RecordKeyGo = kf.GoName

	for _, m := range arch.MemberFields {
		mf, ok := byTF[m]
		if !ok {
			return d, fmt.Errorf("%s: member_field %q not in model", name, m)
		}
		rf, ok := recByJSON[m]
		if !ok {
			return d, fmt.Errorf("%s: record %s has no %s field", name, recType, m)
		}
		d.Members = append(d.Members, assocMember{ModelGo: mf.GoName, TF: m, RecordGo: rf.GoName})
	}

	// Create / update / delete ops.
	create, err := resolveOp("create", ops.Create, idx)
	if err != nil {
		return d, err
	}
	if create == nil {
		return d, fmt.Errorf("%s: association requires a create op", name)
	}
	d.CreateMethod = create.Method.Name
	d.CreateParams = create.Method.ParamsType

	if update, err := resolveOp("update", ops.Update, idx); err != nil {
		return d, err
	} else if update != nil {
		d.HasUpdate = true
		d.UpdateMethod = update.Method.Name
		d.UpdateParams = update.Method.ParamsType
	}
	if del, err := resolveOp("delete", ops.Delete, idx); err != nil {
		return d, err
	} else if del != nil {
		d.HasDeleteOp = true
		d.DeleteMethod = del.Method.Name
		d.DeleteParams = del.Method.ParamsType
	}

	// Parent path-param name (single-field params on the parented ops).
	if d.HasParent {
		if read.Params == nil || len(read.Params.Fields) != 1 {
			return d, fmt.Errorf("%s: parented read must have one path param", name)
		}
		d.ParentArg = read.Params.Fields[0].GoName
	}
	return d, nil
}

// generateAssoc writes the association resource + stub sweeper + resource-only
// service_package (list-member associations expose no data source).
func (g *generator) generateAssoc(dir, name string, ops resOps, idx sdkIndex, arch archetype, model []ModelField, tvPath string, gated, force bool) (int, error) {
	d, err := resolveAssoc(name, ops, idx, arch, model)
	if err != nil {
		return 0, err
	}
	d.Gated = gated

	resourceGo, err := execGoTemplate("association", associationTmpl, d, name+".go")
	if err != nil {
		return 0, err
	}
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newServicePackageData(name, d.Pascal, dataSourceCompanionCtor(name, d.Pascal)), "service_package.go")
	if err != nil {
		return 0, err
	}
	files := []genFile{
		{filepath.Join(dir, name+".go"), resourceGo},
		{filepath.Join(dir, "service_package.go"), pkgGo},
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
		td, err := buildAssocTestData(d, tv)
		if err != nil {
			return 0, err
		}
		test, err := execGoTemplate("assoctest", assocTestTmpl, td, name+"_test.go")
		if err != nil {
			return 0, err
		}
		if err := g.writeFile(filepath.Join(dir, name+"_test.go"), test, force); err != nil {
			return 0, err
		}
	} else {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no test_values entry; skipping acceptance tests\n", name)
	}
	// Some association resources keep a hand-authored data source (association is
	// resource-only); emit its companion files verbatim when registered.
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	if err := g.pruneUnwritten(dir, name); err != nil {
		return 0, err
	}
	return 1, nil
}
