package crud

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ConfigValidatorsPath is the authored list of cross-field API constraints.
const ConfigValidatorsPath = "codegen/config_validators.yaml"

// ConfigValidator is a constraint the API enforces across two or more
// attributes, which an attribute-level schema cannot express: "required" is
// per-attribute, and neither of an either-or pair is required on its own.
//
// Authored rather than derived. The OpenAPI spec carries almost no constraints
// (3 enums, 16 patterns, 5 maxLength across 474 schemas), so the source is the
// portal's own `validate:"atLeastOneFieldPresent=..."` struct tags.
type ConfigValidator struct {
	Body         string   `yaml:"body"`
	AtLeastOneOf []string `yaml:"at_least_one_of"`
	// RequiredWhen covers a constraint conditional on another attribute's
	// value, which none of the framework's own validators express.
	RequiredWhen []RequiredWhen `yaml:"required_when"`
	// RequiredTogether groups attributes that must be set all or none.
	RequiredTogether [][]string `yaml:"required_together"`
}

// RequiredWhen requires Require once the Attribute attribute equals Equals.
type RequiredWhen struct {
	Attribute string   `yaml:"attribute"`
	Equals    int64    `yaml:"equals"`
	Require   []string `yaml:"require"`
}

// ConfigValidatorPolicy answers what a package's resource must validate.
type ConfigValidatorPolicy struct {
	byPkg map[string]ConfigValidator
}

// LoadConfigValidators reads ConfigValidatorsPath. A missing file is an error
// rather than an empty policy: silently dropping every cross-field check
// because a file went missing would turn a plan-time diagnostic back into the
// API error it exists to replace, with nothing to indicate it had happened.
func LoadConfigValidators(root string) (ConfigValidatorPolicy, error) {
	path := filepath.Join(root, ConfigValidatorsPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return ConfigValidatorPolicy{}, fmt.Errorf("reading %s: %w", ConfigValidatorsPath, err)
	}
	var entries map[string]ConfigValidator
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		return ConfigValidatorPolicy{}, fmt.Errorf("parsing %s: %w", ConfigValidatorsPath, err)
	}
	for pkg, v := range entries {
		if len(v.AtLeastOneOf) > 0 && len(v.AtLeastOneOf) < 2 {
			return ConfigValidatorPolicy{}, fmt.Errorf(
				"%s: %s.at_least_one_of needs at least two attributes, got %d",
				ConfigValidatorsPath, pkg, len(v.AtLeastOneOf))
		}
		if len(v.AtLeastOneOf) == 0 && len(v.RequiredWhen) == 0 && len(v.RequiredTogether) == 0 {
			return ConfigValidatorPolicy{}, fmt.Errorf(
				"%s: %s declares none of at_least_one_of, required_when, required_together",
				ConfigValidatorsPath, pkg)
		}
		for _, g := range v.RequiredTogether {
			if len(g) < 2 {
				return ConfigValidatorPolicy{}, fmt.Errorf(
					"%s: %s.required_together groups need at least two attributes, got %v",
					ConfigValidatorsPath, pkg, g)
			}
		}
		for _, rw := range v.RequiredWhen {
			if rw.Attribute == "" || len(rw.Require) == 0 {
				return ConfigValidatorPolicy{}, fmt.Errorf(
					"%s: %s.required_when needs an attribute and at least one require entry",
					ConfigValidatorsPath, pkg)
			}
		}
	}
	return ConfigValidatorPolicy{byPkg: entries}, nil
}

// For returns the attributes pkg must require at least one of, or nil.
func (p ConfigValidatorPolicy) For(pkg string) []string {
	return p.byPkg[pkg].AtLeastOneOf
}

// RequiredTogetherFor returns pkg's all-or-none attribute groups, or nil.
func (p ConfigValidatorPolicy) RequiredTogetherFor(pkg string) [][]string {
	return p.byPkg[pkg].RequiredTogether
}

// RequiredWhenFor returns pkg's value-conditional constraints, or nil.
func (p ConfigValidatorPolicy) RequiredWhenFor(pkg string) []RequiredWhen {
	return p.byPkg[pkg].RequiredWhen
}
