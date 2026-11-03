package crud

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed parentlist.gtpl
var parentListTmpl string

//go:embed parentlisttest.gtpl
var parentListTestTmpl string

// parentListKind is the crud_archetypes.yaml kind for a parent-scoped resource
// with NO by-id GET: every op takes the parent id (from a model attr), the read
// LISTs the children under the parent and finds the one whose id matches state,
// and update/delete additionally take the record id. Create returns a
// CreatedResponse (record_id).
const parentListKind = "parent_list"

// parentListData is the template payload for the parent_list archetype. It reuses
// the entity binding types (fieldBind/sliceBind/respBind/…).
type parentListData struct {
	Pkg, Pascal, Model                              string
	ResConst, ResName, TypeName, SDKAlias           string
	SDKImport                                       string
	IDGo                                            string
	Gated                                           bool
	AtLeastOneOf                                    []string
	ParentIDGo, ParentIDTF, ParentParam, ChildParam string
	// ParentCast/ChildCast are non-empty (e.g. "uint64") when the SDK params
	// field is not int64, so the int64 model/record ids must be converted.
	ParentCast, ChildCast string

	CreateMethod, CreateBody, CreateBodyOpt string
	CreateBodyPtr                           bool
	CreateParams                            string
	CreateBinds                             []fieldBind
	CreateSliceBinds                        []sliceBind

	ReadMethod, ReadParams string
	ResponseType           string
	RecordType, RecordIDGo string
	RecordIDOpt            bool
	RespBinds              []respBind
	RespSliceBinds         []sliceRespBind
	RespObjIDProjs         []objIDProjFlat
	HasRespSlices          bool

	HasUpdate                               bool
	UpdateMethod, UpdateBody, UpdateBodyOpt string
	UpdateBodyPtr                           bool
	UpdateParams                            string
	UpdateBinds                             []fieldBind
	UpdateSliceBinds                        []sliceBind

	DeleteMethod, DeleteParams string

	// Assocs are id lists the update body does not carry, synced through
	// add/remove endpoints addressed by parent and child id (memberships.yaml).
	Assocs []*assocMembershipBind
	// CreateViaUpdate are model fields only the update body carries; when one is
	// configured, Create applies it with an update straight after the create.
	CreateViaUpdate []string

	// AttrTypes is each attribute's model type, for a generated test.
	AttrTypes map[string]string
}

// generateParentList resolves and writes a parent_list resource (resource +
// service_package, resource-only like the association/raw archetypes).
func (g *generator) generateParentList(dir, name string, ops resOps, idx sdkIndex, arch archetype, model []ModelField, tvPath string, gated, force bool) (int, error) {
	d, err := g.resolveParentList(name, ops, idx, arch, model, gated, filepath.Join(dir, name+"_schema_gen.go"))
	if err != nil {
		return 0, err
	}
	resourceGo, err := execGoTemplate("parentlist", parentListTmpl, d, name+".go")
	if err != nil {
		return 0, err
	}
	if err := g.writeFile(filepath.Join(dir, name+".go"), resourceGo, force); err != nil {
		return 0, err
	}
	// A parent_list resource derives no data source; one only exists when a
	// companion supplies it, which is what the registry answers.
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newServicePackageData(name, d.Pascal, dataSourceCompanionCtor(name, d.Pascal)), "service_package.go")
	if err != nil {
		return 0, err
	}
	if err := g.writeFile(filepath.Join(dir, "service_package.go"), pkgGo, force); err != nil {
		return 0, err
	}
	tv, hasTV, err := loadTestValues(tvPath, name)
	if err != nil {
		return 0, err
	}
	if hasTV {
		td, err := buildParentListTestData(d, tv)
		if err != nil {
			return 0, err
		}
		test, err := execGoTemplate("parentlisttest", parentListTestTmpl, td, name+"_test.go")
		if err != nil {
			return 0, err
		}
		if err := g.writeFile(filepath.Join(dir, name+"_test.go"), test, force); err != nil {
			return 0, err
		}
	} else {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no test_values entry; skipping acceptance tests\n", name)
	}
	// parent_list is resource-only; emit its hand-authored companion files.
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	if err := g.pruneUnwritten(dir, name); err != nil {
		return 0, err
	}
	return 1, nil
}

func (g *generator) resolveParentList(name string, ops resOps, idx sdkIndex, arch archetype, model []ModelField, gated bool, schemaGen string) (parentListData, error) {
	pascal := pascalCase(name)
	d := parentListData{
		Pkg: name, Pascal: pascal, Model: pascal + "Model", SDKAlias: "generated", SDKImport: sdkImport(),
		ResConst: "ResName" + pascal, ResName: pascal + " Resource", TypeName: "kion_" + name,
		Gated: gated, AtLeastOneOf: g.configValidators.For(name), RecordType: arch.RecordType, ResponseType: arch.ResponseType,
		ParentParam: arch.ParentParam, ChildParam: arch.ChildParam,
	}
	if arch.RecordType == "" || arch.ResponseType == "" || arch.ParentParam == "" || arch.ChildParam == "" {
		return d, fmt.Errorf("%s: parent_list requires record_type, response_type, parent_param, child_param", name)
	}

	byTF := map[string]ModelField{}
	d.AttrTypes = map[string]string{}
	for _, mf := range model {
		byTF[mf.TFSDK] = mf
		d.AttrTypes[mf.TFSDK] = mf.Type
		if mf.TFSDK == "id" {
			d.IDGo = mf.GoName
		}
		if mf.TFSDK == arch.ParentIDField {
			d.ParentIDGo, d.ParentIDTF = mf.GoName, mf.TFSDK
		}
	}
	if d.IDGo == "" || d.ParentIDGo == "" {
		return d, fmt.Errorf("%s: model missing id or parent %q", name, arch.ParentIDField)
	}

	// Create (POST with parent param; CreatedResponse id).
	create, err := resolveOp("create", ops.Create, idx)
	if err != nil {
		return d, err
	}
	if create == nil || create.Body == nil {
		return d, fmt.Errorf("%s: parent_list requires a create op with a body", name)
	}
	d.CreateMethod, d.CreateBody, d.CreateBodyPtr = create.Method.Name, create.Body.Name, create.Method.BodyPtr
	d.CreateBodyOpt, d.CreateParams = create.Method.BodyType, create.Method.ParamsType
	if d.CreateBinds, d.CreateSliceBinds, err = bodyBinds(create.Body, byTF, "id", nil, g.renames[name]); err != nil {
		return d, fmt.Errorf("%s create body: %w", name, err)
	}

	// Read (list). Resolve the record type's fields for flatten.
	read, err := resolveOp("read", ops.Read, idx)
	if err != nil {
		return d, err
	}
	if read == nil {
		return d, fmt.Errorf("%s: parent_list requires a read (list) op", name)
	}
	d.ReadMethod, d.ReadParams = read.Method.Name, read.Method.ParamsType
	recStruct, ok := idx.structs[arch.RecordType]
	if !ok {
		return d, fmt.Errorf("%s: record type %s not found", name, arch.RecordType)
	}
	for _, f := range recStruct.Fields {
		if f.JSONName == "id" {
			d.RecordIDGo, d.RecordIDOpt = f.GoName, f.Type == "OptUint64" || f.Type == "OptInt64"
		}
	}
	if d.RecordIDGo == "" {
		return d, fmt.Errorf("%s: record %s has no id field", name, arch.RecordType)
	}
	nested, err := resolveNestedFlatten(g.src, schemaGen, recStruct.Fields, byTF, idx, "rec.")
	if err != nil {
		return d, fmt.Errorf("%s record nested: %w", name, err)
	}
	d.RespObjIDProjs = nested.ObjIDProjs
	if d.RespBinds, d.RespSliceBinds, err = respBinds(recStruct.Fields, byTF, "id", "rec.", nested.Names, nil, false, nil); err != nil {
		return d, fmt.Errorf("%s record flatten: %w", name, err)
	}
	d.HasRespSlices = len(d.RespSliceBinds) > 0

	// Update (optional).
	update, err := resolveOp("update", ops.Update, idx)
	if err != nil {
		return d, err
	}
	if update != nil && update.Body != nil {
		d.HasUpdate = true
		d.UpdateMethod, d.UpdateBody, d.UpdateBodyPtr = update.Method.Name, update.Body.Name, update.Method.BodyPtr
		d.UpdateBodyOpt, d.UpdateParams = update.Method.BodyType, update.Method.ParamsType
		if d.UpdateBinds, d.UpdateSliceBinds, err = bodyBinds(update.Body, byTF, "id", nil, g.renames[name]); err != nil {
			return d, fmt.Errorf("%s update body: %w", name, err)
		}
	}

	// Delete.
	del, err := resolveOp("delete", ops.Delete, idx)
	if err != nil {
		return d, err
	}
	if del == nil {
		return d, fmt.Errorf("%s: parent_list requires a delete op", name)
	}
	d.DeleteMethod, d.DeleteParams = del.Method.Name, del.Method.ParamsType

	if mc, ok := g.memberships[name]; ok {
		if mc.Labels != nil || mc.Owners != nil || len(mc.SliceMembers) > 0 {
			return d, fmt.Errorf("%s: parent_list supports only associations memberships", name)
		}
		for i := range mc.Associations {
			a, aerr := resolveAssocMembership(mc.Associations[i], byTF, idx)
			if aerr != nil {
				return d, fmt.Errorf("%s associations: %w", name, aerr)
			}
			d.Assocs = append(d.Assocs, a)
		}
	}
	for _, tf := range arch.CreateViaUpdate {
		mf, ok := byTF[tf]
		if !ok {
			return d, fmt.Errorf("%s: create_via_update field %q not in model", name, tf)
		}
		if !d.HasUpdate {
			return d, fmt.Errorf("%s: create_via_update needs an update op", name)
		}
		d.CreateViaUpdate = append(d.CreateViaUpdate, mf.GoName)
	}

	// The parent/child ids come from the model (types.Int64 -> int64); cast when
	// the SDK params field is a different integer type (e.g. OU uses uint64).
	d.ParentCast = paramCast(read, arch.ParentParam)
	d.ChildCast = paramCast(del, arch.ChildParam)
	return d, nil
}

// paramCast returns a conversion prefix (e.g. "uint64") when op's params field
// named goName is not int64, or "" when it is int64 / unknown.
func paramCast(op *OpModel, goName string) string {
	if op == nil || op.Params == nil {
		return ""
	}
	for _, f := range op.Params.Fields {
		if f.GoName == goName && f.Type != "int64" && f.Type != "" {
			return f.Type
		}
	}
	return ""
}
