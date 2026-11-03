package servicepkg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const servicePackagesContent = `package provider

import (
	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/service/account"
	"terraform-provider-kion/internal/service/billing_rule"
	"terraform-provider-kion/internal/service/label"
	"terraform-provider-kion/internal/service/user_group"
)

// servicePackages returns all registered service packages.
// This function is updated by kgen when scaffolding new resources.
func servicePackages() []conns.ServicePackage {
	return []conns.ServicePackage{
		account.NewServicePackage(),
		billing_rule.NewServicePackage(),
		label.NewServicePackage(),
		user_group.NewServicePackage(),
	}
}
`

func setupTempFile(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	providerDir := filepath.Join(dir, "internal", "provider")
	if err := os.MkdirAll(providerDir, 0750); err != nil {
		t.Fatal(err)
	}
	filePath := filepath.Join(providerDir, "service_packages.go")
	if err := os.WriteFile(filePath, []byte(servicePackagesContent), 0600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func readResult(t *testing.T, root string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(root, "internal", "provider", "service_packages.go"))
	if err != nil {
		t.Fatalf("reading result file: %v", err)
	}
	return string(content)
}

func TestRegisterServicePackage_Middle(t *testing.T) {
	root := setupTempFile(t)
	if err := RegisterServicePackage(root, "cloud_rule"); err != nil {
		t.Fatalf("RegisterServicePackage() error = %v", err)
	}

	text := readResult(t, root)

	if !strings.Contains(text, `"terraform-provider-kion/internal/service/cloud_rule"`) {
		t.Error("import not inserted")
	}
	if !strings.Contains(text, "cloud_rule.NewServicePackage(),") {
		t.Error("registration not inserted")
	}

	// Verify sorted order of imports
	accountIdx := strings.Index(text, `"terraform-provider-kion/internal/service/account"`)
	cloudRuleIdx := strings.Index(text, `"terraform-provider-kion/internal/service/cloud_rule"`)
	labelIdx := strings.Index(text, `"terraform-provider-kion/internal/service/label"`)
	if accountIdx >= cloudRuleIdx || cloudRuleIdx >= labelIdx {
		t.Error("import not in sorted position")
	}

	// Verify sorted order of registrations
	accountRegIdx := strings.Index(text, "account.NewServicePackage(),")
	cloudRuleRegIdx := strings.Index(text, "cloud_rule.NewServicePackage(),")
	labelRegIdx := strings.Index(text, "label.NewServicePackage(),")
	if accountRegIdx >= cloudRuleRegIdx || cloudRuleRegIdx >= labelRegIdx {
		t.Error("registration not in sorted position")
	}
}

func TestRegisterServicePackage_First(t *testing.T) {
	root := setupTempFile(t)
	if err := RegisterServicePackage(root, "aaa"); err != nil {
		t.Fatalf("RegisterServicePackage() error = %v", err)
	}

	text := readResult(t, root)

	aaaIdx := strings.Index(text, `"terraform-provider-kion/internal/service/aaa"`)
	accountIdx := strings.Index(text, `"terraform-provider-kion/internal/service/account"`)
	if aaaIdx >= accountIdx {
		t.Error("import not inserted at first position")
	}

	aaaRegIdx := strings.Index(text, "aaa.NewServicePackage(),")
	accountRegIdx := strings.Index(text, "account.NewServicePackage(),")
	if aaaRegIdx >= accountRegIdx {
		t.Error("registration not inserted at first position")
	}
}

func TestRegisterServicePackage_Last(t *testing.T) {
	root := setupTempFile(t)
	if err := RegisterServicePackage(root, "zzz"); err != nil {
		t.Fatalf("RegisterServicePackage() error = %v", err)
	}

	text := readResult(t, root)

	userGroupIdx := strings.Index(text, `"terraform-provider-kion/internal/service/user_group"`)
	zzzIdx := strings.Index(text, `"terraform-provider-kion/internal/service/zzz"`)
	if userGroupIdx >= zzzIdx {
		t.Error("import not inserted at last position")
	}

	userGroupRegIdx := strings.Index(text, "user_group.NewServicePackage(),")
	zzzRegIdx := strings.Index(text, "zzz.NewServicePackage(),")
	if userGroupRegIdx >= zzzRegIdx {
		t.Error("registration not inserted at last position")
	}
}

func TestRegisterServicePackage_Idempotent(t *testing.T) {
	root := setupTempFile(t)

	if err := RegisterServicePackage(root, "cloud_rule"); err != nil {
		t.Fatalf("first RegisterServicePackage() error = %v", err)
	}

	content1 := readResult(t, root)

	if err := RegisterServicePackage(root, "cloud_rule"); err != nil {
		t.Fatalf("second RegisterServicePackage() error = %v", err)
	}

	content2 := readResult(t, root)

	if content1 != content2 {
		t.Error("calling RegisterServicePackage twice produced different results")
	}
}
