package datasource

import (
	"errors"
	"os"
	"testing"

	"terraform-provider-kion/internal/kgen/kfs/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// swapFS installs a mock FS into the package-level seam for the duration of the
// test and restores the original on cleanup.
func swapFS(t *testing.T) *mocks.MockFS {
	t.Helper()
	m := mocks.NewMockFS(t)
	orig := fsw
	fsw = m
	t.Cleanup(func() { fsw = orig })
	return m
}

func sampleTemplateData() TemplateData {
	return TemplateData{
		DataSource:           "CloudRule",
		DataSourceLower:      "cloudrule",
		DataSourceLowerCamel: "cloudRule",
		DataSourceSnake:      "cloud_rule",
		IncludeComments:      true,
		IncludeTags:          true,
		ServicePackage:       "svc",
		HumanDataSourceName:  "Cloud Rule",
		ProviderResourceName: "kion_cloud_rule",
	}
}

func TestWriteTemplate_WritesWhenNotExist(t *testing.T) {
	m := swapFS(t)
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)

	var gotName string
	var gotContent []byte
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).
		RunAndReturn(func(name string, data []byte, _ os.FileMode) error {
			gotName = name
			gotContent = data
			return nil
		})

	err := writeTemplate("tmpl", "out.go", "hello {{.DataSource}}", false, sampleTemplateData())
	require.NoError(t, err)
	require.Equal(t, "out.go", gotName)
	require.Equal(t, "hello CloudRule", string(gotContent))
}

func TestWriteTemplate_AlreadyExistsNoForce(t *testing.T) {
	m := swapFS(t)
	// Stat returns nil error => file exists. With force=false this should error
	// and must NOT call WriteFile (no WriteFile expectation is set, so a call
	// would fail the mock).
	m.EXPECT().Stat("out.go").Return(nil, nil)

	err := writeTemplate("tmpl", "out.go", "hello", false, sampleTemplateData())
	require.Error(t, err)
	require.Contains(t, err.Error(), "already exists")
	m.AssertNotCalled(t, "WriteFile", mock.Anything, mock.Anything, mock.Anything)
}

func TestWriteTemplate_ForceOverwrites(t *testing.T) {
	m := swapFS(t)
	// File exists, but force=true means we still write.
	m.EXPECT().Stat("out.go").Return(nil, nil)
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).Return(nil)

	err := writeTemplate("tmpl", "out.go", "hi {{.DataSourceSnake}}", true, sampleTemplateData())
	require.NoError(t, err)
}

func TestWriteTemplate_ParseError(t *testing.T) {
	m := swapFS(t)
	// Stat must report not-exist so we reach the parse step.
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)

	err := writeTemplate("tmpl", "out.go", "{{.Bad", false, sampleTemplateData())
	require.Error(t, err)
	require.Contains(t, err.Error(), "parsing template")
	m.AssertNotCalled(t, "WriteFile", mock.Anything, mock.Anything, mock.Anything)
}

func TestWriteTemplate_ExecuteError(t *testing.T) {
	m := swapFS(t)
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)

	// Reference a field that does not exist on TemplateData => execution error.
	err := writeTemplate("tmpl", "out.go", "{{.DoesNotExist}}", false, sampleTemplateData())
	require.Error(t, err)
	require.Contains(t, err.Error(), "executing template")
	m.AssertNotCalled(t, "WriteFile", mock.Anything, mock.Anything, mock.Anything)
}

func TestWriteTemplate_WriteFileErrorPropagates(t *testing.T) {
	m := swapFS(t)
	sentinel := errors.New("disk full")
	m.EXPECT().Stat("out.go").Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile("out.go", mock.Anything, mock.Anything).Return(sentinel)

	err := writeTemplate("tmpl", "out.go", "hello", false, sampleTemplateData())
	require.Error(t, err)
	require.ErrorIs(t, err, sentinel)
	require.Contains(t, err.Error(), "error writing to file")
}

func TestCreate_ValidationErrors(t *testing.T) {
	t.Run("empty name", func(t *testing.T) {
		t.Chdir(t.TempDir())
		swapFS(t)
		err := Create("", "", false, false, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no name given")
	})

	t.Run("not capitalized", func(t *testing.T) {
		t.Chdir(t.TempDir())
		swapFS(t)
		err := Create("cloudrule", "", false, false, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "properly capitalized")
	})

	t.Run("snake name not lower", func(t *testing.T) {
		t.Chdir(t.TempDir())
		swapFS(t)
		err := Create("CloudRule", "Cloud_Rule", false, false, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "snake name should be all lower case")
	})
}

// TestCreate_NoWebsiteDir covers the branch where the website docs dir does not
// exist: Create writes only the data source file and skips the website doc.
func TestCreate_NoWebsiteDir(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	// Every file-existence check reports not-exist so writes proceed.
	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist)

	var written []string
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(name string, _ []byte, _ os.FileMode) error {
			written = append(written, name)
			return nil
		})

	err := Create("CloudRule", "cloud_rule", true, false, true)
	require.NoError(t, err)

	// Only the data source file. No acceptance test (kgen crud owns those)
	// and no website doc.
	require.ElementsMatch(t, []string{
		"cloud_rule_data_source.go",
	}, written)
}

// TestCreate_WithWebsiteDir covers the branch where the website docs dir exists,
// so the website markdown doc is also written.
func TestCreate_WithWebsiteDir(t *testing.T) {
	tmp := t.TempDir()
	t.Chdir(tmp)
	m := swapFS(t)

	websiteDir := "../../../website/docs/d"

	// The website dir Stat succeeds (nil error) => write the doc. All other
	// Stats report not-exist.
	m.EXPECT().Stat(websiteDir).Return(nil, nil)
	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist).Maybe()

	var written []string
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(name string, _ []byte, _ os.FileMode) error {
			written = append(written, name)
			return nil
		})

	// servicePackage is filepath.Base(wd); the temp dir's base name feeds the
	// website filename, so derive it the same way the code does.
	err := Create("CloudRule", "cloud_rule", true, false, true)
	require.NoError(t, err)

	require.Contains(t, written, "cloud_rule_data_source.go")

	// The website doc path is <websiteDir>/<pkg>_cloud_rule.html.markdown.
	var foundDoc bool
	for _, w := range written {
		if len(w) > len(".html.markdown") &&
			w[len(w)-len(".html.markdown"):] == ".html.markdown" {
			foundDoc = true
		}
	}
	require.True(t, foundDoc, "expected a website .html.markdown doc to be written, got %v", written)
}

// TestCreate_DefaultSnakeName covers the branch where no snake name is passed
// and Create derives one from the data source name via convert.ToSnakeCase.
func TestCreate_DefaultSnakeName(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist)

	var written []string
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(name string, _ []byte, _ os.FileMode) error {
			written = append(written, name)
			return nil
		})

	err := Create("CloudRule", "", false, false, false)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{
		"cloud_rule_data_source.go",
	}, written)
}

// TestCreate_WebsiteDocWriteError covers the website-doc writeTemplate error
// branch: the data source file writes fine, but the website doc fails.
func TestCreate_WebsiteDocWriteError(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	websiteDir := "../../../website/docs/d"
	m.EXPECT().Stat(websiteDir).Return(nil, nil)
	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist).Maybe()

	m.EXPECT().WriteFile("cloud_rule_data_source.go", mock.Anything, mock.Anything).Return(nil)
	// The website doc write fails.
	m.EXPECT().WriteFile(mock.MatchedBy(func(name string) bool {
		return len(name) >= len(".html.markdown") &&
			name[len(name)-len(".html.markdown"):] == ".html.markdown"
	}), mock.Anything, mock.Anything).Return(errors.New("boom"))

	err := Create("CloudRule", "cloud_rule", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "writing datasource website doc template")
}

// TestCreate_WriteErrorPropagates ensures a failing WriteFile surfaces from
// Create wrapped with context.
func TestCreate_WriteErrorPropagates(t *testing.T) {
	t.Chdir(t.TempDir())
	m := swapFS(t)

	m.EXPECT().Stat(mock.Anything).Return(nil, os.ErrNotExist)
	m.EXPECT().WriteFile(mock.Anything, mock.Anything, mock.Anything).
		Return(errors.New("boom"))

	err := Create("CloudRule", "cloud_rule", false, false, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "writing datasource template")
}
