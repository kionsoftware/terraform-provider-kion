// Package datasource provides scaffolding generation for Terraform data sources.
package datasource

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"terraform-provider-kion/internal/kgen/convert"
)

//go:embed datasource.gtpl
var datasourceTmpl string

//go:embed websitedoc.gtpl
var websiteTmpl string

// TemplateData holds template variables for code generation.
type TemplateData struct {
	DataSource           string
	DataSourceLower      string
	DataSourceLowerCamel string
	DataSourceSnake      string
	IncludeComments      bool
	IncludeTags          bool
	ServicePackage       string
	HumanDataSourceName  string
	ProviderResourceName string
}

// Create generates scaffold files for a data source.
func Create(dsName, snakeName string, comments, force, tags bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error reading working directory: %w", err)
	}

	servicePackage := filepath.Base(wd)

	if dsName == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if dsName == strings.ToLower(dsName) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., CloudRule)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., cloud_rule)")
	}

	if snakeName == "" {
		snakeName = convert.ToSnakeCase(dsName)
	}

	templateData := TemplateData{
		DataSource:           dsName,
		DataSourceLower:      strings.ToLower(dsName),
		DataSourceLowerCamel: convert.ToLowercasePrefix(dsName),
		DataSourceSnake:      snakeName,
		IncludeComments:      comments,
		IncludeTags:          tags,
		ServicePackage:       servicePackage,
		HumanDataSourceName:  convert.ToHumanResName(dsName),
		ProviderResourceName: convert.ToProviderResourceName(servicePackage, snakeName),
	}

	f := fmt.Sprintf("%s_data_source.go", snakeName)
	if err = writeTemplate("newds", f, datasourceTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing datasource template: %w", err)
	}

	// No acceptance test is written here, see the note in resource.Create.

	wf := fmt.Sprintf("%s_%s.html.markdown", servicePackage, snakeName)
	wfDir := filepath.Join("..", "..", "..", "website", "docs", "d")
	wf = filepath.Join(wfDir, wf)
	if _, statErr := fsw.Stat(wfDir); statErr == nil {
		if err = writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
			return fmt.Errorf("writing datasource website doc template: %w", err)
		}
	}

	return nil
}

func writeTemplate(templateName, filename, tmpl string, force bool, td TemplateData) error {
	if _, err := fsw.Stat(filename); !errors.Is(err, fs.ErrNotExist) && !force {
		return fmt.Errorf("file (%s) already exists and force is not set", filename)
	}

	tplate, err := template.New(templateName).Parse(tmpl)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var buffer bytes.Buffer
	if err := tplate.Execute(&buffer, td); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	if err := fsw.WriteFile(filename, buffer.Bytes(), 0600); err != nil {
		return fmt.Errorf("error writing to file (%s): %w", filename, err)
	}

	return nil
}
