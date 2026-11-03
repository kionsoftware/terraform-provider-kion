package migrate

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"go/format"
	"sort"
	"strings"
	"text/template"
)

//go:embed upgrade_test.gtpl
var upgradeTestTmpl string

type upgradeTestData struct {
	Pkg     string
	Recv    string
	Fixture string // synthesized old-state JSON (a Go raw-string literal body)
}

// GenerateUpgradeTest renders <name>_upgrade_gen_test.go: a decode-clean golden
// that feeds a synthesized representative old-state through the v0 upgrader and
// asserts it decodes against the live schema with no diagnostics. This is the
// per-resource "the generated mapping fits the real schema" guarantee, it
// catches a wrong/missing/extra key or a type mismatch in the emitted new-state
// map. Value-level correctness of the transforms is covered by
// internal/migratehelper unit tests and the two hand-authored deep goldens.
func GenerateUpgradeTest(root, tfType string, t Transform, oldRes, newRes Resource) ([]byte, error) {
	name := strings.TrimPrefix(tfType, "kion_")
	recv, err := parseReceiver(root, name)
	if err != nil {
		return nil, err
	}
	d := upgradeTestData{
		Pkg:     name,
		Recv:    recv,
		Fixture: synthOldState(t, oldRes, newRes),
	}
	var buf bytes.Buffer
	if err := template.Must(template.New("upgradeTest").Parse(upgradeTestTmpl)).Execute(&buf, d); err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}

// synthOldState builds a representative old (SDKv2) state JSON object for the
// upgrader to consume. It populates only the attributes whose exact JSON shape is
// knowable from the transform config, block-sets the upgrader projects, the id
// it may retype, and renamed sources, leaving every other old attribute absent
// so it decodes as null (valid for any type). That keeps the fixture decode-safe
// without needing per-resource nested-object shapes, while still exercising every
// transform-bearing key and forcing the emitted key set to match the new schema.
func synthOldState(t Transform, oldRes, newRes Resource) string {
	fields := map[string]string{} // old JSON key -> JSON value literal

	// Block-sets the upgrader projects to id-lists: [{<field>: n}, …].
	for _, pr := range t.Project {
		if pr.From != "" {
			fields[pr.From] = fmt.Sprintf(`[{%q: 1}, {%q: 2}]`, pr.Field, pr.Field)
		}
	}
	// Single-nested-block unwraps: an empty list unwraps to null, which is
	// shape-free and always decodes. The deep aws_account golden covers a
	// populated unwrap.
	for _, u := range t.Unwrap {
		fields[u] = `[]`
	}
	// Map attributes exploded into object lists. This one MUST be populated: an
	// absent map decodes as null against any type, so an empty fixture would not
	// notice the transform missing, which is exactly how cft's tags shipped
	// passing a map straight into a list-of-objects schema.
	for attr := range t.KVList {
		fields[attr] = mapSample(oldRes.Attrs[attr].TypeJSON)
	}
	// Terraform SDKv2 always stores id as a string; the new schema keeps it a
	// string unless IDInt retypes it to a number (StringToNum handles that).
	if t.IDInt {
		fields["id"] = `"1234"`
	} else if _, ok := oldRes.Attrs["id"]; ok {
		fields["id"] = `"1234"`
	}
	// Renamed sources: give the old key a representative value of the NEW
	// attribute's scalar type so the rename lands under the right key and type.
	for oldA, newA := range t.Rename {
		if v, ok := scalarSample(newRes.Attrs[newA].TypeJSON); ok {
			fields[oldA] = v
		}
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var b strings.Builder
	b.WriteString("{\n")
	for i, k := range keys {
		comma := ","
		if i == len(keys)-1 {
			comma = ""
		}
		fmt.Fprintf(&b, "\t\t%q: %s%s\n", k, fields[k], comma)
	}
	b.WriteString("\t}")
	return b.String()
}

// mapSample returns a two-entry JSON object for an old map attribute's cty
// type-JSON (`["map", <element>]`), with element-typed values so the exploded
// objects decode against the new nested field types. Two entries, so the
// generated fixture also proves the explode handles more than a single pair.
func mapSample(typeJSON string) string {
	elem := `"string"`
	var t []json.RawMessage
	if err := json.Unmarshal([]byte(typeJSON), &t); err == nil && len(t) == 2 {
		elem = string(t[1])
	}
	v, ok := scalarSample(elem)
	if !ok {
		v = `"sample"`
	}
	return fmt.Sprintf(`{"alpha": %s, "beta": %s}`, v, v)
}

// scalarSample returns a representative JSON literal for a simple cty scalar
// type-JSON, and false for anything non-scalar (which the caller then leaves
// absent → null).
func scalarSample(typeJSON string) (string, bool) {
	switch typeJSON {
	case `"string"`:
		return `"sample"`, true
	case `"number"`:
		return `1`, true
	case `"bool"`:
		return `true`, true
	}
	return "", false
}
