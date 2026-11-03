package crud

import (
	"fmt"
	"strings"
)

// readShape declares the JSON shape of a blended resource's private by-id read
// and how it maps onto the model, for reads the flat wire machinery can't
// express: a nested object under a dotted path (with optional key renames), or
// an array that must be exploded into flat model rows. Declared in
// codegen/private_endpoints.yaml under a resource's `read_shape`.
type readShape struct {
	Scalars []readShapeSub    `yaml:"scalars"`  // model scalar attrs at dotted json paths under data
	Objects []readShapeObject `yaml:"objects"`  // nested single-object attrs
	Explode *readShapeExplode `yaml:"explode"`  // one array-exploded list attr
	IDLists []readShapeIDList `yaml:"id_lists"` // set attrs projected from an array of objects
}

// readShapeIDList projects one field out of every element of a JSON array into
// a set of ids. Kion routinely answers with the expanded record where the model
// holds only the ids that were sent -- kion_billing_rule's billing_source_ids
// comes back as billing_criteria.payers[] = [{id, name}, ...]. The typed path
// already has this (see idProjFlat); this is the same idea for a raw read.
type readShapeIDList struct {
	TF    string `yaml:"tf"`    // model attr, e.g. billing_source_ids
	From  string `yaml:"from"`  // dotted json path to the []object under data
	Field string `yaml:"field"` // the id field on each element, e.g. id
}

// readShapeSub is one leaf mapping: a model tfsdk attr (or nested sub-attr) from
// a dotted json path, converted per Kind.
type readShapeSub struct {
	TF   string `yaml:"tf"`   // model tfsdk name (scalar) or Value sub-attr name
	From string `yaml:"from"` // dotted json path (scalars) or leaf key (subs)
	Kind string `yaml:"kind"` // id | string | int_string | int | bool | object

	// Subs are an object sub's own leaves, read one level deeper; only a
	// kind: object sub of a readShapeObject has them.
	Subs []readShapeSub `yaml:"subs"`

	// WriteOnly marks a secret the API accepts on write but never returns
	// truthfully. Kion's private payer read runs domain.Payer.Sanitize(), which
	// replaces every credential with the literal string "REDACTED".
	// Flattening that into state would diff against the
	// configured value on every plan, forever.
	//
	// A write-only field is omitted from the wire struct entirely and, on
	// flatten, carries the prior model value through instead of a wire value,
	// so state keeps what the practitioner configured. It still appears in the
	// schema and is still sent on create/update.
	//
	// Note this does NOT mark the attribute Sensitive: the schema_overrides
	// sidecar has no `sensitive` key, and its nested-attribute path replaces the
	// sibling list rather than merging into it, so an override there would drop
	// the attribute's siblings. A write-only field is therefore never read back,
	// but is not yet redacted from plan output.
	WriteOnly bool `yaml:"write_only"`
}

type readShapeObject struct {
	TF        string         `yaml:"tf"`         // model attr, e.g. aws_connection
	ValueType string         `yaml:"value_type"` // tfplugingen Value type, e.g. AwsConnectionValue
	From      string         `yaml:"from"`       // dotted json path to the object under data
	Subs      []readShapeSub `yaml:"subs"`       // sub.From is a leaf key under From
}

type readShapeExplode struct {
	TF        string         `yaml:"tf"`         // model list attr, e.g. roles
	ValueType string         `yaml:"value_type"` // element Value type, e.g. RolesValue
	From      string         `yaml:"from"`       // dotted json path to the []element under data
	Carry     []readShapeSub `yaml:"carry"`      // element fields copied to every exploded row
	Each      readShapeSub   `yaml:"each"`       // element []scalar exploded one-per-row
}

// kindWire maps a shape Kind to its Go wire type.
func kindWire(kind string) (string, error) {
	switch kind {
	case "id":
		return "uint64", nil
	case "string":
		return "string", nil
	case "int_string", "int", "datecode":
		return "int64", nil
	case "null_int":
		return "struct {\n\t\tInt   int64 `json:\"Int\"`\n\t\tValid bool  `json:\"Valid\"`\n\t}", nil
	case "null_string":
		return "struct {\n\t\tString string `json:\"String\"`\n\t\tValid  bool   `json:\"Valid\"`\n\t}", nil
	case "float":
		return "float64", nil
	case "bool":
		return "bool", nil
	case "opt_int":
		return "*int64", nil
	case "opt_string":
		return "*string", nil
	}
	return "", fmt.Errorf("unknown read_shape kind %q", kind)
}

// kindConv wraps a wire expression as the model attr.Value it flattens to.
func kindConv(kind, expr string) (string, error) {
	switch kind {
	case "id":
		return "types.StringValue(strconv.FormatUint(" + expr + ", 10))", nil
	case "int_string":
		return "types.StringValue(strconv.FormatInt(" + expr + ", 10))", nil
	case "datecode":
		// YYYYMM on the wire, YYYY-MM in the schema. See flex.DatecodeToFramework.
		return "flex.DatecodeToFramework(" + expr + ")", nil
	case "string":
		return "types.StringValue(" + expr + ")", nil
	case "int":
		return "types.Int64Value(" + expr + ")", nil
	case "null_int":
		// Kion serializes sql.NullInt64 as {"Int":1,"Valid":true}; declaring it
		// int64 made the whole read fail to decode.
		return "flex.NullIntToFramework(" + expr + ".Int, " + expr + ".Valid)", nil
	case "null_string":
		return "flex.NullStringToFramework(" + expr + ".String, " + expr + ".Valid)", nil
	case "float":
		return "types.Float64Value(" + expr + ")", nil
	case "bool":
		return "types.BoolValue(" + expr + ")", nil
	case "opt_int":
		// A plain JSON null (a *uint on the API side) stays null, not 0.
		return "types.Int64PointerValue(" + expr + ")", nil
	case "opt_string":
		return "types.StringPointerValue(" + expr + ")", nil
	}
	return "", fmt.Errorf("unknown read_shape kind %q", kind)
}

// goPath converts a dotted json path to the wire struct's Go field path
// (custom_billing_source.aws_connection -> CustomBillingSource.AwsConnection).
func goPath(dotted string) string {
	segs := strings.Split(dotted, ".")
	for i, s := range segs {
		segs[i] = pascalCase(s)
	}
	return strings.Join(segs, ".")
}

// --- wire struct generation (a tree of json keys -> Go struct text) ---

type wireNode struct {
	children map[string]*wireNode
	order    []string
	leaf     string // Go type when this is a leaf ("" for a struct/slice)
	slice    bool   // this node is a []struct (its children are the element fields)
}

func newWireNode() *wireNode { return &wireNode{children: map[string]*wireNode{}} }

// child returns (creating) the child for a json key.
func (n *wireNode) child(key string) *wireNode {
	if c, ok := n.children[key]; ok {
		return c
	}
	c := newWireNode()
	n.children[key] = c
	n.order = append(n.order, key)
	return c
}

// insert walks dotted keys from n, setting the final node's leaf Go type.
func (n *wireNode) insert(dotted, leaf string) {
	cur := n
	for k := range strings.SplitSeq(dotted, ".") {
		cur = cur.child(k)
	}
	cur.leaf = leaf
}

// buildWireTree assembles the tree under "data" for a shape.
func buildWireTree(s readShape) (*wireNode, error) {
	data := newWireNode()
	for _, sc := range s.Scalars {
		if sc.WriteOnly {
			continue // never read back; see readShapeSub.WriteOnly
		}
		w, err := kindWire(sc.Kind)
		if err != nil {
			return nil, err
		}
		data.insert(sc.From, w)
	}
	for _, o := range s.Objects {
		for _, sub := range o.Subs {
			if sub.WriteOnly {
				continue
			}
			if sub.Kind == "object" {
				for _, leaf := range sub.Subs {
					w, err := kindWire(leaf.Kind)
					if err != nil {
						return nil, err
					}
					data.insert(o.From+"."+sub.From+"."+leaf.From, w)
				}
				continue
			}
			w, err := kindWire(sub.Kind)
			if err != nil {
				return nil, err
			}
			data.insert(o.From+"."+sub.From, w)
		}
	}
	if s.Explode != nil {
		// Navigate to the array node and mark it a slice; its children are the
		// element fields (carry leaves + the exploded []scalar leaf).
		elem := data
		for k := range strings.SplitSeq(s.Explode.From, ".") {
			elem = elem.child(k)
		}
		elem.slice = true
		for _, c := range s.Explode.Carry {
			w, err := kindWire(c.Kind)
			if err != nil {
				return nil, err
			}
			elem.child(c.From).leaf = w
		}
		ew, err := kindWire(s.Explode.Each.Kind)
		if err != nil {
			return nil, err
		}
		elem.child(s.Explode.Each.From).leaf = "[]" + ew
	}
	for _, l := range s.IDLists {
		elem := data
		for k := range strings.SplitSeq(l.From, ".") {
			elem = elem.child(k)
		}
		elem.slice = true
		elem.child(l.Field).leaf = "uint64"
	}
	return data, nil
}

// emitStruct renders a struct node's fields (without the enclosing braces).
func emitStruct(n *wireNode) string {
	var b strings.Builder
	for _, key := range n.order {
		c := n.children[key]
		field := pascalCase(key)
		switch {
		case c.leaf != "":
			fmt.Fprintf(&b, "%s %s `json:\"%s\"`\n", field, c.leaf, key)
		case c.slice:
			fmt.Fprintf(&b, "%s []struct {\n%s} `json:\"%s\"`\n", field, emitStruct(c), key)
		default:
			fmt.Fprintf(&b, "%s struct {\n%s} `json:\"%s\"`\n", field, emitStruct(c), key)
		}
	}
	return b.String()
}

// buildWireStruct renders the full `type <pkg>Wire struct { Data struct {…} }`.
func buildWireStruct(pkg string, s readShape) (string, error) {
	tree, err := buildWireTree(s)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("type %sWire struct {\nData struct {\n%s} `json:\"data\"`\n}", pkg, emitStruct(tree)), nil
}

// buildNestedFlatten renders the body of flatten(ctx, w, m) diag.Diagnostics for
// a declared read shape.
func buildNestedFlatten(s readShape, byTF map[string]ModelField) (string, error) {
	var b strings.Builder
	for _, sc := range s.Scalars {
		if sc.WriteOnly {
			continue // leave m.<attr> holding the prior state value
		}
		mf, ok := byTF[sc.TF]
		if !ok {
			return "", fmt.Errorf("read_shape scalar %q not in model", sc.TF)
		}
		conv, err := kindConv(sc.Kind, "w.Data."+goPath(sc.From))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "m.%s = %s\n", mf.GoName, conv)
	}
	for _, o := range s.Objects {
		mf, ok := byTF[o.TF]
		if !ok {
			return "", fmt.Errorf("read_shape object %q not in model", o.TF)
		}
		base := "w.Data." + goPath(o.From)
		fmt.Fprintf(&b, "%sVal, %sValDiags := New%s(%s{}.AttributeTypes(ctx), map[string]attr.Value{\n", mf.GoName, mf.GoName, o.ValueType, o.ValueType)
		for _, sub := range o.Subs {
			// Every sub-attr must appear in the map, the Value constructor
			// requires the full attribute set. So a write-only sub carries the
			// prior model value through rather than being skipped.
			if sub.WriteOnly {
				// No prior state on import: CarryOrNull types the null.
				fmt.Fprintf(&b, "%q: flex.CarryOrNull(ctx, m.%s.%s, %s{}.AttributeTypes(ctx)[%q]),\n",
					sub.TF, mf.GoName, pascalCase(sub.TF), o.ValueType, sub.TF)
				continue
			}
			if sub.Kind == "object" {
				fmt.Fprintf(&b, "%q: flex.ObjectFromAttrs(ctx, &diags, %s{}.AttributeTypes(ctx)[%q], map[string]attr.Value{\n",
					sub.TF, o.ValueType, sub.TF)
				for _, leaf := range sub.Subs {
					conv, err := kindConv(leaf.Kind, base+"."+pascalCase(sub.From)+"."+pascalCase(leaf.From))
					if err != nil {
						return "", err
					}
					fmt.Fprintf(&b, "%q: %s,\n", leaf.TF, conv)
				}
				b.WriteString("}),\n")
				continue
			}
			conv, err := kindConv(sub.Kind, base+"."+pascalCase(sub.From))
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "%q: %s,\n", sub.TF, conv)
		}
		fmt.Fprintf(&b, "})\ndiags.Append(%sValDiags...)\nm.%s = %sVal\n", mf.GoName, mf.GoName, mf.GoName)
	}
	if e := s.Explode; e != nil {
		mf, ok := byTF[e.TF]
		if !ok {
			return "", fmt.Errorf("read_shape explode %q not in model", e.TF)
		}
		eachConv, err := kindConv(e.Each.Kind, "each")
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "var %sElems []%s\n", mf.GoName, e.ValueType)
		fmt.Fprintf(&b, "for _, outer := range w.Data.%s {\n", goPath(e.From))
		fmt.Fprintf(&b, "for _, each := range outer.%s {\n", pascalCase(e.Each.From))
		fmt.Fprintf(&b, "el, elDiags := New%s(%s{}.AttributeTypes(ctx), map[string]attr.Value{\n", e.ValueType, e.ValueType)
		for _, c := range e.Carry {
			conv, err := kindConv(c.Kind, "outer."+pascalCase(c.From))
			if err != nil {
				return "", err
			}
			fmt.Fprintf(&b, "%q: %s,\n", c.TF, conv)
		}
		fmt.Fprintf(&b, "%q: %s,\n", e.Each.TF, eachConv)
		fmt.Fprintf(&b, "})\ndiags.Append(elDiags...)\n%sElems = append(%sElems, el)\n}\n}\n", mf.GoName, mf.GoName)
		fmt.Fprintf(&b, "%sList, %sListDiags := types.ListValueFrom(ctx, %s{}.Type(ctx), %sElems)\n", mf.GoName, mf.GoName, e.ValueType, mf.GoName)
		fmt.Fprintf(&b, "diags.Append(%sListDiags...)\nm.%s = %sList\n", mf.GoName, mf.GoName, mf.GoName)
	}
	for _, l := range s.IDLists {
		mf, ok := byTF[l.TF]
		if !ok {
			return "", fmt.Errorf("read_shape id_list %q not in model", l.TF)
		}
		// Non-nil even when empty: an absent array means "none", which flattens
		// to an empty set rather than null, so a configured [] does not diff.
		fmt.Fprintf(&b, "%sIDs := make([]uint64, 0, len(w.Data.%s))\n", mf.GoName, goPath(l.From))
		fmt.Fprintf(&b, "for _, row := range w.Data.%s {\n%sIDs = append(%sIDs, row.%s)\n}\n",
			goPath(l.From), mf.GoName, mf.GoName, pascalCase(l.Field))
		fmt.Fprintf(&b, "%sSet, %sSetDiags := flex.Uint64SliceToFrameworkSetOrEmpty(ctx, %sIDs)\n",
			mf.GoName, mf.GoName, mf.GoName)
		fmt.Fprintf(&b, "diags.Append(%sSetDiags...)\nm.%s = %sSet\n", mf.GoName, mf.GoName, mf.GoName)
	}
	return b.String(), nil
}
