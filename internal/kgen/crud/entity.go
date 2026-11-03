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

//go:embed entity.gtpl
var entityTmpl string

//go:embed entity_noread.gtpl
var entityNoReadTmpl string

// fieldBind binds one SDK request-body field to a model field + expand converter.
type fieldBind struct {
	SDKField  string // SDK struct field, e.g. "Key"
	ModelGo   string // model field, e.g. "Key"
	Converter string // e.g. "flex.StringValueFromFramework"
}

// sliceBind binds a slice-typed request-body field. The flex slice helpers take
// (ctx) and return diagnostics, so the template emits a prelude computing Var.
type sliceBind struct {
	SDKField string // "BillingSourceIds"
	ModelGo  string // "BillingSourceIds"
	Func     string // "flex.Uint64SliceFromFramework"
	Var      string // local var, "billingSourceIds"
	Wrap     string // ogen nil-aware wrapper (OptNil<T>Array), "" for a bare slice
}

// sliceRespBind binds a slice-typed response field (flatten prelude).
type sliceRespBind struct {
	ModelGo string
	Func    string // "flex.Uint64SliceToFramework"
	SDKPath string // "v.Data.Value.BillingSourceIds"
	Var     string
}

// respBind binds one response-payload field to a model field. The id field is
// special-cased (formatted uint64/int64 -> string) rather than flex-converted.
type respBind struct {
	ModelGo      string
	SDKPath      string // e.g. "v.Data.Value.Key"
	Converter    string // e.g. "flex.StringToFramework" (empty when IsID)
	IsID         bool
	IDFormatFunc string // "FormatUint" | "FormatInt"; "" for a string id (only when IsID)
	IDSDKPath    string // e.g. "v.Data.Value.ID" (only when IsID)
}

// entityData is the template payload for the entity archetype.
type entityData struct {
	Pkg, Pascal, Model string
	ResConst, ResName  string
	TypeName           string // "kion_<name>"
	IDGo, SDKAlias     string
	SDKImport          string
	IDParamType        string // "int64" | "uint64" | "string". The read/update/delete param id Go type
	IDParseBits        int    // strconv bit size for the record id. See idParseBits
	IDFormat           string // "%d" | "%s": the id's verb in diagnostics
	IDVar              string // "idInt" | "id": the local holding the parsed id
	Gated              bool   // emit RequireKionVersionInRange in Create
	// AtLeastOneOf are attributes the API requires at least one of on create.
	// Emitted as framework.AtLeastOneOfOnCreate so the failure lands at plan
	// time without rejecting an existing record that has none.
	AtLeastOneOf []string
	// RequiredWhen are attributes required only for a given value of another.
	RequiredWhen []RequiredWhen
	// RequiredTogether are attribute groups that must be set all or none.
	RequiredTogether [][]string
	// RawDeletePath is the private route to DELETE through when the public spec
	// publishes no delete. Without it the generator emits a Delete that only
	// warns, so every record a configuration creates survives `terraform
	// destroy`: present in Kion, absent from state, unreachable by Terraform.
	RawDeletePath string
	// ImportParentTF is the parent attribute an import id must carry because
	// the read cannot recover it. Delete addresses the record through it, so a
	// bare-id import leaves a resource whose delete targets parent 0.
	ImportParentTF   string
	SchemaVersion    int // >0 bumps resp.Schema.Version (state migration)
	CreateMethod     string
	CreateBodyOpt    string
	CreateBody       string
	CreateBodyPtr    bool
	CreateBinds      []fieldBind
	CreateSliceBinds []sliceBind
	// Literal query-string discriminator params (ogen's __qs__ factory-create
	// routes, e.g. POST /v3/account/__qs__/account-type/google-cloud): the create
	// call takes a params struct whose single field is a compile-time constant
	// (the trailing path segment). "" when create has no params struct.
	CreateParams       string
	CreateParamField   string
	CreateParamLiteral string
	// Create-under-parent: the create-params field named CreateParentParam is set
	// from the model's CreateParentField attribute (cast by CreateParentCast when
	// the SDK field is not int64). Takes precedence over the literal path.
	CreateParentParam string
	CreateParentField string
	CreateParentCast  string
	ConfirmListed     string // see ResourceModel.ConfirmListed
	IDIsParent        bool   // see ResourceModel.IDIsParent
	ConfirmParents    string // see ResourceModel.ConfirmParents
	MissingViaList    string // see ResourceModel.MissingViaList
	// Create id source: "created" uses errs.CreatedID (a *CreatedResponse);
	// "envelope" reads the id from the create response payload (a returned record).
	CreateIDMode       string
	CreateRespType     string // envelope: the create response type, e.g. "ScopeResponse"
	CreateDataPtr      bool   // envelope: Data is *Payload (else an Opt-wrapper)
	CreateIDOpt        bool   // envelope: the id field is an Opt-wrapper (.Set/.Value)
	CreateIDSDKName    string // envelope: the id field GoName, e.g. "ID"
	ReadMethod         string
	ReadParams         string
	ReadIDParam        string
	HasUpdate          bool
	UpdateMethod       string
	UpdateBodyOpt      string
	UpdateBody         string
	UpdateBodyPtr      bool
	UpdateParams       string
	UpdateIDParam      string
	UpdateIDExpr       string // "idInt" or a cast when the update param id type differs from IDParamType
	UpdateBinds        []fieldBind
	UpdateSliceBinds   []sliceBind
	HasDelete          bool
	ParentRead         *parentReadData // no_read resources given a real read (see parentread.go)
	DeleteMethod       string
	DeleteParams       string
	DeleteIDParam      string
	DeleteIDExpr       string // "idInt" or a cast when the delete param id type differs from IDParamType
	DeleteExtraParam   string // asymmetric 2-param delete: the extra param name
	DeleteExtraFieldGo string // model field (Go name) sourcing DeleteExtraParam
	RespType           string
	RespDataPtr        bool
	RespBinds          []respBind
	RespSliceBinds     []sliceRespBind
	HasRespSlices      bool
	// NoRead resources have no single-record GET: Create persists the plan (no
	// read-back), Read is a no-op that keeps state, and no data source is emitted.
	NoRead bool
	// Nested-object binds (expand build preludes for create/update bodies,
	// flatten build preludes for the read payload).
	CreateObjBinds  []objBind
	CreateArrBinds  []arrBind
	CreateFlatSubs  []objSub
	CreateRawValues []rawValueBind
	// RawValueHelpers is the create+update union: the build<Field> helper each
	// bind calls is emitted once per resource.
	RawValueHelpers []rawValueBind
	UpdateObjBinds  []objBind
	UpdateArrBinds  []arrBind
	UpdateFlatSubs  []objSub
	UpdateImplode   *implodeBind
	UpdateRawValues []rawValueBind
	RespObjFlats    []objFlat
	RespArrFlats    []arrFlat
	RespIDProjs     []idProjFlat
	RespObjIDProjs  []objIDProjFlat
	RespSums        []sumFlat // scalars rebuilt by summing an expanded array
	// ReadCompanion is a hand-authored function called after every flatten, for
	// attributes the by-id read does not return at all (see archetype).
	ReadCompanion string
	// RawCreate routes only the create over raw HTTP; see the archetype key.
	RawCreate *rawCreateData
	// Rewritten holds the model field names of attributes the API canonicalizes
	// and echoes back rewritten; Create and Update restore the configured value
	// after the read-back. See archetype's Rewritten.
	Rewritten []string
	// CreateViaUpdate: model fields the create ignores, re-applied by an update
	// when the read-back differs. See archetype's CreateViaUpdate.
	CreateViaUpdate []string
	// CreateOnly are captured from the create response before the read-back
	// and restored after it; Update carries state's value. Never flattened.
	CreateOnly    []createOnlyBind
	HasNestedFlat bool // obj/arr flattens present (drives the attr import)
	HasIDProj     bool // id-projection flattens present
	RespRawValues []rawValueFlat
	// Blended resources (rendered by blended.gtpl, never entity.gtpl) mix typed
	// public ops with raw private ops. These fields are zero for pure-typed
	// resources, so entity.gtpl, which never references them. Is unaffected.
	Blended   bool
	RawRead   *rawReadData  // always set for a blended resource (read is private)
	RawUpdate *rawWriteData // set when update is private (else typed via Update*)
	RawDelete *rawWriteData // set when delete is private (else typed via Delete*)
	// Owner association synced on Update via paired add/remove endpoints (the
	// main update body doesn't carry owners). nil for resources without one.
	Owners *ownerMembershipBind
	// Bulk association syncs + slice member syncs.
	Assocs []*assocMembershipBind
	Labels *labelSyncBind
	// LabelsRespType is the 200 payload of Labels.Get, e.g. OULabelsResponse.
	LabelsRespType string
	SliceMembers   []*sliceMemberBind
	Moves          []moveBind
	HasRawMove     bool
	HasAccountMove bool
	IDTF           string
}

// qsPathMarker is the synthetic path segment fixspec inserts for query-string-
// discriminated operations (POST /v3/account?account-type=aws becomes
// /v3/account/__qs__/account-type/aws). Its presence is what identifies a
// literal-discriminator create; an ordinary create may also take a one-field
// params struct, but that field is a real query parameter.
const qsPathMarker = "/__qs__/"

// renderEntity renders an entity-archetype <name>.go from a ResourceModel.
func renderEntity(rm ResourceModel) ([]byte, error) {
	data, err := buildEntityData(rm)
	if err != nil {
		return nil, err
	}
	tmpl := entityTmpl
	if data.NoRead {
		tmpl = entityNoReadTmpl
	}
	t, err := template.New("entity").Parse(tmpl)
	if err != nil {
		return nil, fmt.Errorf("parse entity template: %w", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("execute entity template for %s: %w", rm.Name, err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated %s.go: %w\n%s", rm.Name, err, buf.Bytes())
	}
	return src, nil
}

// buildEntityData maps a ResourceModel onto the template payload, refusing the
// resource (error) when a field or id shape is outside the entity archetype.
func buildEntityData(rm ResourceModel) (entityData, error) {
	emptyColl := make(map[string]bool, len(rm.EmptyCollections))
	for _, a := range rm.EmptyCollections {
		emptyColl[a] = true
	}
	byTF := map[string]ModelField{rm.IDField.TFSDK: rm.IDField}
	for _, f := range rm.Fields {
		byTF[f.TFSDK] = f
	}

	d := entityData{
		MissingViaList:   rm.MissingViaList,
		Owners:           rm.Owners,
		Assocs:           rm.Assocs,
		Labels:           rm.Labels,
		LabelsRespType:   labelsRespType(rm.Labels),
		SliceMembers:     rm.SliceMembers,
		Moves:            rm.Moves,
		IDTF:             rm.IDField.TFSDK,
		UpdateIDExpr:     "idInt", // overwritten below when the op's param id type differs
		DeleteIDExpr:     "idInt",
		Pkg:              rm.Name,
		Pascal:           rm.Pascal,
		Model:            rm.Model,
		ResConst:         "ResName" + rm.Pascal,
		ResName:          rm.Pascal,
		TypeName:         "kion_" + rm.Name,
		IDGo:             rm.IDField.GoName,
		ParentRead:       rm.ParentRead,
		SDKAlias:         "generated",
		SDKImport:        sdkImport(),
		CreateMethod:     rm.Create.Method.Name,
		CreateBodyOpt:    rm.Create.Method.BodyType,
		ReadMethod:       rm.Read.Method.Name,
		ReadParams:       rm.Read.Method.ParamsType,
		RespType:         rm.Read.RespType,
		Gated:            rm.Gated,
		AtLeastOneOf:     rm.AtLeastOneOf,
		RequiredWhen:     rm.RequiredWhen,
		RequiredTogether: rm.RequiredTogether,
		RawDeletePath:    rm.RawDeletePath,
		ImportParentTF:   rm.ImportParentTF,
		SchemaVersion:    rm.SchemaVersion,
		ReadCompanion:    rm.ReadCompanion,
		Rewritten:        rm.Rewritten,
		CreateViaUpdate:  rm.CreateViaUpdate,
		RawCreate:        rm.RawCreate,
	}
	for _, m := range rm.Moves {
		d.HasRawMove = d.HasRawMove || m.RawPost != ""
		d.HasAccountMove = d.HasAccountMove || m.AccountMove
	}

	if rm.Create.Body == nil {
		return d, fmt.Errorf("%s: create op has no request body", rm.Name)
	}
	d.CreateBody = rm.Create.Body.Name
	d.CreateBodyPtr = rm.Create.Method.BodyPtr

	// Create-under-parent: the create params id comes from a model attribute (the
	// parent id), not a literal. Takes precedence over the __qs__ literal path.
	if rm.CreateParentParam != "" {
		d.CreateParams = rm.Create.Method.ParamsType
		d.CreateParentParam = rm.CreateParentParam
		d.CreateParentField = rm.CreateParentField
		d.ConfirmListed = rm.ConfirmListed
		d.IDIsParent = rm.IDIsParent
		d.ConfirmParents = rm.ConfirmParents
		d.CreateParentCast = rm.CreateParentCast
	} else if rm.Create.Method.ParamsType != "" && !strings.Contains(rm.Create.Method.Path, qsPathMarker) {
		// Params struct with no discriminator and no parent id. Its fields are
		// ordinary query parameters the resource does not set. The SDK signature
		// still requires the struct, so pass it empty.
		d.CreateParams = rm.Create.Method.ParamsType
	} else if rm.Create.Method.ParamsType != "" {
		// __qs__ factory-create routes carry a single-field params struct whose value
		// is a literal discriminator (the trailing path segment).
		//
		// Gated on the marker, not merely on "params has one field": an ordinary
		// create can also take a single-field params struct that is a real query
		// parameter (POST /v3/billing-source/gcp takes Unique OptBool), and
		// filling that with the trailing path segment emits `Unique: "gcp"`.
		if rm.Create.Params == nil || len(rm.Create.Params.Fields) != 1 {
			n := 0
			if rm.Create.Params != nil {
				n = len(rm.Create.Params.Fields)
			}
			return d, fmt.Errorf("%s: create params %s must have exactly one literal-discriminator field, got %d", rm.Name, rm.Create.Method.ParamsType, n)
		}
		d.CreateParams = rm.Create.Method.ParamsType
		d.CreateParamField = rm.Create.Params.Fields[0].GoName
		segs := strings.Split(strings.Trim(rm.Create.Method.Path, "/"), "/")
		d.CreateParamLiteral = segs[len(segs)-1]
	}

	d.NoRead = rm.Read.Method.Name == ""

	var err error
	if d.CreateBinds, d.CreateSliceBinds, err = bodyBinds(rm.Create.Body, byTF, rm.IDField.TFSDK, rm.CreateNested.Names, rm.Renames); err != nil {
		return d, fmt.Errorf("%s create: %w", rm.Name, err)
	}
	d.CreateObjBinds, d.CreateArrBinds, d.CreateFlatSubs = rm.CreateNested.Objs, rm.CreateNested.Arrs, rm.CreateNested.FlatSubs
	d.CreateRawValues = rm.CreateNested.RawValues
	if !d.NoRead {
		if d.ReadIDParam, d.IDParamType, err = idParamName(rm.Read.Params); err != nil {
			return d, fmt.Errorf("%s read: %w", rm.Name, err)
		}
	} else if rm.Delete != nil {
		// No read op: source the id Go type (for ParseInt/ParseUint) from delete.
		if _, d.IDParamType, err = idParamName(rm.Delete.Params); err != nil {
			return d, fmt.Errorf("%s delete: %w", rm.Name, err)
		}
	}
	stringID := d.IDParamType == "string"
	if stringID {
		if err := refuseStringIDFeatures(rm, d.NoRead); err != nil {
			return d, err
		}
		d.UpdateIDExpr, d.DeleteIDExpr = "id", "id"
	}

	// Prefer reading the id from the create response envelope when create returns
	// the record (not a bare *CreatedResponse). Falls back to errs.CreatedID.
	d.CreateIDMode = "created"
	if rm.Create.RespType != "" {
		for _, f := range rm.Create.RespFields {
			if f.JSONName != "id" {
				continue
			}
			opt, ok := idOptKind(f.Type)
			if stringID {
				opt, ok = stringIDOptKind(f.Type)
			}
			if !ok {
				break
			}
			d.CreateIDMode = "envelope"
			d.CreateRespType = rm.Create.RespType
			d.CreateDataPtr = rm.Create.RespDataPtr
			d.CreateIDOpt = opt
			d.CreateIDSDKName = f.GoName
			break
		}
	}
	if stringID && d.CreateIDMode != "envelope" {
		return d, fmt.Errorf("%s: a string id must be read from the create response's id field; errs.CreatedID yields an int64", rm.Name)
	}
	if len(rm.CreateOnly) > 0 {
		if d.CreateOnly, err = entityCreateOnly(rm, d); err != nil {
			return d, err
		}
	}

	if rm.Update != nil {
		if rm.Update.Body == nil {
			return d, fmt.Errorf("%s: update op has no request body", rm.Name)
		}
		d.HasUpdate = true
		d.UpdateMethod = rm.Update.Method.Name
		d.UpdateBodyOpt = rm.Update.Method.BodyType
		d.UpdateBody = rm.Update.Body.Name
		d.UpdateBodyPtr = rm.Update.Method.BodyPtr
		d.UpdateParams = rm.Update.Method.ParamsType
		if d.UpdateBinds, d.UpdateSliceBinds, err = bodyBinds(rm.Update.Body, byTF, rm.IDField.TFSDK, rm.UpdateNested.Names, rm.Renames); err != nil {
			return d, fmt.Errorf("%s update: %w", rm.Name, err)
		}
		d.UpdateObjBinds, d.UpdateArrBinds, d.UpdateFlatSubs = rm.UpdateNested.Objs, rm.UpdateNested.Arrs, rm.UpdateNested.FlatSubs
		d.UpdateImplode = rm.UpdateNested.Implode
		d.UpdateRawValues = rm.UpdateNested.RawValues
		var upType string
		if d.UpdateIDParam, upType, err = idParamName(rm.Update.Params); err != nil {
			return d, fmt.Errorf("%s update: %w", rm.Name, err)
		}
		if d.UpdateIDExpr, err = idExpr(upType, d.IDParamType); err != nil {
			return d, fmt.Errorf("%s update: %w", rm.Name, err)
		}
	}

	if rm.Delete != nil {
		d.HasDelete = true
		d.DeleteMethod = rm.Delete.Method.Name
		d.DeleteParams = rm.Delete.Method.ParamsType
		if rm.DeleteRecordParam != "" {
			// Asymmetric 2-param delete: the record id param plus an extra param
			// sourced from a model field (a parent id the other ops don't use).
			d.DeleteIDParam = rm.DeleteRecordParam
			d.DeleteExtraParam = rm.DeleteExtraParam
			mf, ok := byTF[rm.DeleteExtraField]
			if !ok {
				return d, fmt.Errorf("%s delete: delete_extra_field %q not in model", rm.Name, rm.DeleteExtraField)
			}
			d.DeleteExtraFieldGo = mf.GoName
		} else {
			var delType string
			if d.DeleteIDParam, delType, err = idParamName(rm.Delete.Params); err != nil {
				return d, fmt.Errorf("%s delete: %w", rm.Name, err)
			}
			if d.DeleteIDExpr, err = idExpr(delType, d.IDParamType); err != nil {
				return d, fmt.Errorf("%s delete: %w", rm.Name, err)
			}
		}
	}

	if !d.NoRead {
		topPrefix := "v.Data.Value."
		if rm.Read.RespDataPtr {
			topPrefix = "v.Data."
		}
		flattenSkip := createOnlySkip(rm.CreateOnly, rm.ReadNested.Names)
		if d.RespBinds, d.RespSliceBinds, err = respBinds(rm.Read.RespFields, byTF, rm.IDField.TFSDK, topPrefix, flattenSkip, emptyColl, stringID, rm.Renames); err != nil {
			return d, fmt.Errorf("%s flatten: %w", rm.Name, err)
		}
		// Record-wrapper: the payload nests the bulk of the record under a
		// sub-object (e.g. CFTWithOwnersAndTags.cft), while the model is flat.
		// Flatten those sub-fields from the wrapper path.
		if rm.Read.RespWrapperGo != "" {
			wrapPrefix := topPrefix + rm.Read.RespWrapperGo + "."
			if rm.Read.RespWrapperOpt {
				wrapPrefix += "Value."
			}
			wb, wsb, werr := respBinds(rm.Read.RespWrapperFields, byTF, rm.IDField.TFSDK, wrapPrefix, flattenSkip, emptyColl, stringID, rm.Renames)
			if werr != nil {
				return d, fmt.Errorf("%s flatten (wrapper %s): %w", rm.Name, rm.Read.RespWrapperGo, werr)
			}
			// A field the top-level payload already binds wins over the wrapper
			// (defensive: a true wrapper's fields don't appear at top level).
			bound := make(map[string]bool, len(d.RespBinds)+len(d.RespSliceBinds))
			for _, b := range d.RespBinds {
				bound[b.ModelGo] = true
			}
			for _, b := range d.RespSliceBinds {
				bound[b.ModelGo] = true
			}
			for _, b := range wb {
				if !bound[b.ModelGo] {
					d.RespBinds = append(d.RespBinds, b)
				}
			}
			for _, b := range wsb {
				if !bound[b.ModelGo] {
					d.RespSliceBinds = append(d.RespSliceBinds, b)
				}
			}
		}
		slices.SortFunc(d.RespBinds, func(a, b respBind) int { return cmp.Compare(a.ModelGo, b.ModelGo) })
		slices.SortFunc(d.RespSliceBinds, func(a, b sliceRespBind) int { return cmp.Compare(a.ModelGo, b.ModelGo) })
		d.RespObjFlats = rm.ReadNested.Objs
		ordered := map[string]bool{}
		for _, a := range rm.ReadNested.Arrs {
			a.OrderKey = rm.ListOrderKeys[a.TFSDK]
			if a.OrderKey != "" {
				if !slices.ContainsFunc(a.Subs, func(s flatSub) bool { return s.TF == a.OrderKey }) {
					return entityData{}, fmt.Errorf("%s list_order_keys: %s has no element attribute %q", rm.Name, a.TFSDK, a.OrderKey)
				}
				ordered[a.TFSDK] = true
			}
			d.RespArrFlats = append(d.RespArrFlats, a)
		}
		for tf := range rm.ListOrderKeys {
			if !ordered[tf] {
				return entityData{}, fmt.Errorf("%s list_order_keys: %s is not a nested-object list the read flattens", rm.Name, tf)
			}
		}
		for _, p := range rm.ReadNested.IDProjs {
			p.OrEmpty = emptyColl[p.TFSDK]
			d.RespIDProjs = append(d.RespIDProjs, p)
		}
		d.RespObjIDProjs = rm.ReadNested.ObjIDProjs
		d.RespSums = rm.ReadSums
		d.HasNestedFlat = len(d.RespObjFlats) > 0 || len(d.RespArrFlats) > 0
		d.HasIDProj = len(d.RespIDProjs) > 0 || len(d.RespObjIDProjs) > 0
		d.HasRespSlices = len(d.RespSliceBinds) > 0
		d.RespDataPtr = rm.Read.RespDataPtr
		d.RespRawValues = resolveRawValueFlats(rm.Read.RespFields, byTF, topPrefix, mergeRawValues(rm.CreateNested.RawValues, rm.UpdateNested.RawValues))
	}
	d.RawValueHelpers = mergeRawValues(d.CreateRawValues, d.UpdateRawValues)
	setIDShape(&d)
	if rm.IDIsParent && (d.CreateParentField == "" || d.IDParamType == "string") {
		return d, fmt.Errorf("%s: id_is_parent needs create_parent_field and an integer id", rm.Name)
	}
	if err := checkCreateViaUpdate(d); err != nil {
		return d, fmt.Errorf("%s: %w", rm.Name, err)
	}
	return d, nil
}

// checkCreateViaUpdate refuses a create_via_update the generated Create could
// not honor: the follow-up reuses the update body and its idInt, so each
// attribute must be a scalar that body carries and the id must be an integer.
func checkCreateViaUpdate(d entityData) error {
	if len(d.CreateViaUpdate) == 0 {
		return nil
	}
	if !d.HasUpdate {
		return fmt.Errorf("create_via_update needs an update operation")
	}
	if d.IDParamType == "string" {
		return fmt.Errorf("create_via_update needs an integer id")
	}
	for _, goName := range d.CreateViaUpdate {
		found := false
		for _, b := range d.UpdateBinds {
			if b.ModelGo == goName {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("create_via_update attribute %s is not a scalar the update body carries", goName)
		}
	}
	return nil
}

// entityCreateOnly binds the create_only attributes to the create response's
// payload, refusing a create that has none to capture from.
func entityCreateOnly(rm ResourceModel, d entityData) ([]createOnlyBind, error) {
	var why string
	switch {
	case d.NoRead:
		why = "a resource with no typed read"
	case rm.RawCreate != nil:
		why = "raw_create"
	case d.CreateIDMode != "envelope":
		why = "a create response that returns no record"
	default:
		base := "created.Data"
		if !d.CreateDataPtr {
			base += ".Value"
		}
		return createOnlyBinds(rm.Name, rm.CreateOnly, rm.Create.RespFields, base, d.SDKAlias)
	}
	return nil, fmt.Errorf("%s: create_only attribute %q cannot be captured with %s", rm.Name, rm.CreateOnly[0].TF, why)
}

// setIDShape derives the id's parse bits, diagnostic verb and local name from
// IDParamType.
func setIDShape(d *entityData) {
	d.IDParseBits = idParseBits(d.IDParamType)
	d.IDFormat = idFormat(d.IDParamType)
	d.IDVar = "idInt"
	if d.IDParamType == "string" {
		d.IDVar = "id"
	}
}

// idFormat is the fmt verb for a record id of the given param Go type.
func idFormat(idParamType string) string {
	if idParamType == "string" {
		return "%s"
	}
	return "%d"
}

// refuseStringIDFeatures rejects the entity features that address the record
// through an integer id; a string id would otherwise render code that cannot
// compile.
func refuseStringIDFeatures(rm ResourceModel, noRead bool) error {
	var why string
	switch {
	case noRead:
		why = "no typed read"
	case rm.RawCreate != nil:
		why = "raw_create"
	case rm.RawDeletePath != "":
		why = "a private raw delete"
	case rm.ReadCompanion != "":
		why = "read_companion"
	case rm.Labels != nil:
		why = "label sync"
	case rm.Owners != nil:
		why = "owner sync"
	case len(rm.Assocs) > 0:
		why = "association sync"
	case len(rm.SliceMembers) > 0:
		why = "slice member sync"
	case len(rm.Moves) > 0:
		why = "moves"
	case rm.DeleteRecordParam != "" || rm.ImportParentTF != "":
		why = "a parent-qualified delete"
	default:
		return nil
	}
	return fmt.Errorf("%s: string id is not supported with %s (those paths address the record by an integer id)", rm.Name, why)
}

// idParseBits is the strconv bit size the record id is parsed at.
//
// A uint64 id is parsed at 63, not 64, because one parse feeds two signednesses:
// the SDK's params want uint64 while a blended resource's private read and
// several mixed-type ops want int64 (see idExpr). strconv.ParseUint(s, 10, 64)
// accepts values above MaxInt64, and int64(...) wraps those negative -- an
// unbounded narrowing the generator has no business emitting, whatever Kion's
// ids actually reach. At 63 the value provably fits either type, and an id that
// large fails to parse with a clear error instead of silently going negative.
//
// An int64 id needs no such limit: nothing widens it back.
func idParseBits(idParamType string) int {
	if idParamType == "uint64" {
		return 63
	}
	return 64
}

// bodyBinds maps a request body's fields to model fields via json/tfsdk tags,
// skipping the id field (server-assigned / path param) and fields absent from
// the model. Refuses on an unconvertible scalar type.
// renames maps this resource's API attribute names to its provider ones; a
// renamed attribute is keyed in byTF under the provider name, so a lookup by
// the body field's own JSON name misses without it.
func bodyBinds(body *Struct, byTF map[string]ModelField, idTF string, skip map[string]bool, renames map[string]string) ([]fieldBind, []sliceBind, error) {
	var scalars []fieldBind
	var sliceOut []sliceBind
	for _, f := range body.Fields {
		mf, ok := byTF[f.JSONName]
		if !ok {
			// codegen/renames.yaml gave the attribute a different provider
			// name, so byTF is keyed under that instead.
			if tf, renamed := renames[f.JSONName]; renamed {
				mf, ok = byTF[tf]
			}
		}
		if !ok || mf.TFSDK == idTF || skip[f.JSONName] {
			continue
		}
		if conv, ok := normalizedConverter(f.Type, mf.Type, true); ok {
			scalars = append(scalars, fieldBind{SDKField: f.GoName, ModelGo: mf.GoName, Converter: conv})
			continue
		}
		if conv, ok := expandConverter(f); ok {
			scalars = append(scalars, fieldBind{SDKField: f.GoName, ModelGo: mf.GoName, Converter: conv})
			continue
		}
		if expand, _, wrap, ok := sliceConverter(f.Type, mf.Type); ok {
			sliceOut = append(sliceOut, sliceBind{SDKField: f.GoName, ModelGo: mf.GoName, Func: expand, Var: lowerFirst(f.GoName), Wrap: wrap})
			continue
		}
		return nil, nil, fmt.Errorf("field %q has unsupported type %q (needs crud_override)", f.JSONName, f.Type)
	}
	slices.SortFunc(scalars, func(a, b fieldBind) int { return cmp.Compare(a.SDKField, b.SDKField) })
	slices.SortFunc(sliceOut, func(a, b sliceBind) int { return cmp.Compare(a.SDKField, b.SDKField) })
	return scalars, sliceOut, nil
}

// idParamName returns the single id param field's name and Go type (int64,
// uint64, or string). Other shapes are refused (the caller degrades/skips).
func idParamName(p *Struct) (name, goType string, err error) {
	if p == nil {
		return "", "", fmt.Errorf("op has no params struct")
	}
	// The record id is the single required int64/uint64 field, or a required
	// string named like an id. Extra optional (Opt*) params, e.g. a query filter
	// like PolicyType/Inherited, are omitted (left zero). A second REQUIRED param
	// means a genuine parent+child op the entity path can't source, so refuse.
	var ids []Field
	for _, f := range p.Fields {
		switch {
		case f.Type == "int64" || f.Type == "uint64":
			ids = append(ids, f)
		case f.Type == "string" && strings.HasSuffix(f.GoName, "ID"):
			ids = append(ids, f)
		case strings.HasPrefix(f.Type, "Opt"):
			// optional query param, omitted in the call
		default:
			return "", "", fmt.Errorf("required non-id param %q has type %q", f.GoName, f.Type)
		}
	}
	if len(ids) != 1 {
		return "", "", fmt.Errorf("expected exactly one int64/uint64/string id param, got %d", len(ids))
	}
	return ids[0].GoName, ids[0].Type, nil
}

// idExpr returns the id expression for an op's params: the bare id local when
// the op's param id Go type matches the method's parse type (IDParamType), else
// a cast like "uint64(idInt)". Keeps single-id-type resources byte-identical
// while supporting mixed int64/uint64 op params (e.g. ou: read/update int64,
// delete uint64). A string id cannot be cast to or from an integer one.
func idExpr(opType, parseType string) (string, error) {
	if parseType == "string" && (opType == "" || opType == "string") {
		return "id", nil
	}
	if opType == "string" || parseType == "string" {
		return "", fmt.Errorf("string id mixed with a %s id across ops", cmp.Or(opType, parseType))
	}
	if opType == "" || opType == parseType {
		return "idInt", nil
	}
	return opType + "(idInt)", nil
}

// idOptKind reports whether an id field type is an Opt-wrapper (needs .Set/.Value
// access) and whether it is a handled id type at all.
func idOptKind(t string) (opt, ok bool) {
	switch t {
	case "OptUint64", "OptInt64":
		return true, true
	case "uint64", "int64":
		return false, true
	}
	return false, false
}

// stringIDOptKind is idOptKind for a string id.
func stringIDOptKind(t string) (opt, ok bool) {
	switch t {
	case "OptString":
		return true, true
	case "string":
		return false, true
	}
	return false, false
}

// respBinds maps response-payload fields to model fields, special-casing id.
// prefix is the Go access path to the fields (v.Data.Value. for an Opt envelope,
// v.Data. for a pointer envelope, or v.Data.Value.<Wrapper>.Value. for a record
// nested under a wrapper sub-object). Results are NOT sorted here, the caller
// merges top-level + wrapper binds and sorts once.
// emptyColl names attributes whose empty value the API returns as null; their
// flatten uses the OrEmpty converter so a configured empty collection does not
// come back null and diff for ever. stringID expects the id as an OptString.
func respBinds(fields []Field, byTF map[string]ModelField, idTF, prefix string, skip map[string]bool, emptyColl map[string]bool, stringID bool, renames map[string]string) ([]respBind, []sliceRespBind, error) {
	var out []respBind
	var sliceOut []sliceRespBind
	for _, f := range fields {
		mf, ok := byTF[f.JSONName]
		if !ok {
			// As in bodyBinds: a renamed attribute is keyed under its provider
			// name, and the read may spell it differently from the body.
			if tf, renamed := renames[f.JSONName]; renamed {
				mf, ok = byTF[tf]
			}
		}
		if !ok || skip[f.JSONName] {
			continue
		}
		if mf.TFSDK == idTF {
			var format string
			switch {
			case stringID && f.Type == "OptString":
			case stringID:
				return nil, nil, fmt.Errorf("id response field %q has type %q; expected OptString for a string id", f.JSONName, f.Type)
			case f.Type == "OptUint64" || f.Type == "uint64":
				format = "FormatUint"
			case f.Type == "OptInt64" || f.Type == "int64":
				format = "FormatInt"
			default:
				return nil, nil, fmt.Errorf("id response field %q has type %q; expected Opt(U)Int64", f.JSONName, f.Type)
			}
			out = append(out, respBind{ModelGo: mf.GoName, IsID: true, IDFormatFunc: format, IDSDKPath: prefix + f.GoName})
			continue
		}
		if conv, ok := normalizedConverter(f.Type, mf.Type, false); ok {
			out = append(out, respBind{ModelGo: mf.GoName, SDKPath: prefix + f.GoName, Converter: conv})
			continue
		}
		if conv, ok := flattenConverter(f); ok {
			out = append(out, respBind{ModelGo: mf.GoName, SDKPath: prefix + f.GoName, Converter: conv})
			continue
		}
		if _, flatten, wrap, ok := sliceConverter(f.Type, mf.Type); ok {
			path := prefix + f.GoName
			if wrap != "" {
				path += ".Value" // unwrap the nil-aware slice wrapper
			}
			if emptyColl[mf.TFSDK] {
				flatten += "OrEmpty"
			}
			sliceOut = append(sliceOut, sliceRespBind{ModelGo: mf.GoName, Func: flatten, SDKPath: path, Var: lowerFirst(f.GoName)})
			continue
		}
		return nil, nil, fmt.Errorf("response field %q has unsupported type %q", f.JSONName, f.Type)
	}
	return out, sliceOut, nil
}

// labelsRespType names the 200 payload of the label read: GetOULabels returns
// OULabelsResponse, GetAccountLabels returns AccountLabelsResponse.
func labelsRespType(l *labelSyncBind) string {
	if l == nil {
		return ""
	}
	return strings.TrimPrefix(l.Get, "Get") + "Response"
}

// rawCreateData is a create routed over raw HTTP on an otherwise-typed resource.
// The wire struct is flat and keyed by tfsdk names, like the raw archetype's.
type rawCreateData struct {
	Method  string // conns verb, e.g. "RawPost"
	Path    string
	Fields  []rawField
	Objects []rawCreateObject // nested attributes, declared in the archetype
	IDGo    string            // model field holding the id
}

// rawCreateObject is a nested attribute rendered as a nested wire struct.
type rawCreateObject struct {
	ModelGo string // model field, e.g. "OrganizationalUnit"
	JSON    string // wire key, e.g. "organizational_unit"
	GoType  string // the schema's Value type, e.g. "OrganizationalUnitValue"
	Subs    []rawCreateObjectSub
}

type rawCreateObjectSub struct {
	ModelGo string // sub-field on the Value type, e.g. "OrgUnitId"
	JSON    string // wire key, e.g. "org_unit_id"
}
