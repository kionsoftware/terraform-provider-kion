// Command kalign aligns Terraform resource schemas against the kion-sdk-go
// generated client types. See package terraform-provider-kion/internal/kalign.
//
// Usage:
//
//	kalign check [-sdk DIR] [-version <current>] [-service NAME] [-baseline FILE] [-update]
//	kalign gen   [-sdk DIR] [-version <current>] [-service NAME]
//
// check fails on any finding not recorded in the baseline and on any baseline
// entry that no longer occurs; -update rewrites the baseline instead. -sdk
// defaults to the SDK module go.mod resolves, so no checkout is needed.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"terraform-provider-kion/internal/kalign"
	"terraform-provider-kion/internal/kgen/versions"
)

const (
	sdkModule       = "github.com/kionsoftware/kion-sdk-go"
	defaultBaseline = "codegen/align_baseline.yaml"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// diag is a best-effort diagnostic writer: it remembers the first write error
// so call sites don't each have to check it (mirrors internal/kalign's
// errWriter). Diagnostics are non-critical, so the remembered error is only
// used to short-circuit further writes, never surfaced.
type diag struct {
	w   io.Writer
	err error
}

func (d *diag) printf(format string, a ...any) {
	if d.err != nil {
		return
	}
	_, d.err = fmt.Fprintf(d.w, format, a...)
}

func (d *diag) println(a ...any) {
	if d.err != nil {
		return
	}
	_, d.err = fmt.Fprintln(d.w, a...)
}

// resolveSDKDir is a seam so tests do not shell out to the go tool.
var resolveSDKDir = func() (string, error) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", sdkModule).Output()
	if err != nil {
		return "", fmt.Errorf("resolving %s: %w", sdkModule, err)
	}
	dir := strings.TrimSpace(string(out))
	if dir == "" {
		return "", errors.New("resolving " + sdkModule + ": not in the module cache (run go mod download)")
	}
	return dir, nil
}

// run is the testable entry point: it takes the CLI arguments (excluding the
// program name) and its output streams, and returns the process exit code
// instead of calling os.Exit directly.
func run(args []string, stdout, stderr io.Writer) int {
	d := &diag{w: stderr}
	if len(args) < 1 || (args[0] != "check" && args[0] != "gen") {
		d.println(fmt.Sprintf("usage: kalign <check|gen> [-sdk DIR] [-version %s] [-service NAME] [-baseline FILE] [-update]", versions.CurrentSDK()))
		return 2
	}
	mode := args[0]
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(stderr)
	sdkDir := fs.String("sdk", "", "path to the kion-sdk-go source (default: the module go.mod resolves)")
	version := fs.String("version", versions.CurrentSDK(), "SDK sub-package version to align against")
	only := fs.String("service", "", "limit to a single service (default: all)")
	baseline := fs.String("baseline", defaultBaseline, "accepted findings (check only)")
	update := fs.Bool("update", false, "rewrite the baseline from the current findings (check only)")
	if err := fs.Parse(args[1:]); err != nil {
		d.printf("error: %v\n", err)
		return 2
	}
	if *only != "" {
		// A one-service run would report every other baseline entry as stale.
		*baseline, *update = "", false
	}

	if *sdkDir == "" {
		dir, err := resolveSDKDir()
		if err != nil {
			d.printf("error: %v\n", err)
			return 1
		}
		*sdkDir = dir
	}

	opts := kalign.Options{
		SDKDir:      *sdkDir,
		Version:     *version,
		ServiceRoot: "internal/service",
		FlexDir:     "internal/flex",
		OnlyService: *only,
	}
	src := kalign.NewFileSource()

	if mode == "gen" {
		if _, err := kalign.Gen(src, stdout, opts); err != nil {
			d.printf("error: %v\n", err)
			return 1
		}
		return 0
	}

	findings, err := kalign.Check(src, stdout, opts)
	if err != nil {
		d.printf("error: %v\n", err)
		return 1
	}
	return gate(findings, *baseline, *update, stderr)
}

// gate applies the baseline ratchet. With no baseline every finding fails.
func gate(findings []kalign.Finding, path string, update bool, stderr io.Writer) int {
	d := &diag{w: stderr}
	if path == "" {
		if len(findings) > 0 {
			return 1
		}
		return 0
	}
	prev, err := kalign.LoadBaseline(path)
	if update {
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			d.printf("error: %v\n", err)
			return 1
		}
		if err := kalign.WriteBaseline(path, findings, prev); err != nil {
			d.printf("error: %v\n", err)
			return 1
		}
		d.printf("wrote %s (%d finding(s)); triage any TODO entry\n", path, len(findings))
		return 0
	}
	if err != nil {
		d.printf("error: reading baseline: %v\n", err)
		return 1
	}
	if !kalign.ReportRatchet(stderr, findings, prev, path) {
		return 1
	}
	return 0
}
