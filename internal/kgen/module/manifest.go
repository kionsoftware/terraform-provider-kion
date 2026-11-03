package module

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/resource"

	"terraform-provider-kion/internal/provider"
)

// ManifestFileName is the sidecar's name, written alongside the generated
// module directories (modules/module_manifest.json by default).
const ManifestFileName = "module_manifest.json"

// Manifest is the attribute-to-variable mapping for every generated module. A
// module variable is not always named after the provider attribute: kgen
// module renames any attribute that collides with a `module` block
// meta-argument (see reservedVarNames) by prefixing the resource's short
// name. Two such collisions exist today -- kion_cloud_rule's `source` becomes
// `cloud_rule_source`, kion_compliance_program's `version` becomes
// `compliance_program_version` -- and a naive `source = ...` inside a module
// block would not fail cleanly: Terraform reads `source` as the module's own
// source address and tries to fetch a module from that string.
//
// A consumer that rewrites bare `resource` blocks into module calls (see
// internal/kgen/importmodules) must therefore read this mapping from the
// generator rather than re-deriving it from variables.tf.
type Manifest struct {
	Version int     `json:"version"`
	Modules []Entry `json:"modules"`
}

// Entry is one generated module's identity and its settable
// attributes.
type Entry struct {
	TFType    string             `json:"tf_type"`
	Path      string             `json:"path"`
	Variables []VariableManifest `json:"variables"`
}

// VariableManifest is one settable attribute: the provider attribute name,
// the module variable it is wired to, and whether the attribute is rendered
// as a nested block in the resource schema (so a rewriter knows to convert a
// literal `name { ... }` block into an object literal, rather than copy an
// attribute expression verbatim).
type VariableManifest struct {
	Attr  string `json:"attr"`
	Var   string `json:"var"`
	Block bool   `json:"block"`
}

// BuildManifest computes the current manifest straight from the compiled
// provider schemas -- the same source Generate reads from, so the mapping
// cannot drift from what a fresh `kgen module` run would produce. It does not
// error on an attribute whose type the generator cannot derive (Generate's own
// unknown-type reporting already covers that); a manifest entry only needs the
// attribute and variable names, not the type.
//
// Computed-only attributes (last_updated, and any attribute that is Computed
// without also being Required or Optional) are omitted, matching what
// Generate excludes from a module's inputs.
func BuildManifest() *Manifest {
	ctx := context.Background()
	var modules []Entry

	for _, pkg := range provider.ServicePackages() {
		for _, r := range pkg.Resources(ctx) {
			res := r.Factory()
			typeName := resourceTypeName(ctx, res)

			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

			fields, _, _ := collect(typeName, schemaResp.Schema)
			vars := make([]VariableManifest, 0, len(fields))
			for _, f := range fields {
				vars = append(vars, VariableManifest{Attr: f.Name, Var: f.VarName, Block: f.Block})
			}
			sort.Slice(vars, func(i, j int) bool { return vars[i].Attr < vars[j].Attr })

			modules = append(modules, Entry{
				TFType:    typeName,
				Path:      Name(typeName),
				Variables: vars,
			})
		}
	}

	sort.Slice(modules, func(i, j int) bool { return modules[i].TFType < modules[j].TFType })
	return &Manifest{Version: 1, Modules: modules}
}

// MarshalManifest renders m deterministically: sorted (BuildManifest already
// sorts), two-space indent, no timestamps, trailing newline. Both WriteManifest
// and the drift test that checks the committed file against a freshly built one
// go through this so they can never format differently.
func MarshalManifest(m *Manifest) ([]byte, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshaling module manifest: %w", err)
	}
	return append(data, '\n'), nil
}

// WriteManifest computes and writes the manifest to dir/module_manifest.json.
//
// This is unconditional and independent of Generate's per-module skip/force
// logic: the manifest is a small, global sidecar describing every module, not
// just the ones a filtered or non-forced `kgen module` run happened to
// (re)write, so it is always recomputed in full and always overwritten.
func WriteManifest(dir string) error {
	data, err := MarshalManifest(BuildManifest())
	if err != nil {
		return err
	}
	if err := fsw.WriteFile(filepath.Join(dir, ManifestFileName), data, 0600); err != nil {
		return fmt.Errorf("writing %s: %w", ManifestFileName, err)
	}
	return nil
}
