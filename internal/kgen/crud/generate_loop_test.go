package crud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// The generate loop once regressed so that `make crud` failed on every existing
// resource. Run the real Generate over three resources sharing the label
// fixture's ops: one that renders, one whose files exist (no --force), and one
// whose model is missing, so it fails.
func TestGenerateLoopFailsOnlyTheBrokenResource(t *testing.T) {
	root := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		p := filepath.Join(root, rel)
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o750))
		require.NoError(t, os.WriteFile(p, []byte(body), 0o600))
	}
	read := func(src string) string {
		t.Helper()
		b, err := os.ReadFile(src)
		require.NoError(t, err)
		return string(b)
	}

	for _, f := range []string{FieldPolicyPath, ConfigValidatorsPath, RenamesPath} {
		write(f, "{}\n")
	}
	var cfg strings.Builder
	cfg.WriteString("resources:\n")
	for _, n := range []string{"alpha", "bravo", "charlie"} {
		cfg.WriteString("  " + n + ":\n" +
			"    create: {path: /v3/label, method: POST}\n" +
			"    read: {path: '/v3/label/{id}', method: GET}\n" +
			"    update: {path: '/v3/label/{id}', method: PATCH}\n" +
			"    delete: {path: '/v3/label/{id}', method: DELETE}\n")
	}
	write("codegen/generator_config.yaml", cfg.String())
	for _, f := range []string{"oas_client_gen.go", "oas_schemas_gen.go", "oas_parameters_gen.go"} {
		write(filepath.Join("sdk", "generated", fixtureSDKVersion, f), read(filepath.Join("testdata", f)))
	}
	model := read(filepath.Join("testdata", "label_schema_gen.go"))
	write("internal/service/alpha/alpha_schema_gen.go", strings.ReplaceAll(model, "LabelModel", "AlphaModel"))
	write("internal/service/bravo/bravo_schema_gen.go", strings.ReplaceAll(model, "LabelModel", "BravoModel"))
	write("internal/service/bravo/bravo.go", "package bravo // hand-left\n")
	// charlie's schema declares no CharlieModel, so it cannot render.
	write("internal/service/charlie/charlie_schema_gen.go", model)

	_, err := Generate(Options{
		ProjectRoot: root,
		SDKDir:      "sdk",
		SDKVersion:  fixtureSDKVersion,
		TestValues:  filepath.Join(root, "no_test_values.yaml"),
	})

	require.Error(t, err)
	msg := err.Error()
	require.Contains(t, msg, "1 resource(s) failed to generate")
	require.Contains(t, msg, "\n  charlie: ")
	require.NotContains(t, msg, "alpha")
	require.NotContains(t, msg, "bravo")

	require.Equal(t, "package bravo // hand-left\n", read(filepath.Join(root, "internal/service/bravo/bravo.go")))
	for _, f := range []string{"alpha.go", "alpha_data_source.go", "service_package.go", "sweep.go"} {
		require.FileExists(t, filepath.Join(root, "internal/service/alpha", f))
	}
	require.Contains(t, read(filepath.Join(root, "internal/service/alpha/alpha.go")), "func (r *alphaResource) Create(")
}
