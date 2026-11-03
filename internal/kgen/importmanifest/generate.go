package importmanifest

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"gopkg.in/yaml.v3"

	"terraform-provider-kion/internal/kgen/kfs"
)

// OutputPath is where Generate writes, relative to the project root.
const OutputPath = "codegen/import_manifest.json"

// kindAliases covers the two kinds the provider serves under a tf_type that
// codegen does not use for its resources/data_sources keys: the schema
// snapshot strips kion_aws_cloudformation_template / kion_aws_iam_policy down
// to "aws_cloudformation_template" / "aws_iam_policy", but
// generator_config.yaml keys their entries "cft" / "iam_policy". Resolve
// through this before any generator_config.yaml lookup.
var kindAliases = map[string]string{
	"aws_cloudformation_template": "cft",
	"aws_iam_policy":              "iam_policy",
}

// archetypeInfo is the slice of a crud_archetypes.yaml entry Build needs: the
// archetype kind, the two association fields (association.gtpl reads the
// same two off the same struct — see internal/kgen/crud/assoc.go), and the
// three compound_key_parent_read fields (internal/kgen/crud/compound.go's
// archetype struct reads the same three, under the same yaml names) that
// ShapeNestedCollection carries into the manifest row.
type archetypeInfo struct {
	Kind        string
	KeyField    string
	ParentField string
	// Collection, ParentIDField, ChildIDField are populated only for
	// compound_key_parent_read. Collection is copied verbatim from
	// crud_archetypes.yaml (Go-style, e.g. "CriteriaRecords") -- Build
	// converts it to the JSON key the manifest carries.
	Collection    string
	ParentIDField string
	ChildIDField  string
	// DeleteExtraField marks an entity whose delete addresses the record
	// through a parent the read does not return, so its import id is compound.
	DeleteExtraField string
}

// parentOverrides corrects resources, verified against a real install, where
// placeholder-based derivation picks a parent list that isn't actually
// listable:
//
//	/v4/idms/open-id                     -> 405 (not a listable collection)
//	/v4/idms/open-id/28                  -> 200 (single OpenID-type IDMS)
//	/v4/idms/open-id/28/access-rule       -> 200
//	/v4/idms/open-id/28/group-association -> 200
//	/v3/idms                             -> 200, and it enumerates them
//
// OpenID records are keyed by the IDMS id, not their own -- so the child
// paths derivation would produce are correct (a single-object GET for
// kion_idms_open_id itself, which the reader already handles; the
// create-shaped, not read-shaped, collection endpoints for its two
// sub-resources), only the parent LIST is wrong. This table overrides just
// Kind/ListPath/ParentIDField; ChildPath is the live-verified value.
//
// The two *_cloud_access_role_exemption entries have the same root cause: the
// chosen read path resolves (via private_endpoints.yaml's resources: section,
// since crud_archetypes.yaml declares both no_read) to the internal
// /v1/ou/{id}/ou-cloud-access-role and /v1/project/{id}/ou-cloud-access-role
// reads, so unaided placeholder-based derivation would pick /v1/ou and
// /v1/project as the parent list -- neither is the real, listable collection.
// /v3/ou and /v3/project are. The child route returns the parent's cloud
// access roles with the exemptions beside them, hence ChildRecordsKey.
var parentOverrides = map[string]Parent{
	"kion_idms_open_id": {
		Kind: "idms", ListPath: "/v3/idms",
		ChildPath: "/v4/idms/open-id/{parent_id}", ParentIDField: "idms_id",
	},
	"kion_idms_open_id_access_rule": {
		Kind: "idms", ListPath: "/v3/idms",
		ChildPath: "/v4/idms/open-id/{parent_id}/access-rule", ParentIDField: "idms_id",
	},
	"kion_idms_open_id_group_association": {
		Kind: "idms", ListPath: "/v3/idms",
		ChildPath: "/v4/idms/open-id/{parent_id}/group-association", ParentIDField: "idms_id",
	},
	"kion_ou_cloud_access_role_exemption": {
		Kind: "ou", ListPath: "/v3/ou",
		ChildPath: "/v1/ou/{parent_id}/ou-cloud-access-role", ParentIDField: "ou_id",
		ChildRecordsKey: "ou_exemptions",
	},
	"kion_project_cloud_access_role_exemption": {
		Kind: "project", ListPath: "/v3/project",
		ChildPath: "/v1/project/{parent_id}/ou-cloud-access-role", ParentIDField: "project_id",
		ChildRecordsKey: "project_exemptions",
	},
	// GET /v3/billing-source 405s (POST only); the collection is /v4.
	"kion_billing_source_govcloud": {
		Kind: "billing_source", ListPath: "/v4/billing-source",
		ChildPath: "/v3/billing-source/{parent_id}/govcloud", ParentIDField: "billing_source_id",
	},
}

// importStateSplits reports whether an archetype's generated ImportState parses
// "<parent>/<key>" rather than a bare id.
//
// This is the one fact that decides whether a resource may carry a compound
// import id, and it lived in two places that disagreed. Classify guarded on
// parent_list and compound_key_parent_read; the parentOverrides block below set
// the compound format unconditionally, on the assumption that having a parent
// implied a splitting ImportState. It does not: idms_open_id,
// idms_open_id_access_rule and idms_open_id_group_association are entity
// archetypes reached through a parent, and entity.gtpl emits an id-only
// ImportState. They were handed "28/2", Read parsed the whole string as an
// integer, -generate-config-out wrote no configuration for them, and
// `terraform apply` then failed the entire import with "Configuration for
// import target does not exist".
//
// no_read is in the list because its template does split -- which is why the
// two exemption resources, also reached by override, have always worked.
func importStateSplits(archetype string) bool {
	switch archetype {
	case "parent_list", "compound_key_parent_read", "no_read", "association":
		return true
	}
	return false
}

// requireValidOverrides records the discriminator for collections that mix
// resource kinds; see Resource.RequireValidField. Kept beside parentOverrides
// because it is the same class of authored, live-verified knowledge the schema
// cannot supply.
// kion_custom_variable_override's collections list every custom variable
// visible at the entity, most of them merely inherited; "override" is non-null
// on exactly the ones actually set there. Measured on a demo install: 151
// entities x 12 variables, 21 non-null.
// kion_billing_source is the CUSTOM billing source -- it creates through
// /v3/billing-source/custom and reads /v1/payer/{id}, unwrapping
// custom_billing_source. Its list, /v4/billing-source, is the polymorphic
// collection of every payer type, so without a discriminator it claims records
// it cannot represent: the by-id read then finds no custom_billing_source key,
// flattens everything to zero, and the resource refreshes as id="0". Every
// payer_id pointing at one of those resolves to 0, which on one install planned
// 6,611 accounts to be unlinked from their billing source. Measured across five
// installs, the count reading a real id equals the custom count exactly. The
// AWS, GCP and OCI billing sources list the same collection and need the same
// filter on their own key.
var requireValidOverrides = map[string]string{
	"kion_custom_variable_override": "override",
	"kion_billing_source":           "custom_billing_source",
	"kion_billing_source_aws":       "aws_payer",
	"kion_billing_source_gcp":       "gcp_payer",
	"kion_billing_source_oci":       "oci_payer",
	"kion_billing_source_azure":     "azure_payer",
	"kion_billing_source_anthropic": "anthropic_billing_source",
}

// multiParentOverrides corrects resources enumerable under more than one
// parent collection -- Build's per-resource derivation models exactly one
// Parent, but /v3/budget 405s (there is no flat list) and budgets hang off
// two parents, both verified live: /v3/ou/{id}/budget and
// /v3/project/{id}/budget.
//
// kion_custom_variable_override is the same shape with a third parent. Its
// by-id read takes two ids (/v3/account/{account_id}/custom-variable/{custom_variable_id}),
// which is why derivation gives up on it -- but the SINGULAR read has a plural
// sibling, /v3/{entity}/{id}/custom-variable, that lists every variable at that
// entity and needs only the entity id. Walking the three entity collections
// therefore enumerates the whole cross product in one request per entity rather
// than one per (entity, variable) pair. Each Kind doubles as the resource's own
// entity_type; see FormatKindParentSlashKey.
var multiParentOverrides = map[string][]Parent{
	"kion_budget": {
		{Kind: "ou", ListPath: "/v3/ou", ChildPath: "/v3/ou/{parent_id}/budget", ParentIDField: "ou_id"},
		{Kind: "project", ListPath: "/v3/project", ChildPath: "/v3/project/{parent_id}/budget", ParentIDField: "project_id"},
	},
	"kion_custom_variable_override": {
		{Kind: "account", ListPath: "/v3/account", ChildPath: "/v3/account/{parent_id}/custom-variable", ParentIDField: "entity_id"},
		{Kind: "ou", ListPath: "/v3/ou", ChildPath: "/v3/ou/{parent_id}/custom-variable", ParentIDField: "entity_id"},
		{Kind: "project", ListPath: "/v3/project", ChildPath: "/v3/project/{parent_id}/custom-variable", ParentIDField: "entity_id"},
	},
}

// importIDOverrides fixes the import id for a resource the override tables
// above rescue from ShapeNone, where Classify already ran and returned the
// plain id for a resource it believed unreadable. Only
// kion_custom_variable_override needs it: multiParentOverrides establishes the
// read but not the three-part id its ImportState splits on.
var importIDOverrides = map[string]ImportID{
	"kion_custom_variable_override": {Format: FormatKindParentSlashKey, KeyField: "custom_variable_id"},
}

// explicitParentOverrides is parentOverrides' sibling for a resource whose
// parent-scoped read appears in NEITHER generator_config.yaml NOR
// private_endpoints.yaml -- there is nothing for placeholder-based derivation
// to start from at all, unlike parentOverrides' idms_open_id family, which
// derives a parent-scoped shape but picks the wrong (unlistable) parent list.
// Verified live: kion_funding_source_note's create/read/update/delete are all
// internal /v2, and unlike the note family's other two members (ou_note is
// fully public; project_note is blended, both recorded in
// private_endpoints.yaml), funding_source_note is absent from that file
// entirely.
var explicitParentOverrides = map[string]Parent{
	"kion_funding_source_note": {
		Kind: "funding_source", ListPath: "/v3/funding-source",
		ChildPath: "/v2/funding-source/{parent_id}/funding-source-note", ParentIDField: "funding_source_id",
	},
}

// explicitFlatListOverrides names a flat list endpoint verified live against a
// real install that appears in neither generator_config.yaml nor
// private_endpoints.yaml. kion_dashboard's public create is untyped
// (private_endpoints.yaml records the by-id read at private /v1/dashboard/{id}
// for that reason) but the plural collection, /v1/dashboards, returns every
// dashboard fully typed and is not recorded anywhere in either codegen input.
//
// kion_billing_source_azure's and _anthropic's generator_config reads are schema-only keys, not a
// route (see spec_additions.yaml); like the other typed billing sources it is
// listed from the polymorphic /v4/billing-source and filtered by azure_payer.
//
// kion_billing_rule's /v3/billing-rule answers 400 "invalid rule type" for the
// whole page once any rule is an allocation or custom line item, which the v3
// shape cannot carry. The filter lists only the types the resource manages.
var explicitFlatListOverrides = map[string]string{
	"kion_billing_rule":             "/v3/billing-rule?rule_type_ids=1&rule_type_ids=2&rule_type_ids=3",
	"kion_dashboard":                "/v1/dashboards",
	"kion_billing_source_azure":     "/v4/billing-source",
	"kion_billing_source_anthropic": "/v4/billing-source",
}

// Build assembles the manifest from already-loaded inputs. Pure: no I/O, so the
// table-driven tests do not need a filesystem.
//
//	readPaths           kind -> by-id read path (generator_config.yaml resources:)
//	dataSourcePaths     kind -> list read path  (generator_config.yaml data_sources:)
//	privateListPaths    kind -> flat list path  (private_endpoints.yaml list_read_only:)
//	privateResourcePaths kind -> by-id read path (private_endpoints.yaml resources:)
//	archetypes          kind -> archetype fields (crud_archetypes.yaml: kind, key_field, parent_field)
//	tfTypes             every managed resource the provider serves (schema snapshot)
//	hasNameAttr         tf_type -> whether the schema snapshot declares a top-level "name" attribute
//
// For each resource the CHOSEN PATH is resolved in this order, taking the
// first one that is non-empty:
//
//  1. privateListPaths[kind] -- private_endpoints.yaml's list_read_only:
//     section. Highest priority: it is the one input that names a real,
//     already-flat collection endpoint outright (e.g. kion_aws_resource_tag's
//     /v3/aws-resource-tag), rather than a by-id read that still needs its
//     trailing placeholder stripped.
//  2. dataSourcePaths[kind] -- generator_config.yaml's data_sources: section.
//     When present it is the real collection endpoint, authoritative over any
//     by-id read.
//  3. readPaths[kind] -- generator_config.yaml's resources: section (the by-id
//     read every resource records).
//  4. privateResourcePaths[kind] -- private_endpoints.yaml's resources:
//     section, consulted last since generator_config.yaml already carries the
//     resolved read for anything the CRUD generator itself reads by id; this
//     only fires for a kind absent from generator_config.yaml entirely.
//
// The chosen path's placeholder count then drives shape/list-path/parent,
// independent of archetype except where crud_archetypes.yaml fixes the shape
// outright (no_read, datasource_only, association, parent_list,
// compound_key_parent_read, singleton, raw_http, cv_override -- Classify's
// job, unchanged):
//
//   - 0 placeholders, or exactly 1 that is the path's final segment (a plain
//     by-id GET, e.g. "/v3/app-role/{id}"): flat. ListPath is the path with
//     that trailing placeholder stripped.
//   - exactly 1 placeholder NOT at the end (e.g. "/v3/ou/{id}/ou-note"):
//     parent-scoped. The text before the placeholder is the parent's list
//     path; its segments after the /vN version prefix, joined with "_" (with
//     any "-" also folded to "_"), plus "_id" name the parent id field.
//   - 2+ placeholders (kion_custom_variable_override's
//     /v3/account/{account_id}/custom-variable/{custom_variable_id}): not
//     enumerable -- neither id can be discovered without the other, so the
//     resource is unreadable regardless of what its archetype would otherwise
//     allow.
func Build(readPaths, dataSourcePaths, privateListPaths, privateResourcePaths map[string]string, archetypes map[string]archetypeInfo, tfTypes []string, hasNameAttr map[string]bool) *Manifest {
	out := make([]Resource, 0, len(tfTypes))

	for _, tfType := range tfTypes {
		kind := strings.TrimPrefix(tfType, "kion_")
		lookupKind := kind
		if alias, ok := kindAliases[kind]; ok {
			lookupKind = alias
		}

		// An alias reads the same endpoint as its canonical type. Record the
		// relationship and stop: enumerating both would double-import.
		if canonical, ok := kindAliases[kind]; ok {
			out = append(out, Resource{
				TFType:   tfType,
				Kind:     kind,
				AliasOf:  "kion_" + canonical,
				Readable: false,
				Reason:   "second tf_type for kion_" + canonical + "; importing both would put two resources in charge of one record",
			})
			continue
		}

		info := archetypes[lookupKind]
		archetype := info.Kind
		if archetype == "" {
			archetype = "entity"
		}

		// hasParent decides the import-id FORMAT. It must come from
		// crud_archetypes.yaml's declared parent_field, not from whether a
		// Parent block gets attached below -- internal/kgen/crud/assoc.go
		// derives association.gtpl's HasParent the same way (arch.ParentField
		// != "", erroring "parent_field %q not in model" otherwise), and the
		// manifest exists to mirror that template.
		hasParent := info.ParentField != ""

		chosenPath := privateListPaths[lookupKind]
		if chosenPath == "" {
			chosenPath = dataSourcePaths[lookupKind]
		}
		if chosenPath == "" {
			chosenPath = readPaths[lookupKind]
		}
		if chosenPath == "" {
			chosenPath = privateResourcePaths[lookupKind]
		}

		r := Resource{
			TFType:    tfType,
			Kind:      kind,
			Archetype: archetype,
		}
		if hasNameAttr[tfType] {
			r.NameField = "name"
		}

		placeholders := strings.Count(chosenPath, "{")

		switch {
		case archetype == "datasource_only":
			// Fixed regardless of path: never a readable collection.
			r.ReadShape, r.ImportID.Format, r.Readable, r.Reason = Classify(archetype, chosenPath, hasParent)

		case placeholders >= 2:
			// Two ids in one path (custom_variable_override's
			// .../custom-variable/{custom_variable_id} under
			// account/{account_id}) can't be enumerated from either id alone.
			// This overrides even an archetype Classify would otherwise fix
			// readable (cv_override). Note what this is NOT: the API does
			// have a real read here (verified public,
			// GET .../custom-variable/{custom_variable_id}) -- the resource's
			// identity is simply compound, not missing.
			r.ReadShape = ShapeNone
			r.ImportID.Format = FormatID
			r.Readable = false
			r.Reason = fmt.Sprintf("%s has a compound identity (%s) -- both ids are needed to reach a record and neither can be discovered from the other, so it is not enumerable from a flat list even though the read itself exists", chosenPath, strings.Join(placeholderNames(chosenPath), ", "))

		case placeholders == 1 && !isTrailingPlaceholder(chosenPath):
			parent := parentFromPath(chosenPath)
			r.Parent = &parent
			r.Readable = true
			r.ImportID.Format = FormatID
			if archetype == "association" {
				r.ReadShape = ShapeAssociation
				if hasParent {
					r.ImportID.Format = FormatParentSlashKey
				}
			} else {
				r.ReadShape = ShapeParentList
				// Only an archetype whose ImportState actually splits on "/" may
				// take a compound id. Reaching a record through a parent is not
				// enough: compliance_family and compliance_level declare no
				// archetype, so they get entity.gtpl's id-only ImportState and
				// merely fall back to a parent-scoped read when their flat list
				// 405s. Handing those "<parent>/<id>" made Read parse the whole
				// string as an integer, which is 652 of the 667 "Invalid ID"
				// failures a verification run produced.
				if importStateSplits(archetype) || info.DeleteExtraField != "" {
					r.ImportID.Format = FormatParentSlashKey
					if r.ImportID.KeyField == "" {
						r.ImportID.KeyField = "id"
					}
				}
			}

		default:
			// 0 placeholders, or a single trailing one (a plain by-id GET
			// that collapses to a flat list the same way ListPathFrom always
			// has). Classify's archetype-fixed branches (association,
			// parent_list, compound_key_parent_read, singleton, raw_http,
			// cv_override) and its default entity/blended/cloud_account
			// bucket both apply unchanged.
			r.ReadShape, r.ImportID.Format, r.Readable, r.Reason = Classify(archetype, chosenPath, hasParent)
			r.ListPath = ListPathFrom(chosenPath)
		}

		// association.gtpl's ImportState only splits req.ID on "/" in its
		// {{if .HasParent}} branch; the {{else}} branch (parentless
		// associations, e.g. kion_global_permission_mapping) parses req.ID as
		// a plain integer and assigns it directly to the KEY field (not a
		// synthesized id) -- so for a parentless association the import id
		// IS the key field's value, and the manifest must still carry
		// KeyField so the enumerator knows which raw field to read it from.
		// Format stays FormatID either way; only whether KeyField is set
		// depends on the association having a key field at all.
		if r.ReadShape == ShapeAssociation {
			r.ImportID.KeyField = info.KeyField
		}

		// ShapeNestedCollection's records live inline in the parent payload
		// (Classify never attaches a Parent block for it, unlike
		// ShapeParentList), so the runtime enumerator needs the collection key
		// and the two id fields carried explicitly rather than re-deriving
		// them. Collection is converted from crud_archetypes.yaml's Go-style
		// name (e.g. "CriteriaRecords") to the JSON key the live payload
		// actually uses ("criteria_records") here, once, at generation time --
		// not by the reader on every enumeration.
		if r.ReadShape == ShapeNestedCollection {
			r.Collection = jsonKeyFromGoName(info.Collection)
			r.ParentIDField = info.ParentIDField
			r.ChildIDField = info.ChildIDField
		}

		// Applied after derivation: these two tables replace a wrongly
		// derived parent (or a flat classification that never had one) with
		// the live-verified shape. See their doc comments for the evidence.
		if ov, ok := parentOverrides[tfType]; ok {
			p := ov
			r.Parent = &p
			r.ListPath = ""
			r.ReadShape = ShapeParentList
			r.Readable = true
			r.Reason = ""
			// The override is what establishes the parent, so the import-id
			// format has to be recomputed against it: Classify ran before this
			// with hasParent false and returned the plain id, but the resource's
			// ImportState splits on "/". A bare id leaves the parent zero and
			// every read misses.
			// Only when the ImportState actually splits; see importStateSplits.
			// An entity reached through a parent still gets an id-only
			// ImportState, and handing it a compound id makes the resource
			// unimportable outright.
			if importStateSplits(archetype) {
				r.ImportID.Format = FormatParentSlashKey
				if r.ImportID.KeyField == "" {
					r.ImportID.KeyField = "id"
				}
			}
		}

		if field, ok := requireValidOverrides[tfType]; ok {
			r.RequireValidField = field
		}

		if parents, ok := multiParentOverrides[tfType]; ok {
			ps := make([]Parent, len(parents))
			copy(ps, parents)
			r.Parents = ps
			r.Parent = &ps[0]
			r.ListPath = ""
			r.ReadShape = ShapeParentList
			r.Readable = true
			r.Reason = ""
		}

		// explicitParentOverrides/explicitFlatListOverrides: hand-verified
		// paths present in neither generator_config.yaml nor
		// private_endpoints.yaml. See their doc comments.
		if ov, ok := explicitParentOverrides[tfType]; ok {
			p := ov
			r.Parent = &p
			r.Parents = nil
			r.ListPath = ""
			r.ReadShape = ShapeParentList
			r.Readable = true
			r.Reason = ""
		}

		if listPath, ok := explicitFlatListOverrides[tfType]; ok {
			r.Parent = nil
			r.Parents = nil
			r.ListPath = listPath
			r.ReadShape = ShapeGeneric
			r.Readable = true
			r.Reason = ""
		}

		// Last, so it wins over whatever shape-based derivation or the
		// override tables above computed.
		if id, ok := importIDOverrides[tfType]; ok {
			r.ImportID = id
		}

		out = append(out, r)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].TFType < out[j].TFType })
	return &Manifest{Version: ManifestVersion, Resources: out}
}

// isTrailingPlaceholder reports whether path's single "{...}" placeholder is
// its final path segment (a plain by-id GET, e.g. "/v3/app-role/{id}") rather
// than followed by a nested collection segment (e.g. "/v3/ou/{id}/ou-note").
func isTrailingPlaceholder(path string) bool {
	open := strings.Index(path, "{")
	if open < 0 {
		return false
	}
	closeRel := strings.Index(path[open:], "}")
	if closeRel < 0 {
		return false
	}
	return open+closeRel+1 == len(path)
}

// placeholderNames returns every "{...}" placeholder name in path, in order,
// e.g. "/v3/account/{account_id}/custom-variable/{custom_variable_id}" ->
// ["account_id", "custom_variable_id"].
func placeholderNames(path string) []string {
	var names []string
	rest := path
	for {
		open := strings.Index(rest, "{")
		if open < 0 {
			break
		}
		closeRel := strings.Index(rest[open:], "}")
		if closeRel < 0 {
			break
		}
		names = append(names, rest[open+1:open+closeRel])
		rest = rest[open+closeRel+1:]
	}
	return names
}

// parentFromPath builds a Parent block from a chosen path containing exactly
// one non-trailing "{...}" placeholder, e.g. "/v3/ou/{id}/ou-note" or
// "/v4/compliance/program/{id}/control".
func parentFromPath(path string) Parent {
	open := strings.Index(path, "{")
	closeRel := strings.Index(path[open:], "}")
	closeIdx := open + closeRel + 1

	listPath := strings.TrimSuffix(path[:open], "/")
	childPath := path[:open] + "{parent_id}" + path[closeIdx:]

	// Parent id field: the list path's segments after the /vN (or /beta)
	// version prefix, joined with "_" and any "-" folded to "_", plus "_id".
	// "/v3/ou" -> "ou_id"; "/v4/compliance/program" -> "compliance_program_id";
	// "/v3/funding-source" -> "funding_source_id".
	segments := strings.Split(strings.TrimPrefix(listPath, "/"), "/")
	idField := strings.ReplaceAll(strings.Join(segments[1:], "_"), "-", "_") + "_id"

	return Parent{
		Kind:          strings.TrimSuffix(idField, "_id"),
		ListPath:      listPath,
		ChildPath:     childPath,
		ParentIDField: idField,
	}
}

// jsonKeyFromGoName folds a Go exported field name (as crud_archetypes.yaml's
// collection: key names it, e.g. "CriteriaRecords") into the snake_case JSON
// key the live API actually uses ("criteria_records"): an underscore is
// inserted before every uppercase letter that is not the first character,
// and the whole string is lowercased. This is a plain camel/Pascal->snake
// fold, not an acronym-aware one -- a run of consecutive capitals (e.g. "ID")
// would split as "i_d" rather than staying "id". No field crud_archetypes.yaml
// currently declares as a collection has that shape, but a future one might;
// this limitation is intentional to keep the fold simple and predictable.
func jsonKeyFromGoName(name string) string {
	if name == "" {
		return ""
	}
	var b strings.Builder
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// Generate reads the codegen inputs from root and writes OutputPath.
func Generate(fsw kfs.FS, root string) (*Manifest, error) {
	readPaths, err := loadReadPaths(fsw, filepath.Join(root, "codegen", "generator_config.yaml"))
	if err != nil {
		return nil, err
	}
	dataSourcePaths, err := loadDataSourceReadPaths(fsw, filepath.Join(root, "codegen", "generator_config.yaml"))
	if err != nil {
		return nil, err
	}
	privateListPaths, err := loadPrivateListReadOnlyPaths(fsw, filepath.Join(root, "codegen", "private_endpoints.yaml"))
	if err != nil {
		return nil, err
	}
	privateResourcePaths, err := loadPrivateResourceReadPaths(fsw, filepath.Join(root, "codegen", "private_endpoints.yaml"))
	if err != nil {
		return nil, err
	}
	archetypes, err := loadArchetypeKinds(fsw, filepath.Join(root, "codegen", "crud_archetypes.yaml"))
	if err != nil {
		return nil, err
	}
	tfTypes, hasNameAttr, err := loadTFTypes(fsw, filepath.Join(root, "codegen", "schema_snapshots", "new.json"))
	if err != nil {
		return nil, err
	}

	matches, err := loadDataSourceMatches(fsw, filepath.Join(root, "codegen", "generator_config.yaml"))
	if err != nil {
		return nil, err
	}
	attrs, err := loadResourceAttrs(fsw, filepath.Join(root, "codegen", "schema_snapshots", "new.json"))
	if err != nil {
		return nil, err
	}

	m := Build(readPaths, dataSourcePaths, privateListPaths, privateResourcePaths, archetypes, tfTypes, hasNameAttr)
	if err := applyParentMatches(m, matches, attrs); err != nil {
		return nil, err
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	data = append(data, '\n')
	if err := fsw.WriteFile(filepath.Join(root, OutputPath), data, 0o644); err != nil {
		return nil, err
	}
	return m, nil
}

func loadReadPaths(fsw kfs.FS, path string) (map[string]string, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		Resources map[string]struct {
			Read struct {
				Path string `yaml:"path"`
			} `yaml:"read"`
		} `yaml:"resources"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]string, len(doc.Resources))
	for kind, entry := range doc.Resources {
		out[kind] = entry.Read.Path
	}
	return out, nil
}

// loadDataSourceReadPaths reads generator_config.yaml's data_sources: section,
// mirroring loadReadPaths -- the same shape one level down. Unlike a
// resource's by-id read, a data source's read path is the real list/collection
// endpoint, so it is preferred over the resource fallback whenever present.
func loadDataSourceReadPaths(fsw kfs.FS, path string) (map[string]string, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		DataSources map[string]struct {
			Read struct {
				Path string `yaml:"path"`
			} `yaml:"read"`
		} `yaml:"data_sources"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]string, len(doc.DataSources))
	for kind, entry := range doc.DataSources {
		out[kind] = entry.Read.Path
	}
	return out, nil
}

// loadDataSourceMatches reads each data source read's `match`, authored in
// config_overrides.yaml and carried into generator_config.yaml by kconfig.
func loadDataSourceMatches(fsw kfs.FS, path string) (map[string]map[string]string, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		DataSources map[string]struct {
			Read struct {
				Match map[string]string `yaml:"match"`
			} `yaml:"read"`
		} `yaml:"data_sources"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := map[string]map[string]string{}
	for kind, entry := range doc.DataSources {
		if len(entry.Read.Match) > 0 {
			out[kind] = entry.Read.Match
		}
	}
	return out, nil
}

// applyParentMatches sets Parent.Match from the authored rules. A rule that
// names no parent-scoped manifest row, or a field the parent's schema lacks,
// is an error: it would otherwise filter nothing, or filter every parent out.
func applyParentMatches(m *Manifest, matches map[string]map[string]string, attrs map[string]map[string]bool) error {
	kinds := make([]string, 0, len(matches))
	for k := range matches {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	for _, kind := range kinds {
		tfType := "kion_" + kind
		idx := -1
		for i := range m.Resources {
			if m.Resources[i].TFType == tfType {
				idx = i
			}
		}
		if idx < 0 {
			return fmt.Errorf("match on data source %q: no manifest row for %s", kind, tfType)
		}
		r := &m.Resources[idx]
		if r.Parent == nil || len(r.Parents) > 0 {
			return fmt.Errorf("match on data source %q: %s is not read under a single parent", kind, tfType)
		}
		parentType := "kion_" + r.Parent.Kind
		for field := range matches[kind] {
			if !attrs[parentType][field] {
				return fmt.Errorf("match on data source %q: parent %s has no attribute %q", kind, parentType, field)
			}
		}
		r.Parent.Match = matches[kind]
	}
	return nil
}

// loadResourceAttrs returns every resource's top-level attribute names from
// the schema snapshot.
func loadResourceAttrs(fsw kfs.FS, path string) (map[string]map[string]bool, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]struct {
				Block struct {
					Attributes map[string]json.RawMessage `json:"attributes"`
				} `json:"block"`
			} `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := map[string]map[string]bool{}
	for _, provider := range doc.ProviderSchemas {
		for tfType, schema := range provider.ResourceSchemas {
			set := make(map[string]bool, len(schema.Block.Attributes))
			for a := range schema.Block.Attributes {
				set[a] = true
			}
			out[tfType] = set
		}
	}
	return out, nil
}

func loadArchetypeKinds(fsw kfs.FS, path string) (map[string]archetypeInfo, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc map[string]struct {
		Kind        string `yaml:"kind"`
		KeyField    string `yaml:"key_field"`
		ParentField string `yaml:"parent_field"`
		// compound_key_parent_read's own identity fields -- see
		// internal/kgen/crud/compound.go's archetype struct, which parses the
		// same three yaml keys off the same file for the CRUD generator side.
		Collection    string `yaml:"collection"`
		ParentIDField string `yaml:"parent_id_field"`
		ChildIDField  string `yaml:"child_id_field"`
		// An entity whose delete addresses the record through a parent the read
		// does not return takes a compound import id, because the parent cannot
		// be recovered afterwards. entity.gtpl keys its splitting ImportState on
		// the same field, and the two must agree: a manifest id the resource's
		// ImportState cannot parse is exactly the mismatch that produced 652
		// "Invalid ID" failures in a verification run.
		DeleteExtraField string `yaml:"delete_extra_field"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]archetypeInfo, len(doc))
	for kind, entry := range doc {
		out[kind] = archetypeInfo{
			Kind:             entry.Kind,
			KeyField:         entry.KeyField,
			ParentField:      entry.ParentField,
			Collection:       entry.Collection,
			ParentIDField:    entry.ParentIDField,
			ChildIDField:     entry.ChildIDField,
			DeleteExtraField: entry.DeleteExtraField,
		}
	}
	return out, nil
}

// loadTFTypes returns every managed resource tf_type in the schema snapshot,
// plus hasNameAttr: tf_type -> whether that resource's schema declares a
// top-level "name" attribute (any attribute type -- Build only needs its
// presence, not its shape). NameField should name a real field the read
// response can populate, and "name" is not universal: e.g.
// kion_aws_resource_tag has no such attribute, so leaving it empty there (see
// Build) is correct, not an oversight.
func loadTFTypes(fsw kfs.FS, path string) ([]string, map[string]bool, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		ProviderSchemas map[string]struct {
			ResourceSchemas map[string]struct {
				Block struct {
					Attributes map[string]json.RawMessage `json:"attributes"`
				} `json:"block"`
			} `json:"resource_schemas"`
		} `json:"provider_schemas"`
	}
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, nil, fmt.Errorf("parse %s: %w", path, err)
	}
	var out []string
	hasNameAttr := make(map[string]bool)
	for _, provider := range doc.ProviderSchemas {
		for tfType, schema := range provider.ResourceSchemas {
			out = append(out, tfType)
			_, hasNameAttr[tfType] = schema.Block.Attributes["name"]
		}
	}
	sort.Strings(out)
	return out, hasNameAttr, nil
}

// loadPrivateListReadOnlyPaths reads private_endpoints.yaml's list_read_only:
// section -- kinds where the only read anywhere (public or private) is a flat
// collection endpoint, no by-id GET at all. See Build's doc comment for why
// this is consulted first.
func loadPrivateListReadOnlyPaths(fsw kfs.FS, path string) (map[string]string, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		ListReadOnly map[string]struct {
			Read struct {
				Path string `yaml:"path"`
			} `yaml:"read"`
		} `yaml:"list_read_only"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]string, len(doc.ListReadOnly))
	for kind, entry := range doc.ListReadOnly {
		out[kind] = entry.Read.Path
	}
	return out, nil
}

// loadPrivateResourceReadPaths reads private_endpoints.yaml's resources:
// section -- the real (often private v1/v2) by-id read for a resource whose
// public spec lacks one. Same shape as loadReadPaths, one file over; see
// Build's doc comment for why this is consulted last.
func loadPrivateResourceReadPaths(fsw kfs.FS, path string) (map[string]string, error) {
	raw, err := fsw.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var doc struct {
		Resources map[string]struct {
			Read struct {
				Path string `yaml:"path"`
			} `yaml:"read"`
		} `yaml:"resources"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	out := make(map[string]string, len(doc.Resources))
	for kind, entry := range doc.Resources {
		out[kind] = entry.Read.Path
	}
	return out, nil
}
