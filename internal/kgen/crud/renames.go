package crud

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// RenamesPath is the attribute-rename sidecar the schemas generator applies.
const RenamesPath = "codegen/renames.yaml"

// loadRenames reads RenamesPath as resource -> {api_attribute: tf_attribute}.
//
// crud needs it because bodyBinds matches a request body field to a model
// attribute by the API's own name. A renamed attribute stops matching, and the
// bind was skipped silently -- so renaming an attribute unbound it, and the
// provider accepted the value and discarded it.
//
// A missing file is an error rather than an empty map: every rename would go
// unbound at once, which presents as resources quietly ignoring input.
func loadRenames(root string) (map[string]map[string]string, error) {
	path := filepath.Join(root, RenamesPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", RenamesPath, err)
	}
	var byResource map[string]map[string]string
	if err := yaml.Unmarshal(raw, &byResource); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", RenamesPath, err)
	}
	return byResource, nil
}
