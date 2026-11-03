// Package servicepkg provides scaffolding generation for a complete service package.
package servicepkg

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"terraform-provider-kion/internal/kgen/convert"
	"terraform-provider-kion/internal/kgen/datasource"
	"terraform-provider-kion/internal/kgen/resource"
)

//go:embed servicepkg.gtpl
var servicePkgTmpl string

// TemplateData holds template variables for service package generation.
type TemplateData struct {
	ServicePackage string
	Resource       string
}

// Create generates a complete service package with resource, data source, and registration.
func Create(resName, snakeName string, comments, force, tags bool) error {
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

	projectRoot, err := findProjectRoot()
	if err != nil {
		return err
	}

	servicePackage := snakeName

	// Create service directory
	serviceDir := filepath.Join(projectRoot, "internal", "service", servicePackage)
	if err := fsw.MkdirAll(serviceDir, 0750); err != nil {
		return fmt.Errorf("creating service directory: %w", err)
	}

	// Generate service_package.go
	if err := writeServicePackage(serviceDir, servicePkgTmpl, force, TemplateData{
		ServicePackage: servicePackage,
		Resource:       resName,
	}); err != nil {
		return err
	}

	// Change to service directory to run resource/datasource generators
	origDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting current directory: %w", err)
	}

	if err := os.Chdir(serviceDir); err != nil {
		return fmt.Errorf("changing to service directory: %w", err)
	}
	defer func() {
		// Best-effort restore of working directory; errors are intentionally
		// ignored because there is no meaningful recovery action.
		_ = os.Chdir(origDir) //nolint:errcheck
	}()

	// Generate resource files
	if err := resource.Create(resName, snakeName, comments, force, tags); err != nil {
		return fmt.Errorf("generating resource: %w", err)
	}

	// Generate data source files
	if err := datasource.Create(resName, snakeName, comments, force, tags); err != nil {
		return fmt.Errorf("generating data source: %w", err)
	}

	// Register in service_packages.go
	if err := RegisterServicePackage(projectRoot, servicePackage); err != nil {
		return fmt.Errorf("registering service package: %w", err)
	}

	fmt.Printf("Service package %q created successfully in %s\n", servicePackage, serviceDir)
	return nil
}

func writeServicePackage(dir, tmpl string, force bool, td TemplateData) error {
	filename := filepath.Join(dir, "service_package.go")

	if _, err := fsw.Stat(filename); err == nil && !force {
		return fmt.Errorf("file (%s) already exists and force is not set", filename)
	}

	tplate, err := template.New("servicepkg").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("parsing service package template: %w", err)
	}

	var buf bytes.Buffer
	if err := tplate.Execute(&buf, td); err != nil {
		return fmt.Errorf("executing service package template: %w", err)
	}

	if err := fsw.WriteFile(filename, buf.Bytes(), 0600); err != nil {
		return fmt.Errorf("writing service package file: %w", err)
	}

	return nil
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("getting current directory: %w", err)
	}

	for {
		if _, err := fsw.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}

// RegisterServicePackage adds an import and registration entry to service_packages.go.
func RegisterServicePackage(projectRoot, pkgName string) error {
	filePath := filepath.Join(projectRoot, "internal", "provider", "service_packages.go")

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading service_packages.go: %w", err)
	}

	text := string(content)
	importPath := fmt.Sprintf(`"terraform-provider-kion/internal/service/%s"`, pkgName)
	// Inserted verbatim, so it carries its own indentation: without the tab the
	// import lands hard against the margin and the file fails gofmt, which is a
	// scaffold leaving the repository in a state its own CI rejects.
	importLine := "\t" + importPath
	registrationEntry := fmt.Sprintf("\t\t%s.NewServicePackage(),", pkgName)

	// Check if already registered
	if strings.Contains(text, importPath) {
		return nil
	}

	// Insert import in sorted position
	text, err = insertSorted(text, importLine, `"terraform-provider-kion/internal/service/`)
	if err != nil {
		return fmt.Errorf("inserting import: %w", err)
	}

	// Insert registration entry in sorted position
	text, err = insertSorted(text, registrationEntry, ".NewServicePackage(),")
	if err != nil {
		return fmt.Errorf("inserting registration: %w", err)
	}

	if err := fsw.WriteFile(filePath, []byte(text), 0600); err != nil {
		return fmt.Errorf("writing service_packages.go: %w", err)
	}

	return nil
}

// insertSorted inserts newLine into the file content in sorted order among lines
// that contain matchPattern, before the closing delimiter.
func insertSorted(text, newLine, matchPattern string) (string, error) {
	lines := strings.Split(text, "\n")

	// Find all lines matching the pattern and their indices
	var matchingLines []string
	var matchingIndices []int
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.Contains(trimmed, matchPattern) {
			matchingLines = append(matchingLines, trimmed)
			matchingIndices = append(matchingIndices, i)
		}
	}

	if len(matchingLines) == 0 {
		return "", fmt.Errorf("no lines matching %q found", matchPattern)
	}

	// Add the new line's trimmed form to find where it sorts
	newTrimmed := strings.TrimSpace(newLine)
	allEntries := append(matchingLines, newTrimmed)
	sort.Strings(allEntries)

	// Find where the new entry was sorted to
	insertIdx := 0
	for i, entry := range allEntries {
		if entry == newTrimmed {
			insertIdx = i
			break
		}
	}

	// Determine the line number to insert at
	var insertAtLine int
	if insertIdx >= len(matchingIndices) {
		// Insert after the last matching line
		insertAtLine = matchingIndices[len(matchingIndices)-1] + 1
	} else {
		// Insert before the matching line at this sort position
		insertAtLine = matchingIndices[insertIdx]
	}

	// Insert the new line
	result := make([]string, 0, len(lines)+1)
	result = append(result, lines[:insertAtLine]...)
	result = append(result, newLine)
	result = append(result, lines[insertAtLine:]...)

	return strings.Join(result, "\n"), nil
}
