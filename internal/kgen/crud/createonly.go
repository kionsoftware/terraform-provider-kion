package crud

import (
	"fmt"
	"slices"
)

// createOnlyAttr is one archetype create_only entry resolved against the model.
type createOnlyAttr struct {
	TF      string // "secret"
	ModelGo string // "Secret"
}

// createOnlyBind is a create_only attribute bound to its create-response field.
type createOnlyBind struct {
	ModelGo string
	Capture string // expression yielding the types.String, null when unset or ""
}

// resolveCreateOnly maps an entity archetype's create_only entries onto model
// fields. Each must be a string attribute, and none may also be rewritten:
// both restore after the read-back, and from different sources.
func resolveCreateOnly(name string, arch archetype, model []ModelField) ([]createOnlyAttr, error) {
	if len(arch.CreateOnly) > 0 && arch.ResourceTemplate != "" {
		return nil, fmt.Errorf("%s: create_only is not supported with resource_template %q", name, arch.ResourceTemplate)
	}
	var out []createOnlyAttr
	for _, tf := range arch.CreateOnly {
		i := slices.IndexFunc(model, func(mf ModelField) bool { return mf.TFSDK == tf })
		if i < 0 {
			return nil, fmt.Errorf("%s: create_only attribute %q is not in the model", name, tf)
		}
		if model[i].Type != "types.String" {
			return nil, fmt.Errorf("%s: create_only attribute %q is %s; only string attributes are supported", name, tf, model[i].Type)
		}
		if slices.Contains(arch.Rewritten, tf) {
			return nil, fmt.Errorf("%s: create_only attribute %q is also rewritten", name, tf)
		}
		out = append(out, createOnlyAttr{TF: tf, ModelGo: model[i].GoName})
	}
	return out, nil
}

// createOnlyBinds resolves each create_only attribute to its field on the
// create response payload. base is the payload's Go access path, e.g.
// "created.Data.Value".
func createOnlyBinds(name string, attrs []createOnlyAttr, respFields []Field, base, sdkAlias string) ([]createOnlyBind, error) {
	var out []createOnlyBind
	for _, a := range attrs {
		i := slices.IndexFunc(respFields, func(f Field) bool { return f.JSONName == a.TF })
		if i < 0 {
			return nil, fmt.Errorf("%s: create_only attribute %q is not on the create response", name, a.TF)
		}
		f := respFields[i]
		capture, ok := createOnlyCapture(f.Type, base+"."+f.GoName, sdkAlias)
		if !ok {
			return nil, fmt.Errorf("%s: create_only attribute %q has create-response type %s; only string, OptString and OptNilString are supported", name, a.TF, f.Type)
		}
		out = append(out, createOnlyBind{ModelGo: a.ModelGo, Capture: capture})
	}
	return out, nil
}

// createOnlyCapture is the expression converting a string-typed SDK field to a
// types.String that is null when unset or "" (a public OAuth client has no
// secret). ok is false for any other SDK type.
func createOnlyCapture(sdkType, path, sdkAlias string) (expr string, ok bool) {
	const conv = "flex.OptStringToFrameworkNullIfEmpty"
	switch sdkType {
	case "OptString":
		return conv + "(" + path + ")", true
	case "string":
		return conv + "(" + sdkAlias + ".NewOptString(" + path + "))", true
	case "OptNilString":
		return fmt.Sprintf("%s(%s.OptString{Value: %s.Value, Set: %s.Set && !%s.Null})", conv, sdkAlias, path, path, path), true
	}
	return "", false
}

// createOnlySkip is the set of create_only tfsdk names, for flatten to skip.
func createOnlySkip(attrs []createOnlyAttr, also map[string]bool) map[string]bool {
	if len(attrs) == 0 {
		return also
	}
	out := make(map[string]bool, len(also)+len(attrs))
	for k, v := range also {
		out[k] = v
	}
	for _, a := range attrs {
		out[a.TF] = true
	}
	return out
}
