// Command kconfig derives (and drift-checks) the tfplugingen generator config
// from the service packages. Like kalign it is a standalone binary that
// AST-parses the code without importing it, so it runs even while the provider
// is mid-development.
//
// Usage:
//
//	kconfig gen   [--write] [--spec F] [--config F]   # print (or write) the derived config
//	kconfig check [--spec F] [--config F]             # report drift (non-zero exit)
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"terraform-provider-kion/internal/kgen/config"
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

// run is the testable entry point: it takes the CLI arguments (excluding the
// program name) and its output streams, and returns the process exit code
// instead of calling os.Exit directly.
func run(args []string, stdout, stderr io.Writer) int {
	d := &diag{w: stderr}
	if len(args) < 1 || (args[0] != "gen" && args[0] != "check") {
		d.println("usage: kconfig <gen|check> [--write] [--spec F] [--config F]")
		return 2
	}
	mode := args[0]
	fs := flag.NewFlagSet(mode, flag.ContinueOnError)
	fs.SetOutput(stderr)
	spec := fs.String("spec", "spec/openapi3.json", "OpenAPI 3 spec")
	cfg := fs.String("config", "codegen/generator_config.yaml", "generator config path")
	write := fs.Bool("write", false, "gen: write to the config file instead of stdout")
	if err := fs.Parse(args[1:]); err != nil {
		d.printf("error: %v\n", err)
		return 2
	}

	opts := config.Options{Spec: *spec, ConfigPath: *cfg}
	src := config.NewFileSource()

	switch mode {
	case "gen":
		if !*write {
			if err := config.Gen(src, opts, stdout); err != nil {
				d.printf("error: %v\n", err)
				return 1
			}
			return 0
		}
		var buf bytes.Buffer
		if err := config.Gen(src, opts, &buf); err != nil {
			d.printf("error: %v\n", err)
			return 1
		}
		if err := os.WriteFile(*cfg, buf.Bytes(), 0o600); err != nil {
			d.printf("error: %v\n", err)
			return 1
		}
		d.printf("wrote %s\n", *cfg)
		return 0
	case "check":
		n, err := config.Check(src, opts, stdout)
		if err != nil {
			d.printf("error: %v\n", err)
			return 1
		}
		if n > 0 {
			return 1 // non-zero so CI fails on drift
		}
		return 0
	}
	return 0
}
