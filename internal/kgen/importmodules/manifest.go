// Package importmodules rewrites bare `resource "kion_*" "<label>"` blocks --
// the shape `terraform plan -generate-config-out` writes -- into `module
// "<label>" { source = "<modulesDir>/<path>" ... }` calls against the modules
// under modules/terraform-kion-*.
//
// The rewrite is driven entirely by modules/module_manifest.json (see
// internal/kgen/module.BuildManifest), which is the generator's own record of
// each module's attribute -> variable mapping. This package therefore never
// imports internal/provider: it reads the manifest's JSON, not the compiled
// resource schemas, so it builds and runs without the kgendocs tag and does
// not need to touch the provider dependency graph to do this rewrite.
package importmodules

import (
	"encoding/json"
	"fmt"
	"os"
)

// Manifest is the decoded form of modules/module_manifest.json. Its shape
// mirrors internal/kgen/module.Manifest field for field, but is redeclared
// here (rather than imported) to keep this package's dependency graph free of
// internal/provider.
type Manifest struct {
	Version int      `json:"version"`
	Modules []Module `json:"modules"`
}

// Module is one generated module's identity and its settable attributes.
type Module struct {
	TFType    string     `json:"tf_type"`
	Path      string     `json:"path"`
	Variables []Variable `json:"variables"`
}

// Variable is one settable attribute: the provider attribute name, the module
// variable it is wired to, and whether the resource schema renders it as a
// nested block (so the rewriter knows to fold a literal `name { ... }` block
// into an object literal rather than copy an expression verbatim).
type Variable struct {
	Attr  string `json:"attr"`
	Var   string `json:"var"`
	Block bool   `json:"block"`
}

// ParseManifest decodes a manifest from its JSON form.
func ParseManifest(data []byte) (*Manifest, error) {
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing module manifest: %w", err)
	}
	return &m, nil
}

// LoadManifest reads and parses a manifest file from disk.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading module manifest %s: %w", path, err)
	}
	m, err := ParseManifest(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return m, nil
}

// ByType indexes the manifest's modules by Terraform resource type, the
// lookup Rewrite needs once per resource block.
func (m *Manifest) ByType() map[string]Module {
	idx := make(map[string]Module, len(m.Modules))
	for _, mod := range m.Modules {
		idx[mod.TFType] = mod
	}
	return idx
}

// byAttr indexes one module's variables by provider attribute name.
func (mod Module) byAttr() map[string]Variable {
	idx := make(map[string]Variable, len(mod.Variables))
	for _, v := range mod.Variables {
		idx[v.Attr] = v
	}
	return idx
}
