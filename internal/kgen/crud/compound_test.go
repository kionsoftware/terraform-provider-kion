package crud

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMapParams(t *testing.T) {
	tests := []struct {
		name       string
		fields     []string
		childParam string
		wantParent string
		wantChild  string
		wantErr    bool
	}{
		{
			name:       "create single parent param",
			fields:     []string{"ID"},
			childParam: "CriteriaID",
			wantParent: "ID",
			wantChild:  "",
		},
		{
			name:       "update parent + child (shared parent name)",
			fields:     []string{"ID", "CriteriaID"},
			childParam: "CriteriaID",
			wantParent: "ID",
			wantChild:  "CriteriaID",
		},
		{
			name:       "delete distinct parent name",
			fields:     []string{"ScopeID", "CriteriaID"},
			childParam: "CriteriaID",
			wantParent: "ScopeID",
			wantChild:  "CriteriaID",
		},
		{
			name:       "child param ordered first",
			fields:     []string{"CriteriaID", "ScopeID"},
			childParam: "CriteriaID",
			wantParent: "ScopeID",
			wantChild:  "CriteriaID",
		},
		{
			name:       "ambiguous: two non-child ids",
			fields:     []string{"ID", "OtherID"},
			childParam: "CriteriaID",
			wantErr:    true,
		},
		{
			name:       "no parent: only the child id",
			fields:     []string{"CriteriaID"},
			childParam: "CriteriaID",
			wantErr:    true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := &Struct{Name: "P"}
			for _, f := range tc.fields {
				s.Fields = append(s.Fields, Field{GoName: f})
			}
			parent, child, err := mapParams(s, tc.childParam)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got parent=%q child=%q", parent, child)
				}
				return
			}
			if err != nil {
				t.Fatalf("mapParams: %v", err)
			}
			if parent != tc.wantParent || child != tc.wantChild {
				t.Errorf("parent/child = %q/%q, want %q/%q", parent, child, tc.wantParent, tc.wantChild)
			}
		})
	}
}

func TestMapParams_nilStruct(t *testing.T) {
	if _, _, err := mapParams(nil, "CriteriaID"); err == nil {
		t.Fatal("want error for nil params struct")
	}
}

func TestLoadArchetypes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "crud_archetypes.yaml")
	yaml := "scope_criteria:\n" +
		"  kind: compound_key_parent_read\n" +
		"  parent_id_field: scope_id\n" +
		"  child_id_field: criteria_id\n" +
		"  child_id_param: CriteriaID\n" +
		"  collection: CriteriaRecords\n" +
		"  record_id_field: ID\n" +
		"  json_fields:\n" +
		"    - criteria\n"
	if err := os.WriteFile(path, []byte(yaml), 0o644); err != nil {
		t.Fatal(err)
	}
	m, err := loadArchetypes(path)
	if err != nil {
		t.Fatalf("loadArchetypes: %v", err)
	}
	a, ok := m["scope_criteria"]
	if !ok {
		t.Fatal("scope_criteria not loaded")
	}
	if a.Kind != compoundKindParentRead {
		t.Errorf("kind = %q", a.Kind)
	}
	if a.ParentIDField != "scope_id" || a.ChildIDField != "criteria_id" || a.ChildIDParam != "CriteriaID" {
		t.Errorf("ids = %q/%q/%q", a.ParentIDField, a.ChildIDField, a.ChildIDParam)
	}
	if a.Collection != "CriteriaRecords" || a.RecordIDField != "ID" {
		t.Errorf("collection/record = %q/%q", a.Collection, a.RecordIDField)
	}
	if len(a.JSONFields) != 1 || a.JSONFields[0] != "criteria" {
		t.Errorf("json_fields = %v", a.JSONFields)
	}
}

func TestLoadArchetypes_missingFile(t *testing.T) {
	m, err := loadArchetypes(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatalf("missing file should be a no-op, got %v", err)
	}
	if len(m) != 0 {
		t.Errorf("want empty map, got %d entries", len(m))
	}
}
