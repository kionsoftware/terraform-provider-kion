package module

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// atLeastOneOfByType duplicates codegen/config_validators.yaml, because kgen
// module builds from the provider schema alone and reads no codegen inputs.
// Duplication is the compromise; silent divergence is not.
//
// A missing entry is not a compile error or a lint failure -- it surfaces as a
// generated module whose plan-only test trips the resource's own
// AtLeastOneOf validator, which reads as the validator being wrong rather than
// the fixture being incomplete.
func TestAtLeastOneOfMatchesCodegenDeclaration(t *testing.T) {
	t.Parallel()

	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "codegen", "config_validators.yaml"))
	require.NoError(t, err)

	var declared map[string]struct {
		AtLeastOneOf []string `yaml:"at_least_one_of"`
		RequiredWhen []struct {
			Attribute string `yaml:"attribute"`
		} `yaml:"required_when"`
	}
	require.NoError(t, yaml.Unmarshal(raw, &declared))
	require.NotEmpty(t, declared)

	// Aliases render their own module from the aliased resource's schema, so
	// they carry the constraint without appearing in the codegen file.
	aliases := map[string]string{
		"kion_aws_cloudformation_template": "cft",
		"kion_aws_iam_policy":              "iam_policy",
	}

	for pkg, d := range declared {
		// A required_when validator only fires once its trigger attribute is
		// set, and module fixtures do not set one, so an entry declaring only
		// those needs no fixture value.
		if len(d.AtLeastOneOf) == 0 {
			continue
		}
		typeName := "kion_" + pkg
		got, ok := atLeastOneOfByType[typeName]
		if !assert.Truef(t, ok,
			"codegen/config_validators.yaml declares %s but atLeastOneOfByType has no entry; "+
				"its generated module test will fail the resource's own validator", pkg) {
			continue
		}
		assert.Containsf(t, d.AtLeastOneOf, got,
			"%s: atLeastOneOfByType names %q, which is not in the declared group %v",
			pkg, got, d.AtLeastOneOf)
	}

	for typeName, attr := range atLeastOneOfByType {
		if src, isAlias := aliases[typeName]; isAlias {
			assert.Containsf(t, declared[src].AtLeastOneOf, attr,
				"%s aliases %s, whose declared group does not contain %q", typeName, src, attr)
			continue
		}
		pkg := typeName[len("kion_"):]
		_, ok := declared[pkg]
		assert.Truef(t, ok,
			"atLeastOneOfByType has %s with no entry in codegen/config_validators.yaml", typeName)
	}
}
