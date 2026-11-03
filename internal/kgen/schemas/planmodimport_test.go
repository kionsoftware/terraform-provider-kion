package schemas

import "testing"

// The provider's own modifiers (framework.CreateOnlyString) import from
// internal/framework, not from the terraform-plugin-framework planmodifier tree.
func TestPlanModifierImportProviderFramework(t *testing.T) {
	for call, want := range map[string]string{
		"stringplanmodifier.RequiresReplace()": planModifierBase + "stringplanmodifier",
		"framework.CreateOnlyString()":         "terraform-provider-kion/internal/framework",
	} {
		got, err := planModifierImport(call)
		if err != nil || got != want {
			t.Errorf("planModifierImport(%q) = %q, %v; want %q", call, got, err, want)
		}
	}
}
