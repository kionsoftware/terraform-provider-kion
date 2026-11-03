package importmanifest

import "strings"

// ListPathFrom strips a trailing "/{...}" id template off a read path to get
// the list endpoint. A read path that is already a list (billing_source's
// /v4/billing-source) is returned unchanged.
func ListPathFrom(readPath string) string {
	if readPath == "" {
		return ""
	}
	if i := strings.Index(readPath, "/{"); i >= 0 {
		return readPath[:i]
	}
	return readPath
}

// Classify maps a crud_archetypes.yaml archetype onto a read shape and an
// import-id format. The format must match what the matching crud template
// generates in ImportState -- see the package doc.
func Classify(archetype string, readPath string, hasParent bool) (ReadShape, IDFormat, bool, string) {
	switch archetype {
	case "no_read":
		// "no_read" means no by-id GET, not "unreadable" -- e.g.
		// kion_aws_resource_tag has no by-id read but IS listable at a flat
		// collection endpoint (private_endpoints.yaml's list_read_only
		// section). Build resolves readPath through the same four-source
		// fallback chain it uses for every archetype (see Build's doc
		// comment); it is not narrowed specially for no_read. In practice a
		// no_read kind's generator_config.yaml resources/data_sources entries
		// are typically empty -- internal/kgen/crud/noread.go builds a
		// no_read resource's model from Create/Update/Delete only, so the
		// CRUD generator never populates a read path for it there -- which is
		// why a non-empty readPath reaching this branch usually came from
		// list_read_only or, as a last resort, private_endpoints.yaml's own
		// resources section. That is a fact about the input data, though, not
		// a rule this function or Build enforces.
		if ListPathFrom(readPath) == "" {
			return ShapeNone, FormatID, false,
				"crud_archetypes.yaml declares kind: no_read (no by-id GET) and no list-only or private read path was found either"
		}
		// A no_read resource with a parent-scoped `parent_read` gets the
		// compound ImportState from entity_noread.gtpl (it has to: the read is a
		// parent collection and the parent id cannot be discovered from the
		// record id), so its import id must match. One without a parent --
		// kion_aws_resource_tag, whose collection is flat -- keeps the plain id.
		if hasParent {
			return ShapeParentList, FormatParentSlashKey, true, ""
		}
		return ShapeGeneric, FormatID, true, ""
	case "datasource_only":
		return ShapeNone, FormatID, false,
			"crud_archetypes.yaml declares kind: datasource_only - not a managed resource"
	case "association":
		// association.gtpl's {{if .HasParent}} branch splits req.ID on "/" into
		// (ParentTF, KeyTF); the {{else}} branch treats req.ID as a plain id.
		// The format mirrors this conditional.
		if hasParent {
			return ShapeAssociation, FormatParentSlashKey, true, ""
		}
		return ShapeAssociation, FormatID, true, ""
	case "parent_list":
		// Looks the record up under its parent via a second HTTP call per
		// parent id, so it splits req.ID on "/" in ImportState. A bare id
		// leaves the parent zero and every read 404s, which is how 187
		// enforcement imports failed against a real install.
		return ShapeParentList, FormatParentSlashKey, true, ""
	case "compound_key_parent_read":
		// Unlike parent_list, the child records are not behind a second
		// endpoint -- they live inline inside the parent's own list payload
		// (e.g. /beta/scope's criteria_records). entity_compound.gtpl's
		// ImportState still splits req.ID on "/" the same way, so the id
		// format matches parent_list even though the read is structurally
		// different (one HTTP call, not two).
		return ShapeNestedCollection, FormatParentSlashKey, true, ""
	case "singleton", "raw_http", "cv_override":
		return ShapeSpecial, FormatID, true, ""
	}

	if ListPathFrom(readPath) == "" {
		return ShapeNone, FormatID, false,
			"no read path in generator_config.yaml or importmanifest paths table"
	}
	return ShapeGeneric, FormatID, true, ""
}
