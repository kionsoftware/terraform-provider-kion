package crud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every missing_via_list declared in private_endpoints.yaml must reach the
// generated read. Dropping it silently brings back the 500 that broke refresh
// of a resource deleted outside Terraform.
func TestMissingViaListReachesGeneratedRead(t *testing.T) {
	t.Parallel()
	eps, err := loadPrivateEndpoints(filepath.Join("..", "..", "..", "codegen", "private_endpoints.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for pkg, pe := range eps {
		if pe.MissingViaList == "" {
			continue
		}
		n++
		src, err := os.ReadFile(filepath.Join("..", "..", "service", pkg, pkg+".go"))
		if err != nil {
			t.Fatal(err)
		}
		want := `RawIDListed(ctx, "` + pe.MissingViaList + `", id)`
		if !strings.Contains(string(src), want) {
			t.Errorf("%s declares missing_via_list %s but its read does not consult it; run `make crud-force`", pkg, pe.MissingViaList)
		}
	}
	if n == 0 {
		t.Fatal("no resource declares missing_via_list; app_role should")
	}
}

// The entity archetype's missing_via_list must reach the generated read too.
func TestEntityMissingViaListReachesGeneratedRead(t *testing.T) {
	t.Parallel()
	archs, err := loadArchetypes(filepath.Join("..", "..", "..", "codegen", "crud_archetypes.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for pkg, a := range archs {
		if a.MissingViaList == "" {
			continue
		}
		n++
		src, err := os.ReadFile(filepath.Join("..", "..", "service", pkg, pkg+".go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), `Listed(ctx, "`+a.MissingViaList+`", `) {
			t.Errorf("%s declares missing_via_list %s but its read does not consult it; run `make crud-force`", pkg, a.MissingViaList)
		}
	}
	if n == 0 {
		t.Fatal("no entity declares missing_via_list; oauth2_client should")
	}
}

// A payer-kind min_kion_version must become a gate in the generated resource:
// version_support cannot derive one from a private route, so without it an old
// install fails at apply with an opaque API error.
func TestPayerMinKionVersionReachesResource(t *testing.T) {
	t.Parallel()
	eps, err := loadPrivateEndpoints(filepath.Join("..", "..", "..", "codegen", "private_endpoints.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for pkg, pe := range eps {
		if pe.MinKionVersion == "" {
			continue
		}
		n++
		src, err := os.ReadFile(filepath.Join("..", "..", "service", pkg, pkg+".go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{
			`conns.MustParseKionVersion("` + pe.MinKionVersion + `")`,
			"func (r *" + pkg + "Resource) ModifyPlan(",
		} {
			if !strings.Contains(string(src), want) {
				t.Errorf("%s declares min_kion_version %s but its resource lacks %q; run `make crud-force`", pkg, pe.MinKionVersion, want)
			}
		}
	}
	if n == 0 {
		t.Fatal("no resource declares min_kion_version; billing_source_anthropic should")
	}
}
