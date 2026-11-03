package crud

import (
	"cmp"
	_ "embed"
	"fmt"
	"os"
	"slices"

	"gopkg.in/yaml.v3"
)

//go:embed entity_compound.gtpl
var entityCompoundTmpl string

//go:embed datasource_compound.gtpl
var dataSourceCompoundTmpl string

// archetype is a compound-key / parent-read declaration from
// codegen/crud_archetypes.yaml, the shape the op-set and SDK AST cannot
// express on their own (see the file's header comment).
// archMove declares one move: raw_post is a private route taking the new value
// as its bare JSON body; account is POST /v3/account/{id}/move.
type archMove struct {
	Attr    string `yaml:"attr"`
	RawPost string `yaml:"raw_post"`
	Account bool   `yaml:"account"`
}

type archetype struct {
	Kind          string `yaml:"kind"`
	ParentIDField string `yaml:"parent_id_field"`
	ChildIDField  string `yaml:"child_id_field"`
	ChildIDParam  string `yaml:"child_id_param"`
	Collection    string `yaml:"collection"`
	RecordIDField string `yaml:"record_id_field"`
	// DeleteRetains (kind: compound_key_parent_read) marks a record the API
	// refuses to remove because its parent must always have one.
	//
	// kion_scope_criteria is the case: POST /beta/scope/{id}/criteria REPLACES
	// the scope's single criteria record rather than adding to it, and the
	// delete answers "A scope must retain at least one criteria record". So the
	// resource could be created and updated but never destroyed -- `terraform
	// destroy` failed, which means Terraform could not manage its lifecycle at
	// all.
	//
	// With this set, Delete forgets the record instead of calling the API, and
	// says so with a warning. Silently succeeding would be its own lie; the
	// warning is what tells the practitioner the criteria is still on the scope,
	// which is the only state it can be in.
	DeleteRetains bool     `yaml:"delete_retains"`
	JSONFields    []string `yaml:"json_fields"`
	// EmptyCollections are attributes whose empty value the API stores and
	// returns as null (or, for strings, as [""]), so null and empty are one
	// state on the wire. The read
	// collapses both to an empty collection; keeping them distinct makes a
	// configured `[]` diff for ever, and fails the apply outright when the
	// attribute is Required or explicitly set.
	EmptyCollections []string `yaml:"empty_collections"`
	// ListOrderKeys maps a nested-object list to the element attribute that
	// identifies an element. The API returns such a list in an order of its
	// own, so the read puts its elements back in the prior state's order,
	// matched on that attribute, rather than failing the apply.
	ListOrderKeys map[string]string `yaml:"list_order_keys"`
	// Asymmetric 2-param delete (kind: entity). See ResourceModel.DeleteRecordParam.
	DeleteRecordParam string `yaml:"delete_record_param"`
	DeleteExtraParam  string `yaml:"delete_extra_param"`
	DeleteExtraField  string `yaml:"delete_extra_field"`
	// Create-under-parent (kind: entity): the create op takes a parent-id path
	// param sourced from a model attribute, while read/update/delete address the
	// record by its own id (e.g. idms open-id access-rule / group-association).
	CreateParentParam string `yaml:"create_parent_param"` // SDK create-params field naming the parent id
	CreateParentField string `yaml:"create_parent_field"` // model attribute holding the parent id
	// ConfirmListed (kind: entity) is a parent collection path, "{id}" being
	// create_parent_field's value. A record the by-id read returns but this list
	// omits is treated as deleted: for by-id reads that serve soft-deleted rows.
	ConfirmListed string `yaml:"confirm_listed"`
	// ConfirmParents lists every possible parent. On import the parent id is
	// unknown, so the read finds the one whose confirm_listed list holds the
	// record and sets create_parent_field from it.
	ConfirmParents string `yaml:"confirm_parents"`
	// MissingViaList (kind: entity) is a list path consulted when the by-id read
	// answers a 500: an id the list omits is treated as deleted. For reads that
	// answer a missing record with a 500 rather than a 404.
	MissingViaList string `yaml:"missing_via_list"`
	// CreateViaUpdate (kind: parent_list) lists attributes only the update body
	// carries; a configured value is applied by an update right after create.
	CreateViaUpdate []string `yaml:"create_via_update"`
	// Moves (kind: entity) are attributes the update body cannot carry, changed
	// through a dedicated endpoint before the PATCH.
	Moves []archMove `yaml:"moves"`
	// Association (kind: association), one list-member row keyed by KeyField,
	// scoped by an optional ParentField, over a bulk read + replace-list API.
	KeyField     string   `yaml:"key_field"`
	ParentField  string   `yaml:"parent_field"`
	MemberFields []string `yaml:"member_fields"`
	// Parent-list (kind: parent_list), a parent-scoped resource with no by-id
	// GET; the read lists children under the parent and finds by id.
	ParentParam  string `yaml:"parent_param"`  // SDK param naming the parent id (all ops)
	ChildParam   string `yaml:"child_param"`   // SDK param naming the record id (update/delete)
	RecordType   string `yaml:"record_type"`   // the list element struct, e.g. ProjectEnforcement
	ResponseType string `yaml:"response_type"` // the read response struct, e.g. ProjectEnforcementResponse
	// Collapse (kind: entity): a scalar the write expands into an array the read
	// returns in its place, rebuilt by summing one field of that array. See
	// sumFlat; kion_budget's amount is the case it exists for.
	SumFrom []sumFromDecl `yaml:"sum_from"`
	// Read companion (kind: entity): a hand-authored function, registered in
	// companionsByName and emitted beside the resource, called after every
	// flatten to fill in attributes the by-id read does not return at all
	// because they live behind other endpoints (kion_funding_source's owners and
	// permission scheme). Which endpoints those are, and how to read an
	// attribute back out of them, is knowledge no spec carries.
	ReadCompanion string `yaml:"read_companion"`
	// RawCreate (kind: entity) routes ONLY the create over raw HTTP, leaving
	// read, update and delete on the SDK.
	//
	// It exists for a create whose typed client cannot be used while the rest of
	// the resource is perfectly ordinary. Two causes so far, both spec defects
	// rather than private endpoints: a declared status the server does not send
	// (POST /v3/project-line-item answers 201, the spec declares only 200, so
	// ogen refuses to decode a SUCCESSFUL create and the new record's id is
	// lost), and a response shape that does not match what the server sends.
	//
	// The body is built from the model's scalar attributes, keyed by their
	// tfsdk names, and the id is read from the {"record_id": N} envelope.
	RawCreate *rawOp `yaml:"raw_create"`
	// Rewritten (kind: entity) names attributes the API stores in a canonical
	// form of its own and echoes back rewritten, so the read-back after a write
	// does not textually match what the practitioner configured. Terraform
	// compares the post-apply value to the plan verbatim and rejects the
	// mismatch outright:
	//
	//	Provider produced inconsistent result after apply
	//	.role_permissions: was {"actions":[...],"notActions":[]}, but now
	//	{"actions":[...],"dataActions":[],"notActions":[],"notDataActions":[]}
	//
	// and on a later plan the same difference is a diff that can never
	// converge. Listed attributes keep the configured value through Create and
	// Update; Read still flattens them, so drift on an unmanaged change is
	// still detected. This is the labels restore generalized -- same shape, but
	// driven by the archetype rather than hardcoded.
	//
	// Use it only for an attribute whose rewrite is a canonicalization, never
	// to paper over one the provider sends wrong.
	Rewritten []string `yaml:"rewritten"`
	// CreateOnly (kind: entity) names string attributes only the create
	// response carries (an OAuth client's secret). Create captures each from
	// that response; flatten never assigns it, so Read and Update keep state's.
	CreateOnly []string `yaml:"create_only"`
	// Sweeper (kind: entity): the data-source collection is parent-scoped, so it
	// takes a required parent-id param the data source cannot supply. A sweeper
	// can: it enumerates SweepParent's own collection first. Without these the
	// resource gets no sweeper at all.
	SweepParent      string `yaml:"sweep_parent"`       // parent resource name, e.g. "ou"
	SweepParentParam string `yaml:"sweep_parent_param"` // list param taking the parent id, e.g. "ID"
	// SweepParentMatch keeps only parents whose integer field equals the value,
	// e.g. {idms_type_id: "3"}: a child list that fails on other parents.
	SweepParentMatch map[string]string `yaml:"sweep_parent_match"`
	// ResourceTemplate (kind: entity) names a verbatim template in
	// resourceTemplates that replaces the derived <name>.go. Everything else the
	// entity path produces -- data source, sweeper, acceptance tests -- stays
	// derived, and so does the schema. It exists for a resource whose CRUD
	// control flow the entity archetype cannot express while the rest of it is
	// perfectly ordinary; project, which dispatches create on the install's
	// financial mode, is the only one.
	ResourceTemplate string `yaml:"resource_template"`
}

// sumFromDecl declares one rebuilt-by-summing model attribute: which attribute,
// which read-payload array, and which element field adds up to it.
type sumFromDecl struct {
	TF    string `yaml:"tf"`    // model attribute, "amount"
	From  string `yaml:"from"`  // read payload array field, "data"
	Field string `yaml:"field"` // element field to sum, "amount"
}

const (
	compoundKindParentRead = "compound_key_parent_read"
	// entityKind routes to the normal entity path but applies archetype tweaks
	// (currently an asymmetric 2-param delete).
	entityKind = "entity"
)

// loadArchetypes reads codegen/crud_archetypes.yaml. A missing file yields an
// empty map (no resource uses the compound archetype).
func loadArchetypes(path string) (map[string]archetype, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]archetype{}, nil
		}
		return nil, fmt.Errorf("reading archetypes %s: %w", path, err)
	}
	var m map[string]archetype
	if err := yaml.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parsing archetypes %s: %w", path, err)
	}
	return m, nil
}

// bind is one TF-model <-> SDK-field binding rendered as a full Go expression.
type bind struct {
	SDKField string // SDK struct field, e.g. "StartMonth" (body binds)
	ModelGo  string // model field, e.g. "StartMonth" (flatten binds)
	Expr     string // the complete RHS Go expression
}

// dsCompoundAttr is one data-source schema attribute. The key fields
// (parent/child id) are Required inputs; the rest are Computed.
type dsCompoundAttr struct {
	TFName     string
	SchemaType string // "Int64Attribute" | "StringAttribute" | "BoolAttribute"
	Required   bool
	CustomType string // "jsontypes.NormalizedType{}" for json fields, else ""
}

// compoundData is the template payload for a compound-key/parent-read resource.
type compoundData struct {
	Pkg, Pascal, Model string
	SDKAlias           string
	SDKImport          string
	ResConst, ResName  string
	DSConst, DSName    string
	Gated              bool
	DSGated            bool   // data source Read checks the version gate
	TypeName           string // "kion_scope_criteria"

	IDGo string // composite-id model field ("Id")
	// DeleteRetains: Delete forgets the record instead of calling the API,
	// because the API refuses to remove it. See the archetype field.
	DeleteRetains bool
	ParentIDGo    string // "ScopeId"
	ParentIDTF    string // "scope_id"
	ChildIDGo     string // "CriteriaId"
	ChildIDTF     string // "criteria_id"

	// Parent read (extract child from a collection).
	ReadMethod    string // "GetScopeByID"
	ReadParams    string // "GetScopeByIDParams"
	ReadParentArg string // "ID"
	RespType      string // "ScopeResponse"
	RespDataPtr   bool   // envelope Data is *Payload (false = Opt-wrapper)
	Collection    string // "CriteriaRecords"
	CollectionNil bool   // collection is an OptNil<T>Array wrapper (.Value)
	RecordType    string // "ScopeCriteriaRecord"
	RecordIDGo    string // "ID"

	// Create.
	CreateMethod    string
	CreateBody      string // "ScopeCriteriaCreate"
	CreateBodyPtr   bool
	CreateParams    string
	CreateParentArg string
	CreateBinds     []bind

	// Update (optional).
	HasUpdate       bool
	UpdateMethod    string
	UpdateBody      string
	UpdateBodyPtr   bool
	UpdateParams    string
	UpdateParentArg string
	UpdateChildArg  string
	UpdateBinds     []bind

	// Delete.
	DeleteMethod    string
	DeleteParams    string
	DeleteParentArg string
	DeleteChildArg  string

	// Flatten (record -> model).
	Flattens []bind

	// Data-source schema attributes (key fields Required, rest Computed).
	DSAttrs []dsCompoundAttr

	NeedsJx bool // a json_field is present (import go-faster/jx)
}

// resolveCompound assembles a compoundData from the op-set, the SDK index, the
// archetype declaration, and the generated model. It refuses (error) any shape
// it can't map so the caller logs a skip rather than emit broken code.
func resolveCompound(name string, ops resOps, idx sdkIndex, arch archetype, model []ModelField, gated bool) (compoundData, error) {
	pascal := pascalCase(name)
	d := compoundData{
		Pkg:           name,
		Pascal:        pascal,
		Model:         pascal + "Model",
		SDKAlias:      "generated",
		SDKImport:     sdkImport(),
		ResConst:      "ResName" + pascal,
		ResName:       pascal + " Resource",
		DeleteRetains: arch.DeleteRetains,
		DSConst:       "DSName" + pascal,
		DSName:        pascal + " Data Source",
		Gated:         gated,
		TypeName:      "kion_" + name,
	}

	// Model fields: locate the composite id, parent id, and child id; the rest
	// are body/flatten payload fields.
	byTF := map[string]ModelField{}
	for _, mf := range model {
		byTF[mf.TFSDK] = mf
		switch mf.TFSDK {
		case "id":
			d.IDGo = mf.GoName
		case arch.ParentIDField:
			d.ParentIDGo, d.ParentIDTF = mf.GoName, mf.TFSDK
		case arch.ChildIDField:
			d.ChildIDGo, d.ChildIDTF = mf.GoName, mf.TFSDK
		}
	}
	if d.IDGo == "" || d.ParentIDGo == "" || d.ChildIDGo == "" {
		return d, fmt.Errorf("%s: model missing id/%s/%s", name, arch.ParentIDField, arch.ChildIDField)
	}
	jsonField := map[string]bool{}
	for _, jf := range arch.JSONFields {
		jsonField[jf] = true
		d.NeedsJx = true
	}

	// Parent read op + envelope (Scope). resolveOp("read") gives the payload
	// (Scope) fields; the archetype names the child collection within them.
	read, err := resolveOp("read", ops.Read, idx)
	if err != nil {
		return d, err
	}
	if read == nil {
		return d, fmt.Errorf("%s: compound archetype requires a read (parent) op", name)
	}
	d.ReadMethod = read.Method.Name
	d.ReadParams = read.Method.ParamsType
	d.RespType = read.RespType
	d.RespDataPtr = read.RespDataPtr
	parentArg, _, err := mapParams(read.Params, arch.ChildIDParam)
	if err != nil {
		return d, fmt.Errorf("%s read params: %w", name, err)
	}
	d.ReadParentArg = parentArg

	// The collection field on the parent payload, and its element (record) type.
	var collType string
	for _, f := range read.RespFields {
		if f.GoName == arch.Collection {
			collType = f.Type
		}
	}
	if collType == "" {
		return d, fmt.Errorf("%s: parent payload %s has no %s field", name, read.RespPayload, arch.Collection)
	}
	elem, nilAware, ok := wrapperSliceElem(collType, idx)
	if !ok {
		return d, fmt.Errorf("%s: collection %s (%s) is not a slice", name, arch.Collection, collType)
	}
	d.Collection, d.CollectionNil, d.RecordType, d.RecordIDGo = arch.Collection, nilAware, elem, arch.RecordIDField

	recStruct, ok := idx.structs[elem]
	if !ok {
		return d, fmt.Errorf("%s: record type %s not found", name, elem)
	}

	// Create.
	create, err := resolveOp("create", ops.Create, idx)
	if err != nil {
		return d, err
	}
	if create == nil || create.Body == nil {
		return d, fmt.Errorf("%s: compound archetype requires a create op with a body", name)
	}
	d.CreateMethod = create.Method.Name
	d.CreateBody = create.Method.BodyType
	d.CreateBodyPtr = create.Method.BodyPtr
	d.CreateParams = create.Method.ParamsType
	if d.CreateParentArg, _, err = mapParams(create.Params, arch.ChildIDParam); err != nil {
		return d, fmt.Errorf("%s create params: %w", name, err)
	}
	if d.CreateBinds, err = bodyBindsCompound(create.Body, byTF, jsonField); err != nil {
		return d, fmt.Errorf("%s create body: %w", name, err)
	}

	// Update (optional).
	update, err := resolveOp("update", ops.Update, idx)
	if err != nil {
		return d, err
	}
	if update != nil && update.Body != nil {
		d.HasUpdate = true
		d.UpdateMethod = update.Method.Name
		d.UpdateBody = update.Method.BodyType
		d.UpdateBodyPtr = update.Method.BodyPtr
		d.UpdateParams = update.Method.ParamsType
		if d.UpdateParentArg, d.UpdateChildArg, err = mapParams(update.Params, arch.ChildIDParam); err != nil {
			return d, fmt.Errorf("%s update params: %w", name, err)
		}
		if d.UpdateChildArg == "" {
			return d, fmt.Errorf("%s update params %s lack the child id %q", name, d.UpdateParams, arch.ChildIDParam)
		}
		if d.UpdateBinds, err = bodyBindsCompound(update.Body, byTF, jsonField); err != nil {
			return d, fmt.Errorf("%s update body: %w", name, err)
		}
	}

	// Delete. A delete_retains resource has no delete endpoint by declaration:
	// the API keeps the record and the template's Delete forgets it with a
	// warning, so requiring an op here contradicts the archetype and skipped the
	// whole resource.
	del, err := resolveOp("delete", ops.Delete, idx)
	if err != nil && !arch.DeleteRetains {
		return d, err
	}
	switch {
	case arch.DeleteRetains:
	case del == nil:
		return d, fmt.Errorf("%s: compound archetype requires a delete op unless delete_retains is set", name)
	default:
		d.DeleteMethod = del.Method.Name
		d.DeleteParams = del.Method.ParamsType
		if d.DeleteParentArg, d.DeleteChildArg, err = mapParams(del.Params, arch.ChildIDParam); err != nil {
			return d, fmt.Errorf("%s delete params: %w", name, err)
		}
		if d.DeleteChildArg == "" {
			return d, fmt.Errorf("%s delete params %s lack the child id %q", name, d.DeleteParams, arch.ChildIDParam)
		}
	}

	// Flatten binds (record -> model): the child id comes from the record's
	// id field; every other non-key model field maps by json/tfsdk name.
	recByJSON := map[string]Field{}
	for _, f := range recStruct.Fields {
		recByJSON[f.JSONName] = f
	}
	for _, mf := range model {
		switch mf.TFSDK {
		case "id", arch.ParentIDField:
			continue
		case arch.ChildIDField:
			rf, ok := fieldByGo(recStruct, arch.RecordIDField)
			if !ok {
				return d, fmt.Errorf("%s: record %s has no id field %s", name, elem, arch.RecordIDField)
			}
			conv, ok := flattenConverter(rf)
			if !ok {
				return d, fmt.Errorf("%s: record id field type %q unsupported", name, rf.Type)
			}
			d.Flattens = append(d.Flattens, bind{ModelGo: mf.GoName, Expr: conv + "(rec." + rf.GoName + ")"})
			continue
		}
		rf, ok := recByJSON[mf.TFSDK]
		if !ok {
			return d, fmt.Errorf("%s: record %s has no field for model %q", name, elem, mf.TFSDK)
		}
		if jsonField[mf.TFSDK] {
			// flex.NormalizedToFramework, not jsontypes.NewNormalizedValue:
			// the value has to be canonicalised on the way in. Semantic
			// equality is not consulted when a generated configuration is
			// planned, so the API's own spacing compared against the compact,
			// sorted JSON `jsonencode({…})` produces reported a change on
			// every record carrying JSON -- 9 kion_scope_criteria on one
			// install, updating nothing. See flex.canonicalJSON, which the
			// entity path already goes through.
			d.Flattens = append(d.Flattens, bind{ModelGo: mf.GoName, Expr: "flex.NormalizedToFramework(rec." + rf.GoName + ")"})
			continue
		}
		conv, ok := flattenConverter(rf)
		if !ok {
			return d, fmt.Errorf("%s: record field %q type %q unsupported", name, mf.TFSDK, rf.Type)
		}
		d.Flattens = append(d.Flattens, bind{ModelGo: mf.GoName, Expr: conv + "(rec." + rf.GoName + ")"})
	}

	// Data-source schema attributes: the parent + child ids are Required inputs;
	// every other model field is Computed.
	for _, mf := range model {
		st, ct, ok := dsSchemaType(mf.Type)
		if !ok {
			return d, fmt.Errorf("%s: model field %q type %q unsupported for data source", name, mf.TFSDK, mf.Type)
		}
		d.DSAttrs = append(d.DSAttrs, dsCompoundAttr{
			TFName:     mf.TFSDK,
			SchemaType: st,
			CustomType: ct,
			Required:   mf.TFSDK == arch.ParentIDField || mf.TFSDK == arch.ChildIDField,
		})
	}
	slices.SortFunc(d.DSAttrs, func(a, b dsCompoundAttr) int { return cmp.Compare(a.TFName, b.TFName) })

	return d, nil
}

// dsSchemaType maps a framework model type to its data-source schema.Attribute
// kind and (for a framework custom type) the CustomType expression.
func dsSchemaType(modelType string) (schemaType, customType string, ok bool) {
	switch modelType {
	case "types.Int64":
		return "Int64Attribute", "", true
	case "types.String":
		return "StringAttribute", "", true
	case "types.Bool":
		return "BoolAttribute", "", true
	case "jsontypes.Normalized":
		return "StringAttribute", "jsontypes.NormalizedType{}", true
	}
	return "", "", false
}

// bodyBindsCompound maps a request body's fields to model fields by json/tfsdk
// name. json_fields become jx.Raw(plan.X.ValueString()); scalars use the
// flex expand converters.
func bodyBindsCompound(body *Struct, byTF map[string]ModelField, jsonField map[string]bool) ([]bind, error) {
	var binds []bind
	for _, f := range body.Fields {
		mf, ok := byTF[f.JSONName]
		if !ok {
			continue // body field with no model attribute (e.g. server-set), skip
		}
		if jsonField[f.JSONName] {
			binds = append(binds, bind{SDKField: f.GoName, Expr: "jx.Raw(plan." + mf.GoName + ".ValueString())"})
			continue
		}
		conv, ok := expandConverter(f)
		if !ok {
			return nil, fmt.Errorf("body field %q type %q unsupported", f.JSONName, f.Type)
		}
		binds = append(binds, bind{SDKField: f.GoName, Expr: conv + "(plan." + mf.GoName + ")"})
	}
	return binds, nil
}

// mapParams splits a params struct into its parent-id arg and (optional)
// child-id arg. The child arg is the field named childParam; the parent arg is
// the single remaining field.
func mapParams(p *Struct, childParam string) (parent, child string, err error) {
	if p == nil {
		return "", "", fmt.Errorf("op has no params struct")
	}
	for _, f := range p.Fields {
		if f.GoName == childParam {
			child = f.GoName
			continue
		}
		if parent != "" {
			return "", "", fmt.Errorf("params %s has multiple non-child id fields (%s, %s)", p.Name, parent, f.GoName)
		}
		parent = f.GoName
	}
	if parent == "" {
		return "", "", fmt.Errorf("params %s has no parent id field", p.Name)
	}
	return parent, child, nil
}

func fieldByGo(s Struct, goName string) (Field, bool) {
	for _, f := range s.Fields {
		if f.GoName == goName {
			return f, true
		}
	}
	return Field{}, false
}

func renderCompoundEntity(d compoundData) ([]byte, error) {
	return execGoTemplate("entity_compound", entityCompoundTmpl, d, d.Pkg+".go")
}

func renderCompoundDataSource(d compoundData) ([]byte, error) {
	return execGoTemplate("datasource_compound", dataSourceCompoundTmpl, d, d.Pkg+"_data_source.go")
}
