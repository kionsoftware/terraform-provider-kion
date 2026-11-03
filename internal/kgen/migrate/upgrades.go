package migrate

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ProjectRule maps an old set/list-of-objects attribute to a new scalar-id list
// by extracting Field from each element.
type ProjectRule struct {
	From  string `yaml:"from"`
	Field string `yaml:"field"`
}

// KVListRule names the two fields of the object that each entry of an old map
// attribute becomes, for a map(...) attribute the new schema declares as a
// list/set of two-field objects.
type KVListRule struct {
	KeyField string `yaml:"key_field"`
	ValField string `yaml:"val_field"`
}

// Transform is the per-resource old→new state mapping. Attributes not named by
// any rule fall through by name (same name in old → passthrough; new-only → null).
type Transform struct {
	Rename  map[string]string      `yaml:"rename"`   // old_attr -> new_attr, value unchanged
	IDInt   bool                   `yaml:"id_int"`   // id string -> number
	Project map[string]ProjectRule `yaml:"project"`  // new_attr -> {from, field}
	KVList  map[string]KVListRule  `yaml:"kv_list"`  // attr -> {key_field, val_field}: map {k: v} -> [{key_field: k, val_field: v}]
	Unwrap  []string               `yaml:"unwrap"`   // block-set → single object ([{…}] -> {…})
	Drop    []string               `yaml:"drop"`     // old attrs to ignore explicitly
	OldType string                 `yaml:"old_type"` // alias resources: the old state's type name differs from the new (keyed) type
}

// LoadUpgrades reads codegen/state_upgrades.yaml keyed by resource type
// (e.g. "kion_user_group").
func LoadUpgrades(path string) (map[string]Transform, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseUpgrades(raw, path)
}

// ParseUpgrades decodes state_upgrades.yaml from bytes, so callers holding an
// embedded copy do not have to go through the filesystem. name is used only to
// identify the source in an error.
func ParseUpgrades(raw []byte, name string) (map[string]Transform, error) {
	var doc struct {
		Resources map[string]Transform `yaml:"resources"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse %s: %w", name, err)
	}
	return doc.Resources, nil
}
