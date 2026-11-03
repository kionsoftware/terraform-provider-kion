package crud

import (
	"cmp"
	"fmt"
	"strings"
)

// objSub is one sub-attribute binding inside a nested object: the SDK struct
// field and the full expand expression for it.
type objSub struct {
	SDKField string // "Name"
	Expr     string // "flex.OptStringFromFramework(<prefix>.Description)"
}

// objBind is a single nested-object body field (e.g. azure_policy). The template
// emits a prelude building the SDK struct into Var, then references it in the body.
type objBind struct {
	SDKField string // body field GoName, "AzurePolicy"
	Var      string // local var, "azurePolicy"
	SDKType  string // "AzurePolicyDefinitionCreate"
	OptType  string // "OptAzurePolicyDefinitionCreate" ("" if the body field is *T / bare)
	Ptr      bool   // body field is *SDKType
	ModelGo  string // model field GoName, "AzureConnection"
	Subs     []objSub

	// Guard sends this object only when its model attribute is configured.
	//
	// Set when a request body carries more than one optional nested object,
	// which in Kion's API means they are alternatives rather than companions:
	// a custom billing source takes aws_connection OR azure_connection, and
	// sending both is a 400. Emitting both unconditionally,
	// the behavior before this existed. Makes every create fail.
	//
	// Bodies with a single nested object are unaffected, so this changes
	// nothing for the resources that predate it.
	Guard bool
}

// arrBind is a nested object-array body field (e.g. roles). The prelude iterates
// the model list, building []SDKElem into Var.
type arrBind struct {
	SDKField  string // "Roles"
	Var       string // "roles"
	ModelGo   string // "Roles"
	ElemType  string // "AppRolePermission"
	ValueType string // "RolesValue" (the tfplugingen list-element Value type)
	Wrap      string // "OptNilAppRolePermissionArray" ("" for a bare []T)
	Subs      []objSub
}

// implodeBind regroups a flat model list into the grouped array a write body
// takes: the inverse of a read_shape explode. The read declares that the API
// returns {permission_id, role_ids[]} and the model stores one row per
// (permission_id, role_id) pair; writing has to put the pairs back together.
//
// Without this the body field matched no model attribute, so it was skipped in
// silence -- which, for permission_scheme, left the entire update body empty.
type implodeBind struct {
	SDKField  string // "PermissionRoles"
	Var       string // "permissionRoles"
	ModelGo   string // model list attr, "Roles"
	ElemType  string // "AppRoleArrayPermission"
	ValueType string // "RolesValue"
	Wrap      string // "OptNilAppRoleArrayPermissionArray" ("" for a bare []T)

	KeySDK   string // element key field, "PermissionID"
	KeyValue string // Value-type sub-field holding the key, "PermissionId"
	KeyConv  string // key expression built from the loop var k, e.g. "flex.NilUint64FromFramework(types.Int64Value(k))"

	MemberSDK   string // element slice field, "RoleIds"
	MemberValue string // Value-type sub-field holding one member, "RoleId"
	MemberGo    string // slice element Go type, "uint64"
}

// nestedFieldSet is the set of body json-names handled as nested (so bodyBinds
// skips them).
type nestedResult struct {
	Objs      []objBind
	Arrs      []arrBind
	Implode   *implodeBind
	RawValues []rawValueBind
	// FlatSubs are body scalar fields with no top-level model attribute that map
	// to a sub-field of a nested model Value type, the inverse of Objs. This
	// covers a flat update body (e.g. AzurePolicyDefinitionUpdate{description,
	// name, ...}) against a model that nests those fields under one object
	// attribute (azure_policy).
	FlatSubs []objSub
	Names    map[string]bool
}

// subFieldSource identifies where a nested model sub-attribute lives: the model
// object attribute holding it and the sub-field's Go name and Go type.
type subFieldSource struct {
	NestedGo string // "AzurePolicy" (the model object attribute Go name)
	SubGo    string // "Description" (the sub-field Go name in the Value type)
}

// nestedSubIndex maps a nested sub-attribute tfsdk name to its source, for every
// model attribute whose type is a tfplugingen Value type. Used to source a flat
// body field from a nested model sub-attribute.
func nestedSubIndex(src Source, schemaGen string, byTF map[string]ModelField) map[string]subFieldSource {
	idx := map[string]subFieldSource{}
	for _, mf := range byTF {
		if !strings.HasSuffix(mf.Type, "Value") || !valueTypeExists(src, schemaGen, mf.Type) {
			continue
		}
		subs, err := src.ModelFields(schemaGen, mf.Type)
		if err != nil {
			continue
		}
		for _, sf := range subs {
			if sf.TFSDK == "" {
				continue // the unexported state field
			}
			// First writer wins; ambiguous sub-names across two nested objects
			// are rare and would need an override.
			if _, exists := idx[sf.TFSDK]; !exists {
				idx[sf.TFSDK] = subFieldSource{NestedGo: mf.GoName, SubGo: sf.GoName}
			}
		}
	}
	return idx
}

// recordSubFields returns the model sub-attributes of the record-wrapper object
// attribute named by wrapperGo, promoted to top-level ModelFields.
//
// It is non-empty only for a model that nests the WHOLE record under one object
// attribute (azure_policy) instead of flattening it. Such a model exposes none
// of the record's fields at the top level, so the data source and sweeper,
// which project the model's attributes onto the record. Would otherwise see an
// empty attribute set and produce a `list` carrying nothing but the id. Promoting
// them mirrors the SDKv2 provider, whose `list` rows were flat.
//
// Returns nil for the common flat-model case (the wrapper is not a model
// attribute at all) and for a wrapper with no tfplugingen Value type.
func recordSubFields(src Source, schemaGen, wrapperGo string, byTF map[string]ModelField) []ModelField {
	if wrapperGo == "" {
		return nil
	}
	var wrapper ModelField
	for _, mf := range byTF {
		if mf.GoName == wrapperGo {
			wrapper = mf
			break
		}
	}
	if wrapper.TFSDK == "" || !strings.HasSuffix(wrapper.Type, "Value") {
		return nil
	}
	subs, err := src.ModelFields(schemaGen, wrapper.Type)
	if err != nil {
		return nil
	}
	var out []ModelField
	for _, sf := range subs {
		if sf.TFSDK == "" {
			continue // the unexported state field
		}
		if _, clash := byTF[sf.TFSDK]; clash {
			continue // a top-level attribute of the same name wins
		}
		out = append(out, ModelField{GoName: sf.GoName, TFSDK: sf.TFSDK, Type: frameworkModelType(sf.Type)})
	}
	return out
}

// frameworkModelType normalizes a tfplugingen Value sub-field's type spelling
// (basetypes.StringValue) to the types.* alias the schema/attr-type tables key
// on. Identity for anything already in that form.
func frameworkModelType(t string) string {
	switch t {
	case "basetypes.StringValue":
		return "types.String"
	case "basetypes.Int64Value":
		return "types.Int64"
	case "basetypes.BoolValue":
		return "types.Bool"
	case "basetypes.Float64Value":
		return "types.Float64"
	}
	return t
}

// valueTypeExists reports whether the schema_gen file declares the named
// tfplugingen Value type. It is the discriminator between a genuinely nested
// tfplugingen attribute and a plain scalar collection: a body/response field
// whose SDK type is a struct(-array) but whose model attribute is a bare
// types.List (e.g. an array of record ids) has no <Pascal>Value type and must
// not be treated as nested.
func valueTypeExists(src Source, schemaGen, valueType string) bool {
	_, err := src.ModelFields(schemaGen, valueType)
	return err == nil
}

// nestedOpts carries what the nested resolver needs beyond the body itself.
type nestedOpts struct {
	// NoGuard disables the "alternatives, not companions" guard on a body with
	// more than one optional nested object.
	NoGuard bool
	// IDAttr is the model's id attribute tfsdk name, and IDVar the local holding
	// the parsed record id. Together they source an `id` sub-field of a wrapped
	// object. Empty on create, where the record has no id yet.
	IDAttr string
	IDVar  string
	// IDType is the record id's Go type ("int64", "uint64" or "string"); ""
	// is int64.
	IDType string
	// Implode inverts a declared read_shape explode; nil where none applies.
	Implode *readShapeExplode
}

// resolveNested finds nested-object and object-array fields in a request body
// and builds their expand binds, reading the model's tfplugingen Value type
// sub-fields (via src) and the SDK struct sub-fields (via idx).
func resolveNested(src Source, schemaGen string, body *Struct, byTF map[string]ModelField, idx sdkIndex, opts nestedOpts) (nestedResult, error) {
	res := nestedResult{Names: map[string]bool{}}
	if body == nil {
		return res, nil
	}
	res.RawValues = resolveRawValues(body, byTF)
	for _, rv := range res.RawValues {
		res.Names[rv.Attr] = true
	}
	subIdx := nestedSubIndex(src, schemaGen, byTF)
	for _, f := range body.Fields {
		if opts.Implode != nil && f.JSONName == opts.Implode.From {
			ib, err := resolveImplode(src, schemaGen, f, byTF, idx, *opts.Implode)
			if err != nil {
				return res, fmt.Errorf("implode field %q: %w", f.JSONName, err)
			}
			res.Implode = ib
			res.Names[f.JSONName] = true
			continue
		}
		mf, ok := byTF[f.JSONName]
		if !ok {
			// Not a top-level model attribute. If it names a sub-field of a
			// nested model object, source it from there (flat body over a nested
			// model, e.g. a flat update definition against a wrapping create).
			if s, isSub := subIdx[f.JSONName]; isSub {
				if conv, okConv := expandConverter(f); okConv {
					res.FlatSubs = append(res.FlatSubs, objSub{
						SDKField: f.GoName,
						Expr:     conv + "(plan." + s.NestedGo + "." + s.SubGo + ")",
					})
					res.Names[f.JSONName] = true
				}
				continue
			}
			// Or it is a nested object whose OWN sub-fields are the model's
			// top-level attributes: a nested body over a flat model, the mirror
			// of FlatSubs. PermissionSchemeUpdateRequest.app_policy wraps id,
			// name and type this way, and matching nothing left the body empty.
			ob, isWrap, err := wrapObject(src, schemaGen, f, byTF, idx, opts)
			if err != nil {
				return res, fmt.Errorf("object field %q: %w", f.JSONName, err)
			}
			if isWrap {
				res.Objs = append(res.Objs, ob)
				res.Names[f.JSONName] = true
			}
			continue
		}
		// Object array: OptNil<T>Array / []T whose element is a struct.
		if elem, _, isSlice := wrapperSliceElem(f.Type, idx); isSlice {
			es, isStruct := idx.structs[elem]
			if !isStruct || len(es.Fields) == 0 {
				continue // scalar slice, handled by sliceConverter
			}
			valueType := pascalCase(f.JSONName) + "Value"
			if !valueTypeExists(src, schemaGen, valueType) {
				continue // modeled as a bare list (e.g. array of ids), not nested
			}
			subs, err := nestedSubs(src, schemaGen, valueType, es, "elem")
			if err != nil {
				return res, fmt.Errorf("array field %q: %w", f.JSONName, err)
			}
			wrap := ""
			if strings.HasPrefix(f.Type, "Opt") {
				wrap = f.Type
			}
			res.Arrs = append(res.Arrs, arrBind{
				SDKField: f.GoName, Var: lowerFirst(f.GoName), ModelGo: mf.GoName,
				ElemType: elem, ValueType: valueType, Wrap: wrap, Subs: subs,
			})
			res.Names[f.JSONName] = true
			continue
		}
		// Single nested object: Opt<T> / *T / T whose base is a struct.
		base := strings.TrimPrefix(f.Type, "Opt")
		bs, isStruct := idx.structs[base]
		if !isStruct || len(bs.Fields) == 0 {
			continue
		}
		valueType := pascalCase(f.JSONName) + "Value"
		if !valueTypeExists(src, schemaGen, valueType) {
			continue // not a tfplugingen nested attribute
		}
		subs, err := nestedSubs(src, schemaGen, valueType, bs, "plan."+mf.GoName)
		if err != nil {
			return res, fmt.Errorf("object field %q: %w", f.JSONName, err)
		}
		optType := ""
		if strings.HasPrefix(f.Type, "Opt") {
			optType = f.Type
		}
		res.Objs = append(res.Objs, objBind{
			SDKField: f.GoName, Var: lowerFirst(f.GoName), SDKType: base,
			OptType: optType, Ptr: f.Ptr, ModelGo: mf.GoName, Subs: subs,
		})
		res.Names[f.JSONName] = true
	}
	// Multiple optional nested objects in one body are alternatives, not
	// companions, send only the configured one. See objBind.Guard.
	if len(res.Objs) > 1 && !opts.NoGuard {
		for i := range res.Objs {
			if res.Objs[i].OptType != "" {
				res.Objs[i].Guard = true
			}
		}
	}
	return res, nil
}

// wrapObject builds the bind for a body field that is a struct the model does
// not mirror, whose sub-fields ARE the model's top-level attributes.
//
// It reports false for anything else, including a struct the model nests under
// an attribute of the same name (resolveNested's ordinary object path owns
// that) and a struct none of whose sub-fields the model names -- a server-owned
// object the resource has no business filling in.
func wrapObject(src Source, schemaGen string, f Field, byTF map[string]ModelField, idx sdkIndex, opts nestedOpts) (objBind, bool, error) {
	base := trimOptWrappers(f.Type)
	bs, isStruct := idx.structs[base]
	if !isStruct || len(bs.Fields) == 0 {
		return objBind{}, false, nil
	}
	if valueTypeExists(src, schemaGen, pascalCase(f.JSONName)+"Value") {
		return objBind{}, false, nil // a tfplugingen nested attribute whose model attr is merely missing
	}

	var subs []objSub
	var idErr error
	modelSourced := 0
	for _, sf := range bs.Fields {
		if opts.IDAttr != "" && sf.JSONName == opts.IDAttr {
			expr, ok, err := idSubExpr(sf.Type, opts.IDVar, opts.IDType)
			if ok {
				subs = append(subs, objSub{SDKField: sf.GoName, Expr: expr})
			}
			idErr = err
			continue
		}
		mf, ok := byTF[sf.JSONName]
		if !ok {
			continue // left zero: the model does not carry it
		}
		conv, ok := expandConverter(sf)
		if !ok {
			continue // no one-level converter; leaving it zero beats refusing the resource
		}
		subs = append(subs, objSub{SDKField: sf.GoName, Expr: conv + "(plan." + mf.GoName + ")"})
		modelSourced++
	}
	if modelSourced == 0 {
		return objBind{}, false, nil
	}
	// The object is sent, so an id sub-field it cannot source would go out zero.
	if idErr != nil {
		return objBind{}, false, idErr
	}

	optType := ""
	if strings.HasPrefix(f.Type, "Opt") {
		optType = f.Type
	}
	return objBind{
		SDKField: f.GoName, Var: lowerFirst(f.GoName), SDKType: base,
		OptType: optType, Ptr: f.Ptr, Subs: subs,
	}, true, nil
}

// idSubExpr sources a wrapped object's id sub-field from the already-parsed
// record id. The conversion is written out unconditionally because the local's
// own type follows the op's params (int64 or uint64) and either converts.
// It is not ok, with no error, on create (empty idVar), where there is no id yet.
func idSubExpr(sdkType, idVar, idType string) (string, bool, error) {
	if idVar == "" {
		return "", false, nil
	}
	if idType == "string" {
		switch sdkType {
		case "string":
			return idVar, true, nil
		case "OptString":
			return "generated.OptString{Value: " + idVar + ", Set: true}", true, nil
		}
	} else {
		switch sdkType {
		case "uint64", "int64":
			return sdkType + "(" + idVar + ")", true, nil
		case "OptUint64":
			return "generated.OptUint64{Value: uint64(" + idVar + "), Set: true}", true, nil
		case "OptInt64":
			return "generated.OptInt64{Value: int64(" + idVar + "), Set: true}", true, nil
		}
	}
	return "", false, fmt.Errorf("id sub-field type %s cannot be sourced from a %s record id", sdkType, cmp.Or(idType, "int64"))
}

// trimOptWrappers strips the ogen optional/nullable prefixes from a type name.
func trimOptWrappers(t string) string {
	t = strings.TrimPrefix(t, "*")
	for _, p := range []string{"OptNilPointer", "OptPointer", "OptNil", "Opt", "Nil"} {
		if strings.HasPrefix(t, p) && len(t) > len(p) {
			return t[len(p):]
		}
	}
	return t
}

// resolveImplode builds the regrouping bind for a write-body array that is the
// inverse of a declared read_shape explode.
//
// Everything it needs is already authored in the read shape -- which model attr
// holds the exploded rows, which element field is the group key, which is the
// member -- so inverting it needs no second declaration, and the two cannot
// drift apart.
func resolveImplode(src Source, schemaGen string, f Field, byTF map[string]ModelField, idx sdkIndex, ex readShapeExplode) (*implodeBind, error) {
	elem, _, isSlice := wrapperSliceElem(f.Type, idx)
	if !isSlice {
		return nil, fmt.Errorf("body field is %q, not an array", f.Type)
	}
	es, isStruct := idx.structs[elem]
	if !isStruct {
		return nil, fmt.Errorf("array element %q is not an SDK struct", elem)
	}
	mf, ok := byTF[ex.TF]
	if !ok {
		return nil, fmt.Errorf("read_shape explode attr %q is not in the model", ex.TF)
	}
	if len(ex.Carry) != 1 {
		return nil, fmt.Errorf("expected exactly one carried key, got %d", len(ex.Carry))
	}

	valFields, err := src.ModelFields(schemaGen, ex.ValueType)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", ex.ValueType, err)
	}
	valByTF := map[string]ModelField{}
	for _, vf := range valFields {
		valByTF[vf.TFSDK] = vf
	}
	keyVal, ok := valByTF[ex.Carry[0].TF]
	if !ok {
		return nil, fmt.Errorf("%s has no sub-attribute %q", ex.ValueType, ex.Carry[0].TF)
	}
	memberVal, ok := valByTF[ex.Each.TF]
	if !ok {
		return nil, fmt.Errorf("%s has no sub-attribute %q", ex.ValueType, ex.Each.TF)
	}

	keySDK, ok := fieldByJSON(es, ex.Carry[0].From)
	if !ok {
		return nil, fmt.Errorf("%s has no field %q", elem, ex.Carry[0].From)
	}
	memberSDK, ok := fieldByJSON(es, ex.Each.From)
	if !ok {
		return nil, fmt.Errorf("%s has no field %q", elem, ex.Each.From)
	}
	memberGo, isMemberSlice := sliceElem(memberSDK.Type)
	if !isMemberSlice || (memberGo != "uint64" && memberGo != "int64") {
		return nil, fmt.Errorf("member field %q is %q; expected []uint64 or []int64", ex.Each.From, memberSDK.Type)
	}
	keyConv, ok := expandConverter(keySDK)
	if !ok {
		return nil, fmt.Errorf("key field %q has unsupported type %q", ex.Carry[0].From, keySDK.Type)
	}

	wrap := ""
	if strings.HasPrefix(f.Type, "Opt") {
		wrap = f.Type
	}
	return &implodeBind{
		SDKField: f.GoName, Var: lowerFirst(f.GoName), ModelGo: mf.GoName,
		ElemType: elem, ValueType: ex.ValueType, Wrap: wrap,
		KeySDK: keySDK.GoName, KeyValue: keyVal.GoName,
		KeyConv:   keyConv + "(types.Int64Value(k))",
		MemberSDK: memberSDK.GoName, MemberValue: memberVal.GoName, MemberGo: memberGo,
	}, nil
}

func fieldByJSON(s Struct, jsonName string) (Field, bool) {
	for _, f := range s.Fields {
		if f.JSONName == jsonName {
			return f, true
		}
	}
	return Field{}, false
}

// flatSub is one sub-attribute of a nested object on the flatten (read) side:
// the model tfsdk name and the framework attr.Value expression built from the
// SDK struct field.
type flatSub struct {
	TF   string // "tag_key"
	Expr string // "flex.OptStringToFramework(elem.TagKey)"
}

// objFlat flattens a single nested-object response field into the model's
// tfplugingen Value type.
type objFlat struct {
	ModelGo   string // "AzurePolicy"
	ValueType string // "AzurePolicyValue"
	Var       string // "azurePolicyVal"
	Subs      []flatSub
}

// arrFlat flattens a nested object-array response field into a types.List of the
// element Value type.
type arrFlat struct {
	ModelGo   string // "Tags"
	ValueType string // "TagsValue"
	Var       string // "tags"
	SrcPath   string // "v.Data.Value.Tags.Value" (the []SDKElem to iterate)
	Subs      []flatSub
	TFSDK     string // model attribute, matched against list_order_keys
	OrderKey  string // element attribute to restore the prior order by; "" keeps the API's
}

// idProjFlat flattens a response struct-array into a plain list/set of int64
// ids by projecting each element's id field. It handles the write-ids /
// read-objects asymmetry: the create body takes owner_users []uint64, but the
// read payload returns owner_users []User{id}, and the model attribute is a
// bare types.List/types.Set of Int64.
type idProjFlat struct {
	ModelGo string // "OwnerUserGroups"
	Var     string // "ownerUserGroupsIDs"
	SrcPath string // "v.Data.Value.OwnerUserGroups.Value" (the []Elem to iterate)
	IDGo    string // element id field GoName, "ID"
	IDOpt   bool   // element id field is an Opt-wrapper (guard on .Set, read .Value)
	IDExpr  string // per-element int64 expression, "int64(elem.ID.Value)"
	Coll    string // "List" | "Set" (the model collection kind)
	TFSDK   string // model attribute, matched against empty_collections
	OrEmpty bool   // flatten no ids as an empty collection, not null
}

// objIDProjFlat flattens a single nested-object response field into the model's
// scalar int id attribute, projecting the object's id (read returns the object,
// e.g. aws_iam_permissions_boundary -> IAMPolicy; the model stores its id).
type objIDProjFlat struct {
	ModelGo string // "AwsIamPermissionsBoundary"
	SrcPath string // "v.Data.Value.AwsIamPermissionsBoundary" (the OptNil<Struct>)
	IDExpr  string // "int64(<src>.Value.ID.Value)", the int64 id expression
	Opt     bool   // the object field is Opt/OptNil (guard on .Set, else Int64Null)
}

// sumFlat rebuilds a scalar model attribute by summing one field of a response
// array, the inverse of a write the server expands.
//
// kion_budget is the case it exists for: POST /v3/budget takes a single amount
// and creates one data row per month with it distributed across them, so the
// read returns the rows and no amount at all. Nothing in the payload or the
// spec says the rows add back up to what was sent, which is why this is
// declared in crud_archetypes.yaml rather than derived.
type sumFlat struct {
	ModelGo  string // "Amount"
	Var      string // "amountSum"
	SrcPath  string // "v.Data.Value.Data.Value" (the []Elem to iterate)
	ElemExpr string // per-element float64 expression, "elem.Amount.Value"
	Round    bool   // model attr is Int64: round the float sum rather than truncate
}

type nestedFlatResult struct {
	Objs       []objFlat
	Arrs       []arrFlat
	IDProjs    []idProjFlat
	ObjIDProjs []objIDProjFlat
	Names      map[string]bool
}

// idArrayRename maps a payload object-array json-name to the model's id-list
// attribute name by stripping a trailing "s" and appending "_ids"
// (owner_user_groups -> owner_user_group_ids, owner_users -> owner_user_ids).
func idArrayRename(jsonName string) string {
	return strings.TrimSuffix(jsonName, "s") + "_ids"
}

// idObjectRename maps a payload single-object json-name to the model's scalar
// id attribute name (cloud_rule -> cloud_rule_id, service -> service_id).
//
// The write takes an id; the read returns the whole record. Where the SDK
// happens to name the response field cloud_rule_id anyway (ProjectEnforcement
// does, for the same OptCloudRule) the names already match and this is not
// reached -- which is exactly why project_enforcement unwrapped its cloud rule
// while ou_enforcement and funding_source_enforcement silently dropped theirs.
func idObjectRename(jsonName string) string {
	return jsonName + "_id"
}

// collKind maps a model field Go type to its collection kind ("List"/"Set"),
// or "" when the field is not a list/set.
func collKind(modelType string) string {
	switch modelType {
	case "types.List":
		return "List"
	case "types.Set":
		return "Set"
	}
	return ""
}

// structIDField returns a struct's id field (json name "id", or GoName ID/Id).
func structIDField(s Struct) (Field, bool) {
	for _, f := range s.Fields {
		if f.JSONName == "id" || f.GoName == "ID" || f.GoName == "Id" {
			return f, true
		}
	}
	return Field{}, false
}

// idProjExpr builds the per-element int64 expression for an id field of the
// given SDK type, and reports whether the field is an Opt-wrapper (guarded).
func idProjExpr(idGo, idType string) (expr string, opt bool, ok bool) {
	switch idType {
	case "OptUint64":
		return "int64(elem." + idGo + ".Value)", true, true
	case "OptInt64":
		return "elem." + idGo + ".Value", true, true
	case "uint64":
		return "int64(elem." + idGo + ")", false, true
	case "int64":
		return "elem." + idGo, false, true
	}
	return "", false, false
}

// resolveNestedFlatten finds nested-object / object-array fields in a read
// payload and builds their flatten binds (SDK struct -> model Value), reading
// the Value type sub-fields via src and the SDK struct sub-fields via idx.
func resolveNestedFlatten(src Source, schemaGen string, fields []Field, byTF map[string]ModelField, idx sdkIndex, prefix string) (nestedFlatResult, error) {
	res := nestedFlatResult{Names: map[string]bool{}}
	for _, f := range fields {
		mf, ok := byTF[f.JSONName]
		if !ok {
			// A payload object-array `foos` may map to the model's id-list `foo_ids`
			// (create takes owner_user_group_ids; read returns owner_user_groups).
			if _, _, isSlice := wrapperSliceElem(f.Type, idx); isSlice {
				if rmf, rok := byTF[idArrayRename(f.JSONName)]; rok {
					mf, ok = rmf, true
				}
			} else if renamed := idObjectRename(f.JSONName); !hasJSONFieldNamed(fields, renamed) {
				// Or a payload object `foo` maps to the model's scalar `foo_id`
				// (create takes cloud_rule_id; read returns the cloud rule). Skipped
				// when the payload carries a flat `foo_id` of its own, which owns
				// that attribute through the ordinary scalar path.
				if rmf, rok := byTF[renamed]; rok {
					mf, ok = rmf, true
				}
			}
			if !ok {
				continue
			}
		}
		if elem, _, isSlice := wrapperSliceElem(f.Type, idx); isSlice {
			es, isStruct := idx.structs[elem]
			if !isStruct || len(es.Fields) == 0 {
				continue
			}
			valueType := pascalCase(f.JSONName) + "Value"
			if !valueTypeExists(src, schemaGen, valueType) {
				// Not a nested tfplugingen attribute. If the model attribute is a
				// bare int list/set and the element struct carries an id, project
				// the ids (write-ids / read-objects asymmetry). Otherwise leave it
				// to respBinds (which will refuse an unconvertible type loudly).
				coll := collKind(mf.Type)
				if idf, hasID := structIDField(es); coll != "" && hasID {
					if expr, opt, okExpr := idProjExpr(idf.GoName, idf.Type); okExpr {
						srcPath := prefix + f.GoName
						if strings.HasPrefix(f.Type, "Opt") {
							srcPath += ".Value"
						}
						res.IDProjs = append(res.IDProjs, idProjFlat{
							ModelGo: mf.GoName, Var: lowerFirst(f.GoName) + "IDs", SrcPath: srcPath,
							IDGo: idf.GoName, IDOpt: opt, IDExpr: expr, Coll: coll, TFSDK: mf.TFSDK,
						})
						res.Names[f.JSONName] = true
					}
				}
				continue
			}
			subs, err := flatSubs(src, schemaGen, valueType, es, "elem")
			if err != nil {
				return res, fmt.Errorf("array field %q: %w", f.JSONName, err)
			}
			src := prefix + f.GoName
			if strings.HasPrefix(f.Type, "Opt") {
				src += ".Value" // unwrap the OptNil<T>Array
			}
			res.Arrs = append(res.Arrs, arrFlat{ModelGo: mf.GoName, ValueType: valueType, Var: lowerFirst(f.GoName) + "Vals", SrcPath: src, Subs: subs, TFSDK: mf.TFSDK})
			res.Names[f.JSONName] = true
			continue
		}
		base := strings.TrimPrefix(strings.TrimPrefix(f.Type, "Opt"), "Nil")
		bs, isStruct := idx.structs[base]
		if !isStruct || len(bs.Fields) == 0 {
			continue
		}
		valueType := pascalCase(f.JSONName) + "Value"
		if !valueTypeExists(src, schemaGen, valueType) {
			// Single object with no Value type: if the model attr is a scalar int
			// and the struct carries an id, project the object's id (read returns
			// the object; the model stores its id, e.g. aws_iam_permissions_boundary).
			if mf.Type == "types.Int64" {
				if idf, hasID := structIDField(bs); hasID {
					opt := strings.HasPrefix(f.Type, "Opt")
					idBase := prefix + f.GoName
					if opt {
						idBase += ".Value"
					}
					if idExpr, _, ok := idProjExpr(idf.GoName, idf.Type); ok {
						// idProjExpr builds "int64(elem.ID...)"; rebase "elem" onto idBase.
						idExpr = strings.Replace(idExpr, "elem.", idBase+".", 1)
						res.ObjIDProjs = append(res.ObjIDProjs, objIDProjFlat{
							ModelGo: mf.GoName, SrcPath: prefix + f.GoName, IDExpr: idExpr, Opt: opt,
						})
						res.Names[f.JSONName] = true
					}
				}
			}
			continue // not a tfplugingen nested attribute
		}
		objPrefix := prefix + f.GoName
		if strings.HasPrefix(f.Type, "Opt") {
			objPrefix += ".Value" // unwrap the Opt<T> envelope to reach sub-fields
		}
		subs, err := flatSubs(src, schemaGen, valueType, bs, objPrefix)
		if err != nil {
			return res, fmt.Errorf("object field %q: %w", f.JSONName, err)
		}
		res.Objs = append(res.Objs, objFlat{ModelGo: mf.GoName, ValueType: valueType, Var: lowerFirst(f.GoName) + "Val", Subs: subs})
		res.Names[f.JSONName] = true
	}
	return res, nil
}

// flatSubs maps a nested struct's SDK sub-fields to flatten expressions building
// framework attr.Values, keyed by the model tfsdk name.
func flatSubs(src Source, schemaGen, valueType string, sdkStruct Struct, prefix string) ([]flatSub, error) {
	valFields, err := src.ModelFields(schemaGen, valueType)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", valueType, err)
	}
	valByTF := map[string]bool{}
	for _, vf := range valFields {
		valByTF[vf.TFSDK] = true
	}
	var subs []flatSub
	for _, sf := range sdkStruct.Fields {
		if !valByTF[sf.JSONName] {
			continue
		}
		conv, ok := flattenConverter(sf)
		if !ok {
			return nil, fmt.Errorf("sub-field %q has unsupported type %q", sf.JSONName, sf.Type)
		}
		subs = append(subs, flatSub{TF: sf.JSONName, Expr: conv + "(" + prefix + "." + sf.GoName + ")"})
	}
	return subs, nil
}

// resolveSumFlats builds the sum binds declared by an archetype's sum_from,
// resolving each against the read payload's array field and the model attribute
// it rebuilds. Every declaration must resolve: a stale one would leave the
// attribute dropped again, which is the class of bug this exists to fix.
func resolveSumFlats(decls []sumFromDecl, fields []Field, byTF map[string]ModelField, idx sdkIndex, prefix string) ([]sumFlat, error) {
	var out []sumFlat
	for _, d := range decls {
		mf, ok := byTF[d.TF]
		if !ok {
			return nil, fmt.Errorf("sum_from tf %q is not a model attribute", d.TF)
		}
		if mf.Type != "types.Int64" && mf.Type != "types.Float64" {
			return nil, fmt.Errorf("sum_from tf %q is %s; expected types.Int64 or types.Float64", d.TF, mf.Type)
		}
		src, ok := fieldByJSONName(fields, d.From)
		if !ok {
			return nil, fmt.Errorf("sum_from from %q is not a field of the read payload", d.From)
		}
		elem, _, isSlice := wrapperSliceElem(src.Type, idx)
		if !isSlice {
			return nil, fmt.Errorf("sum_from from %q is %q, not an array", d.From, src.Type)
		}
		es, isStruct := idx.structs[elem]
		if !isStruct {
			return nil, fmt.Errorf("sum_from from %q element %q is not an SDK struct", d.From, elem)
		}
		ef, ok := fieldByJSON(es, d.Field)
		if !ok {
			return nil, fmt.Errorf("sum_from field %q is not a field of %s", d.Field, elem)
		}
		expr, ok := sumElemExpr(ef)
		if !ok {
			return nil, fmt.Errorf("sum_from field %q has unsupported type %q", d.Field, ef.Type)
		}
		srcPath := prefix + src.GoName
		if strings.HasPrefix(src.Type, "Opt") {
			srcPath += ".Value"
		}
		out = append(out, sumFlat{
			ModelGo: mf.GoName, Var: lowerFirst(mf.GoName) + "Sum", SrcPath: srcPath,
			ElemExpr: expr, Round: mf.Type == "types.Int64",
		})
	}
	return out, nil
}

// fieldByJSONName is fieldByJSON over a bare field slice.
func fieldByJSONName(fields []Field, jsonName string) (Field, bool) {
	for _, f := range fields {
		if f.JSONName == jsonName {
			return f, true
		}
	}
	return Field{}, false
}

// hasJSONFieldNamed reports whether fields carries a field with that json name.
func hasJSONFieldNamed(fields []Field, jsonName string) bool {
	_, ok := fieldByJSONName(fields, jsonName)
	return ok
}

// sumElemExpr builds the per-element float64 expression for a summable field.
func sumElemExpr(f Field) (string, bool) {
	switch f.Type {
	case "float64":
		return "elem." + f.GoName, true
	case "int64", "uint64":
		return "float64(elem." + f.GoName + ")", true
	case "OptFloat64", "OptNilFloat64":
		return "elem." + f.GoName + ".Value", true
	case "OptInt64", "OptNilInt64", "OptUint64", "OptNilUint64":
		return "float64(elem." + f.GoName + ".Value)", true
	}
	return "", false
}

// nestedSubs maps a nested struct's SDK sub-fields to expand expressions reading
// the corresponding tfplugingen Value sub-field (matched by json/tfsdk name).
func nestedSubs(src Source, schemaGen, valueType string, sdkStruct Struct, prefix string) ([]objSub, error) {
	valFields, err := src.ModelFields(schemaGen, valueType)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", valueType, err)
	}
	byTF := map[string]ModelField{}
	for _, vf := range valFields {
		byTF[vf.TFSDK] = vf
	}
	var subs []objSub
	for _, sf := range sdkStruct.Fields {
		vf, ok := byTF[sf.JSONName]
		if !ok {
			continue
		}
		if expr, ok := objectSubExpr(sf, vf, prefix); ok {
			subs = append(subs, objSub{SDKField: sf.GoName, Expr: expr})
			continue
		}
		conv, ok := expandConverter(sf)
		if !ok {
			// A sub-field the generator cannot express, in practice a struct
			// nested inside an already-nested object, which the one-level
			// expand machinery has no converter for. Skip it rather than
			// failing the whole resource: the alternative is that the resource
			// cannot be generated at all, which is strictly worse than one
			// unsettable attribute. Warned so it is not silent.
			fmt.Printf("kgen crud: %s.%s: no expand converter for type %q; attribute will not be settable\n",
				valueType, sf.JSONName, sf.Type)
			continue
		}
		subs = append(subs, objSub{SDKField: sf.GoName, Expr: conv + "(" + prefix + "." + vf.GoName + ")"})
	}
	return subs, nil
}
