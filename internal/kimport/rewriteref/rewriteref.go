// Package rewriteref turns the literal foreign keys in a generated Terraform
// configuration into references between resource blocks.
//
// `terraform plan -generate-config-out` writes every attribute as a literal,
// because Terraform cannot know that `ou_id = 11` points at anything:
//
//	resource "kion_project" "kion_project_9" {
//	  ou_id = 11
//	}
//
// That is correct for the install it was imported from, and it is the whole
// problem everywhere else. The configuration does not express its own
// dependency graph, so destroy-and-reapply does not order correctly and
// recreating an OU does not cascade; and applied against a DIFFERENT install
// the literal either dangles or -- worse -- silently binds to whatever record
// happens to hold id 11 there. Rewritten to a reference:
//
//	resource "kion_project" "kion_project_9" {
//	  ou_id = kion_ou.kion_ou_11.id
//	}
//
// Terraform's graph then does the remapping itself, and gets ordering and
// parallelism for free. This is why cloning an install needs no id map on the
// Terraform path: references are the id map.
//
// The id -> label index cannot come from generated.tf, which contains no ids
// (they are Computed, so `-generate-config-out` omits them). It comes from the
// import blocks, where `to = kion_ou.kion_ou_11` and `id = "11"` sit together.
// Both files are therefore required.
package rewriteref

import (
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"

	"terraform-provider-kion/internal/kgen/references"
)

// Index maps a resource id to the configuration label that manages it,
// per tf_type: index["kion_ou"]["11"] == "kion_ou_11".
type Index map[string]map[string]string

// Unresolved is one foreign key that could not be rewritten. These are the
// boundary of the configuration: the referent is not managed here, so the
// literal is left in place and reported rather than silently kept.
type Unresolved struct {
	Resource string // block label, e.g. "kion_project.kion_project_9"
	Attr     string // "owner_user_ids"
	Target   string // "kion_user"
	Value    string // the literal that could not be resolved
}

// Cyclic is one foreign key whose referent IS managed here, but whose rewrite
// would close a dependency cycle Terraform refuses to plan. The literal stays,
// and is reported separately from Unresolved because the cause and the remedy
// differ: nothing is missing, the relation is simply expressible from both
// sides and only one of them can carry the reference.
type Cyclic struct {
	Resource string   // block label, e.g. "kion_user.kion_user_119"
	Attr     string   // "user_group_ids"
	Target   string   // "kion_user_group"
	Value    string   // the literal left in place
	Ref      string   // what it would have become, e.g. "kion_user_group.devs"
	Path     []string // existing path from Ref back to Resource that the edge would close
}

// Result reports what a rewrite did.
type Result struct {
	Rewritten  int          // literals replaced with references
	Unresolved []Unresolved // literals left alone because the referent is not managed here
	Cycles     []Cyclic     // literals left alone because a reference would be a cycle
	Blocks     int          // resource blocks visited
}

// TargetCounts summarizes Unresolved by target type, which is the actionable
// view: "47 references to kion_user could not be resolved" tells an operator
// they need to import users, or supply them another way.
func (r Result) TargetCounts() map[string]int {
	out := map[string]int{}
	for _, u := range r.Unresolved {
		out[u.Target]++
	}
	return out
}

// CycleCounts summarizes Cycles by the attribute that had to keep its literal,
// which is the actionable view: mutual relations cluster on one attribute pair.
func (r Result) CycleCounts() map[string]int {
	out := map[string]int{}
	for _, c := range r.Cycles {
		tfType, _, _ := strings.Cut(c.Resource, ".")
		out[tfType+"."+c.Attr]++
	}
	return out
}

// BuildIndex reads import blocks and returns the id -> label index.
//
// An import id may be compound ("11/4" for a parent-scoped resource, where 4 is
// the record's own id). Foreign keys hold the plain record id, so the trailing
// segment is what gets indexed; the full id is indexed too, harmlessly, so a
// compound value would also resolve.
func BuildIndex(importsHCL []byte, filename string) (Index, error) {
	p := hclparse.NewParser()
	f, diags := p.ParseHCL(importsHCL, filename)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing %s: %s", filename, diags.Error())
	}
	body, ok := f.Body.(*hclsyntax.Body)
	if !ok {
		return nil, fmt.Errorf("parsing %s: unexpected body type", filename)
	}

	idx := Index{}
	for _, blk := range body.Blocks {
		if blk.Type != "import" {
			continue
		}
		var to, id string
		for name, attr := range blk.Body.Attributes {
			switch name {
			case "to":
				// `to` is a traversal (kion_ou.kion_ou_11), not a string.
				if t, d := hcl.AbsTraversalForExpr(attr.Expr); !d.HasErrors() {
					parts := make([]string, 0, len(t))
					for _, step := range t {
						switch s := step.(type) {
						case hcl.TraverseRoot:
							parts = append(parts, s.Name)
						case hcl.TraverseAttr:
							parts = append(parts, s.Name)
						}
					}
					to = strings.Join(parts, ".")
				}
			case "id":
				if v, d := attr.Expr.Value(nil); !d.HasErrors() && v.Type() == cty.String {
					id = v.AsString()
				}
			}
		}
		tfType, label, ok := strings.Cut(to, ".")
		if !ok || id == "" {
			continue
		}
		if idx[tfType] == nil {
			idx[tfType] = map[string]string{}
		}
		idx[tfType][id] = label
		if _, rec, compound := strings.Cut(id, "/"); compound {
			idx[tfType][rec] = label
		}
	}
	return idx, nil
}

// refAttr is one reference-carrying attribute of one resource block, held with
// its tokens so the block can be visited once and rewritten once.
type refAttr struct {
	name   string
	target string
	tokens hclwrite.Tokens
}

// refBlock is one resource block, with its reference attributes in sorted
// order. Sorting matters: hclwrite returns attributes in a Go map, and which
// edge of a cycle gets refused depends on the order edges are offered.
type refBlock struct {
	node  string // "kion_user.kion_user_119"
	blk   *hclwrite.Block
	attrs []refAttr
}

// candidate is one integer literal that resolves to a block managed here, and
// so could become a reference -- unless doing so would close a cycle.
type candidate struct {
	block  int    // index into []refBlock
	attr   int    // index into that block's attrs
	token  int    // index into that attribute's tokens
	id     string // the literal
	tfType string // the family member holding the id
	label  string // the referent's block label
	node   string // "<tfType>.<label>"
}

// Rewrite replaces resolvable literal foreign keys in generatedHCL with
// references, returning the new file content and a report.
//
// Rewriting is not a per-attribute decision: a reference is an edge in
// Terraform's dependency graph, and an edge that closes a cycle produces a
// configuration Terraform refuses to plan. User<->group membership is settable
// from both sides, so rewriting both directions loops. Edges are therefore
// offered to a growing graph in a fixed order -- block order in the file, then
// attribute name, then position in the expression -- and any edge that would
// close a cycle is refused and reported. Fixed order is what makes the refusal
// reproducible: the same edge loses on every run.
func Rewrite(generatedHCL []byte, filename string, idx Index, refs *references.Map) ([]byte, Result, error) {
	f, diags := hclwrite.ParseConfig(generatedHCL, filename, hcl.InitialPos)
	if diags.HasErrors() {
		return nil, Result{}, fmt.Errorf("parsing %s: %s", filename, diags.Error())
	}

	var res Result
	var blocks []refBlock
	graph := refGraph{}

	for _, blk := range f.Body().Blocks() {
		if blk.Type() != "resource" || len(blk.Labels()) != 2 {
			continue
		}
		tfType, label := blk.Labels()[0], blk.Labels()[1]
		res.Blocks++
		node := tfType + "." + label

		rb := refBlock{node: node, blk: blk}
		for attrName, attr := range blk.Body().Attributes() {
			tokens := attr.Expr().BuildTokens(nil)
			// Seed the graph with references the configuration already holds,
			// so a second run cannot add an edge that closes a cycle with one
			// the first run created.
			for _, dst := range existingRefs(tokens, idx) {
				graph.add(node, dst)
			}
			if target, isRef := refs.Target(tfType, attrName); isRef {
				rb.attrs = append(rb.attrs, refAttr{name: attrName, target: target, tokens: tokens})
			}
		}
		sort.Slice(rb.attrs, func(i, j int) bool { return rb.attrs[i].name < rb.attrs[j].name })
		blocks = append(blocks, rb)
	}

	var cands []candidate
	for bi, rb := range blocks {
		for ai, ra := range rb.attrs {
			for ti, tok := range ra.tokens {
				if tok.Type != hclsyntax.TokenNumberLit {
					continue
				}
				id := string(tok.Bytes)
				// Kion ids are positive, so a literal 0 never names a record:
				// it means "none". `parent_ou_id = 0` is the case that matters
				// -- it creates a top-level OU, and a reference there would
				// invent a parent. Left alone, and NOT reported as a boundary
				// gap, because nothing is missing.
				if id == "0" {
					continue
				}
				tfType, label, ok := resolve(idx, refs.Members(ra.target), id)
				if !ok {
					res.Unresolved = append(res.Unresolved, Unresolved{
						Resource: rb.node, Attr: ra.name, Target: ra.target, Value: id,
					})
					continue
				}
				cands = append(cands, candidate{
					block: bi, attr: ai, token: ti,
					id: id, tfType: tfType, label: label, node: tfType + "." + label,
				})
			}
		}
	}

	// map[blockIdx]map[attrIdx]map[tokenIdx]traversal
	accepted := map[int]map[int]map[int]hclwrite.Tokens{}
	for _, c := range cands {
		rb, ra := blocks[c.block], blocks[c.block].attrs[c.attr]
		if path, cyclic := graph.wouldCycle(rb.node, c.node); cyclic {
			res.Cycles = append(res.Cycles, Cyclic{
				Resource: rb.node, Attr: ra.name, Target: ra.target,
				Value: c.id, Ref: c.node, Path: path,
			})
			continue
		}
		graph.add(rb.node, c.node)
		if accepted[c.block] == nil {
			accepted[c.block] = map[int]map[int]hclwrite.Tokens{}
		}
		if accepted[c.block][c.attr] == nil {
			accepted[c.block][c.attr] = map[int]hclwrite.Tokens{}
		}
		accepted[c.block][c.attr][c.token] = traversalTokens(c.tfType, c.label)
		res.Rewritten++
	}

	for bi := range blocks {
		for ai, ra := range blocks[bi].attrs {
			if byToken := accepted[bi][ai]; byToken != nil {
				blocks[bi].blk.Body().SetAttributeRaw(ra.name, spliceTokens(ra.tokens, byToken))
			}
		}
	}

	// Stable sorts over a deterministic append order: entries that tie on every
	// key still come out in the same place on every run.
	sort.SliceStable(res.Unresolved, func(i, j int) bool {
		a, b := res.Unresolved[i], res.Unresolved[j]
		if a.Target != b.Target {
			return a.Target < b.Target
		}
		if a.Resource != b.Resource {
			return a.Resource < b.Resource
		}
		if a.Attr != b.Attr {
			return a.Attr < b.Attr
		}
		return a.Value < b.Value
	})
	sort.SliceStable(res.Cycles, func(i, j int) bool {
		a, b := res.Cycles[i], res.Cycles[j]
		if a.Resource != b.Resource {
			return a.Resource < b.Resource
		}
		if a.Attr != b.Attr {
			return a.Attr < b.Attr
		}
		return a.Value < b.Value
	})
	return f.Bytes(), res, nil
}

// spliceTokens rebuilds an expression with the accepted literals replaced by
// traversals, preserving everything else (brackets, commas, whitespace) exactly
// as written. It works on the token stream rather than re-encoding a parsed
// value because an expression may already hold references, or a value this
// package should not touch.
// resolve finds id under the first of types that the index holds it for.
func resolve(idx Index, types []string, id string) (tfType, label string, ok bool) {
	for _, t := range types {
		if label, ok := idx[t][id]; ok {
			return t, label, true
		}
	}
	return "", "", false
}

func spliceTokens(tokens hclwrite.Tokens, replace map[int]hclwrite.Tokens) hclwrite.Tokens {
	out := make(hclwrite.Tokens, 0, len(tokens)+2*len(replace))
	for i, tok := range tokens {
		if sub, ok := replace[i]; ok {
			out = append(out, sub...)
			continue
		}
		out = append(out, tok)
	}
	return out
}

// existingRefs returns the nodes an expression already references, recognizing
// the `<tf_type>.<label>.id` shape this package writes. Only a tf_type present
// in the index counts, which keeps a two-segment traversal like a local value
// from being mistaken for a resource.
func existingRefs(tokens hclwrite.Tokens, idx Index) []string {
	var out []string
	for i := 0; i+4 < len(tokens); i++ {
		if tokens[i].Type != hclsyntax.TokenIdent ||
			tokens[i+1].Type != hclsyntax.TokenDot ||
			tokens[i+2].Type != hclsyntax.TokenIdent ||
			tokens[i+3].Type != hclsyntax.TokenDot ||
			tokens[i+4].Type != hclsyntax.TokenIdent {
			continue
		}
		tfType := string(tokens[i].Bytes)
		if _, ok := idx[tfType]; !ok {
			continue
		}
		out = append(out, tfType+"."+string(tokens[i+2].Bytes))
	}
	return out
}

// refGraph is the dependency graph of the configuration being rewritten: an
// edge src -> dst means src's configuration mentions dst, which is exactly the
// edge Terraform builds.
type refGraph map[string]map[string]bool

func (g refGraph) add(src, dst string) {
	if g[src] == nil {
		g[src] = map[string]bool{}
	}
	g[src][dst] = true
}

// wouldCycle reports whether adding src -> dst would close a cycle, and if so
// the existing path from dst back to src that it would close.
func (g refGraph) wouldCycle(src, dst string) ([]string, bool) {
	if src == dst {
		return []string{src}, true
	}
	if path := g.path(dst, src); path != nil {
		return path, true
	}
	return nil, false
}

// path returns a path from -> ... -> to, or nil. Neighbors are visited in
// sorted order so the reported path is identical on every run.
func (g refGraph) path(from, to string) []string {
	type step struct {
		node string
		prev int
	}
	seen := map[string]bool{from: true}
	queue := []step{{node: from, prev: -1}}
	for i := 0; i < len(queue); i++ {
		cur := queue[i]
		if cur.node == to {
			var rev []string
			for at := i; at >= 0; at = queue[at].prev {
				rev = append(rev, queue[at].node)
			}
			slices.Reverse(rev)
			return rev
		}
		next := make([]string, 0, len(g[cur.node]))
		for n := range g[cur.node] {
			if !seen[n] {
				next = append(next, n)
			}
		}
		sort.Strings(next)
		for _, n := range next {
			seen[n] = true
			queue = append(queue, step{node: n, prev: i})
		}
	}
	return nil
}

// traversalTokens builds the tokens for `<tfType>.<label>.id`.
func traversalTokens(tfType, label string) hclwrite.Tokens {
	ident := func(s string) *hclwrite.Token {
		return &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(s)}
	}
	dot := func() *hclwrite.Token {
		return &hclwrite.Token{Type: hclsyntax.TokenDot, Bytes: []byte(".")}
	}
	return hclwrite.Tokens{ident(tfType), dot(), ident(label), dot(), ident("id")}
}
