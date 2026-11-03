package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testManifestJSON = `{
  "version": 1,
  "modules": [
    {"tf_type":"kion_ou","path":"terraform-kion-ou","variables":[
      {"attr":"name","var":"name"},
      {"attr":"description","var":"description"}
    ]},
    {"tf_type":"kion_cloud_rule","path":"terraform-kion-cloud-rule","variables":[
      {"attr":"name","var":"name"},
      {"attr":"source","var":"cloud_rule_source"}
    ]}
  ]
}`

const testGeneratedTF = `resource "kion_ou" "example" {
  name         = "engineering"
  description  = "eng org unit"
  last_updated = "2026-01-01T00:00:00Z"
}

resource "kion_cloud_rule" "collides" {
  name   = "my-rule"
  source = "aws"
}

resource "not_kion_thing" "untouched" {
  a = 1
}
`

// runImportModules invokes the subcommand through the root, which is how cobra
// dispatches: calling Execute on a child re-enters from its parent and would
// print root help instead of running anything. Flag vars are package-level, so
// they are restored afterwards to keep tests order-independent.
func runImportModules(t *testing.T, args ...string) (string, error) {
	t.Helper()
	t.Cleanup(func() {
		importModulesIn, importModulesOut = "generated.tf", "modules.tf"
		importModulesManifest = filepath.Join("modules", "module_manifest.json")
		importModulesModulesDir, importModulesForce = "./modules", false
		rootCmd.SetArgs(nil)
		rootCmd.SetOut(nil)
		rootCmd.SetErr(nil)
	})
	var out bytes.Buffer
	rootCmd.SetOut(&out)
	rootCmd.SetErr(&out)
	rootCmd.SetArgs(append([]string{"import-modules"}, args...))
	// Cobra prints usage on RunE errors by default, which buries a one-line
	// failure under the whole help text in test output.
	importModulesCmd.SilenceUsage = true
	err := rootCmd.Execute()
	return out.String(), err
}

// norm collapses whitespace runs so assertions do not pin hclwrite's column
// alignment, which shifts with the longest attribute name in each block.
func norm(s string) string { return strings.Join(strings.Fields(s), " ") }

// countAttrIn counts assignments of exactly name to value. A plain Contains
// would over-count: the renamed variable this test exists to check,
// cloud_rule_source, ends with the very name whose absence is asserted.
func countAttrIn(s, name, value string) int {
	re := regexp.MustCompile(`(^|[^0-9A-Za-z_])` + regexp.QuoteMeta(name+" = "+value))
	return len(re.FindAllString(s, -1))
}

func writeFixture(t *testing.T) (dir, manifest, in, outPath string) {
	t.Helper()
	dir = t.TempDir()
	manifest = filepath.Join(dir, "module_manifest.json")
	in = filepath.Join(dir, "generated.tf")
	outPath = filepath.Join(dir, "modules.tf")
	require.NoError(t, os.WriteFile(manifest, []byte(testManifestJSON), 0o600))
	require.NoError(t, os.WriteFile(in, []byte(testGeneratedTF), 0o600))
	return dir, manifest, in, outPath
}

func TestImportModules_RegisteredOnRoot(t *testing.T) {
	cmd := findSubcommand(rootCmd, "import-modules")
	require.NotNil(t, cmd, "import-modules should be registered on the root command")
	assert.NotEmpty(t, cmd.Short)
	assert.NotNil(t, cmd.RunE)
}

func TestImportModules_Rewrites(t *testing.T) {
	_, manifest, in, outPath := writeFixture(t)

	stdout, err := runImportModules(t,
		"--in", in, "--out", outPath, "--manifest", manifest, "--modules-dir", "./modules")
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	s := string(got)

	// hclwrite column-aligns `=`, so assertions run against a space-collapsed
	// copy rather than pinning whatever padding the current block happens to
	// produce.
	n := norm(s)

	assert.Contains(t, n, `module "example" {`)
	assert.Contains(t, n, `source = "./modules/terraform-kion-ou"`)
	// The meta-argument collision: the resource's own `source` must be renamed,
	// and the module's `source` must still be the module path. Counted at a
	// token boundary because "cloud_rule_source = " ends with "source = ".
	assert.Contains(t, n, `cloud_rule_source = "aws"`)
	assert.Contains(t, n, `source = "./modules/terraform-kion-cloud-rule"`)
	assert.Equal(t, 0, countAttrIn(n, "source", `"aws"`),
		`a bare source = "aws" would be read as the module's own source address`)
	// A type with no manifest entry is left alone entirely.
	assert.Contains(t, n, `resource "not_kion_thing" "untouched"`)
	assert.NotContains(t, n, `resource "kion_ou"`)

	// An attribute with no module variable is dropped AND reported; a silent
	// drop is how a rewrite would quietly change what is under management.
	assert.NotContains(t, s, "last_updated")
	assert.Contains(t, stdout, "last_updated")
	assert.Contains(t, stdout, "2 block(s) rewritten, 1 left untouched, 1 attribute(s) dropped")
}

func TestImportModules_RefusesToClobber(t *testing.T) {
	_, manifest, in, outPath := writeFixture(t)
	require.NoError(t, os.WriteFile(outPath, []byte("# precious\n"), 0o600))

	_, err := runImportModules(t, "--in", in, "--out", outPath, "--manifest", manifest)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--force")

	kept, readErr := os.ReadFile(outPath)
	require.NoError(t, readErr)
	assert.Equal(t, "# precious\n", string(kept), "the existing file must be untouched")
}

func TestImportModules_ForceOverwrites(t *testing.T) {
	_, manifest, in, outPath := writeFixture(t)
	require.NoError(t, os.WriteFile(outPath, []byte("# stale\n"), 0o600))

	_, err := runImportModules(t, "--in", in, "--out", outPath, "--manifest", manifest, "--force")
	require.NoError(t, err)

	got, err := os.ReadFile(outPath)
	require.NoError(t, err)
	assert.Contains(t, string(got), `module "example"`)
}

// In-place is the one case the clobber guard must not block: --out equal to
// --in is a deliberate rewrite, not an accident, and requiring --force for it
// would make the obvious invocation fail.
func TestImportModules_InPlaceNeedsNoForce(t *testing.T) {
	_, manifest, in, _ := writeFixture(t)

	_, err := runImportModules(t, "--in", in, "--out", in, "--manifest", manifest)
	require.NoError(t, err)

	got, err := os.ReadFile(in)
	require.NoError(t, err)
	assert.Contains(t, string(got), `module "example"`)
	assert.NotContains(t, string(got), `resource "kion_ou"`)
}

func TestImportModules_MissingManifestExplainsItself(t *testing.T) {
	_, _, in, outPath := writeFixture(t)

	_, err := runImportModules(t,
		"--in", in, "--out", outPath, "--manifest", filepath.Join(t.TempDir(), "absent.json"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "make modules",
		"a missing manifest should name the command that produces it")
}
