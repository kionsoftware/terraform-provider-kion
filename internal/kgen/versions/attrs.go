package versions

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"terraform-provider-kion/internal/kgen/crud"
	"terraform-provider-kion/internal/kgen/kfs"
)

// Resource gating asks "does this operation exist here". That misses the other
// half: an operation can exist on an older Kion while a field on its request
// body does not. Portal decodes with json.Unmarshal and no DisallowUnknownFields,
// so an unknown field is not rejected. It is dropped. The value never lands,
// the read returns without it, and Terraform reports an inconsistent result or
// diffs forever, neither of which names the real cause.
//
// deriveAttrWindows answers, per resource, the tracked-version window in which
// each tfsdk attribute is accepted. It walks the create body struct in every
// tracked SDK version and records the first and last one carrying each field.
// Attributes present for the whole life of the operation are omitted.
func deriveAttrWindows(src crud.Source, sdkDir, serviceRoot string, entries map[string]entry, logw io.Writer) map[string]map[string][]attrWindow {
	out := map[string]map[string][]attrWindow{}

	for name, e := range entries {
		if e.Create == nil {
			continue // nothing to send, nothing to gate
		}
		key := strings.ToUpper(e.Create.Method) + " " + e.Create.Path

		// Go field name -> first and last tracked version index carrying it.
		earliest := map[string]int{}
		latest := map[string]int{}
		// Union across versions, so a field the newest release dropped is still
		// considered.
		anyField := map[string]crud.Field{}
		// Newest version carrying the op, so a dropped attribute is not confused
		// with a dropped endpoint (the resource gate covers that).
		lastOpAt := -1

		for i, v := range trackedVersions {
			gen := filepath.Join(sdkDir, "generated", v.dir)
			methods, err := src.ClientMethods(filepath.Join(gen, "oas_client_gen.go"))
			if err != nil {
				continue // version not present locally; treated as no data
			}
			var body string
			for _, m := range methods {
				if strings.ToUpper(m.HTTPMethod)+" "+m.Path == key {
					body = strings.TrimPrefix(m.BodyType, "Opt")
					break
				}
			}
			if body == "" {
				continue // op absent in this version, resource gating covers that
			}
			structs, err := src.Structs(filepath.Join(gen, "oas_schemas_gen.go"))
			if err != nil {
				continue
			}
			st, ok := structs[body]
			if !ok {
				continue
			}
			lastOpAt = i
			for _, f := range st.Fields {
				if _, seen := earliest[f.GoName]; !seen {
					earliest[f.GoName] = i
				}
				latest[f.GoName] = i
				anyField[f.GoName] = f
			}
		}
		if lastOpAt < 0 {
			continue
		}

		// Map SDK field -> tfsdk attribute via the generated model, so the
		// diagnostic names what the practitioner actually wrote.
		pascal := pascalFor(name)
		models, err := src.ModelFields(
			filepath.Join(serviceRoot, name, name+"_schema_gen.go"), pascal+"Model")
		if err != nil {
			fmt.Fprintf(logw, "attr-versions: %s: no model (%v); skipping\n", name, err)
			continue
		}
		// Keyed case-insensitively: the SDK writes acronyms in Go style
		// (CloudAccessRoleTypeID, AWSPartition) and tfplugingen does not
		// (CloudAccessRoleTypeId, AwsPartition), so an exact match silently
		// skips every attribute carrying one, foreign keys included.
		tfByGo := make(map[string]string, len(models))
		for _, m := range models {
			tfByGo[strings.ToLower(m.GoName)] = m.TFSDK
		}

		windows := map[string][]attrWindow{}
		for goName := range anyField {
			first, ok := earliest[goName]
			if !ok {
				continue
			}
			last := latest[goName]
			// Present for the whole life of the operation: no gate needed.
			if first == 0 && last == lastOpAt {
				continue
			}
			tf, ok := tfByGo[strings.ToLower(goName)]
			if !ok {
				continue // not surfaced as a Terraform attribute
			}
			var w attrWindow
			if first > 0 {
				w.Min = versionString(trackedVersions[first])
			}
			if last < lastOpAt {
				// Exclusive: carried through trackedVersions[last], so the bound
				// is the start of the next tracked line.
				w.Before = versionString(trackedVersions[last+1])
			}
			windows[tf] = []attrWindow{w}
		}
		if len(windows) > 0 {
			out[name] = windows
		}
	}
	return out
}

// attrWindow is one derived range: Min inclusive, Before exclusive. An empty
// value means unbounded on that side.
type attrWindow struct {
	Min    string
	Before string
}

// pascalFor converts a snake_case package name to the PascalCase prefix the
// generated model uses: ou_note -> OuNote.
func pascalFor(name string) string {
	parts := strings.Split(name, "_")
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

// renderAttrMins renders the sorted attribute->version map literal body.
func renderAttrMins(windows map[string][]attrWindow) string {
	if len(windows) == 0 {
		return ""
	}
	names := make([]string, 0, len(windows))
	for k := range windows {
		names = append(names, k)
	}
	sort.Strings(names)

	var b strings.Builder
	for _, n := range names {
		rs := windows[n]
		parts := make([]string, 0, len(rs))
		for _, w := range rs {
			var fields []string
			if w.Min != "" {
				fields = append(fields, fmt.Sprintf("Min: conns.MustParseKionVersion(%q)", w.Min))
			}
			if w.Before != "" {
				fields = append(fields, fmt.Sprintf("Before: conns.MustParseKionVersion(%q)", w.Before))
			}
			if len(fields) == 0 {
				fields = append(fields, "Min: conns.KionVersion{}")
			}
			parts = append(parts, "{"+strings.Join(fields, ", ")+"}")
		}
		if bounded := upperBound(rs); bounded != "" {
			fmt.Fprintf(&b, "\t// Not accepted from %s; sending it there is silently ignored.\n", bounded)
		}
		fmt.Fprintf(&b, "\t%q: {%s},\n", n, strings.Join(parts, ", "))
	}
	return b.String()
}

// upperBound is the version from which no range admits the attribute, or ""
// when one is open-ended. Authored ranges keep file order, so this does not
// assume the last one is the highest.
func upperBound(rs []attrWindow) string {
	top := ""
	for _, w := range rs {
		if w.Before == "" {
			return ""
		}
		if top == "" || versionLess(top, w.Before) {
			top = w.Before
		}
	}
	return top
}

// versionLess compares dotted numeric versions ("3.16.5" < "3.17.0").
func versionLess(a, b string) bool {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		if x, y := versionPart(as, i), versionPart(bs, i); x != y {
			return x < y
		}
	}
	return false
}

// versionPart is the i'th numeric component, 0 when absent or not a number.
func versionPart(parts []string, i int) int {
	if i >= len(parts) {
		return 0
	}
	n, err := strconv.Atoi(parts[i])
	if err != nil {
		return 0
	}
	return n
}

// pruneRedundant drops attribute minimums at or below the resource's own
// minimum. The resource gate already refuses those, and emitting both would
// report one cause twice, every field of a 3.14-only resource is trivially
// "3.14+". What remains is the interesting case: a field newer than the
// resource carrying it.
func pruneRedundant(windows map[string][]attrWindow, resourceMin string) map[string][]attrWindow {
	if len(windows) == 0 {
		return nil
	}
	floor := minorOf(resourceMin)
	out := make(map[string][]attrWindow, len(windows))
	for attr, rs := range windows {
		keep := false
		for _, w := range rs {
			// An upper bound is never redundant with a floor.
			if w.Before != "" || minorOf(w.Min) > floor {
				keep = true
				break
			}
		}
		if keep {
			out[attr] = rs
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// minorOf extracts NN from "3.NN.0"; an unparseable or empty string is 0, so an
// ungated resource keeps every attribute minimum.
func minorOf(v string) int {
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return 0
	}
	n := 0
	for _, r := range parts[1] {
		if r < '0' || r > '9' {
			return 0
		}
		n = n*10 + int(r-'0')
	}
	return n
}

// attrOverridesFile is codegen/attr_version_overrides.yaml.
type attrOverridesFile struct {
	Resources map[string]map[string][]struct {
		Min    string `yaml:"min"`
		Before string `yaml:"before"`
	} `yaml:"resources"`
	ResourceFloors map[string]string `yaml:"resource_floors"`
}

// applyAttrOverrides replaces derived ranges with authored ones, in place. The
// SDK ships one package per Kion minor, so derivation cannot express a floor
// like 3.15.13, nor a different floor per support line. A missing file is not
// an error; an entry naming an attribute derivation did not find is reported,
// since a window nothing reads is worse than no window.
func applyAttrOverrides(windows map[string]map[string][]attrWindow, path string, fs kfs.FS, logw io.Writer) {
	b, err := fs.ReadFile(path)
	if err != nil {
		return
	}
	f, err := decodeAttrOverrides(b)
	if err != nil {
		fmt.Fprintf(logw, "attr-versions: %s: %v; ignoring\n", path, err)
		return
	}

	for _, res := range sortedResourceNames(f.Resources) {
		for _, attr := range sortedAttrNames(f.Resources[res]) {
			ranges := f.Resources[res][attr]
			if len(ranges) == 0 {
				continue
			}
			if _, ok := windows[res][attr]; !ok {
				fmt.Fprintf(logw, "attr-versions: %s.%s: override matches no derived attribute; check the name\n", res, attr)
				continue
			}
			out := make([]attrWindow, 0, len(ranges))
			for _, r := range ranges {
				out = append(out, attrWindow{Min: r.Min, Before: r.Before})
			}
			windows[res][attr] = out
		}
	}
}

func sortedResourceNames[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func sortedAttrNames[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
