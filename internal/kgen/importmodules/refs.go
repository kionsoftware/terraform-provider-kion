package importmodules

import (
	"fmt"
	"sort"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

// DanglingRef is a reference into a converted resource that Rewrite could NOT
// retarget, and which therefore points at a resource that no longer exists.
//
// The common case is not this: `kion-import rewrite-refs` emits exactly
// `<type>.<label>.id`, and retargetRefs turns those into `module.<label>.id`,
// which is well-defined because every generated module exposes an `id` output.
// What is left here is a traversal into some OTHER attribute of a converted
// resource -- `kion_ou.parent.name`, say. A module exposes outputs, not
// attributes, and there may be no output of that name, so retargeting it would
// be a guess that plans against the wrong thing.
//
// Those are reported rather than rewritten. A configuration Terraform refuses
// to load is a better outcome than one that loads and manages the wrong
// record.
type DanglingRef struct {
	From   string // block making the reference, e.g. `module "child"`
	Attr   string // attribute holding it, e.g. "parent_ou_id"
	Target string // the reference as written, e.g. "kion_ou.parent"
}

func (d DanglingRef) String() string {
	return fmt.Sprintf("%s: %s reaches into %s, which this rewrite turns into a module "+
		"(only .id can be retargeted, to module.<label>.id)", d.From, d.Attr, d.Target)
}

// FindDanglingRefs reports references to resource blocks that Rewrite converts.
// It parses src independently of Rewrite -- hclwrite carries tokens, not
// evaluated expressions, and traversals are far easier to read out of the
// syntax tree.
//
// A reference to a resource that is NOT rewritten (a type with no manifest
// entry) is fine and not reported: that block is left exactly where it was.
func FindDanglingRefs(src []byte, manifest *Manifest) ([]DanglingRef, error) {
	f, diags := hclsyntax.ParseConfig(src, "generated.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("parse: %s", diags.Error())
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("parse: unexpected body type %T", f.Body)
	}

	byType := manifest.ByType()

	// Every resource block the rewrite will convert, as "<type>.<label>".
	converted := map[string]bool{}
	for _, b := range body.Blocks {
		if b.Type != "resource" || len(b.Labels) < 2 {
			continue
		}
		if _, known := byType[b.Labels[0]]; known {
			converted[b.Labels[0]+"."+b.Labels[1]] = true
		}
	}
	if len(converted) == 0 {
		return nil, nil
	}

	var out []DanglingRef
	for _, b := range body.Blocks {
		if len(b.Labels) < 2 {
			continue
		}
		from := fmt.Sprintf("%s %q", b.Type, b.Labels[1])
		if b.Type == "resource" {
			if _, known := byType[b.Labels[0]]; known {
				from = fmt.Sprintf("module %q", b.Labels[1])
			}
		}
		names := make([]string, 0, len(b.Body.Attributes))
		for name := range b.Body.Attributes {
			names = append(names, name)
		}
		sort.Strings(names) // deterministic report order
		for _, name := range names {
			for _, tr := range b.Body.Attributes[name].Expr.Variables() {
				if len(tr) < 2 {
					continue
				}
				root, ok := tr[0].(hcl.TraverseRoot)
				if !ok {
					continue
				}
				step, ok := tr[1].(hcl.TraverseAttr)
				if !ok {
					continue
				}
				target := root.Name + "." + step.Name
				if !converted[target] {
					continue
				}
				// `.id` is retargeted to module.<label>.id by retargetRefs, so
				// it is not dangling; anything else into a converted resource is.
				if len(tr) >= 3 {
					if attrStep, ok := tr[2].(hcl.TraverseAttr); ok && attrStep.Name == "id" {
						continue
					}
				}
				out = append(out, DanglingRef{From: from, Attr: name, Target: target})
			}
		}
	}
	return out, nil
}
