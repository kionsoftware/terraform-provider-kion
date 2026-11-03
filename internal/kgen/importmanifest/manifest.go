// Package importmanifest derives, from the same codegen inputs the CRUD
// generator consumes, everything an enumerator needs to walk a live Kion
// install and emit Terraform `import` blocks: the list endpoint per resource,
// the read shape, and the import-id format.
//
// The import-id format exists nowhere else. internal/kgen/crud/*.gtpl generate
// ImportState per archetype -- entity.gtpl and parentlist.gtpl set id from
// req.ID directly, association.gtpl splits req.ID on "/" into (parent, key) --
// and a consumer outside this repo cannot know that. Deriving the manifest from
// the same archetype table that drives those templates is what keeps the two
// from diverging silently.
package importmanifest

// ManifestVersion is bumped when the JSON shape changes incompatibly.
const ManifestVersion = 2

// ReadShape is how a resource's records are enumerated.
type ReadShape string

// ReadShape constants.
const (
	ShapeGeneric    ReadShape = "generic"     // flat list at ListPath
	ShapeParentList ReadShape = "parent_list" // one collection per parent, fetched
	// with a second HTTP call per parent id
	ShapeAssociation ReadShape = "association" // membership rows under a parent
	// ShapeNestedCollection is a parent list whose records live inline in the
	// parent's own list payload -- e.g. /beta/scope returns scope objects, each
	// carrying a CriteriaRecords array of its criteria. There is no second HTTP
	// call: Collection names the JSON key of the nested array on each parent
	// record, and ParentIDField/ChildIDField name the fields on each nested
	// element that make up its "<parent>/<child>" import id.
	ShapeNestedCollection ReadShape = "nested_collection"
	ShapeSpecial          ReadShape = "special" // bespoke endpoint (singleton/raw)
	ShapeNone             ReadShape = "none"    // cannot be enumerated
)

// IDFormat is how a record's Terraform import id is built. It mirrors the
// ImportState implementation the matching crud template generates.
type IDFormat string

// IDFormat constants.
const (
	FormatID             IDFormat = "id"               // req.ID is the record id
	FormatParentSlashKey IDFormat = "parent_slash_key" // "<parent_id>/<key_id>"
	// FormatKindParentSlashKey is "<parent_kind>/<parent_id>/<key_id>", for a
	// resource whose parent may be one of several entity types and whose
	// ImportState therefore needs the type named as well as the id --
	// kion_custom_variable_override hangs off an account, an OU or a project,
	// and cvoverride.gtpl splits req.ID into exactly those three parts. Unlike
	// the two-part format the parent kind is load-bearing here, so Parent.Kind
	// must match the entity_type the resource's own CRUD dispatches on.
	FormatKindParentSlashKey IDFormat = "kind_parent_slash_key"
)

// Parent describes a parent-scoped or association read.
type Parent struct {
	// Kind is diagnostic only for every IDFormat except
	// FormatKindParentSlashKey, where it becomes the first segment of the
	// import id and must match the resource's own entity_type vocabulary.
	Kind          string `json:"kind"`
	ListPath      string `json:"list_path"`
	ChildPath     string `json:"child_path"` // contains "{parent_id}"
	ParentIDField string `json:"parent_id_field"`

	// ParentIDJSON is the record's own key for its owning parent, for a child
	// collection that also returns records other parents own. When set, the
	// enumerator takes the parent from this key rather than the path.
	ParentIDJSON string `json:"parent_id_json,omitempty"`

	// Match restricts which parents have their child collection read: a parent
	// is read only when every named field on its list record equals the given
	// value. /v3/idms/{id}/group-association answers only for a SAML IDMS and
	// fails for every other type, so reading the rest reports errors that say
	// nothing about the resource. Parents that do not match are skipped.
	// Authored as `match:` on the data source read in config_overrides.yaml.
	Match map[string]string `json:"match,omitempty"`

	// ChildRecordsKey names the key inside the child response's data object
	// that holds the records, for a route answering {"data":{"<key>":[...]}}
	// alongside other arrays the generic unwrapping cannot tell apart.
	ChildRecordsKey string `json:"child_records_key,omitempty"`
}

// ImportID describes how to build the id for an `import` block.
type ImportID struct {
	Format   IDFormat `json:"format"`
	KeyField string   `json:"key_field,omitempty"`
}

// Resource is one row of the manifest.
type Resource struct {
	TFType string `json:"tf_type"`
	// Kind is diagnostic only -- not consumed by the runtime enumerator (which
	// keys everything off TFType), kept so the generated JSON is legible
	// without cross-referencing tf_type back to a codegen kind by hand.
	Kind string `json:"kind"`
	// Archetype is diagnostic only -- not consumed by the runtime enumerator.
	// It records which crud_archetypes.yaml kind produced this row, useful
	// for tracing a derivation decision back to its source, but ReadShape and
	// ImportID are what the enumerator actually reads.
	Archetype string    `json:"archetype"`
	ReadShape ReadShape `json:"read_shape"`
	Readable  bool      `json:"readable"`
	ListPath  string    `json:"list_path,omitempty"`
	NameField string    `json:"name_field,omitempty"`
	Parent    *Parent   `json:"parent,omitempty"`

	// RequireValidField names a SQL-null-wrapper key that must be Valid for a
	// listed record to belong to this resource. Some collections mix kinds: the
	// cloud-access-role exemption endpoints return cloud RULE exemptions
	// alongside them, and only 6 of 22 records on a live install were actually
	// this type. Without the filter the other 16 imported as this type while
	// being something else.
	RequireValidField string `json:"require_valid_field,omitempty"`
	// Parents holds every parent set for a resource enumerable under more than
	// one parent collection (e.g. kion_budget under both /v3/ou and
	// /v3/project). Parent always mirrors Parents[0] when Parents is set, so a
	// reader that only knows about Parent still gets a correct (if partial) read.
	Parents []Parent `json:"parents,omitempty"`
	// Collection, ParentIDField, and ChildIDField are populated only when
	// ReadShape is ShapeNestedCollection. Collection is the JSON key (already
	// converted from the archetype's Go-style name, e.g. "criteria_records")
	// of the nested array on each record at ListPath. ParentIDField and
	// ChildIDField name the fields read off each element of that array to
	// build its "<parent>/<child>" import id.
	Collection    string   `json:"collection,omitempty"`
	ParentIDField string   `json:"parent_id_field,omitempty"`
	ChildIDField  string   `json:"child_id_field,omitempty"`
	ImportID      ImportID `json:"import_id"`
	Reason        string   `json:"reason,omitempty"`
	// AliasOf names the tf_type this one is a second name for. The provider
	// serves kion_aws_iam_policy and kion_iam_policy from one implementation over
	// one endpoint, so enumerating both reads the same objects twice and would
	// put two Terraform resources in charge of one Kion record. The row stays so
	// every tf_type the provider serves is still accounted for, but it is not
	// enumerated.
	AliasOf string `json:"alias_of,omitempty"`
}

// Manifest is the generated document. Resources are sorted by TFType so a
// regeneration diff means a real change.
type Manifest struct {
	Version   int        `json:"version"`
	Resources []Resource `json:"resources"`
}
