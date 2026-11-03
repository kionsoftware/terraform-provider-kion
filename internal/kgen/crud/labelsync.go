package crud

import "fmt"

// labelSyncBind is what the template needs to sync a resource's labels.
//
// Kion carries labels on a per-resource sub-resource (GET/PUT
// /v3/<type>/{id}/labels), never in the resource's own request body, so a
// resource exposing `labels` has to read and write them separately. Until it
// does, the attribute is accepted and silently discarded.
type labelSyncBind struct {
	ModelGo string // model field, e.g. "Labels"
	Get     string // SDK read op
	Put     string // SDK write op
	Params  string // params field holding the parent id
	Element string // per-resource record type the GET returns
}

// resolveLabelSync validates the declaration against the model.
//
// byTF is nil for a bespoke archetype, which dispatches before the schema_gen
// model is read. The attribute name then cannot be checked, so a typo in the
// declaration surfaces as a compile error in the generated code rather than a
// generator error -- noisier, but not silent.
func resolveLabelSync(l membershipLabels, byTF map[string]ModelField) (*labelSyncBind, error) {
	attr := l.Attr
	if attr == "" {
		attr = "labels"
	}
	modelGo := pascalCase(attr)
	if byTF != nil {
		mf, ok := byTF[attr]
		if !ok {
			return nil, fmt.Errorf("attribute %q is not in the model", attr)
		}
		if mf.Type != "types.Map" {
			return nil, fmt.Errorf("attribute %q is %s, expected types.Map", attr, mf.Type)
		}
		modelGo = mf.GoName
	}
	for name, v := range map[string]string{"get": l.Get, "put": l.Put, "params": l.Params, "element": l.Element} {
		if v == "" {
			return nil, fmt.Errorf("%s is required", name)
		}
	}
	return &labelSyncBind{ModelGo: modelGo, Get: l.Get, Put: l.Put, Params: l.Params, Element: l.Element}, nil
}
