// Package module generates a standalone Terraform module per registered
// resource, straight from the compiled Plugin Framework schema.
//
// Generating from the same schema the provider serves keeps module inputs,
// types, requiredness and descriptions from drifting away from it.
//
// Each module is emitted as a self-contained directory:
//
//	modules/terraform-kion-<name>/
//	  main.tf                  resource, every settable attribute wired to a var
//	  variables.tf             one variable per settable attribute
//	  outputs.tf               computed attributes
//	  versions.tf              provider requirement
//	  README.md                terraform-docs markers, filled by `terraform-docs`
//	  examples/complete/       runnable example using only required inputs
//	  tests/plan.tftest.hcl    plan-only test (no credentials needed)
package module

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"terraform-provider-kion/internal/kgen/kfs"
	"terraform-provider-kion/internal/provider"
)

// fsw is the filesystem seam the generator writes through, matching the
// convention in the examples package so tests can swap in a mock.
var fsw kfs.FS = kfs.OS{}

// ProviderSource is the address callers put in required_providers.
const ProviderSource = "kionsoftware/kion"

// field is one settable attribute, flattened to what the templates need.
type field struct {
	Name      string // provider attribute name
	VarName   string // module variable name; differs when Name is reserved
	TFType    string // Terraform type expression, e.g. list(number)
	Desc      string
	Required  bool
	Unknown   bool     // type could not be derived; needs a human
	Block     bool     // rendered as a dynamic block, not a plain attribute
	Sensitive bool     // schema marks the attribute sensitive
	Sample    string   // placeholder accepted by the attribute's validators, "" if none
	Sub       []string // block sub-attribute names, in order
}

// Generate writes a module per registered resource.
//
// filterResource limits generation to one resource type. providerVersion is
// written into versions.tf. When force is false, existing modules are skipped.
func Generate(outDir, providerVersion, filterResource string, force bool) error {
	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}
	if outDir == "" {
		outDir = filepath.Join(projectRoot, "modules")
	}

	ctx := context.Background()
	var generated, skipped int
	var unknowns []string

	for _, pkg := range provider.ServicePackages() {
		for _, r := range pkg.Resources(ctx) {
			res := r.Factory()
			typeName := resourceTypeName(ctx, res)
			if filterResource != "" && typeName != filterResource {
				continue
			}

			dir := filepath.Join(outDir, Name(typeName))
			if !force {
				if _, err := fsw.Stat(filepath.Join(dir, "main.tf")); err == nil {
					fmt.Printf("  skip %s (exists, use --force to overwrite)\n", dir)
					skipped++
					continue
				}
			}

			schemaResp := &resource.SchemaResponse{}
			res.Schema(ctx, resource.SchemaRequest{}, schemaResp)

			fields, outs, unk := collect(typeName, schemaResp.Schema)
			for _, u := range unk {
				unknowns = append(unknowns, typeName+"."+u)
			}
			if err := write(dir, typeName, providerVersion, fields, outs); err != nil {
				return fmt.Errorf("generating module for %s: %w", typeName, err)
			}
			generated++
		}
	}

	if filterResource != "" && generated == 0 && skipped == 0 {
		return fmt.Errorf("no resource found matching %q", filterResource)
	}

	// Always written in full, regardless of --resource or --force: the
	// manifest describes every module, not just the ones this run happened to
	// (re)write, and it is cheap enough to recompute every time.
	if err := WriteManifest(outDir); err != nil {
		return fmt.Errorf("writing module manifest: %w", err)
	}

	fmt.Printf("Generated %d module(s), skipped %d\n", generated, skipped)
	if len(unknowns) > 0 {
		// Surfaced loudly: a mistyped variable fails later and less clearly.
		fmt.Printf("\n%d attribute(s) had a type this generator does not handle and were\n"+
			"emitted as `any` with a TODO. Fix the generator or the schema:\n", len(unknowns))
		for _, u := range unknowns {
			fmt.Printf("  - %s\n", u)
		}
		// An error, not just a warning: callers (CI) should fail on the exit
		// code rather than by matching this text.
		return fmt.Errorf("%d attribute(s) with a type the generator cannot derive", len(unknowns))
	}
	return nil
}

// reservedVarNames are meta-arguments of a `module` block, so Terraform rejects
// a variable declared with any of these names. A provider attribute that
// collides gets the resource's short name as a prefix.
var reservedVarNames = map[string]bool{
	"source": true, "version": true, "providers": true, "count": true,
	"for_each": true, "lifecycle": true, "depends_on": true, "locals": true,
}

// varName returns the module variable name for a provider attribute.
func varName(typeName, attr string) string {
	if !reservedVarNames[attr] {
		return attr
	}
	return strings.TrimPrefix(typeName, "kion_") + "_" + attr
}

// Name maps a resource type to its module directory name, following the
// Terraform Registry convention terraform-<PROVIDER>-<NAME>.
func Name(typeName string) string {
	return "terraform-kion-" + strings.ReplaceAll(strings.TrimPrefix(typeName, "kion_"), "_", "-")
}

// output is a computed attribute exposed by the module. Terraform rejects a
// root-module output holding sensitive data unless it is marked sensitive.
type output struct {
	Name      string
	Sensitive bool
}

// collect flattens a schema into settable input fields and computed outputs.
func collect(typeName string, s rsschema.Schema) (fields []field, outputs []output, unknown []string) {
	for _, name := range sortedKeys(s.Attributes) {
		a := s.Attributes[name]
		req, opt, comp := classify(a)

		// Computed-only attributes are outputs, never inputs. `id` is always
		// computed and always worth exposing.
		if comp && !req && !opt {
			outputs = append(outputs, output{Name: name, Sensitive: a.IsSensitive()})
			continue
		}
		// last_updated is provider bookkeeping, not user input.
		if name == "last_updated" {
			continue
		}

		t, ok := tfType(a)
		if !ok {
			unknown = append(unknown, name)
		}
		fields = append(fields, field{
			Name:      name,
			VarName:   varName(typeName, name),
			TFType:    t,
			Desc:      describe(a),
			Required:  req,
			Unknown:   !ok,
			Sensitive: a.IsSensitive(),
			Sample:    sampleSatisfying(a),
		})
	}

	// Blocks are rare (most attributes are plain), but dropping one silently
	// would leave the module unable to configure it.
	for _, name := range sortedKeys(s.Blocks) {
		t, ok := blockType(s.Blocks[name])
		if !ok {
			unknown = append(unknown, "block:"+name)
		}
		fields = append(fields, field{
			Name:     name,
			VarName:  varName(typeName, name),
			TFType:   t,
			Desc:     blockDescribe(s.Blocks[name]),
			Required: false,
			Unknown:  !ok,
			Block:    true,
			Sub:      blockAttrNames(s.Blocks[name]),
		})
	}

	if !slices.ContainsFunc(outputs, func(o output) bool { return o.Name == "id" }) {
		outputs = append(outputs, output{Name: "id"})
	}
	sort.Slice(outputs, func(i, j int) bool { return outputs[i].Name < outputs[j].Name })
	return fields, outputs, unknown
}

func write(dir, typeName, providerVersion string, fields []field, outputs []output) error {
	for _, sub := range []string{dir, filepath.Join(dir, "examples", "complete"), filepath.Join(dir, "tests")} {
		if err := fsw.MkdirAll(sub, 0750); err != nil {
			return fmt.Errorf("creating %s: %w", sub, err)
		}
	}
	files := map[string]string{
		filepath.Join(dir, "main.tf"):                         mainTF(typeName, fields),
		filepath.Join(dir, "variables.tf"):                    variablesTF(fields),
		filepath.Join(dir, "outputs.tf"):                      outputsTF(typeName, outputs),
		filepath.Join(dir, "versions.tf"):                     versionsTF(providerVersion),
		filepath.Join(dir, "README.md"):                       readme(typeName, fields),
		filepath.Join(dir, "examples", "complete", "main.tf"): exampleTF(fields),
		filepath.Join(dir, "tests", "plan.tftest.hcl"):        testHCL(typeName, fields),
		filepath.Join(dir, ".gitignore"):                      gitignore(),
	}
	for path, content := range files {
		data := []byte(content)
		// hclwrite is the engine behind `terraform fmt`, so output satisfies
		// `terraform fmt -check`.
		if ext := filepath.Ext(path); ext == ".tf" || ext == ".hcl" {
			data = hclwrite.Format(data)
		}
		if err := fsw.WriteFile(path, data, 0600); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}
	}
	fmt.Printf("  wrote %s (%d inputs, %d outputs)\n", dir, len(fields), len(outputs))
	return nil
}

func mainTF(typeName string, fields []field) string {
	var b strings.Builder
	b.WriteString(header())
	fmt.Fprintf(&b, "resource %q \"this\" {\n", typeName)
	for _, f := range fields {
		if f.Block {
			continue
		}
		fmt.Fprintf(&b, "  %s = var.%s\n", f.Name, f.VarName)
	}
	for _, f := range fields {
		if !f.Block {
			continue
		}
		// Single-nested block: an object variable, or null to omit it entirely.
		fmt.Fprintf(&b, "\n  dynamic %q {\n", f.Name)
		fmt.Fprintf(&b, "    for_each = var.%s == null ? [] : [var.%s]\n", f.VarName, f.VarName)
		b.WriteString("    content {\n")
		for _, sub := range f.Sub {
			fmt.Fprintf(&b, "      %s = %s.value.%s\n", sub, f.Name, sub)
		}
		b.WriteString("    }\n  }\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func variablesTF(fields []field) string {
	var b strings.Builder
	b.WriteString(header())
	for i, f := range fields {
		if i > 0 {
			b.WriteString("\n")
		}
		if f.Unknown {
			b.WriteString("# TODO: the generator could not derive this type from the schema.\n")
		}
		fmt.Fprintf(&b, "variable %q {\n", f.VarName)
		if f.Desc != "" {
			fmt.Fprintf(&b, "  description = %q\n", f.Desc)
		}
		fmt.Fprintf(&b, "  type        = %s\n", f.TFType)
		if f.Sensitive {
			b.WriteString("  sensitive   = true\n")
		}
		if !f.Required {
			// null, not a zero value: an unset optional attribute must stay
			// unset so the provider keeps whatever it computes.
			b.WriteString("  default     = null\n")
		}
		b.WriteString("}\n")
	}
	return b.String()
}

func outputsTF(typeName string, outputs []output) string {
	var b strings.Builder
	b.WriteString(header())
	for i, o := range outputs {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "output %q {\n", o.Name)
		fmt.Fprintf(&b, "  description = %q\n", fmt.Sprintf("%s of the %s.", o.Name, typeName))
		fmt.Fprintf(&b, "  value       = %s.this.%s\n", typeName, o.Name)
		if o.Sensitive {
			b.WriteString("  sensitive   = true\n")
		}
		b.WriteString("}\n")
	}
	return b.String()
}

// DefaultProviderConstraint is a lower bound, not a pin: an exact version makes
// a module usable with one provider release only, and cannot be satisfied beside
// the ">= <release>" kion-import writes into the same root.
const DefaultProviderConstraint = ">= 1.0.0"

func versionsTF(providerVersion string) string {
	return fmt.Sprintf(`%sterraform {
  required_version = ">= 1.0"

  required_providers {
    kion = {
      source  = %q
      version = %q
    }
  }
}
`, header(), ProviderSource, providerVersion)
}

func exampleTF(fields []field) string {
	var b strings.Builder
	b.WriteString(header())
	b.WriteString("module \"this\" {\n  source = \"../..\"\n\n")
	for _, f := range fields {
		if f.Required {
			fmt.Fprintf(&b, "  %s = %s\n", f.VarName, sampleForField(f))
		}
	}
	b.WriteString("}\n")
	return b.String()
}

// testHCL emits a plan-only test: it needs no credentials, so it runs in CI on
// every change, and it still exercises the provider's own validation (required
// attributes, conflicting sets, type coercion).
// atLeastOneOfByType mirrors codegen/config_validators.yaml: a resource-level
// AtLeastOneOf rejects a plan that sets none of the listed attributes, and none
// of them is Required, so a test built from Required fields alone does not
// satisfy it. Only the first attribute of each pair is needed.
//
// Duplicated here rather than read from the YAML because kgen module builds
// from the provider schema alone and takes no codegen inputs; a test guards the
// two against drifting apart.
var atLeastOneOfByType = map[string]string{
	"kion_ami": "owner_user_ids",
	// Back-compat aliases render their own module and carry the same
	// constraint as the resource they alias.
	"kion_aws_cloudformation_template": "owner_user_ids",
	"kion_aws_iam_policy":              "owner_user_ids",
	"kion_azure_arm_template":          "owner_user_ids",
	"kion_azure_policy":                "owner_users",
	"kion_azure_role":                  "owner_user_ids",
	"kion_cft":                         "owner_user_ids",
	"kion_cloud_rule":                  "owner_user_ids",
	"kion_compliance_check":            "owner_user_ids",
	"kion_compliance_standard":         "owner_user_ids",
	"kion_funding_source":              "owner_user_ids",
	"kion_funding_source_enforcement":  "user_ids",
	"kion_gcp_iam_role":                "role_permissions",
	"kion_iam_policy":                  "owner_user_ids",
	"kion_ou":                          "owner_user_ids",
	"kion_ou_enforcement":              "user_ids",
	"kion_project":                     "owner_user_ids",
	"kion_project_enforcement":         "user_ids",
	"kion_service_catalog":             "owner_user_ids",
	"kion_service_control_policy":      "owner_user_ids",
	"kion_user_group":                  "owner_user_ids",
	"kion_webhook":                     "owner_user_ids",
}

func testHCL(typeName string, fields []field) string {
	var b strings.Builder
	// The provider does not contact Kion during Configure, so a placeholder
	// endpoint is enough to plan. Keeps the test runnable with no credentials.
	b.WriteString("# Plan-only. The provider is configured with a placeholder endpoint and is\n" +
		"# never contacted, so no Kion credentials are required.\n\n" +
		"provider \"kion\" {\n  api_url = \"http://127.0.0.1:1\"\n  api_key = \"test\"\n}\n\n")
	b.WriteString("run \"plan\" {\n  command = plan\n")
	need := atLeastOneOfByType[typeName]
	var req []field
	for _, f := range fields {
		if f.Required || (need != "" && f.Name == need) {
			req = append(req, f)
		}
	}
	if len(req) > 0 {
		b.WriteString("\n  variables {\n")
		for _, f := range req {
			fmt.Fprintf(&b, "    %s = %s\n", f.VarName, sampleForField(f))
		}
		b.WriteString("  }\n")
	}
	b.WriteString("}\n")
	return b.String()
}

func readme(typeName string, fields []field) string {
	// Built as HCL and formatted, so the snippet in the README matches the style
	// of the .tf files around it.
	var usage strings.Builder
	// The real module name is written up front rather than substituted into the
	// formatted output. A post-format strings.Replace on a literal `module "x"`
	// would silently leave the placeholder in every README if hclwrite.Format
	// ever changed its quoting or spacing.
	fmt.Fprintf(&usage, "module %q {\n  source = \"...\"\n\n", strings.TrimPrefix(typeName, "kion_"))
	for _, f := range fields {
		if f.Required {
			fmt.Fprintf(&usage, "  %s = %s\n", f.VarName, sampleForField(f))
		}
	}
	usage.WriteString("}\n")
	snippet := string(hclwrite.Format([]byte(usage.String())))
	return fmt.Sprintf(`# %s

Terraform module for `+"`%s`"+`, generated from the provider schema by
`+"`kgen module`"+`. Do not edit by hand -- regenerate instead.

## Usage

`+"```hcl"+`
%s`+"```"+`

<!-- BEGIN_TF_DOCS -->
<!-- END_TF_DOCS -->
`, Name(typeName), typeName, snippet)
}

func gitignore() string {
	return ".terraform/\n.terraform.lock.hcl\n*.tfstate\n*.tfstate.*\n*.tfplan\ncrash.log\n"
}

func header() string {
	return "# Generated by `kgen module` from the provider schema. Do not edit.\n\n"
}

// --- type mapping -----------------------------------------------------------

// tfType renders the Terraform type expression for an attribute. The bool
// reports whether the type was understood; callers surface the misses rather
// than guessing.
func tfType(a rsschema.Attribute) (string, bool) {
	switch v := a.(type) {
	case rsschema.StringAttribute:
		return "string", true
	case rsschema.Int64Attribute, rsschema.Float64Attribute, rsschema.NumberAttribute:
		return "number", true
	case rsschema.BoolAttribute:
		return "bool", true
	case rsschema.ListAttribute:
		return wrap("list", v.ElementType)
	case rsschema.SetAttribute:
		return wrap("set", v.ElementType)
	case rsschema.MapAttribute:
		return wrap("map", v.ElementType)
	case rsschema.SingleNestedAttribute:
		return nested(v.Attributes)
	case rsschema.ListNestedAttribute:
		inner, ok := nested(v.NestedObject.Attributes)
		return "list(" + inner + ")", ok
	case rsschema.SetNestedAttribute:
		inner, ok := nested(v.NestedObject.Attributes)
		return "set(" + inner + ")", ok
	case rsschema.MapNestedAttribute:
		inner, ok := nested(v.NestedObject.Attributes)
		return "map(" + inner + ")", ok
	default:
		return "any", false
	}
}

func wrap(kind string, et attr.Type) (string, bool) {
	inner, ok := elemType(et)
	return kind + "(" + inner + ")", ok
}

func elemType(t attr.Type) (string, bool) {
	switch t.(type) {
	case basetypes.StringType:
		return "string", true
	case basetypes.Int64Type, basetypes.Float64Type, basetypes.NumberType:
		return "number", true
	case basetypes.BoolType:
		return "bool", true
	}
	return "any", false
}

func nested(attrs map[string]rsschema.Attribute) (string, bool) {
	ok := true
	var parts []string
	for _, n := range sortedKeys(attrs) {
		t, o := tfType(attrs[n])
		if !o {
			ok = false
		}
		req, _, _ := classify(attrs[n])
		if !req {
			// Anything the schema does not require must be optional() in the
			// object type, or callers are forced to supply attributes the
			// provider treats as optional.
			parts = append(parts, fmt.Sprintf("%s = optional(%s)", n, t))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s = %s", n, t))
	}
	return "object({ " + strings.Join(parts, ", ") + " })", ok
}

// blockType renders a single-nested block as an object type. Other nesting
// modes are reported rather than guessed at.
func blockType(b rsschema.Block) (string, bool) {
	switch v := b.(type) {
	case rsschema.SingleNestedBlock:
		return nested(v.Attributes)
	default:
		return "any", false
	}
}

func blockDescribe(b rsschema.Block) string {
	d := b.GetMarkdownDescription()
	if d == "" {
		d = b.GetDescription()
	}
	return strings.Join(strings.Fields(d), " ")
}

// blockAttrNames returns a block's attribute names from the schema. Taken
// structurally rather than parsed back out of the rendered type string, which
// would mis-split a nested object's own separators.
func blockAttrNames(b rsschema.Block) []string {
	if v, ok := b.(rsschema.SingleNestedBlock); ok {
		return sortedKeys(v.Attributes)
	}
	return nil
}

// sampleFor picks a placeholder that satisfies the provider's attribute
// validators. A bare "example" fails regex, JSON and format checks, which
// surfaces as a failing generated test rather than anything useful.
// sampleSatisfying returns the first candidate placeholder that the attribute's
// own validators accept, or "" when the attribute has none (or none match).
//
// The validators are run rather than introspected: stringvalidator's concrete
// types are unexported and its descriptions do not reliably carry the pattern,
// but every validator is directly callable, which is exact.
func sampleSatisfying(a rsschema.Attribute) string {
	sa, ok := a.(rsschema.StringAttribute)
	if !ok || len(sa.Validators) == 0 {
		return ""
	}
	ctx := context.Background()
	for _, cand := range patternCandidates {
		okAll := true
		for _, v := range sa.Validators {
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, validator.StringRequest{
				Path:        path.Root("x"),
				ConfigValue: types.StringValue(cand),
			}, resp)
			if resp.Diagnostics.HasError() {
				okAll = false
				break
			}
		}
		if okAll {
			return cand
		}
	}
	return ""
}

// Ordered so the most specific shapes win; every entry is a plausible value for
// some attribute in this provider.
var patternCandidates = []string{
	"2026-01", "2026-01-01", "2026-01-01T00:00:00Z",
	"user@example.com", "https://example.com", "#1a2b3c",
	"{}", "example", "example-1", "Example", "1", "000000000000",
}

// sampleForField picks a placeholder for a field, preferring one the attribute's
// validators actually accept over a name-based guess. A value that fails
// validation surfaces as a failing generated module test, which is how
// billing_start_date (a YYYY-MM regex, missed by the _datecode name heuristic)
// broke the terraform test stage.
func sampleForField(f field) string {
	if f.Sample != "" {
		return `"` + f.Sample + `"`
	}
	return sampleFor(f.Name, f.TFType)
}

func sampleFor(name, tfType string) string {
	if tfType == "string" {
		switch {
		case strings.HasSuffix(name, "_datecode"), strings.HasSuffix(name, "_start_date"),
			strings.HasSuffix(name, "_end_date"):
			return `"2026-01"`
		case strings.Contains(name, "email"):
			return `"user@example.com"`
		case strings.Contains(name, "url"):
			return `"https://example.com"`
		case name == "color":
			return `"#1a2b3c"`
		case name == "criteria", name == "policy", name == "body",
			strings.HasSuffix(name, "_json"), strings.HasSuffix(name, "parameters"),
			strings.HasSuffix(name, "_template"), strings.HasSuffix(name, "permissions"),
			strings.HasSuffix(name, "_headers"), strings.HasSuffix(name, "_body"):
			return `"{}"`
		}
	}
	return sample(tfType)
}

func sample(tfType string) string {
	switch {
	case tfType == "string":
		return `"example"`
	case tfType == "number":
		return "1"
	case tfType == "bool":
		return "false"
	case strings.HasPrefix(tfType, "list("), strings.HasPrefix(tfType, "set("):
		return "[]"
	case strings.HasPrefix(tfType, "map("):
		return "{}"
	case strings.HasPrefix(tfType, "object("):
		return objectSample(tfType)
	default:
		return "null"
	}
}

// objectSample builds a literal for an object() type carrying every non-optional
// field, so a required nested attribute gets a usable value instead of null.
// Terraform rejects null for a required attribute ("Missing Configuration for
// Required Attribute"), which is how the gcp module's generated test failed.
// optional(...) fields are omitted. The point is a minimal valid object.
func objectSample(tfType string) string {
	inner := strings.TrimSuffix(strings.TrimPrefix(tfType, "object({"), "})")
	var parts []string
	for _, f := range splitTopLevel(inner) {
		name, typ, ok := strings.Cut(f, "=")
		if !ok {
			continue
		}
		name, typ = strings.TrimSpace(name), strings.TrimSpace(typ)
		if strings.HasPrefix(typ, "optional(") {
			continue
		}
		parts = append(parts, name+" = "+sampleFor(name, typ))
	}
	if len(parts) == 0 {
		return "{}"
	}
	return "{ " + strings.Join(parts, ", ") + " }"
}

// splitTopLevel splits an object body on commas that are not inside nested
// parens or braces, so a nested object() is treated as one field.
func splitTopLevel(s string) []string {
	var out []string
	depth, start := 0, 0
	for i, r := range s {
		switch r {
		case '(', '{', '[':
			depth++
		case ')', '}', ']':
			depth--
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if tail := strings.TrimSpace(s[start:]); tail != "" {
		out = append(out, tail)
	}
	return out
}

// --- schema helpers ---------------------------------------------------------

func classify(a rsschema.Attribute) (required, optional, computed bool) {
	return a.IsRequired(), a.IsOptional(), a.IsComputed()
}

func describe(a rsschema.Attribute) string {
	d := a.GetMarkdownDescription()
	if d == "" {
		d = a.GetDescription()
	}
	return strings.Join(strings.Fields(d), " ")
}

func resourceTypeName(ctx context.Context, r resource.Resource) string {
	req := resource.MetadataRequest{ProviderTypeName: "kion"}
	resp := &resource.MetadataResponse{}
	r.Metadata(ctx, req, resp)
	return resp.TypeName
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}
	for {
		if _, err := fsw.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
