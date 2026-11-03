package flex

import (
	"encoding/json"
	"github.com/go-faster/jx"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	generated "github.com/kionsoftware/kion-sdk-go/generated/v3_17"
	"strings"
)

// Converters between a Terraform jsontypes.Normalized attribute (a JSON string
// compared for semantic equality) and the SDK's jx.Raw (raw JSON bytes).

// NormalizedFromFramework converts a jsontypes.Normalized to jx.Raw.
// Null or unknown values produce a nil jx.Raw.
func NormalizedFromFramework(v jsontypes.Normalized) jx.Raw {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	return jx.Raw(v.ValueString())
}

// NormalizedToFramework converts jx.Raw to a jsontypes.Normalized.
// Empty raw JSON produces a null value.
func NormalizedToFramework(v jx.Raw) jsontypes.Normalized {
	// See canonicalJSON.
	if len(v) == 0 {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(canonicalJSON(string(v)))
}

// NormalizedStringFromFramework converts a jsontypes.Normalized to the plain
// string the SDK carries for JSON fields it does not model as jx.Raw.
func NormalizedStringFromFramework(v jsontypes.Normalized) string {
	return v.ValueString()
}

// NormalizedStringToFramework is its inverse, canonicalising on the way in.
func NormalizedStringToFramework(v string) jsontypes.Normalized {
	return jsontypes.NewNormalizedValue(canonicalJSON(v))
}

// canonicalJSON re-encodes a JSON string the way Terraform's jsonencode does:
// compact, with object keys sorted, which is what encoding/json produces from a
// decoded map.
//
// Semantic equality is not consulted when a generated configuration is planned:
// `terraform plan -generate-config-out` writes the attribute as jsonencode({…}),
// and comparing that against the API's own spacing reported a change on every
// record. 1,348 IAM policies, 827 Azure roles and 293 ARM templates planned an
// update that changed nothing. Byte equality does not depend on the framework
// consulting anything.
//
// A value that is not JSON is returned unchanged: CloudFormation templates are
// often YAML, and rewriting one would be worse than leaving it alone.
func canonicalJSON(v string) string {
	var any0 any
	if err := json.Unmarshal([]byte(v), &any0); err != nil {
		return v
	}
	out, err := json.Marshal(any0)
	if err != nil {
		return v
	}
	return string(out)
}

// OptNormalizedStringFromFramework converts to the SDK's optional string.
func OptNormalizedStringFromFramework(v jsontypes.Normalized) generated.OptString {
	if v.IsNull() || v.IsUnknown() {
		return generated.OptString{}
	}
	return generated.NewOptString(v.ValueString())
}

// OptNormalizedStringToFramework is its inverse; an unset value becomes null.
func OptNormalizedStringToFramework(v generated.OptString) jsontypes.Normalized {
	if !v.Set {
		return jsontypes.NewNormalizedNull()
	}
	return jsontypes.NewNormalizedValue(canonicalJSON(v.Value))
}

// CanonicalJSONString re-encodes a JSON object or array the way Terraform's
// jsonencode does, and returns anything else untouched.
//
// `terraform plan -generate-config-out` writes a JSON-valued attribute as
// jsonencode({…}), which evaluates to compact JSON with sorted keys. Comparing
// that against the API's own spacing reported a change on every record that
// carried JSON. Restricting the rewrite to values that begin with "{" or "["
// keeps it away from ordinary strings that happen to parse, such as "123" or a
// bare word, and leaves YAML CloudFormation templates alone.
func CanonicalJSONString(v string) string {
	t := strings.TrimSpace(v)
	if !strings.HasPrefix(t, "{") && !strings.HasPrefix(t, "[") {
		return v
	}
	return canonicalJSON(t)
}
