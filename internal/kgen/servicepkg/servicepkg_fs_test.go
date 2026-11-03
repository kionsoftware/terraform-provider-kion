package servicepkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"terraform-provider-kion/internal/kgen/kfs/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// swapFS installs a mock FS in place of the package-level fsw seam for the
// duration of the test and restores the original on cleanup.
func swapFS(t *testing.T) *mocks.MockFS {
	t.Helper()
	m := mocks.NewMockFS(t)
	orig := fsw
	fsw = m
	t.Cleanup(func() { fsw = orig })
	return m
}

// writeGoMod creates a minimal go.mod in dir so findProjectRoot (which stats
// through the real OS when fsw is OS, or through the mock otherwise) can locate
// a project root.
func writeGoMod(t *testing.T, dir string) {
	t.Helper()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/tmp\n\ngo 1.25\n"), 0600))
}

// ---------------------------------------------------------------------------
// findProjectRoot
// ---------------------------------------------------------------------------

func TestFindProjectRoot_FoundInCwd(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := swapFS(t)
	// go.mod exists directly in the cwd.
	m.EXPECT().Stat(filepath.Join(dir, "go.mod")).Return(nil, nil).Once()

	root, err := findProjectRoot()
	require.NoError(t, err)
	require.Equal(t, dir, root)
}

func TestFindProjectRoot_FoundInParent(t *testing.T) {
	parent := t.TempDir()
	child := filepath.Join(parent, "internal", "service", "thing")
	require.NoError(t, os.MkdirAll(child, 0750))
	t.Chdir(child)

	// Resolve paths through the real cwd so they match what findProjectRoot
	// derives from os.Getwd() (macOS temp dirs are symlinked).
	cwd, err := os.Getwd()
	require.NoError(t, err)
	realParent := filepath.Dir(filepath.Dir(filepath.Dir(cwd)))
	parentGoMod := filepath.Join(realParent, "go.mod")

	m := swapFS(t)
	// Not found walking up until we reach the parent that holds go.mod.
	m.EXPECT().Stat(mock.MatchedBy(func(name string) bool { return name != parentGoMod })).
		Return(nil, os.ErrNotExist).Maybe()
	m.EXPECT().Stat(parentGoMod).Return(nil, nil).Once()

	root, err := findProjectRoot()
	require.NoError(t, err)
	require.Equal(t, realParent, root)
}

func TestFindProjectRoot_NotFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	m := swapFS(t)
	// go.mod never exists anywhere up to the filesystem root.
	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist)

	_, err := findProjectRoot()
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not find project root")
}

// ---------------------------------------------------------------------------
// writeServicePackage
// ---------------------------------------------------------------------------

func TestWriteServicePackage_Success(t *testing.T) {
	dir := t.TempDir()
	m := swapFS(t)

	target := filepath.Join(dir, "service_package.go")
	m.EXPECT().Stat(target).Return(nil, os.ErrNotExist).Once()

	var written []byte
	m.EXPECT().WriteFile(target, mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ os.FileMode) { written = data }).
		Return(nil).Once()

	err := writeServicePackage(dir, servicePkgTmpl, false, TemplateData{
		ServicePackage: "cloud_rule",
		Resource:       "CloudRule",
	})
	require.NoError(t, err)

	rendered := string(written)
	require.Contains(t, rendered, "package cloud_rule")
	require.Contains(t, rendered, "NewCloudRuleResource")
	require.Contains(t, rendered, "NewCloudRuleDataSource")
	require.Contains(t, rendered, `return "cloud_rule"`)
}

func TestWriteServicePackage_ExistsNoForce(t *testing.T) {
	dir := t.TempDir()
	m := swapFS(t)

	target := filepath.Join(dir, "service_package.go")
	// File already exists (Stat returns nil error) and force is false.
	m.EXPECT().Stat(target).Return(nil, nil).Once()

	err := writeServicePackage(dir, servicePkgTmpl, false, TemplateData{
		ServicePackage: "cloud_rule",
		Resource:       "CloudRule",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestWriteServicePackage_ExistsWithForce(t *testing.T) {
	dir := t.TempDir()
	m := swapFS(t)

	target := filepath.Join(dir, "service_package.go")
	// File exists but force is true -> it should overwrite (WriteFile called).
	m.EXPECT().Stat(target).Return(nil, nil).Once()
	m.EXPECT().WriteFile(target, mock.Anything, mock.Anything).Return(nil).Once()

	err := writeServicePackage(dir, servicePkgTmpl, true, TemplateData{
		ServicePackage: "cloud_rule",
		Resource:       "CloudRule",
	})
	require.NoError(t, err)
}

func TestWriteServicePackage_TemplateParseError(t *testing.T) {
	dir := t.TempDir()
	m := swapFS(t)

	target := filepath.Join(dir, "service_package.go")
	m.EXPECT().Stat(target).Return(nil, os.ErrNotExist).Once()

	err := writeServicePackage(dir, "{{ .Unterminated", false, TemplateData{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing service package template")
}

func TestWriteServicePackage_TemplateExecError(t *testing.T) {
	dir := t.TempDir()
	m := swapFS(t)

	target := filepath.Join(dir, "service_package.go")
	m.EXPECT().Stat(target).Return(nil, os.ErrNotExist).Once()

	// Reference a method that does not exist on the data -> execution error.
	err := writeServicePackage(dir, "{{ .Missing.Field }}", false, TemplateData{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "executing service package template")
}

func TestWriteServicePackage_WriteError(t *testing.T) {
	dir := t.TempDir()
	m := swapFS(t)

	target := filepath.Join(dir, "service_package.go")
	m.EXPECT().Stat(target).Return(nil, os.ErrNotExist).Once()
	m.EXPECT().WriteFile(target, mock.Anything, mock.Anything).Return(fmt.Errorf("disk full")).Once()

	err := writeServicePackage(dir, servicePkgTmpl, false, TemplateData{
		ServicePackage: "cloud_rule",
		Resource:       "CloudRule",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "writing service package file")
}

// ---------------------------------------------------------------------------
// Create - validation error paths (no FS interaction expected)
// ---------------------------------------------------------------------------

func TestCreate_EmptyName(t *testing.T) {
	swapFS(t) // no expectations: validation fails before any FS call.
	err := Create("", "", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no name given")
}

func TestCreate_LowercaseName(t *testing.T) {
	swapFS(t)
	err := Create("cloudrule", "", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "properly capitalized")
}

func TestCreate_BadSnakeName(t *testing.T) {
	swapFS(t)
	err := Create("CloudRule", "CloudRule", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "snake name should be all lower case")
}

// ---------------------------------------------------------------------------
// Create - MkdirAll / writeServicePackage error paths
// ---------------------------------------------------------------------------

func TestCreate_MkdirAllError(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)
	t.Chdir(dir)

	m := swapFS(t)
	// findProjectRoot stats go.mod (found in cwd).
	m.EXPECT().Stat(filepath.Join(dir, "go.mod")).Return(nil, nil).Once()

	serviceDir := filepath.Join(dir, "internal", "service", "cloud_rule")
	m.EXPECT().MkdirAll(serviceDir, mock.Anything).Return(fmt.Errorf("permission denied")).Once()

	err := Create("CloudRule", "", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "creating service directory")
}

func TestCreate_WriteServicePackageError(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)
	t.Chdir(dir)

	m := swapFS(t)
	m.EXPECT().Stat(filepath.Join(dir, "go.mod")).Return(nil, nil).Once()

	serviceDir := filepath.Join(dir, "internal", "service", "cloud_rule")
	m.EXPECT().MkdirAll(serviceDir, mock.Anything).Return(nil).Once()

	// service_package.go does not exist yet, but WriteFile fails.
	pkgFile := filepath.Join(serviceDir, "service_package.go")
	m.EXPECT().Stat(pkgFile).Return(nil, os.ErrNotExist).Once()
	m.EXPECT().WriteFile(pkgFile, mock.Anything, mock.Anything).Return(fmt.Errorf("disk full")).Once()

	err := Create("CloudRule", "", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "writing service package file")
}

// TestCreate_SnakeDerivedFromResName verifies the snake name is derived from
// resName when not supplied, by observing the service directory path passed to
// MkdirAll.
func TestCreate_SnakeDerivedFromResName(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)
	t.Chdir(dir)

	m := swapFS(t)
	m.EXPECT().Stat(filepath.Join(dir, "go.mod")).Return(nil, nil).Once()

	// ToSnakeCase("CloudRule") == "cloud_rule".
	serviceDir := filepath.Join(dir, "internal", "service", "cloud_rule")
	// Fail at MkdirAll so we stop early; we only care about the derived path.
	m.EXPECT().MkdirAll(serviceDir, mock.Anything).Return(fmt.Errorf("stop")).Once()

	err := Create("CloudRule", "", false, false, false)
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Create - full happy path (resource/datasource write to real disk under cwd)
// ---------------------------------------------------------------------------

// TestCreate_FullSuccess exercises the whole Create flow. The servicepkg fsw
// seam is mocked; MkdirAll actually creates the real directory (so the
// subsequent os.Chdir succeeds), and the resource/datasource generators, which
// use their own OS-backed fsw. Write real files into the temp service dir.
func TestCreate_FullSuccess(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir)

	// service_packages.go must exist for RegisterServicePackage's os.ReadFile.
	providerDir := filepath.Join(dir, "internal", "provider")
	require.NoError(t, os.MkdirAll(providerDir, 0750))
	require.NoError(t, os.WriteFile(filepath.Join(providerDir, "service_packages.go"),
		[]byte(servicePackagesContent), 0600))

	t.Chdir(dir)

	m := swapFS(t)
	m.EXPECT().Stat(filepath.Join(dir, "go.mod")).Return(nil, nil).Once()

	serviceDir := filepath.Join(dir, "internal", "service", "cloud_rule")
	// Actually create the dir so os.Chdir(serviceDir) works.
	m.EXPECT().MkdirAll(serviceDir, mock.Anything).
		Run(func(path string, perm os.FileMode) {
			if err := os.MkdirAll(path, perm); err != nil {
				t.Errorf("creating dir %s: %v", path, err)
			}
		}).
		Return(nil).Once()

	// writeServicePackage: file absent -> write succeeds.
	pkgFile := filepath.Join(serviceDir, "service_package.go")
	m.EXPECT().Stat(pkgFile).Return(nil, os.ErrNotExist).Once()
	m.EXPECT().WriteFile(pkgFile, mock.Anything, mock.Anything).Return(nil).Once()

	// RegisterServicePackage writes service_packages.go back through the mock.
	spFile := filepath.Join(providerDir, "service_packages.go")
	var registered []byte
	m.EXPECT().WriteFile(spFile, mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ os.FileMode) { registered = data }).
		Return(nil).Once()

	err := Create("CloudRule", "", false, false, false)
	require.NoError(t, err)

	// The registration edit should mention the new package.
	require.Contains(t, string(registered), "cloud_rule.NewServicePackage(),")
	require.Contains(t, string(registered), `"terraform-provider-kion/internal/service/cloud_rule"`)
}

// ---------------------------------------------------------------------------
// RegisterServicePackage - error branches
// ---------------------------------------------------------------------------

func TestRegisterServicePackage_ReadError(t *testing.T) {
	swapFS(t)
	// No file at this root -> os.ReadFile fails.
	err := RegisterServicePackage(t.TempDir(), "cloud_rule")
	require.Error(t, err)
	require.Contains(t, err.Error(), "reading service_packages.go")
}

func TestRegisterServicePackage_WriteError(t *testing.T) {
	root := setupTempFile(t)
	m := swapFS(t)

	spFile := filepath.Join(root, "internal", "provider", "service_packages.go")
	m.EXPECT().WriteFile(spFile, mock.Anything, mock.Anything).Return(fmt.Errorf("read-only fs")).Once()

	err := RegisterServicePackage(root, "cloud_rule")
	require.Error(t, err)
	require.Contains(t, err.Error(), "writing service_packages.go")
}

// TestRegisterServicePackage_MissingImportBlock covers the insertSorted error
// path when the file has no matching import lines to sort against.
func TestRegisterServicePackage_MissingImportBlock(t *testing.T) {
	dir := t.TempDir()
	providerDir := filepath.Join(dir, "internal", "provider")
	require.NoError(t, os.MkdirAll(providerDir, 0750))
	// Content lacks any "terraform-provider-kion/internal/service/" line.
	require.NoError(t, os.WriteFile(filepath.Join(providerDir, "service_packages.go"),
		[]byte("package provider\n\nfunc servicePackages() {}\n"), 0600))

	swapFS(t) // WriteFile should never be reached.

	err := RegisterServicePackage(dir, "cloud_rule")
	require.Error(t, err)
	require.Contains(t, err.Error(), "inserting import")
}

// TestRegisterServicePackage_MissingRegistrationBlock covers the insertSorted
// error path for the registration entry when imports exist but no
// ".NewServicePackage()," lines do.
func TestRegisterServicePackage_MissingRegistrationBlock(t *testing.T) {
	dir := t.TempDir()
	providerDir := filepath.Join(dir, "internal", "provider")
	require.NoError(t, os.MkdirAll(providerDir, 0750))
	content := `package provider

import (
	"terraform-provider-kion/internal/service/account"
)

func servicePackages() {}
`
	require.NoError(t, os.WriteFile(filepath.Join(providerDir, "service_packages.go"),
		[]byte(content), 0600))

	swapFS(t) // WriteFile should never be reached.

	err := RegisterServicePackage(dir, "cloud_rule")
	require.Error(t, err)
	require.Contains(t, err.Error(), "inserting registration")
}

// ---------------------------------------------------------------------------
// insertSorted - direct unit coverage
// ---------------------------------------------------------------------------

func TestInsertSorted_NoMatch(t *testing.T) {
	_, err := insertSorted("a\nb\nc\n", "new", "nomatch")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no lines matching")
}

func TestInsertSorted_InsertsInOrder(t *testing.T) {
	text := "x_aaa\nx_ccc\n"
	out, err := insertSorted(text, "x_bbb", "x_")
	require.NoError(t, err)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	require.Equal(t, []string{"x_aaa", "x_bbb", "x_ccc"}, lines)
}
