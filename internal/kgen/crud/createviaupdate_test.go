package crud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An entity's create_via_update must reach its generated Create: without the
// follow-up update, Kion keeps the value it stored on create and the apply
// fails with "Provider produced inconsistent result".
func TestEntityCreateViaUpdateReachesGeneratedCreate(t *testing.T) {
	t.Parallel()
	archs, err := loadArchetypes(filepath.Join("..", "..", "..", "codegen", "crud_archetypes.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for pkg, a := range archs {
		if a.Kind != "entity" || len(a.CreateViaUpdate) == 0 {
			continue
		}
		n++
		src, err := os.ReadFile(filepath.Join("..", "..", "service", pkg, pkg+".go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(src), "create_via_update: the create ignores these") {
			t.Errorf("%s declares create_via_update %v but its Create never re-applies it; run `make crud-force`", pkg, a.CreateViaUpdate)
		}
	}
	if n == 0 {
		t.Fatal("no entity declares create_via_update; project_cloud_access_role should")
	}
}

func TestCheckCreateViaUpdate(t *testing.T) {
	t.Parallel()
	ok := entityData{HasUpdate: true, IDParamType: "int64", CreateViaUpdate: []string{"Flag"},
		UpdateBinds: []fieldBind{{ModelGo: "Flag"}}}
	if err := checkCreateViaUpdate(ok); err != nil {
		t.Errorf("valid create_via_update refused: %v", err)
	}
	for name, d := range map[string]entityData{
		"no update":      {IDParamType: "int64", CreateViaUpdate: []string{"Flag"}},
		"string id":      {HasUpdate: true, IDParamType: "string", CreateViaUpdate: []string{"Flag"}},
		"not in update":  {HasUpdate: true, IDParamType: "int64", CreateViaUpdate: []string{"Flag"}},
		"wrong bind set": {HasUpdate: true, IDParamType: "int64", CreateViaUpdate: []string{"Flag"}, UpdateBinds: []fieldBind{{ModelGo: "Other"}}},
	} {
		if err := checkCreateViaUpdate(d); err == nil {
			t.Errorf("%s: accepted, want an error", name)
		}
	}
}
