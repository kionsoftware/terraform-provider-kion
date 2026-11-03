// Command acctestconfig validates the HCL that acceptance tests feed to Terraform.
//
// Acceptance tests only run against a live Kion instance, so their config strings
// are never exercised by CI, nothing checks that they even parse, let alone that
// they match the provider's schema. That gap let real defects sit in the tree:
// twelve configs used an `owner_users { id = 1 }` block on resources whose schema
// declares an `owner_user_ids` list attribute, and two set a read-only `id`. Every
// one of those would have failed the moment someone ran the acceptance suite.
//
// This closes the gap without a Kion instance or any API call. It extracts each
// test's config string, substitutes the fmt verbs, and runs `terraform validate`
// against a real build of the provider, which is enough to catch a block that
// should be an attribute, a required argument that is missing, a read-only
// attribute being set, or a type mismatch.
//
// dev_overrides is deliberate here: it bypasses `terraform init` entirely, so
// validating ~85 configs costs one provider build and no network. (Elsewhere in
// this project a filesystem mirror is preferred precisely because it does NOT
// skip init. But here skipping it is the point.)
//
// Usage:
//
//	go run ./scripts/acctestconfig            # validate every config
//	go run ./scripts/acctestconfig -v         # also print each config as validated
//
// Exits non-zero, listing every offending function, if any config fails.
package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Config-producing helpers are named testAcc<Thing>Config<_variant> by convention
// throughout the service packages.
var configFuncRe = regexp.MustCompile(`^testAcc.*Config`)

// Indexed and plain fmt verbs. Values only need to be type-plausible: validate
// checks the schema, not the API, so any string/number of the right shape works.
var (
	verbIndexedRe = regexp.MustCompile(`%\[\d+\](\w)`)
	verbPlainRe   = regexp.MustCompile(`%(\w)`)
)

type finding struct {
	pkg, fn, file string
	detail        string
}

func main() {
	verbose := flag.Bool("v", false, "print each config as it is validated")
	flag.Parse()

	root, err := os.Getwd()
	if err != nil {
		fatal(err)
	}

	tfrc, bin, cleanup, err := buildProvider(root)
	if err != nil {
		fatal(err)
	}
	defer cleanup()
	fmt.Printf("provider built: %s\n", bin)

	configs, err := collect(filepath.Join(root, "internal", "service"))
	if err != nil {
		fatal(err)
	}
	fmt.Printf("found %d acceptance-test config(s)\n\n", len(configs))

	work, err := os.MkdirTemp("", "acctestcfg")
	if err != nil {
		fatal(err)
	}
	defer func() {
		if rerr := os.RemoveAll(work); rerr != nil {
			fmt.Fprintf(os.Stderr, "acctestconfig: cleaning %s: %v\n", work, rerr)
		}
	}()

	var bad []finding
	for _, c := range configs {
		if *verbose {
			fmt.Printf("── %s.%s ──\n%s\n", c.pkg, c.fn, c.hcl)
		}
		if detail := validate(work, tfrc, c); detail != "" {
			bad = append(bad, finding{c.pkg, c.fn, c.file, detail})
		}
	}

	if len(bad) > 0 {
		fmt.Printf("\n✗ %d of %d config(s) failed validation:\n\n", len(bad), len(configs))
		for _, b := range bad {
			fmt.Printf("  %s.%s\n    %s\n    %s\n\n", b.pkg, b.fn, b.file, b.detail)
		}
		os.Exit(1)
	}
	fmt.Printf("✓ all %d config(s) valid against the provider schema\n", len(configs))
}

type config struct {
	pkg, fn, file, hcl string
}

// collect parses every *_test.go under dir and returns the HCL each config
// helper produces, with fmt verbs already substituted.
func collect(dir string) ([]config, error) {
	var out []config
	fset := token.NewFileSet()

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, perr := parser.ParseFile(fset, path, nil, 0)
		if perr != nil {
			return fmt.Errorf("parse %s: %w", path, perr)
		}
		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !configFuncRe.MatchString(fn.Name.Name) {
				continue
			}
			hcl, ok := extractHCL(fn)
			if !ok {
				continue
			}
			out = append(out, config{
				pkg:  filepath.Base(filepath.Dir(path)),
				fn:   fn.Name.Name,
				file: path,
				hcl:  substituteVerbs(hcl),
			})
		}
		return nil
	})
	sort.Slice(out, func(i, j int) bool {
		if out[i].pkg != out[j].pkg {
			return out[i].pkg < out[j].pkg
		}
		return out[i].fn < out[j].fn
	})
	return out, err
}

// extractHCL pulls the returned string out of `return <lit>` or
// `return fmt.Sprintf(<lit>, ...)`. Anything else (a composed config, a helper
// call) is skipped rather than guessed at.
func extractHCL(fn *ast.FuncDecl) (string, bool) {
	for _, stmt := range fn.Body.List {
		ret, ok := stmt.(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			continue
		}
		switch e := ret.Results[0].(type) {
		case *ast.BasicLit:
			return unquote(e), true
		case *ast.CallExpr:
			if len(e.Args) == 0 {
				continue
			}
			if lit, ok := e.Args[0].(*ast.BasicLit); ok {
				return unquote(lit), true
			}
		}
	}
	return "", false
}

func unquote(lit *ast.BasicLit) string {
	if lit.Kind != token.STRING {
		return ""
	}
	if s, err := strconv.Unquote(lit.Value); err == nil {
		return s
	}
	return strings.Trim(lit.Value, "`")
}

// substituteVerbs replaces fmt verbs with values of a plausible shape.
//
// Whether a verb sits inside a quoted HCL string decides what it can become. In
// `name = "%[1]s-group"` it is part of a string and any bare token will do; as a
// whole right-hand side (`saml_idms_id = %[2]s`, where the caller passes a
// numeric id read from the environment) a bare token parses as a variable
// reference and validate rejects it. So: inside quotes, emit a token; outside,
// emit a self-contained value.
func substituteVerbs(s string) string {
	var b strings.Builder
	for line := range strings.SplitSeq(s, "\n") {
		b.WriteString(substituteLine(line))
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func substituteLine(line string) string {
	// A verb is inside a string when an odd number of unescaped quotes precede it.
	inString := func(idx int) bool {
		n := 0
		for i := range idx {
			if line[i] == '"' && (i == 0 || line[i-1] != '\\') {
				n++
			}
		}
		return n%2 == 1
	}
	replace := func(re *regexp.Regexp) {
		for {
			loc := re.FindStringIndex(line)
			if loc == nil {
				return
			}
			verb := re.FindStringSubmatch(line[loc[0]:loc[1]])[1]
			var val string
			switch {
			case inString(loc[0]):
				val = "test-acc-example"
			case verb == "d":
				val = "1"
			case verb == "q":
				val = `"test-acc-example"`
			default:
				// A bare %s/%v right-hand side is an id in every current caller.
				val = "1"
			}
			line = line[:loc[0]] + val + line[loc[1]:]
		}
	}
	replace(verbIndexedRe)
	replace(verbPlainRe)
	return line
}

// buildProvider compiles the provider and writes a dev_overrides CLI config
// pointing at it.
func buildProvider(root string) (tfrc, bin string, cleanup func(), err error) {
	dir, err := os.MkdirTemp("", "acctestprov")
	if err != nil {
		return "", "", func() {}, err
	}
	cleanup = func() {
		if rerr := os.RemoveAll(dir); rerr != nil {
			fmt.Fprintf(os.Stderr, "acctestconfig: cleaning %s: %v\n", dir, rerr)
		}
	}

	bin = filepath.Join(dir, "terraform-provider-kion")
	build := exec.Command("go", "build", "-o", bin, ".")
	build.Dir = root
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		cleanup()
		return "", "", func() {}, fmt.Errorf("build provider: %w", err)
	}

	tfrc = filepath.Join(dir, "dev.tfrc")
	body := fmt.Sprintf(`provider_installation {
  dev_overrides { "kionsoftware/kion" = %q }
  direct {}
}
`, dir)
	if err := os.WriteFile(tfrc, []byte(body), 0o600); err != nil {
		cleanup()
		return "", "", func() {}, err
	}
	return tfrc, bin, cleanup, nil
}

const providerBlock = `terraform {
  required_providers {
    kion = {
      source = "kionsoftware/kion"
    }
  }
}

provider "kion" {
  url    = "https://validate.invalid"
  apikey = "validate"
}
`

// validate writes one config into its own directory, resource addresses repeat
// across packages, so they cannot share one. And returns the diagnostic text if
// terraform rejects it.
func validate(work, tfrc string, c config) string {
	dir := filepath.Join(work, c.pkg+"_"+c.fn)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err.Error()
	}
	if err := os.WriteFile(filepath.Join(dir, "provider.tf"), []byte(providerBlock), 0o600); err != nil {
		return err.Error()
	}
	if err := os.WriteFile(filepath.Join(dir, "config.tf"), []byte(c.hcl), 0o600); err != nil {
		return err.Error()
	}

	cmd := exec.Command("terraform", "validate", "-no-color")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "TF_CLI_CONFIG_FILE="+tfrc, "TF_IN_AUTOMATION=true")
	out, err := cmd.CombinedOutput()
	if err == nil {
		return ""
	}
	return condense(string(out))
}

// condense trims terraform's box drawing down to the lines that name the problem.
func condense(s string) string {
	var keep []string
	for line := range strings.SplitSeq(s, "\n") {
		l := strings.TrimSpace(strings.TrimLeft(line, "│╷╵ "))
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "Error:") || strings.HasPrefix(l, "on ") ||
			strings.Contains(l, "is required") || strings.Contains(l, "not expected here") ||
			strings.Contains(l, "Read-Only") || strings.Contains(l, "Inappropriate value") {
			keep = append(keep, l)
		}
	}
	if len(keep) == 0 {
		return strings.TrimSpace(s)
	}
	return strings.Join(keep, "\n    ")
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "acctestconfig: %v\n", err)
	os.Exit(1)
}
