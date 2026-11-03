package crud

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"go/format"
	"slices"
	"strings"
	"text/template"
)

//go:embed datasource.gtpl
var dataSourceTmpl string

//go:embed datasource_list.gtpl
var dataSourceListTmpl string

// dsField binds one data-source attribute: schema type, model type, flatten
// converter, and the SDK response field it reads.
type dsField struct {
	ModelGo    string // model field, e.g. "Key"
	TFName     string // tfsdk name, e.g. "key"
	SchemaType string // "StringAttribute" | "Int64Attribute" | "BoolAttribute"
	ModelType  string // "types.String" | "types.Int64" | "types.Bool"
	Flatten    string // "flex.StringToFramework"
	SDKGo      string // access path from the record, e.g. "Key" or "Cft.Value.Key"
}

// NullExpr is the framework null value for this attribute's type, used where the
// data source cannot populate it (filter mode).
func (f dsField) NullExpr() string { return nullExpr(f.ModelType) }

// dsData is the template payload for the id-only data source.
type dsData struct {
	Pkg, Pascal, Model       string
	DSConst, DSName          string
	SDKAlias                 string
	SDKImport                string
	ReadMethod, ReadParams   string
	ReadIDParam              string
	IDParamType              string // "int64" | "uint64" | "string"
	IDGo, IDFlatten, IDSDKGo string
	HasRespID                bool   // response echoes the id (re-flatten it); else keep config id
	DataPtr                  bool   // envelope Data is a pointer (*Payload) not an Opt-wrapper
	PayloadPath              string // "" | ".Webhook.Value", descent from Data to the record
	RespType                 string
	Gated                    bool // Read checks the package's version gate first
	TypeName                 string
	Attrs                    []dsField
	// OuterAttrs are attributes read from the payload above the record (see
	// recordView.Outer): available by id, null under a filter.
	OuterAttrs []dsField
}

// PayloadExpr is the Go expression for the read payload (the level above the
// record), used to source OuterAttrs.
func (d dsData) PayloadExpr() string {
	if d.DataPtr {
		return "api.Data"
	}
	return "api.Data.Value"
}

// needsPayload reports whether the read response is consulted at all, false for
// a degenerate read (empty payload, no echoed id), where the data source only
// returns the config-provided id.
func (d dsData) NeedsPayload() bool {
	return d.HasRespID || len(d.Attrs) > 0 || len(d.OuterAttrs) > 0
}

// renderDataSource renders <name>_data_source.go. When the resource has a
// resolved list endpoint it emits the full dual-mode data source (id + legacy
// filter blocks + list); otherwise, or if the dual-mode build fails, it falls
// back to the id-only data source and reports the downgrade.
//
// The returned downgrade string is non-empty exactly when a collection read was
// configured but the filter-capable data source could not be built; drops lists
// the fields the dual-mode build could not project.
func renderDataSource(rm ResourceModel, policy FieldPolicy) (out []byte, downgrade string, drops []dropped, err error) {
	out, downgrade, drops, _, err = renderDataSourceAttrs(rm, policy)
	return out, downgrade, drops, err
}

// renderDataSourceAttrs is renderDataSource that also returns the top-level
// attributes the rendered schema has, besides id.
func renderDataSourceAttrs(rm ResourceModel, policy FieldPolicy) (out []byte, downgrade string, drops []dropped, attrs []string, err error) {
	if rm.List != nil {
		data, d, berr := buildListDSData(rm, policy)
		if berr == nil {
			b, err := execGoTemplate("datasource_list", dataSourceListTmpl, data, rm.Name+"_data_source.go")
			return b, "", d, data.attrNames(), err
		}
		downgrade = berr.Error()
	} else {
		downgrade = rm.ListDowngrade
	}
	data, err := buildDSData(rm)
	if err != nil {
		return nil, downgrade, nil, nil, err
	}
	b, err := execGoTemplate("datasource", dataSourceTmpl, data, rm.Name+"_data_source.go")
	return b, downgrade, nil, data.attrNames(), err
}

// attrNames are the schema's top-level attributes besides id.
func (d dsData) attrNames() []string {
	var out []string
	for _, f := range append(slices.Clone(d.Attrs), d.OuterAttrs...) {
		out = append(out, f.TFName)
	}
	return out
}

// renderIDOnlyDataSource renders an id-only <name>_data_source.go (single read
// by id, no filter/list). The degradation path for resources without a
// resolvable list endpoint.
func renderIDOnlyDataSource(rm ResourceModel) ([]byte, error) {
	data, err := buildDSData(rm)
	if err != nil {
		return nil, err
	}
	return execGoTemplate("datasource", dataSourceTmpl, data, rm.Name+"_data_source.go")
}

// recordView tells the data-source builders which SDK struct to treat as "the
// record" and how to reach each of its scalar fields.
//
// Read payloads come in two arrangements. Usually the payload is the record
// (Label, Project), but many are a "with owners" envelope that nests the record
// under a sub-object (WebhookWithOwners.webhook, CFTWithOwnersAndTags.cft)
// alongside sibling collections. When the index endpoint's element type is that
// sub-object, the data source descends into it (PayloadPath) so the by-id and
// by-filter modes share one field mapping; otherwise the payload stays the
// record and the sub-object's fields are reached through it.
type recordView struct {
	PayloadPath string            // "" | ".Webhook.Value"
	Fields      map[string]Field  // json name -> SDK field
	Paths       map[string]string // json name -> access path from the record
	// Outer holds fields that live on the read payload but NOT on the record,
	// only possible when the record is a sub-object (e.g. UserWithUGroups.mfa
	// sits beside the User record). They are readable by id but absent from the
	// collection, so they stay out of the `list` objects and go null under a
	// filter. Paths are relative to the payload, not the record.
	Outer      map[string]Field
	OuterPaths map[string]string
}

func buildRecordView(rm ResourceModel) recordView {
	v := recordView{
		Fields: map[string]Field{}, Paths: map[string]string{},
		Outer: map[string]Field{}, OuterPaths: map[string]string{},
	}
	wrapPrefix := ""
	if rm.Read.RespWrapperGo != "" {
		wrapPrefix = rm.Read.RespWrapperGo + "."
		if rm.Read.RespWrapperOpt {
			wrapPrefix += "Value."
		}
	}
	if rm.List != nil && rm.List.ElemIsWrapper {
		v.PayloadPath = "." + strings.TrimSuffix(wrapPrefix, ".")
		for _, f := range rm.Read.RespWrapperFields {
			v.Fields[f.JSONName] = f
			v.Paths[f.JSONName] = f.GoName
		}
		for _, f := range rm.Read.RespFields {
			if _, onRecord := v.Fields[f.JSONName]; onRecord || f.GoName == rm.Read.RespWrapperGo {
				continue
			}
			v.Outer[f.JSONName] = f
			v.OuterPaths[f.JSONName] = f.GoName
		}
		return v
	}
	// Wrapper fields first so a same-named top-level field wins.
	for _, f := range rm.Read.RespWrapperFields {
		v.Fields[f.JSONName] = f
		v.Paths[f.JSONName] = wrapPrefix + f.GoName
	}
	for _, f := range rm.Read.RespFields {
		v.Fields[f.JSONName] = f
		v.Paths[f.JSONName] = f.GoName
	}
	return v
}

func buildDSData(rm ResourceModel) (dsData, error) {
	view := buildRecordView(rm)
	respByJSON := view.Fields

	d := dsData{
		Pkg:         rm.Name,
		Pascal:      rm.Pascal,
		Model:       rm.Name + "DataSourceModel",
		DSConst:     "DSName" + rm.Pascal,
		DSName:      rm.Pascal + " Data Source",
		SDKAlias:    "generated",
		SDKImport:   sdkImport(),
		ReadMethod:  rm.Read.Method.Name,
		ReadParams:  rm.Read.Method.ParamsType,
		RespType:    rm.Read.RespType,
		DataPtr:     rm.Read.RespDataPtr,
		PayloadPath: view.PayloadPath,
		IDGo:        rm.IDField.GoName,
		Gated:       rm.DSGated,
		TypeName:    "kion_" + rm.Name,
	}

	readIDParam, readIDType, err := idParamName(rm.Read.Params)
	if err != nil {
		return d, fmt.Errorf("%s data source read: %w", rm.Name, err)
	}
	d.ReadIDParam = readIDParam
	d.IDParamType = readIDType

	// The id may be absent from the read response (some APIs don't echo it). Then
	// the data source keeps the config-provided id rather than re-flattening it.
	if idResp, ok := respByJSON["id"]; ok {
		idFlatten, ok := flattenConverter(idResp)
		if !ok {
			return d, fmt.Errorf("%s data source: id response field type %q unsupported", rm.Name, idResp.Type)
		}
		d.HasRespID = true
		d.IDFlatten = idFlatten
		d.IDSDKGo = view.Paths["id"]
	}

	for _, mf := range rm.projectedFields() {
		rf, onRecord := respByJSON[mf.TFSDK]
		path := view.Paths[mf.TFSDK]
		if !onRecord {
			if rf, onRecord = view.Outer[mf.TFSDK]; !onRecord {
				continue
			}
			path = view.OuterPaths[mf.TFSDK]
		}
		// The data source can only expose scalar fields it knows how to render;
		// skip anything else (e.g. list/set/object) rather than refuse the whole
		// resource. The resource itself is unaffected.
		sType, ok := schemaAttrType(mf.Type)
		if !ok {
			continue
		}
		flat, ok := flattenConverter(rf)
		if !ok {
			continue
		}
		f := dsField{
			ModelGo:    mf.GoName,
			TFName:     mf.TFSDK,
			SchemaType: sType,
			ModelType:  mf.Type,
			Flatten:    flat,
			SDKGo:      path,
		}
		if _, isOuter := view.Outer[mf.TFSDK]; isOuter {
			d.OuterAttrs = append(d.OuterAttrs, f)
			continue
		}
		d.Attrs = append(d.Attrs, f)
	}
	byName := func(a, b dsField) int { return cmp.Compare(a.TFName, b.TFName) }
	slices.SortFunc(d.Attrs, byName)
	slices.SortFunc(d.OuterAttrs, byName)
	return d, nil
}

// schemaAttrType maps a framework model type to its schema.Attribute kind.
func schemaAttrType(modelType string) (string, bool) {
	switch modelType {
	case "types.String":
		return "StringAttribute", true
	case "types.Int64":
		return "Int64Attribute", true
	case "types.Bool":
		return "BoolAttribute", true
	}
	return "", false
}

// attrTypeFor maps a framework model type to its attr.Type expression, used for
// the `list` nested-object attribute types.
func attrTypeFor(modelType string) (string, bool) {
	switch modelType {
	case "types.String":
		return "types.StringType", true
	case "types.Int64":
		return "types.Int64Type", true
	case "types.Bool":
		return "types.BoolType", true
	}
	return "", false
}

// execGoTemplate parses, executes, and gofmts a Go-source template.
func execGoTemplate(name, tmpl string, data any, outName string) ([]byte, error) {
	t, err := template.New(name).Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse %s template: %w", name, err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute %s template: %w", name, err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated %s: %w\n%s", outName, err, buf.Bytes())
	}
	return src, nil
}

// --- dual-mode (id + filter + list) data source ---

// listObjField describes one attribute of the `list` nested object (and its
// listObjectAttrTypes entry), plus how to build its value from an SDK element.
type listObjField struct {
	TFName     string // "id","key"
	SchemaType string // "Int64Attribute","StringAttribute" (nested schema)
	AttrType   string // "types.Int64Type","types.StringType" (object attr types)
	IsID       bool
	SDKGo      string // element field GoName: "ID","Key"
	ObjExpr    string // non-id object value: "types.StringValue(lbl.Key)"
	RowExpr    string // non-id filter-row value: "lbl.Key"
}

// listScalarAssign drives readByID (Flatten) and readByFilter (NullExpr) for the
// top-level scalar attributes.
type listScalarAssign struct {
	ModelGo  string // "Id","Key"
	Flatten  string // "flex.OptUint64ToFramework" / "flex.StringToFramework"
	SDKGo    string // "ID","Key"
	NullExpr string // "types.Int64Null()" / "types.StringNull()"
	IsID     bool
}

type listDSData struct {
	dsData
	listAccess
	ElemType     string // "Label", the list element type
	ListMethod   string
	ListParams   string
	ListRespType string
	PageParam    string
	CountParam   string
	ObjFields    []listObjField
	Assigns      []listScalarAssign
}

// dropped is one field a list data source could not project into its `list`
// objects, so the loss is reported at the end of a run rather than absorbed.
type dropped struct {
	Field  string
	Reason string
}

func buildListDSData(rm ResourceModel, policy FieldPolicy) (listDSData, []dropped, error) {
	var drops []dropped
	base, err := buildDSData(rm)
	if err != nil {
		return listDSData{}, nil, err
	}
	view := buildRecordView(rm)
	// The dual-mode list objects carry an id; without one in the response the
	// list can't expose it, so degrade to the id-only data source.
	if !base.HasRespID {
		// An empty field set means the read payload's OpenAPI schema declares no
		// properties at all. Nothing the generator can do, the spec must be fixed.
		if len(view.Fields) == 0 {
			return listDSData{}, nil, fmt.Errorf("%s dual-mode data source: read payload %s has no fields at all (empty OpenAPI schema)", rm.Name, rm.Read.RespPayload)
		}
		return listDSData{}, nil, fmt.Errorf("%s dual-mode data source: response has no id field", rm.Name)
	}
	lm := rm.List
	// The id is special-cased: it is exposed as an Int64 (a String for a string
	// id) and the list objects key off it, so an id the templates cannot unwrap
	// is fatal here.
	if !idFieldOK(view.Fields["id"].Type, base.IDParamType) {
		return listDSData{}, nil, fmt.Errorf("%s dual-mode data source: id field type %q unsupported for list", rm.Name, view.Fields["id"].Type)
	}
	d := listDSData{
		dsData:       base,
		listAccess:   lm.access(),
		ElemType:     lm.ElemType,
		ListMethod:   lm.Method.Name,
		ListParams:   lm.Method.ParamsType,
		ListRespType: lm.RespType,
		PageParam:    lm.PageParam,
		CountParam:   lm.CountParam,
	}

	idSchema, idAttr, idNull := "Int64Attribute", "types.Int64Type", "types.Int64Null()"
	if base.IDParamType == "string" {
		idSchema, idAttr, idNull = "StringAttribute", "types.StringType", "types.StringNull()"
	}
	d.ObjFields = append(d.ObjFields, listObjField{
		TFName: "id", SchemaType: idSchema, AttrType: idAttr,
		IsID: true, SDKGo: base.IDSDKGo,
	})
	d.Assigns = append(d.Assigns, listScalarAssign{
		ModelGo: base.IDGo, Flatten: base.IDFlatten, SDKGo: base.IDSDKGo,
		NullExpr: idNull, IsID: true,
	})

	// A field the `list` objects cannot render is dropped from the data source
	// entirely rather than failing it, losing one attribute beats losing the
	// whole filter block. The drop is recorded so the run reports it.
	attrs := make([]dsField, 0, len(base.Attrs))
	for _, a := range base.Attrs {
		pf, ok := view.Fields[a.TFName]
		if !ok {
			return listDSData{}, nil, fmt.Errorf("field %q not in response payload", a.TFName)
		}
		attrType, objExpr, rowExpr, ok := listValueExprs("lbl."+a.SDKGo, pf.Type, a.ModelType)
		if !ok {
			drops = append(drops, dropped{a.TFName, fmt.Sprintf("resource attribute: SDK type %q has no `list` projection", pf.Type)})
			continue
		}
		attrs = append(attrs, a)
		d.ObjFields = append(d.ObjFields, listObjField{
			TFName: a.TFName, SchemaType: a.SchemaType, AttrType: attrType,
			SDKGo: a.SDKGo, ObjExpr: objExpr, RowExpr: rowExpr,
		})
		d.Assigns = append(d.Assigns, listScalarAssign{
			ModelGo: a.ModelGo, Flatten: a.Flatten, SDKGo: a.SDKGo,
			NullExpr: nullExpr(a.ModelType),
		})
	}
	d.Attrs = attrs

	// Response-only scalars: everything above came from the resource's
	// attributes, which are generated from the create request body, so a field
	// the server owns has not been considered yet. Those reach the `list`
	// objects and the filter row but deliberately NOT d.Attrs or d.Assigns --
	// they are not settable, so they have no place among the data source's own
	// scalar arguments.
	seen := make(map[string]bool, len(d.ObjFields))
	for _, f := range d.ObjFields {
		seen[f.TFName] = true
	}
	names := make([]string, 0, len(view.Fields))
	for name := range view.Fields {
		names = append(names, name)
	}
	slices.Sort(names) // generated output must not depend on map order
	for _, name := range names {
		if seen[name] {
			continue
		}
		rf := view.Fields[name]
		ls, known := listScalars[rf.Type]
		if !known {
			continue // not a scalar the list objects can render
		}
		if policy.Hidden(rm.Name, name, ls.model) {
			continue
		}
		// Past this point the type IS renderable, so a failure is an internal
		// inconsistency rather than an expected shape. Report those.
		path, ok := view.Paths[name]
		if !ok {
			drops = append(drops, dropped{name, "response field: renderable scalar with no access path"})
			continue
		}
		attrType, objExpr, rowExpr, ok := listValueExprs("lbl."+path, rf.Type, ls.model)
		if !ok {
			drops = append(drops, dropped{name, fmt.Sprintf("response field: SDK type %q has no `list` projection", rf.Type)})
			continue
		}
		schemaType, ok := schemaAttrType(ls.model)
		if !ok {
			drops = append(drops, dropped{name, fmt.Sprintf("response field: model type %q has no schema attribute", ls.model)})
			continue
		}
		d.ObjFields = append(d.ObjFields, listObjField{
			TFName: name, SchemaType: schemaType, AttrType: attrType,
			SDKGo: path, ObjExpr: objExpr, RowExpr: rowExpr,
		})
	}
	return d, drops, nil
}

// idFieldOK is listIDOK for a record whose id param is idParamType: a string id
// is unwrapped from an OptString.
func idFieldOK(sdkType, idParamType string) bool {
	if idParamType == "string" {
		return sdkType == "OptString"
	}
	return listIDOK(sdkType)
}

// listIDOK reports whether an SDK id field is an Opt-wrapped integer, the shape
// the list templates unwrap (`lbl.ID.Set` / `int64(lbl.ID.Value)`).
func listIDOK(sdkType string) bool {
	switch sdkType {
	case "OptInt64", "OptUint64", "OptNilInt64", "OptNilUint64":
		return true
	}
	return false
}

// listScalar says how one SDK scalar type projects into a list object: the
// framework model type it must line up with, the Go expression that unwraps it
// (%s is the access path), and the types.* constructor that boxes it.
type listScalar struct {
	model string
	value string
	ctor  string
}

// listScalars covers the SDK scalar types a list element can expose. Anything
// absent (nested objects, slices, jx.Raw, enums) is skipped for the `list`
// attribute rather than failing the whole data source.
var listScalars = map[string]listScalar{
	"string":       {"types.String", "%s", "types.StringValue"},
	"OptString":    {"types.String", `%s.Or("")`, "types.StringValue"},
	"OptNilString": {"types.String", `%s.Or("")`, "types.StringValue"},
	"bool":         {"types.Bool", "%s", "types.BoolValue"},
	"OptBool":      {"types.Bool", "%s.Or(false)", "types.BoolValue"},
	"OptNilBool":   {"types.Bool", "%s.Or(false)", "types.BoolValue"},
	"NilBool":      {"types.Bool", "%s.Or(false)", "types.BoolValue"},
	"int64":        {"types.Int64", "%s", "types.Int64Value"},
	"OptInt64":     {"types.Int64", "%s.Or(0)", "types.Int64Value"},
	"OptNilInt64":  {"types.Int64", "%s.Or(0)", "types.Int64Value"},
	"uint64":       {"types.Int64", "int64(%s)", "types.Int64Value"},
	"OptUint64":    {"types.Int64", "int64(%s.Or(0))", "types.Int64Value"},
	"OptNilUint64": {"types.Int64", "int64(%s.Or(0))", "types.Int64Value"},
	"NilUint64":    {"types.Int64", "int64(%s.Or(0))", "types.Int64Value"},
	// uint64-backed ogen enum; its resource attribute is an Int64.
	"OptGCPRoleLaunchStage": {"types.Int64", "int64(%s.Or(0))", "types.Int64Value"},
	// SQL-style nullable timestamp; the flex converter renders it as RFC3339.
	"OptNullTime": {"types.String", "flex.OptNullTimeToFramework(%s).ValueString()", "types.StringValue"},
}

// listValueExprs returns the object attr type plus the object/row value
// expressions for one scalar field of a list element. modelType pins the
// framework type the schema declares, so a projection that would yield a
// different type is refused rather than silently mismatched. ok=false for a
// type the list path can't render.
func listValueExprs(access, sdkType, modelType string) (attrType, objExpr, rowExpr string, ok bool) {
	ls, known := listScalars[sdkType]
	if !known || ls.model != modelType {
		return "", "", "", false
	}
	attrType, ok = attrTypeFor(modelType)
	if !ok {
		return "", "", "", false
	}
	rowExpr = fmt.Sprintf(ls.value, access)
	return attrType, ls.ctor + "(" + rowExpr + ")", rowExpr, true
}

func nullExpr(modelType string) string {
	switch modelType {
	case "types.Int64":
		return "types.Int64Null()"
	case "types.Bool":
		return "types.BoolNull()"
	default:
		return "types.StringNull()"
	}
}
