package migrate

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"text/template"
)

func contains(ss []string, s string) bool { return slices.Contains(ss, s) }

// addedNestedFields returns the fields a new nested-object attribute has that
// the old block it passes through from lacks, sorted.
func addedNestedFields(oa, na Attr) []string {
	if oa.Kind != "block" || !na.NestedObj {
		return nil
	}
	var extra []string
	for _, f := range na.NestedAttrs {
		if !contains(oa.NestedAttrs, f) {
			extra = append(extra, f)
		}
	}
	sort.Strings(extra)
	return extra
}

func quoteJoin(ss []string) string {
	q := make([]string, len(ss))
	for i, s := range ss {
		q[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(q, ", ")
}

//go:embed upgrade.gtpl
var upgradeTmpl string

type fieldEmit struct {
	TF   string // new attribute name (JSON key)
	Expr string // json.RawMessage expression for the new value
}

type upgradeData struct {
	Pkg    string
	Recv   string // resource receiver type, e.g. "user_groupResource" or "awsAccountResource"
	Fields []fieldEmit
}

var metadataRecvRe = regexp.MustCompile(`func \(\w+ \*(\w+Resource)\) Metadata\(`)

// parseReceiver finds the resource receiver type name in <name>.go (handles both
// the entity convention <name>Resource and bespoke lowerCamel names).
func parseReceiver(root, name string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(root, "internal", "service", name, name+".go"))
	if err != nil {
		return "", err
	}
	if m := metadataRecvRe.FindSubmatch(raw); m != nil {
		return string(m[1]), nil
	}
	return "", fmt.Errorf("%s: could not find resource receiver in %s.go", name, name)
}

// GenerateUpgrade renders <name>_upgrade_gen.go for one migrated resource. The
// upgrader transforms the old-state JSON map into a new-schema JSON map and
// decodes it with tftypes.ValueFromJSON, so every attribute type migrates
// uniformly, and mismatches (e.g. an object shape that changed) fail loudly in
// ValueFromJSON rather than silently.
func GenerateUpgrade(root, tfType string, t Transform, oldRes, newRes Resource) ([]byte, error) {
	name := strings.TrimPrefix(tfType, "kion_")
	recv, err := parseReceiver(root, name)
	if err != nil {
		return nil, err
	}

	renameSrc := map[string]string{} // new_attr -> old_attr
	for oldA, newA := range t.Rename {
		renameSrc[newA] = oldA
	}

	tfs := make([]string, 0, len(newRes.Attrs))
	for a := range newRes.Attrs {
		tfs = append(tfs, a)
	}
	sort.Strings(tfs)

	d := upgradeData{Pkg: name, Recv: recv}
	for _, tf := range tfs {
		e := fieldEmit{TF: tf}
		switch {
		case t.Project[tf].From != "":
			pr := t.Project[tf]
			e.Expr = fmt.Sprintf("migratehelper.ProjectIDs(old[%q], %q)", pr.From, pr.Field)
		case t.KVList[tf].KeyField != "":
			kv := t.KVList[tf]
			e.Expr = fmt.Sprintf("migratehelper.MapToObjectList(old[%q], %q, %q)", tf, kv.KeyField, kv.ValField)
		case tf == "id" && t.IDInt:
			e.Expr = `migratehelper.StringToNum(old["id"])`
		case contains(t.Unwrap, tf):
			e.Expr = fmt.Sprintf("migratehelper.Unwrap(old[%q])", tf)
		default:
			src := ""
			if o, ok := renameSrc[tf]; ok {
				src = o
			} else if _, ok := oldRes.Attrs[tf]; ok {
				src = tf
			}
			if src == "" {
				e.Expr = "migratehelper.Null"
			} else if extra := addedNestedFields(oldRes.Attrs[src], newRes.Attrs[tf]); len(extra) > 0 {
				e.Expr = fmt.Sprintf("migratehelper.PadObjectKeys(old[%q], %s)", src, quoteJoin(extra))
			} else {
				e.Expr = fmt.Sprintf("migratehelper.OrNull(old[%q])", src)
			}
		}
		d.Fields = append(d.Fields, e)
	}

	var buf bytes.Buffer
	if err := template.Must(template.New("upgrade").Parse(upgradeTmpl)).Execute(&buf, d); err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}
