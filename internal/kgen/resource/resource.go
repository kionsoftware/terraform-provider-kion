// Package resource provides scaffolding generation for Terraform resources.
package resource

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

//go:embed resource.gtpl
var resourceTmpl string

//go:embed websitedoc.gtpl
var websiteTmpl string

// TemplateData holds template variables for code generation.
type TemplateData struct {
	Resource             string
	ResourceLower        string
	ResourceLowerCamel   string
	ResourceSnake        string
	IncludeComments      bool
	IncludeTags          bool
	ServicePackage       string
	HumanResourceName    string
	ProviderResourceName string
}

// Create generates scaffold files for a resource.
func Create(resName, snakeName string, comments, force, tags bool) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error reading working directory: %w", err)
	}

	servicePackage := filepath.Base(wd)

	if resName == "" {
		return fmt.Errorf("error checking: no name given")
	}

	if resName == strings.ToLower(resName) {
		return fmt.Errorf("error checking: name should be properly capitalized (e.g., CloudRule)")
	}

	if snakeName != "" && snakeName != strings.ToLower(snakeName) {
		return fmt.Errorf("error checking: snake name should be all lower case with underscores, if needed (e.g., cloud_rule)")
	}

	if snakeName == "" {
		snakeName = convert.ToSnakeCase(resName)
	}

	templateData := TemplateData{
		Resource:             resName,
		ResourceLower:        strings.ToLower(resName),
		ResourceLowerCamel:   convert.ToLowercasePrefix(resName),
		ResourceSnake:        snakeName,
		IncludeComments:      comments,
		IncludeTags:          tags,
		ServicePackage:       servicePackage,
		HumanResourceName:    convert.ToHumanResName(resName),
		ProviderResourceName: convert.ToProviderResourceName(servicePackage, snakeName),
	}

	f := fmt.Sprintf("%s.go", snakeName)
	if err = writeTemplate("newres", f, resourceTmpl, force, templateData); err != nil {
		return fmt.Errorf("writing resource template: %w", err)
	}

	// No acceptance test is written here: `kgen crud` generates it from the
	// resource's codegen/test_values.yaml entry once the real schema exists.

	wf := fmt.Sprintf("%s_%s.html.markdown", servicePackage, snakeName)
	wfDir := filepath.Join("..", "..", "..", "website", "docs", "r")
	wf = filepath.Join(wfDir, wf)
	if _, statErr := fsw.Stat(wfDir); statErr == nil {
		if err = writeTemplate("webdoc", wf, websiteTmpl, force, templateData); err != nil {
			return fmt.Errorf("writing resource website doc template: %w", err)
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
