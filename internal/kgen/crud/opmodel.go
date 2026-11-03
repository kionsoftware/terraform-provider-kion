package crud

// ClientMethod is one SDK client method, e.g. PostLabel.
type ClientMethod struct {
	Name       string // "PostLabel"
	HTTPMethod string // "POST" (from the ogen doc-comment route)
	Path       string // "/v3/label"
	BodyType   string // "OptCreateLabel" / "OUNoteCreate"  ("" if no request body)
	BodyPtr    bool   // the body param is *T (pointer) rather than an Opt-wrapper
	ParamsType string // "GetLabelParams"  ("" if the op has no params struct)
	ResultType string // "PostLabelRes"
}

// Field is one field of an SDK body/params struct.
type Field struct {
	GoName   string // "Color"
	JSONName string // "color"        (from the json tag; "" falls back to GoName)
	Type     string // "string", "OptString", "OptUint64", "int64", ...
	Optional bool   // true iff Type is an Opt*-wrapper
	Ptr      bool   // true iff the field is a pointer (*T), e.g. a *Payload data envelope
}

// Struct is a parsed SDK struct (a request body, a response, or a params struct).
type Struct struct {
	Name   string
	Fields []Field
}

// ModelField is one field of a generated <Name>Model (from <name>_schema_gen.go).
type ModelField struct {
	GoName string // "Color", "Id"
	TFSDK  string // "color", "id"   (from the tfsdk tag)
	Type   string // "types.String", "types.Int64", "types.Bool", "types.List", "types.Set"
}

// OpModel is a fully-resolved operation ready for templating.
type OpModel struct {
	Verb        string       // "create" | "read" | "update" | "delete"
	Method      ClientMethod //
	Body        *Struct      // request body struct (nil if none)
	Params      *Struct      // params struct (nil if none)
	RespType    string       // success response type name for flatten (read), e.g. "LabelResponse"
	RespPayload string       // domain payload type inside the envelope, e.g. "Label"
	RespDataPtr bool         // the envelope's Data field is a pointer (*Payload), not an Opt-wrapper
	RespFields  []Field      // fields of the payload inside the envelope (Data.Value.* or Data.*)
	// Record wrapper: some read payloads nest the bulk of the record under a
	// sub-object (e.g. CFTWithOwnersAndTags.cft) alongside sibling collections
	// (owners, tags), while the model is flat. These name that sub-object and its
	// fields so the flatten can source the flat model from the wrapper path.
	RespWrapperGo     string  // wrapper field Go name, e.g. "Cft" ("" when none)
	RespWrapperOpt    bool    // the wrapper field is an Opt-wrapper (access .Value)
	RespWrapperFields []Field // the wrapper struct's fields
}

// listModel describes a resolved collection endpoint (the data-source read),
// enabling the dual-mode data source and the real sweeper. nil when unresolved.
//
// Two envelope shapes exist in the SDK and both are supported:
//
//	shape A: type LabelListPaginatedResponse struct { Data OptLabelListPaginated }
//	         type LabelListPaginated        struct { Items OptNilLabelArray; Total OptInt64 }
//	shape B: type UserListResponse          struct { Data []User }
//
// In shape A the Data field wraps a struct that holds the items slice (and
// usually a Total); in shape B the Data field IS the items slice. DataDirect
// distinguishes them.
type listModel struct {
	Method     ClientMethod // e.g. GetLabelIndex
	Params     *Struct      // e.g. GetLabelIndexParams (nil when the op has no params)
	RespType   string       // "LabelListPaginatedResponse"
	DataInner  string       // "LabelListPaginated" ("" for shape B)
	DataOpt    bool         // the Data field is an Opt-wrapper (guard on .Set, read .Value)
	DataDirect bool         // shape B: the Data field IS the items slice
	ItemsGo    string       // "Items" (field in DataInner); "Data" for shape B
	ItemsOpt   bool         // the items field is an Opt-wrapper, not a bare slice
	ItemsNil   bool         // items wrapper has a .Null field (OptNil*)
	ElemType   string       // "Label", the SDK list element type
	// ElemIsWrapper reports that ElemType is not the read payload but the read
	// payload's record-wrapper sub-object (e.g. the read returns
	// WebhookWithOwners while the index returns []Webhook, and WebhookWithOwners
	// carries the record under .Webhook). The data source then treats that
	// sub-object as the record, so one field mapping serves both modes.
	ElemIsWrapper bool
	TotalGo       string // "Total" ("" if absent)
	Paginated     bool   // the op takes Page+Count (else it is a single-shot list)
	PageParam     string // "Page" ("" when not paginated)
	CountParam    string // "Count" ("" when not paginated)
}

// access renders the Go expressions the list templates use to read one page of
// this envelope from a response variable named resp.
func (lm *listModel) access() listAccess {
	a := listAccess{
		Paginated: lm.Paginated,
		HasParams: lm.Method.ParamsType != "",
		ItemsExpr: "resp." + lm.ItemsGo,
	}
	if !lm.DataDirect {
		prefix := "resp.Data."
		if lm.DataOpt {
			prefix = "resp.Data.Value."
			a.EnvelopeGuard = "resp.Data.Set"
		}
		a.ItemsExpr = prefix + lm.ItemsGo
		if lm.TotalGo != "" {
			a.TotalExpr = prefix + lm.TotalGo
		}
	} else if lm.TotalGo != "" {
		a.TotalExpr = "resp." + lm.TotalGo
	}
	if lm.ItemsOpt {
		a.ItemsSlice = "items.Value"
		a.ItemsGuard = "items.Set"
		a.ItemsBreak = "!items.Set || len(items.Value) == 0"
		if lm.ItemsNil {
			a.ItemsGuard = "items.Set && !items.Null"
			a.ItemsBreak = "!items.Set || items.Null || len(items.Value) == 0"
		}
		return a
	}
	a.ItemsSlice = "items"
	a.ItemsBreak = "len(items) == 0"
	return a
}

// listAccess is the template-facing projection of a listModel's envelope: the
// Go expressions that walk one response page. Keeping them here rather than
// branching inside the templates lets one template body serve every shape.
type listAccess struct {
	Paginated     bool   // emit the Page/Count loop (else a single call)
	HasParams     bool   // the op takes a params struct at all
	EnvelopeGuard string // "resp.Data.Set" ("" when Data needs no guard)
	ItemsExpr     string // "resp.Data.Value.Items" | "resp.Data"
	ItemsGuard    string // "items.Set && !items.Null" ("" for a bare slice)
	ItemsSlice    string // "items.Value" | "items"
	ItemsBreak    string // last-page test, e.g. "len(items) == 0"
	TotalExpr     string // "resp.Data.Value.Total" ("" when the envelope has no total)
}

// ResourceModel is everything the templates need for one resource.
type ResourceModel struct {
	// ParentRead gives a no_read resource a real read over a private
	// collection; see parentread.go.
	ParentRead *parentReadData
	Name       string // snake, e.g. "label"
	// AtLeastOneOf are attributes codegen/config_validators.yaml says the API
	// requires at least one of, emitted as a resource-level ConfigValidator.
	AtLeastOneOf []string
	// RequiredWhen are value-conditional constraints from the same file.
	RequiredWhen []RequiredWhen
	// RequiredTogether are all-or-none attribute groups from the same file.
	RequiredTogether [][]string
	RawDeletePath    string
	ImportParentTF   string
	// EmptyCollections: see archetype.EmptyCollections.
	EmptyCollections []string
	// ListOrderKeys: see archetype.ListOrderKeys.
	ListOrderKeys map[string]string
	// Renames maps this resource's API attribute names to its provider ones.
	Renames map[string]string
	Pascal  string       // "Label"
	Model   string       // model type, e.g. "LabelModel"
	IDField ModelField   // the tfsdk:"id" model field
	Fields  []ModelField // model fields excluding id
	// RecordSubFields are the sub-attributes of a record-wrapper object attribute
	// (azure_policy's AzurePolicyValue{description,name,parameters,policy}),
	// promoted to top-level ModelFields. Non-empty only when the model nests the
	// whole record under one object attribute; see recordSubFields.
	RecordSubFields []ModelField
	Create          OpModel    //
	Read            OpModel    //
	Update          *OpModel   // nil if resource has no update
	Delete          *OpModel   // nil if resource has no delete
	List            *listModel // nil if no resolvable list endpoint
	// ListDowngrade explains why List is nil even though the config declared a
	// data-source collection read, i.e. why this data source lost its `filter`
	// block and fell back to id-only. Empty when List resolved (or when no
	// collection read was declared at all).
	ListDowngrade string
	// SweepList and SweepParent serve the sweeper alone: a parent-scoped
	// collection the data source cannot call (it has no parent id to pass) but a
	// sweeper can, by enumerating the parents first. Both nil for a flat
	// collection, where the sweeper reuses List.
	SweepList     *listModel
	SweepParent   *parentSweep
	Gated         bool // resource is version-gated (emit RequireKionVersionInRange)
	DSGated       bool // data source is version-gated (same vars, checked in Read)
	SchemaVersion int  // >0 when the resource migrates old SDKv2 state (bumps schema.Version)
	// Asymmetric 2-param delete (e.g. compliance_control's
	// DELETE /v4/compliance/program/{id}/control/{controlID}): the record-id
	// param plus an extra param sourced from a model field. All "" for the
	// common single-param delete.
	DeleteRecordParam string // delete param holding the record id (e.g. "ControlID")
	DeleteExtraParam  string // the other delete param (e.g. "ID")
	DeleteExtraField  string // model tfsdk field sourcing DeleteExtraParam (e.g. "program_id")
	// Create-under-parent: create's id param is the parent id from a model field
	// (not a literal). All "" for the common create. CreateParentField is the Go
	// model field name; CreateParentCast is a conversion prefix when not int64.
	CreateParentParam string
	CreateParentField string
	CreateParentCast  string
	// ConfirmListed: see the archetype field. ConfirmListedTF is the parent
	// attribute's tfsdk name, for the acceptance test's destroy check.
	ConfirmListed   string
	IDIsParent      bool // see the archetype field
	ConfirmListedTF string
	ConfirmParents  string
	MissingViaList  string // see the archetype field
	// Nested-object fields: create/update body objects expanded to SDK structs,
	// read-payload objects flattened into tfplugingen Value types.
	CreateNested nestedResult
	UpdateNested nestedResult
	ReadNested   nestedFlatResult
	// Scalars the read returns only as an expanded array, rebuilt by summing it
	// (declared by the archetype's sum_from).
	ReadSums []sumFlat
	// ReadCompanion is a hand-authored function called after every flatten to
	// read attributes the by-id GET does not return (archetype read_companion).
	ReadCompanion string
	// RawCreate routes only the create over raw HTTP; see the archetype key.
	RawCreate *rawCreateData
	// Rewritten holds model field names whose configured value Create and Update
	// restore after the read-back; see the archetype's Rewritten.
	Rewritten []string
	// CreateViaUpdate holds model field names the create ignores; see the
	// archetype's CreateViaUpdate.
	CreateViaUpdate []string
	// CreateOnly are attributes captured from the create response and never
	// flattened; see the archetype's CreateOnly.
	CreateOnly []createOnlyAttr
	// Owner association synced on Update via paired add/remove endpoints.
	Owners *ownerMembershipBind
	// Bulk association syncs (each a struct body of id-lists via one add/remove pair).
	Assocs []*assocMembershipBind
	// Labels is the per-resource label sub-resource, when the resource has one.
	Labels *labelSyncBind
	// Slice member syncs ([]int64 add/remove endpoints, e.g. user_group users).
	SliceMembers []*sliceMemberBind
	// Moves change an attribute through a dedicated endpoint before the PATCH.
	Moves []moveBind
}

// moveBind is one attribute the update body cannot carry, moved through its own
// endpoint: a raw POST whose body is the new value, or the account move.
type moveBind struct {
	ModelGo     string
	TF          string
	RawPost     string // path with an {id} placeholder
	AccountMove bool   // POST /v3/account/{id}/move; the account gets a new id
}

// sweepList is the collection the sweeper enumerates: the parent-scoped one
// when the archetype declares it, otherwise the data source's own.
func (rm ResourceModel) sweepList() *listModel {
	if rm.SweepList != nil {
		return rm.SweepList
	}
	return rm.List
}

// projectedFields is the model attribute set the data source and sweeper project
// onto the record: the flat model fields plus any promoted record-wrapper
// sub-attributes. Identical to Fields for every flat-model resource.
func (rm ResourceModel) projectedFields() []ModelField {
	if len(rm.RecordSubFields) == 0 {
		return rm.Fields
	}
	out := make([]ModelField, 0, len(rm.Fields)+len(rm.RecordSubFields))
	out = append(out, rm.Fields...)
	return append(out, rm.RecordSubFields...)
}

// Source is the AST boundary, mockable, reads only.
type Source interface {
	// ClientMethods parses oas_client_gen.go: one entry per (c *Client) method.
	ClientMethods(clientFile string) ([]ClientMethod, error)
	// Structs parses a schemas or parameters file into name->Struct.
	Structs(file string) (map[string]Struct, error)
	// MarkerImpls maps an unexported marker-method name (e.g. "getLabelRes")
	// to the concrete types that implement it (e.g. ["LabelResponse"]).
	MarkerImpls(schemaFile string) (map[string][]string, error)
	// ModelFields parses <name>_schema_gen.go for the named model struct.
	ModelFields(schemaGenFile, modelType string) ([]ModelField, error)
}
