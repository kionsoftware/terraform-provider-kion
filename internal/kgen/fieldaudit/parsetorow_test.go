package fieldaudit

import (
	"os"
	"path/filepath"
	"testing"
)

// writeDS writes src to a temp file and parses it with parseToRow.
func writeDS(t *testing.T, src string) (elem string, keys, read, reached map[string]bool) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "x_data_source.go")
	if err := os.WriteFile(path, []byte(src), 0o600); err != nil {
		t.Fatal(err)
	}
	elem, keys, read, reached, err := parseToRow(path)
	if err != nil {
		t.Fatal(err)
	}
	return elem, keys, read, reached
}

// TestParseToRowRecordsRenamedFields is the case the audit used to miscount: a
// data source may publish an SDK field under a different name, and comparing
// JSON names alone reports that as a gap. The bespoke cloud_account data source
// emits AccountEmail as `email` and AccountName as `name`.
func TestParseToRowRecordsRenamedFields(t *testing.T) {
	t.Parallel()
	elem, keys, read, _ := writeDS(t, `package x

func accountToRow(a generated.Account) map[string]any {
	row := map[string]any{
		"email": a.AccountEmail.Value,
		"name":  a.AccountName.Value,
	}
	if a.ID.Set {
		row["id"] = int64(a.ID.Value)
	}
	return row
}
`)

	if elem != "Account" {
		t.Errorf("elem = %q, want Account", elem)
	}
	for _, k := range []string{"email", "name", "id"} {
		if !keys[k] {
			t.Errorf("key %q not recorded", k)
		}
	}
	// The point of the fix: the SDK field is recorded even though no emitted
	// key is spelled account_email.
	for _, f := range []string{"AccountEmail", "AccountName", "ID"} {
		if !read[f] {
			t.Errorf("SDK field %q not recorded as read", f)
		}
	}
	if read["Value"] || read["Set"] {
		t.Error("ogen wrapper accessors must not be recorded as SDK fields")
	}
	if read["AccountNumber"] {
		t.Error("a field the row never reads must not be recorded as read")
	}
}

// TestParseToRowReachesThroughWrapper covers the augmented shapes, where the
// record is nested under the list element and the parameter is not named `a`.
func TestParseToRowReachesThroughWrapper(t *testing.T) {
	t.Parallel()
	_, _, read, reached := writeDS(t, `package x

func amiToRow(lbl generated.AMIWithOwners) map[string]any {
	return map[string]any{
		"name": lbl.Ami.Value.Name.Or(""),
	}
}
`)

	if !reached["Ami"] {
		t.Error("nested wrapper field Ami not recorded as reached")
	}
	if !read["Name"] {
		t.Error("leaf field Name not recorded as read")
	}
	if read["Or"] {
		t.Error("method calls must not be recorded as SDK fields")
	}
}
