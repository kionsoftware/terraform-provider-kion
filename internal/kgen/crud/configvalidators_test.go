package crud_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Every constraint declared in codegen/config_validators.yaml must reach the
// generated resource, and every emitted validator must be declared.
//
// The declaration is the reviewed artifact; the generated method is derived. A
// one-way check would let a regeneration silently drop a constraint and turn a
// plan-time diagnostic back into the API error it exists to replace -- which
// looks like nothing at all until someone writes a configuration missing both
// attributes.
func TestConfigValidatorsMatchDeclaration(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "codegen", "config_validators.yaml"))
	require.NoError(t, err)

	var declared map[string]struct {
		Body         string   `yaml:"body"`
		AtLeastOneOf []string `yaml:"at_least_one_of"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &declared))
	require.NotEmpty(t, declared)

	emitted := map[string][]string{}
	root := filepath.Join("..", "..", "service")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)

	// framework.AtLeastOneOfOnCreate(&resp.Schema, "a", "b") in Schema().
	atLeastOneOfCall := regexp.MustCompile(`framework\.AtLeastOneOfOnCreate\(&resp\.Schema((?:, "[a-z0-9_]+")+)\)`)
	quoted := regexp.MustCompile(`"([a-z0-9_]+)"`)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		files, err := filepath.Glob(filepath.Join(root, e.Name(), "*.go"))
		require.NoError(t, err)
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			src, err := os.ReadFile(f)
			require.NoError(t, err)
			body := string(src)
			g := atLeastOneOfCall.FindStringSubmatch(body)
			if g == nil {
				continue
			}
			var attrs []string
			for _, m := range quoted.FindAllStringSubmatch(g[1], -1) {
				attrs = append(attrs, m[1])
			}
			sort.Strings(attrs)
			emitted[e.Name()] = attrs
		}
	}

	for pkg, d := range declared {
		if len(d.AtLeastOneOf) == 0 {
			continue // required_when / required_together only; guarded below
		}
		want := append([]string(nil), d.AtLeastOneOf...)
		sort.Strings(want)
		got, ok := emitted[pkg]
		assert.Truef(t, ok,
			"%s declares at_least_one_of %v but its resource emits no AtLeastOneOfOnCreate; "+
				"run `make crud-force`", pkg, want)
		if ok {
			assert.Equalf(t, want, got, "%s: declared and emitted constraints differ", pkg)
		}
	}

	for pkg := range emitted {
		_, ok := declared[pkg]
		assert.Truef(t, ok,
			"%s emits AtLeastOneOfOnCreate with no entry in codegen/config_validators.yaml", pkg)
	}
}

// A validator naming an attribute the schema does not declare is a runtime
// panic, not a compile error -- and the API body's spelling is not always the
// provider's (kion_project_enforcement exposes user_group_ids where its two
// siblings expose ugroup_ids), so this is a real way to get it wrong.
func TestConfigValidatorAttributesExist(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "codegen", "config_validators.yaml"))
	require.NoError(t, err)

	var declared map[string]struct {
		AtLeastOneOf []string `yaml:"at_least_one_of"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &declared))

	attrDecl := regexp.MustCompile(`"([a-z0-9_]+)":\s*schema\.`)

	for pkg, d := range declared {
		matches, err := filepath.Glob(filepath.Join("..", "..", "service", pkg, "*_schema_gen.go"))
		require.NoError(t, err)
		require.NotEmptyf(t, matches, "%s: no schema file", pkg)

		src, err := os.ReadFile(matches[0])
		require.NoError(t, err)
		body := string(src)

		// Scope to the resource schema; the data source declares its own set.
		if i := strings.Index(body, "ResourceSchema(ctx context.Context) schema.Schema"); i >= 0 {
			body = body[i:]
		}
		attrs := map[string]bool{}
		for _, m := range attrDecl.FindAllStringSubmatch(body, -1) {
			attrs[m[1]] = true
		}
		for _, a := range d.AtLeastOneOf {
			assert.Truef(t, attrs[a],
				"%s: config_validators.yaml names %q, which its resource schema does not declare", pkg, a)
		}
	}
}

// TestRequiredWhenMatchDeclaration is the other half of
// TestConfigValidatorsMatchDeclaration: that one reads AtLeastOneOfOnCreate, which
// required_when does not emit, so without this a dropped rule passes both.
func TestRequiredWhenMatchDeclaration(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "codegen", "config_validators.yaml"))
	require.NoError(t, err)

	var declared map[string]struct {
		RequiredWhen []struct {
			Attribute string   `yaml:"attribute"`
			Equals    int64    `yaml:"equals"`
			Require   []string `yaml:"require"`
		} `yaml:"required_when"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &declared))

	// framework.RequiredWhenInt64("attr", N, followed by the required names.
	call := regexp.MustCompile(`framework\.RequiredWhenInt64\("([a-z0-9_]+)",\s*(\d+),([^)]*)\)`)
	name := regexp.MustCompile(`"([a-z0-9_]+)"`)

	root := filepath.Join("..", "..", "service")
	for pkg, d := range declared {
		if len(d.RequiredWhen) == 0 {
			continue
		}
		files, err := filepath.Glob(filepath.Join(root, pkg, "*.go"))
		require.NoError(t, err)

		emitted := map[string][]string{} // "attr=N" -> required names
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			src, err := os.ReadFile(f)
			require.NoError(t, err)
			for _, m := range call.FindAllStringSubmatch(string(src), -1) {
				var req []string
				for _, n := range name.FindAllStringSubmatch(m[3], -1) {
					req = append(req, n[1])
				}
				sort.Strings(req)
				emitted[m[1]+"="+m[2]] = req
			}
		}

		for _, rw := range d.RequiredWhen {
			key := fmt.Sprintf("%s=%d", rw.Attribute, rw.Equals)
			want := append([]string(nil), rw.Require...)
			sort.Strings(want)
			got, ok := emitted[key]
			assert.Truef(t, ok,
				"%s declares required_when %s but its resource emits no matching validator; run `make crud-force`",
				pkg, key)
			if ok {
				assert.Equalf(t, want, got, "%s %s: declared and emitted differ", pkg, key)
			}
		}
	}
}

// TestRequiredTogetherMatchDeclaration checks required_together both ways, like
// the two tests above do for their kinds.
func TestRequiredTogetherMatchDeclaration(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "codegen", "config_validators.yaml"))
	require.NoError(t, err)

	var declared map[string]struct {
		RequiredTogether [][]string `yaml:"required_together"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &declared))

	call := regexp.MustCompile(`(?s)resourcevalidator\.RequiredTogether\((.*?)\n\s*\),`)
	matchRoot := regexp.MustCompile(`path\.MatchRoot\("([a-z0-9_]+)"\)`)

	root := filepath.Join("..", "..", "service")
	entries, err := os.ReadDir(root)
	require.NoError(t, err)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		pkg := e.Name()
		files, err := filepath.Glob(filepath.Join(root, pkg, "*.go"))
		require.NoError(t, err)
		var emitted []string
		for _, f := range files {
			if strings.HasSuffix(f, "_test.go") {
				continue
			}
			src, err := os.ReadFile(f)
			require.NoError(t, err)
			for _, m := range call.FindAllStringSubmatch(string(src), -1) {
				var group []string
				for _, n := range matchRoot.FindAllStringSubmatch(m[1], -1) {
					group = append(group, n[1])
				}
				sort.Strings(group)
				emitted = append(emitted, strings.Join(group, ","))
			}
		}
		var want []string
		for _, g := range declared[pkg].RequiredTogether {
			g = append([]string(nil), g...)
			sort.Strings(g)
			want = append(want, strings.Join(g, ","))
		}
		sort.Strings(want)
		sort.Strings(emitted)
		assert.Equalf(t, want, emitted,
			"%s: declared and emitted required_together differ; run `make crud-force`", pkg)
	}
}
