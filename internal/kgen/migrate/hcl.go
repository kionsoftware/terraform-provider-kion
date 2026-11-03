package migrate

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/hashicorp/hcl/v2/hclwrite"
)

// BlockToListAttr lists old nested-object blocks that became list/set nested
// ATTRIBUTES in the new schema, so `X { … }` must become `X = [{ … }]`. Unlike
// the ownership id-list projections, these carry a full object body, not just an
// id. (The four *_account resources also have block→single-nested changes but
// migrate via Layer 3 import, not config rewrite, so they are omitted.)
var BlockToListAttr = map[string][]string{
	"kion_project": {"project_funding", "budget", "move_ou_settings"},
}

// objFold folds a set of old top-level scalar attributes into a single new
// nested-object attribute.
type objFold struct {
	Target string   // new nested-object attribute name
	Fields []string // old top-level attributes to move inside it
}

// AttrsToObject lists resources whose old top-level attributes were folded into a
// nested object. kion_azure_policy's name/description/policy/parameters became a
// single `azure_policy { … }` object with those exact fields.
var AttrsToObject = map[string]objFold{
	"kion_azure_policy": {Target: "azure_policy", Fields: []string{"name", "description", "policy", "parameters"}},
}

// RequiredAdditions lists attributes the new schema requires that the old schema
// lacked or left optional. These are not renames, the rewrite handles those, so
// there is no old value to carry over and kmigrate cannot supply one. It reports
// the blocks that need them instead, which is the difference between a config
// that fails `terraform plan` with "Missing required argument" and one the
// practitioner has been told to finish.
//
// TestRequiredAdditions_matchSchemas keeps this in step with the snapshots.
var RequiredAdditions = map[string][]string{
	// The old kion_user had one computed attribute, id, with read and delete and
	// no create: it could only adopt an existing user for deletion. The new
	// resource is full CRUD.
	"kion_user": {"email", "first_name", "idms_id", "last_name", "username"},
	// Optional in the old provider, which could leave these accounts in the
	// account cache; the new resources always place one in a project.
	"kion_azure_account":  {"project_id", "start_datecode"},
	"kion_custom_account": {"project_id", "start_datecode"},
	"kion_gcp_account":    {"project_id", "start_datecode"},
}

// requiredAdditionHints follows a RequiredAdditions report where the obvious
// fix is wrong: adding project_id to a cached account moves it into a project.
var requiredAdditionHints = map[string]string{
	"kion_azure_account":  cachedAccountHint,
	"kion_custom_account": cachedAccountHint,
	"kion_gcp_account":    cachedAccountHint,
}

const cachedAccountHint = "if the account is in the account cache, move it to kion_account_cache " +
	"instead (see the upgrade guide); setting project_id would move it into that project"

// RewriteFile applies the input-attribute half of the state_upgrades transforms
// to one .tf file's `resource "kion_*"` blocks: attribute renames and the
// block-set → id-list restructure (owner_users { id = x } → owner_user_ids = [x]),
// and the kv_list map explode (tags = { k = v } → tags = [{ tag_key = "k", … }]).
// It preserves the id expressions verbatim (literals or references) and returns a
// human-readable list of the changes made. Comments and formatting are preserved
// by hclwrite. Computed-only transforms (id_int, unwrap, drops) are state-only
// and never appear in config, so they are ignored here.
//
// Both spellings of a repeatable block are handled: the literal block and the
// `dynamic "name" { for_each = … content { … } }` that generates it. An attribute
// has no dynamic form, so a dynamic block collapses into a for expression over
// the same for_each, `dynamic "owner_users" { for_each = var.ids; content { id =
// owner_users.value } }` becomes `owner_user_ids = [for owner_users in var.ids :
// owner_users]`, and a resource mixing both spellings gets a concat.
//
// The third return is manual follow-ups: things the new schema demands that no
// rewrite can produce. They are kept apart from changes so the CLI does not count
// them as edits it made.
func RewriteFile(src []byte, ups map[string]Transform) ([]byte, []string, []string, error) {
	return RewriteNamedFile("config.tf", src, ups)
}

// RewriteNamedFile is RewriteFile with the source's real path, so a parse
// diagnostic names the file the practitioner has to go and fix. RewriteFile
// passed a literal "config.tf" for every input, which meant a failure over a
// directory reported a filename that did not exist anywhere in the tree.
func RewriteNamedFile(name string, src []byte, ups map[string]Transform) ([]byte, []string, []string, error) {
	f, diags := hclwrite.ParseConfig(src, name, hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, nil, nil, fmt.Errorf("parse: %s", diags.Error())
	}
	// Alias resources: customers' config still uses the old type name (e.g.
	// kion_aws_iam_policy), but the transform is keyed by the new primary type.
	// Map each old type to its transform so those blocks get rewritten too.
	byOldType := map[string]Transform{}
	for _, tr := range ups {
		if tr.OldType != "" {
			byOldType[tr.OldType] = tr
		}
	}

	var changes, actions []string
	changes = append(changes, rewriteProviderConstraint(f.Body())...)

	for _, block := range f.Body().Blocks() {
		if block.Type() != "resource" || len(block.Labels()) < 2 {
			continue
		}
		rtype, rname := block.Labels()[0], block.Labels()[1]
		t, hasTransform := ups[rtype]
		if !hasTransform {
			t, hasTransform = byOldType[rtype]
		}
		_, hasDrops := ConfigDrops[rtype]
		_, hasRODrops := ReadOnlyDrops[rtype]
		_, hasBlockConv := BlockToListAttr[rtype]
		_, hasFold := AttrsToObject[rtype]
		_, hasRequired := RequiredAdditions[rtype]
		if !hasTransform && !hasDrops && !hasRODrops && !hasBlockConv && !hasFold && !hasRequired {
			continue
		}
		body := block.Body()

		for oldA, newA := range t.Rename {
			if attr := body.GetAttribute(oldA); attr != nil {
				toks := attr.Expr().BuildTokens(nil)
				body.RemoveAttribute(oldA)
				body.SetAttributeRaw(newA, toks)
				changes = append(changes, fmt.Sprintf("%s.%s: %s → %s", rtype, rname, oldA, newA))
			}
		}

		// Drop obsolete attributes the new schema no longer accepts.
		for _, dropA := range ConfigDrops[rtype] {
			if body.GetAttribute(dropA) != nil {
				body.RemoveAttribute(dropA)
				changes = append(changes, fmt.Sprintf("%s.%s: dropped obsolete attribute %s", rtype, rname, dropA))
			}
		}

		// Drop attributes the new schema kept but made read-only (computed): the
		// provider supplies the value now, so config may no longer set it.
		for _, dropA := range ReadOnlyDrops[rtype] {
			if body.GetAttribute(dropA) != nil {
				body.RemoveAttribute(dropA)
				changes = append(changes, fmt.Sprintf(
					"%s.%s: dropped %s: read-only in the new schema", rtype, rname, dropA))
			}
		}

		// Sorted so the rewritten attributes land in a stable order regardless of
		// map iteration.
		for _, newA := range sortedKeys(t.Project) {
			pr := t.Project[newA]
			// Each source block contributes a chunk: a static block one element
			// (its id expression), a `dynamic` block a whole list (a for
			// expression over its for_each).
			var chunks []listChunk
			var consumed []*hclwrite.Block
			var stuck []string
			dynamic := false
			for _, b := range collectSources(body, pr.From) {
				if d := parseDynamic(b); d != nil {
					if err := d.check(); err != nil {
						stuck = append(stuck, fmt.Sprintf("dynamic %q: %v", pr.From, err))
						continue
					}
					fa := d.content.Body().GetAttribute(pr.Field)
					if fa == nil {
						stuck = append(stuck, fmt.Sprintf("dynamic %q has no %q in its content", pr.From, pr.Field))
						continue
					}
					toks, err := d.forExpr(fa.Expr().BuildTokens(nil))
					if err != nil {
						stuck = append(stuck, fmt.Sprintf("dynamic %q: %v", pr.From, err))
						continue
					}
					chunks = append(chunks, listChunk{toks: toks, isList: true})
					consumed = append(consumed, b)
					dynamic = true
					continue
				}
				if b.Type() != pr.From {
					// A `dynamic` block we could not parse at all.
					stuck = append(stuck, fmt.Sprintf("dynamic %q is not a convertible for_each/content pair", pr.From))
					continue
				}
				fa := b.Body().GetAttribute(pr.Field)
				if fa == nil {
					stuck = append(stuck, fmt.Sprintf("block %q has no %q", pr.From, pr.Field))
					continue
				}
				chunks = append(chunks, listChunk{toks: fa.Expr().BuildTokens(nil)})
				consumed = append(consumed, b)
			}
			for _, s := range stuck {
				actions = append(actions, fmt.Sprintf("%s.%s: %s. Convert it to %s = [...] by hand", rtype, rname, s, newA))
			}
			if len(chunks) == 0 {
				continue
			}
			combined, err := combineChunks(chunks)
			if err != nil {
				actions = append(actions, fmt.Sprintf("%s.%s: could not build %s: %v", rtype, rname, newA, err))
				continue
			}
			// Remove only the blocks actually folded in (collect-then-remove;
			// RemoveBlock is safe per block).
			for _, b := range consumed {
				body.RemoveBlock(b)
			}
			body.SetAttributeRaw(newA, combined)
			src := pr.From + " { " + pr.Field + " }"
			if dynamic {
				src = "dynamic \"" + pr.From + "\" { … }"
			}
			changes = append(changes, fmt.Sprintf("%s.%s: %s → %s = [...]", rtype, rname, src, newA))
		}

		// Convert nested-object blocks that became list/set nested attributes:
		// project_funding { … } → project_funding = [{ … }].
		for _, blkName := range BlockToListAttr[rtype] {
			var chunks []listChunk
			var consumed []*hclwrite.Block
			for _, b := range collectSources(body, blkName) {
				if d := parseDynamic(b); d != nil {
					err := d.check()
					var toks hclwrite.Tokens
					if err == nil {
						toks, err = d.forExpr(blockBodyToObject(d.content))
					}
					if err != nil {
						actions = append(actions, fmt.Sprintf(
							"%s.%s: dynamic %q: %v. Convert it to %s = [{…}] by hand", rtype, rname, blkName, err, blkName))
						continue
					}
					chunks = append(chunks, listChunk{toks: toks, isList: true})
					consumed = append(consumed, b)
					continue
				}
				if b.Type() != blkName {
					continue
				}
				chunks = append(chunks, listChunk{toks: blockBodyToObject(b)})
				consumed = append(consumed, b)
			}
			if len(chunks) == 0 {
				continue
			}
			combined, err := combineChunks(chunks)
			if err != nil {
				actions = append(actions, fmt.Sprintf("%s.%s: could not build %s: %v", rtype, rname, blkName, err))
				continue
			}
			for _, b := range consumed {
				body.RemoveBlock(b)
			}
			body.SetAttributeRaw(blkName, combined)
			changes = append(changes, fmt.Sprintf("%s.%s: %s { … } → %s = [{…}]", rtype, rname, blkName, blkName))
		}

		// Explode a map attribute into a list of key/value objects:
		// tags = { k = v } → tags = [{ tag_key = "k", tag_value = v }].
		// Driven by the shared kv_list rule, the same one the state upgrader
		// reads, so config and state can never restructure it differently.
		for _, attrName := range sortedKeys(t.KVList) {
			fold := t.KVList[attrName]
			attr := body.GetAttribute(attrName)
			if attr == nil {
				continue
			}
			toks, err := mapToObjectList(attr.Expr().BuildTokens(nil), fold)
			switch {
			case err != nil:
				actions = append(actions, fmt.Sprintf(
					"%s.%s: %s. Rewrite it as a list of { %s, %s } objects by hand",
					rtype, rname, err, fold.KeyField, fold.ValField))
			case toks == nil: // already a list; nothing to do
			default:
				body.SetAttributeRaw(attrName, toks)
				changes = append(changes, fmt.Sprintf("%s.%s: %s = {…} → %s = [{%s, %s}]",
					rtype, rname, attrName, attrName, fold.KeyField, fold.ValField))
			}
		}

		// Fold scattered top-level attributes into a nested object:
		// azure_policy name/description/policy/parameters → azure_policy = { … }.
		if fold, ok := AttrsToObject[rtype]; ok {
			toks := hclwrite.Tokens{
				{Type: hclsyntax.TokenOBrace, Bytes: []byte("{")},
				{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
			}
			var present []string
			for _, f := range fold.Fields {
				if attr := body.GetAttribute(f); attr != nil {
					toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(f)})
					toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte(" = ")})
					toks = append(toks, attr.Expr().BuildTokens(nil)...)
					toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
					present = append(present, f)
				}
			}
			if len(present) > 0 {
				toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")})
				for _, f := range present {
					body.RemoveAttribute(f)
				}
				body.SetAttributeRaw(fold.Target, toks)
				changes = append(changes, fmt.Sprintf("%s.%s: %v → %s = {…}", rtype, rname, present, fold.Target))
			}
		}

		// Checked last: a rename above may have just supplied one of these.
		var missing []string
		for _, req := range RequiredAdditions[rtype] {
			if body.GetAttribute(req) == nil {
				missing = append(missing, req)
			}
		}
		if len(missing) > 0 {
			action := fmt.Sprintf("%s.%s: add %v: required by the new schema, and kmigrate has no value to supply",
				rtype, rname, missing)
			if hint := requiredAdditionHints[rtype]; hint != "" {
				action += "; " + hint
			}
			actions = append(actions, action)
		}
	}
	return f.Bytes(), changes, actions, nil
}

// blockBodyToObject renders a block's attribute body as an HCL object-constructor
// token sequence: { name = expr\n … }. Nested blocks within the body (rare for
// these funding/settings objects) are not carried over.
func blockBodyToObject(b *hclwrite.Block) hclwrite.Tokens {
	toks := hclwrite.Tokens{
		{Type: hclsyntax.TokenOBrace, Bytes: []byte("{")},
		{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")},
	}
	names := make([]string, 0, len(b.Body().Attributes()))
	for name := range b.Body().Attributes() {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		attr := b.Body().GetAttribute(name)
		toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenIdent, Bytes: []byte(name)})
		toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenEqual, Bytes: []byte(" = ")})
		toks = append(toks, attr.Expr().BuildTokens(nil)...)
		toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenNewline, Bytes: []byte("\n")})
	}
	toks = append(toks, &hclwrite.Token{Type: hclsyntax.TokenCBrace, Bytes: []byte("}")})
	return toks
}

// collectSources returns the blocks in body that hold values for the old
// repeatable block `name`: the literal `name { … }` blocks plus the
// `dynamic "name" { … }` blocks that generated them. Customers write both, a
// static block for a fixed owner, a dynamic one for a variable list, and a
// resource can carry a mix of the two.
func collectSources(body *hclwrite.Body, name string) []*hclwrite.Block {
	var out []*hclwrite.Block
	for _, b := range body.Blocks() {
		if b.Type() == name || (b.Type() == "dynamic" && len(b.Labels()) == 1 && b.Labels()[0] == name) {
			out = append(out, b)
		}
	}
	return out
}

// dynBlock is a parsed `dynamic "NAME" { for_each = … [iterator = it] content { … } }`.
// The new schema has no `dynamic` equivalent for an attribute, so the whole
// construct collapses into one for expression over the same for_each.
type dynBlock struct {
	iter    string          // iterator name, the label unless `iterator` overrides it
	forEach hclwrite.Tokens // the for_each expression, verbatim
	content *hclwrite.Block
}

// parseDynamic recognizes a `dynamic` block and pulls out the parts the rewrite
// needs. It returns nil for anything that is not a dynamic block; a dynamic block
// whose shape cannot be converted faithfully still parses, and check() says why,
// so the caller can tell "not a dynamic block" from "a dynamic block I must hand
// back to the practitioner".
func parseDynamic(b *hclwrite.Block) *dynBlock {
	if b.Type() != "dynamic" || len(b.Labels()) != 1 {
		return nil
	}
	d := &dynBlock{iter: b.Labels()[0]}
	if fe := b.Body().GetAttribute("for_each"); fe != nil {
		d.forEach = fe.Expr().BuildTokens(nil)
	}
	if it := b.Body().GetAttribute("iterator"); it != nil {
		d.iter = identExpr(it.Expr().BuildTokens(nil))
	}
	for _, ib := range b.Body().Blocks() {
		if ib.Type() == "content" {
			d.content = ib
			break
		}
	}
	return d
}

// identExpr returns the expression's text when it is a single bare identifier
// (`iterator = p`), else "".
func identExpr(toks hclwrite.Tokens) string {
	var ident string
	for _, t := range toks {
		switch t.Type {
		case hclsyntax.TokenIdent:
			if ident != "" {
				return ""
			}
			ident = string(t.Bytes)
		case hclsyntax.TokenComment, hclsyntax.TokenNewline:
		default:
			return ""
		}
	}
	return ident
}

// check reports why the dynamic block cannot be rewritten as a for expression,
// or nil when it can. Callers must run it before reading content: a `dynamic`
// block is only well-formed by convention, and a customer's file may not be.
func (d *dynBlock) check() error {
	switch {
	case d.iter == "":
		return fmt.Errorf("iterator is not a bare identifier")
	case len(d.forEach) == 0:
		return fmt.Errorf("no for_each")
	case d.content == nil:
		return fmt.Errorf("no content block")
	}
	return nil
}

// forExpr turns the dynamic block into `[for v in <for_each> : <elem>]`, where
// elem is the per-iteration expression with the block's iterator references
// resolved to the loop variable(s). The two-variable form is used only when the
// body reads `.key`, which for a list for_each is the index and for a map the
// key, the same meaning the for expression gives it.
//
// The result is round-tripped through the HCL parser, so a rewrite that would not
// parse is reported as an error rather than written to a customer's file.
func (d *dynBlock) forExpr(elem hclwrite.Tokens) (hclwrite.Tokens, error) {
	if err := d.check(); err != nil {
		return nil, err
	}
	keyVar, valVar := d.iter+"_key", d.iter
	body, usedKey := substIterator(elem, d.iter, keyVar, valVar)
	vars := valVar
	if usedKey {
		vars = keyVar + ", " + valVar
	}
	return parseExpr(fmt.Sprintf("[for %s in %s : %s]",
		vars, trimExpr(d.forEach), trimExpr(body)))
}

// substIterator rewrites `<iter>.value` → valVar and `<iter>.key` → keyVar in an
// expression, which is what turns a dynamic block's content into the body of a
// for expression. It reports whether `.key` was used. A qualified reference
// (`foo.<iter>.value`) is left alone, only the iterator's own root matches.
func substIterator(toks hclwrite.Tokens, iter, keyVar, valVar string) (hclwrite.Tokens, bool) {
	out := make(hclwrite.Tokens, 0, len(toks))
	usedKey := false
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		isRoot := i == 0 || toks[i-1].Type != hclsyntax.TokenDot
		if isRoot && t.Type == hclsyntax.TokenIdent && string(t.Bytes) == iter &&
			i+2 < len(toks) && toks[i+1].Type == hclsyntax.TokenDot &&
			toks[i+2].Type == hclsyntax.TokenIdent {
			var repl string
			switch string(toks[i+2].Bytes) {
			case "value":
				repl = valVar
			case "key":
				repl, usedKey = keyVar, true
			}
			if repl != "" {
				out = append(out, &hclwrite.Token{
					Type: hclsyntax.TokenIdent, Bytes: []byte(repl), SpacesBefore: t.SpacesBefore,
				})
				i += 2
				continue
			}
		}
		out = append(out, t)
	}
	return out, usedKey
}

// mapToObjectList turns a map-valued expression into the list of two-field
// objects the new schema wants. A literal map is expanded entry by entry, which
// is what a practitioner would have written; anything else, a variable, a
// merge(), a locals lookup. Becomes a for expression over the same value, which
// is correct without needing to know what it evaluates to.
//
// It returns (nil, nil) when the expression is already a list, so re-running
// kmigrate over a migrated file is a no-op.
func mapToObjectList(toks hclwrite.Tokens, f KVListRule) (hclwrite.Tokens, error) {
	src := trimExpr(toks)
	if strings.HasPrefix(src, "[") {
		return nil, nil
	}
	expr, diags := hclsyntax.ParseExpression([]byte(src), "kmigrate-input.tf", hcl.InitialPos)
	if diags.HasErrors() {
		return nil, fmt.Errorf("cannot parse %q: %s", src, diags.Error())
	}
	obj, ok := expr.(*hclsyntax.ObjectConsExpr)
	if !ok {
		return parseExpr(fmt.Sprintf("[for k, v in %s : { %s = k, %s = v }]", src, f.KeyField, f.ValField))
	}
	elems := make([]string, 0, len(obj.Items))
	for _, item := range obj.Items {
		elems = append(elems, fmt.Sprintf("{ %s = %s, %s = %s }",
			f.KeyField, objectKeySource(item.KeyExpr, src),
			f.ValField, exprSource(item.ValueExpr, src)))
	}
	return parseExpr("[" + strings.Join(elems, ", ") + "]")
}

// objectKeySource renders an object key as an expression usable as a value. A
// bare identifier key is a literal string in HCL (`{env = 1}` has the key
// "env"), so it has to be quoted once it moves into value position; anything
// already written as an expression carries over verbatim.
func objectKeySource(k hclsyntax.Expression, src string) string {
	if kw := hcl.ExprAsKeyword(k); kw != "" {
		return strconv.Quote(kw)
	}
	return exprSource(k, src)
}

// exprSource slices an expression's own source back out of the text it was
// parsed from, so comments, string escapes and interpolation survive untouched.
func exprSource(e hclsyntax.Expression, src string) string {
	r := e.Range()
	return src[r.Start.Byte:r.End.Byte]
}

// listChunk is one contribution to a rewritten list attribute: a single element
// expression (from a static block) or an entire list expression (from a dynamic
// block's for expression).
type listChunk struct {
	toks   hclwrite.Tokens
	isList bool
}

// combineChunks renders the chunks as one list-valued expression: the plain
// `[a, b, c]` when they are all single elements, the for expression alone when
// that is all there is, and `concat(…)` when a resource mixes static and dynamic
// blocks of the same name.
func combineChunks(chunks []listChunk) (hclwrite.Tokens, error) {
	lists := 0
	for _, c := range chunks {
		if c.isList {
			lists++
		}
	}
	if lists == 0 {
		elems := make([]hclwrite.Tokens, 0, len(chunks))
		for _, c := range chunks {
			elems = append(elems, c.toks)
		}
		return listTokens(elems), nil
	}
	if len(chunks) == 1 {
		return chunks[0].toks, nil
	}
	// Group runs of single elements into one literal list so concat's arguments
	// stay readable.
	var parts []string
	var pending []string
	flush := func() {
		if len(pending) > 0 {
			parts = append(parts, "["+strings.Join(pending, ", ")+"]")
			pending = nil
		}
	}
	for _, c := range chunks {
		if c.isList {
			flush()
			parts = append(parts, trimExpr(c.toks))
			continue
		}
		pending = append(pending, trimExpr(c.toks))
	}
	flush()
	return parseExpr("concat(" + strings.Join(parts, ", ") + ")")
}

// trimExpr renders tokens as source text without the leading/trailing whitespace
// hclwrite carries on an attribute's expression.
func trimExpr(toks hclwrite.Tokens) string {
	return strings.TrimSpace(string(toks.Bytes()))
}

// parseExpr parses generated expression source back into tokens, so every
// construction in this file is validated by the same parser Terraform uses before
// it can reach a customer's file.
func parseExpr(src string) (hclwrite.Tokens, error) {
	f, diags := hclwrite.ParseConfig([]byte("x = "+src+"\n"), "kmigrate-generated.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil, fmt.Errorf("generated %q does not parse: %s", src, diags.Error())
	}
	attr := f.Body().GetAttribute("x")
	if attr == nil {
		return nil, fmt.Errorf("generated %q does not parse as an expression", src)
	}
	return attr.Expr().BuildTokens(nil), nil
}

// sortedKeys returns a map's keys in a stable order.
func sortedKeys[V any](m map[string]V) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// listTokens wraps element expressions into a `[a, b, c]` token sequence.
func listTokens(elems []hclwrite.Tokens) hclwrite.Tokens {
	out := hclwrite.Tokens{{Type: hclsyntax.TokenOBrack, Bytes: []byte("[")}}
	for i, e := range elems {
		if i > 0 {
			out = append(out, &hclwrite.Token{Type: hclsyntax.TokenComma, Bytes: []byte(", ")})
		}
		out = append(out, e...)
	}
	out = append(out, &hclwrite.Token{Type: hclsyntax.TokenCBrack, Bytes: []byte("]")})
	return out
}

// NewProviderConstraint is what a migrated configuration must ask for. The new
// provider is a new major version at the same registry address, so an unbumped
// constraint resolves to the OLD provider and every rewrite kmigrate just made
// becomes invalid config.
const NewProviderConstraint = ">= 1.0.0"

// rewriteProviderConstraint bumps `version` inside
// terraform { required_providers { kion = { ... } } }.
//
// Without this kmigrate produced configuration that could not init: it rewrote
// every resource to the new schema while leaving the constraint pinned to the
// old provider. The migration guide asks the practitioner to bump it by hand and
// shows a single block. But a real configuration has one per directory (the
// import tool emits eight), so doing it by hand means finding all of them.
//
// Matched on the `source` inside the object rather than the attribute name, so a
// provider aliased to something other than `kion` is still found and a different
// provider that happens to be named `kion` is left alone.
func rewriteProviderConstraint(body *hclwrite.Body) []string {
	var changes []string
	for _, tf := range body.Blocks() {
		if tf.Type() != "terraform" {
			continue
		}
		for _, rp := range tf.Body().Blocks() {
			if rp.Type() != "required_providers" {
				continue
			}
			for name, attr := range rp.Body().Attributes() {
				toks := attr.Expr().BuildTokens(nil)
				if !tokensMentionKion(toks) {
					continue
				}
				newToks, old, ok := setVersionInObject(toks, NewProviderConstraint)
				if !ok || old == NewProviderConstraint {
					continue
				}
				rp.Body().SetAttributeRaw(name, newToks)
				changes = append(changes, fmt.Sprintf(
					"required_providers.%s: version %q → %q", name, old, NewProviderConstraint))
			}
		}
	}
	return changes
}

func tokensMentionKion(toks hclwrite.Tokens) bool {
	for _, t := range toks {
		if t.Type == hclsyntax.TokenQuotedLit && strings.Contains(string(t.Bytes), "kionsoftware/kion") {
			return true
		}
	}
	return false
}

// setVersionInObject replaces the string value of the object's `version` key,
// returning the previous value. Token surgery rather than re-emitting the object
// so every other key, and the author's formatting, survive untouched.
func setVersionInObject(toks hclwrite.Tokens, want string) (hclwrite.Tokens, string, bool) {
	for i, t := range toks {
		if t.Type != hclsyntax.TokenIdent || string(t.Bytes) != "version" {
			continue
		}
		// version = "x"  ->  ident, equals, oquote, lit, cquote
		for j := i + 1; j < len(toks) && j < i+4; j++ {
			if toks[j].Type != hclsyntax.TokenQuotedLit {
				continue
			}
			old := string(toks[j].Bytes)
			out := make(hclwrite.Tokens, len(toks))
			copy(out, toks)
			repl := *toks[j]
			repl.Bytes = []byte(want)
			out[j] = &repl
			return out, old, true
		}
	}
	return toks, "", false
}
