package crud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Kion returns an empty membership list as null, and a configured [] then failed
// the apply. These are the sets a live 0.3.26 -> 1.8.1 migration failed on; each
// must keep a configured [] through the generated flatten.
func TestEmptySetsKeepConfiguredEmpty(t *testing.T) {
	t.Parallel()
	for pkg, fields := range map[string][]string{
		"project_cloud_access_role": {"AccountIds", "AwsIamPolicies", "AzureRoleDefinitions", "GcpIamRoles", "UserGroupIds", "UserIds"},
		"ou_cloud_access_role":      {"AwsIamPolicies", "AzureRoleDefinitions", "GcpIamRoles", "UserGroupIds", "UserIds"},
		"azure_role":                {"OwnerUserGroupIds", "OwnerUserIds"},
	} {
		src, err := os.ReadFile(filepath.Join("..", "..", "service", pkg, pkg+".go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range fields {
			if !strings.Contains(string(src), "EmptyLikePrior(ctx, model."+f+",") {
				t.Errorf("%s.%s: the read does not keep a configured [] when Kion returns null; run `make crud-force`", pkg, f)
			}
		}
	}
}
