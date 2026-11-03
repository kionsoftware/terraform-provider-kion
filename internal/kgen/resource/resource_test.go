package resource

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"terraform-provider-kion/internal/kgen/kfs/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// swapFS installs a mock FS into the package-level fsw seam for the duration of
// the test and restores the original when the test completes.
func swapFS(t *testing.T) *mocks.MockFS {
	t.Helper()
	m := mocks.NewMockFS(t)
	orig := fsw
	fsw = m
	t.Cleanup(func() { fsw = orig })
	return m
}

func sampleData() TemplateData {
	return TemplateData{
		Resource:             "CloudRule",
		ResourceLower:        "cloudrule",
		ResourceLowerCamel:   "cloudRule",
		ResourceSnake:        "cloud_rule",
		IncludeComments:      true,
		IncludeTags:          true,
		ServicePackage:       "kion",
		HumanResourceName:    "Cloud Rule",
		ProviderResourceName: "kion_cloud_rule",
	}
}

func TestWriteTemplate_Success(t *testing.T) {
	m := swapFS(t)
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).Return(nil)

	err := writeTemplate("newres", "out.go", "package {{.ServicePackage}}", false, sampleData())
	require.NoError(t, err)
}

func TestWriteTemplate_WritesRenderedContent(t *testing.T) {
	m := swapFS(t)
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)
	var got []byte
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).
		Run(func(_ string, data []byte, _ os.FileMode) { got = data }).
		Return(nil)

	err := writeTemplate("newres", "out.go", "package {{.ServicePackage}}", false, sampleData())
	require.NoError(t, err)
	require.Equal(t, "package kion", string(got))
}

func TestWriteTemplate_SkipIfExists(t *testing.T) {
	m := swapFS(t)
	// Stat returns (nil, nil) => file exists; force is false => skip with error
	// and NO WriteFile call (asserted by the mock's cleanup, since none is set up).
	m.EXPECT().Stat("out.go").Return(nil, nil)

	err := writeTemplate("newres", "out.go", "package p", false, sampleData())
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
}

func TestWriteTemplate_ForceOverwrites(t *testing.T) {
	m := swapFS(t)
	// File exists, but force=true => proceed to write anyway.
	m.EXPECT().Stat("out.go").Return(nil, nil)
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).Return(nil)

	err := writeTemplate("newres", "out.go", "package p", true, sampleData())
	require.NoError(t, err)
}

func TestWriteTemplate_ParseError(t *testing.T) {
	m := swapFS(t)
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)
	// Malformed template => parse error before any WriteFile.

	err := writeTemplate("newres", "out.go", "{{.Unclosed", false, sampleData())
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing template")
}

func TestWriteTemplate_ExecuteError(t *testing.T) {
	m := swapFS(t)
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)
	// References a field that does not exist on TemplateData => execute error.

	err := writeTemplate("newres", "out.go", "{{.DoesNotExist}}", false, sampleData())
	require.Error(t, err)
	require.Contains(t, err.Error(), "executing template")
}

func TestWriteTemplate_WriteFileError(t *testing.T) {
	m := swapFS(t)
	sentinel := errors.New("disk full")
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).Return(sentinel)

	err := writeTemplate("newres", "out.go", "package p", false, sampleData())
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
}

// --- Create ---

// docDir is the website docs directory Create checks/writes relative to cwd.
var docDir = filepath.Join("..", "..", "..", "website", "docs", "r")

func TestCreate_Success_NoWebsiteDir(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// Only the resource .go file is written; it does not exist yet.
	m.EXPECT().Stat("cloud_rule.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("cloud_rule.go", mock.Anything, mock.Anything).Return(nil)

	// Website docs dir does NOT exist => the doc file is skipped entirely.
	m.EXPECT().Stat(docDir).Return(nil, os.ErrNotExist)

	err := Create("CloudRule", "", true, false, true)
	require.NoError(t, err)
}

func TestCreate_Success_WithWebsiteDir(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)
	m := swapFS(t)

	servicePackage := filepath.Base(wd)
	wdocFile := filepath.Join(docDir, servicePackage+"_cloud_rule.html.markdown")

	m.EXPECT().Stat("cloud_rule.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("cloud_rule.go", mock.Anything, mock.Anything).Return(nil)

	// Website docs dir EXISTS => the doc file is written.
	m.EXPECT().Stat(docDir).Return(nil, nil)
	m.EXPECT().Stat(wdocFile).Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile(wdocFile, mock.Anything, mock.Anything).Return(nil)

	err := Create("CloudRule", "", true, false, true)
	require.NoError(t, err)
}

func TestCreate_DerivesSnakeName(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// snakeName is empty => derived from resName via convert.ToSnakeCase.
	m.EXPECT().Stat("my_widget.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("my_widget.go", mock.Anything, mock.Anything).Return(nil)
	m.EXPECT().Stat(docDir).Return(nil, os.ErrNotExist)

	err := Create("MyWidget", "", false, false, false)
	require.NoError(t, err)
}

func TestCreate_EmptyName(t *testing.T) {
	t.Chdir(t.TempDir())
	swapFS(t)
	err := Create("", "", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no name given")
}

func TestCreate_LowercaseName(t *testing.T) {
	t.Chdir(t.TempDir())
	swapFS(t)
	err := Create("cloudrule", "", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "properly capitalized")
}

func TestCreate_BadSnakeName(t *testing.T) {
	t.Chdir(t.TempDir())
	swapFS(t)
	err := Create("CloudRule", "Cloud_Rule", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "snake name should be all lower case")
}

func TestCreate_ResourceFileExists_Errors(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// First template's file already exists, force=false => Create errors out
	// before writing anything.
	m.EXPECT().Stat("cloud_rule.go").Return(nil, nil)

	err := Create("CloudRule", "", true, false, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "writing resource template")
}

func TestCreate_WebsiteDocWriteError(t *testing.T) {
	wd := t.TempDir()
	t.Chdir(wd)
	m := swapFS(t)

	servicePackage := filepath.Base(wd)
	wdocFile := filepath.Join(docDir, servicePackage+"_cloud_rule.html.markdown")
	sentinel := errors.New("boom")

	m.EXPECT().Stat("cloud_rule.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("cloud_rule.go", mock.Anything, mock.Anything).Return(nil)
	m.EXPECT().Stat(docDir).Return(nil, nil)
	m.EXPECT().Stat(wdocFile).Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile(wdocFile, mock.Anything, mock.Anything).Return(sentinel)

	err := Create("CloudRule", "", true, false, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "website doc")
}
