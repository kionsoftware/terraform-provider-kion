package kalign

import (
	"fmt"
	"io"
	"path/filepath"
)

// Options configures a Run.
type Options struct {
	SDKDir      string // path to the kion-sdk-go checkout
	Version     string // SDK sub-package version, e.g. "v3_16"
	ServiceRoot string // provider service root, e.g. "internal/service"
	FlexDir     string // flex package directory, e.g. "internal/flex"
	OnlyService string // limit to one service; "" means all
}

// SDKFile returns the path to the SDK schemas file the options point at.
func (o Options) SDKFile() string {
	return filepath.Join(o.SDKDir, "generated", o.Version, "oas_schemas_gen.go")
}

// Check resolves every service model via src and writes a drift report to w. It
// returns every actionable finding across all models.
func Check(src Source, w io.Writer, o Options) ([]Finding, error) {
	resolved, err := resolveAll(src, o)
	if err != nil {
		return nil, err
	}
	ew := &errWriter{w: w}
	var findings []Finding
	for _, r := range resolved {
		checkOne(ew, r)
		findings = append(findings, r.ActionableFindings()...)
	}
	ew.Fprintf("\n%d model(s) checked, %d finding(s)\n", len(resolved), len(findings))
	return findings, ew.err
}

// Gen resolves every service model via src and writes flatten converters to w.
// It returns the number of fields that could not be emitted (nested / missing
// flex), which are written as TODO comments.
func Gen(src Source, w io.Writer, o Options) (int, error) {
	resolved, err := resolveAll(src, o)
	if err != nil {
		return 0, err
	}
	ew := &errWriter{w: w}
	todos := 0
	for _, r := range resolved {
		todos += genOne(ew, r)
	}
	return todos, ew.err
}

// resolveAll loads inputs from src and resolves every service model.
func resolveAll(src Source, o Options) ([]Resolved, error) {
	sdkTypes, err := src.SDKStructs(o.SDKFile())
	if err != nil {
		return nil, fmt.Errorf("reading SDK types: %w", err)
	}
	flexFuncs, err := src.FlexFuncs(o.FlexDir)
	if err != nil {
		return nil, fmt.Errorf("reading flex package: %w", err)
	}
	models, err := src.ServiceModels(o.ServiceRoot, o.OnlyService)
	if err != nil {
		return nil, fmt.Errorf("reading service models: %w", err)
	}
	out := make([]Resolved, 0, len(models))
	for _, m := range models {
		out = append(out, Resolve(m, sdkTypes, flexFuncs))
	}
	return out, nil
}
