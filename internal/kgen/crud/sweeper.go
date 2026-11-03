package crud

import (
	_ "embed"
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"
)

//go:embed sweep_none.gtpl
var sweepNoneTmpl string

//go:embed sweep_list.gtpl
var sweepRealTmpl string

// renderSweep renders sweep.go. With a resolvable collection endpoint and a
// delete op it emits a real list+delete-by-prefix sweeper (self-contained: its
// own traversal of the resolved envelope, no coupling to the data source).
//
// Otherwise it emits a file that registers NOTHING and says why. A registered
// sweeper whose body returns nil reports success to `make sweep` while orphaned
// test-acc records accumulate, which is worse than having no sweeper at all.
//
// The returned reason is non-empty exactly when no sweeper was registered.
func renderSweep(rm ResourceModel) (out []byte, reason string, err error) {
	if rm.sweepList() != nil && rm.Delete != nil {
		b, berr := renderRealSweep(rm)
		if berr == nil {
			return b, "", nil
		}
		reason = berr.Error()
	} else {
		reason = sweepBlocker(rm)
	}
	b, err := renderNoSweep(rm.Name, "kion_"+rm.Name, reason)
	return b, reason, err
}

// sweepBlocker explains in one line why a resource cannot be swept, for the
// generated file and for the end-of-run report.
func sweepBlocker(rm ResourceModel) string {
	switch {
	case rm.Delete == nil && rm.RawDeletePath == "" && rm.sweepList() == nil:
		return "the API exposes neither a delete endpoint nor a resolvable collection endpoint"
	case rm.Delete == nil && rm.RawDeletePath == "":
		return "the API exposes no delete endpoint, so orphans cannot be removed"
	case rm.ListDowngrade != "":
		return "no resolvable collection endpoint: " + rm.ListDowngrade
	default:
		return "no collection endpoint is configured, so test resources cannot be enumerated"
	}
}

// --- no sweeper registered ---

type noSweepData struct {
	Pkg          string
	ResourceType string
	Reason       string
}

func renderNoSweep(pkg, resourceType, reason string) ([]byte, error) {
	d := noSweepData{Pkg: pkg, ResourceType: resourceType, Reason: reason}
	return execGoTemplate("sweep_none", sweepNoneTmpl, d, "sweep.go")
}

// --- real sweeper ---

// parentSweep drives the outer loop of a parent-scoped sweeper: the collection
// is only listable under a parent, so the sweeper enumerates parent ids first
// and lists children once per parent.
type parentSweep struct {
	listAccess
	Resource   string // parent resource name, e.g. "ou"
	Method     string // parent list method, e.g. "GetOUs"
	Params     string // parent list params type ("" when the op takes none)
	RespType   string // parent list success response type
	PageParam  string
	CountParam string
	IDGo       string // parent element id field, e.g. "ID"
	IDOpt      bool   // that field is Opt-wrapped
	Param      string // child list param taking the parent id, e.g. "ID"
	ParamType  string // that param's Go type, e.g. "int64"
	Match      string // Go condition a parent item must meet; "" keeps all
}

type realSweepData struct {
	listAccess
	Pkg, Pascal, ResourceType  string
	SDKAlias                   string
	SDKImport                  string
	ElemType                   string
	ListMethod, ListParams     string
	ListRespType               string
	PageParam, CountParam      string
	DeleteMethod, DeleteParams string
	// RawDeletePath sweeps through the private route when the public spec
	// publishes no delete; the SDK has no method for it.
	RawDeletePath string
	DeleteIDParam string
	DeleteIDType  string   // "int64" | "uint64" | "string", delete param id Go type
	IDSDKGo       string   // element id access path, e.g. "ID" or "Cft.Value.ID"
	MatchExprs    []string // string field accessors, e.g. "item.Key"
	Prefix        string
	Parent        *parentSweep // nil for a flat collection
}

func renderRealSweep(rm ResourceModel) ([]byte, error) {
	d, err := buildRealSweepData(rm)
	if err != nil {
		return nil, err
	}
	return execGoTemplate("sweep_real", sweepRealTmpl, d, "sweep.go")
}

func buildRealSweepData(rm ResourceModel) (realSweepData, error) {
	lm := rm.sweepList()
	delIDParam, delIDType, err := idParamName(rm.Delete.Params)
	if err != nil {
		return realSweepData{}, fmt.Errorf("%s sweeper: %w", rm.Name, err)
	}

	// rm is a copy: point the record view at the collection this sweeper walks,
	// which for a parent-scoped resource is not the data source's.
	rm.List = lm
	view := buildRecordView(rm)
	idField, ok := view.Fields["id"]
	if !ok {
		if len(view.Fields) == 0 {
			return realSweepData{}, fmt.Errorf("%s sweeper: read payload %s has no fields at all (empty OpenAPI schema)", rm.Name, rm.Read.RespPayload)
		}
		return realSweepData{}, fmt.Errorf("%s sweeper: response has no id field", rm.Name)
	}
	if !idFieldOK(idField.Type, delIDType) {
		return realSweepData{}, fmt.Errorf("%s sweeper: id field type %q unsupported", rm.Name, idField.Type)
	}
	if delIDType == "string" && (rm.SweepParent != nil || deleteMethodName(rm) == "") {
		return realSweepData{}, fmt.Errorf("%s sweeper: a string id is supported only for a flat collection with a typed delete", rm.Name)
	}

	d := realSweepData{
		listAccess:    lm.access(),
		Pkg:           rm.Name,
		Pascal:        rm.Pascal,
		ResourceType:  "kion_" + rm.Name,
		SDKAlias:      "generated",
		SDKImport:     sdkImport(),
		ElemType:      lm.ElemType,
		ListMethod:    lm.Method.Name,
		ListParams:    lm.Method.ParamsType,
		ListRespType:  lm.RespType,
		PageParam:     lm.PageParam,
		CountParam:    lm.CountParam,
		RawDeletePath: rm.RawDeletePath,
		DeleteMethod:  deleteMethodName(rm),
		DeleteParams:  rm.Delete.Method.ParamsType,
		DeleteIDParam: delIDParam,
		DeleteIDType:  delIDType,
		IDSDKGo:       view.Paths["id"],
		Prefix:        "test-acc",
		Parent:        rm.SweepParent,
	}

	// String-valued payload fields drive the prefix match.
	for _, mf := range rm.projectedFields() {
		pf, ok := view.Fields[mf.TFSDK]
		if !ok {
			continue
		}
		switch pf.Type {
		case "string":
			d.MatchExprs = append(d.MatchExprs, "item."+view.Paths[mf.TFSDK])
		case "OptString":
			d.MatchExprs = append(d.MatchExprs, "item."+view.Paths[mf.TFSDK]+`.Or("")`)
		}
	}
	if len(d.MatchExprs) == 0 {
		return realSweepData{}, fmt.Errorf("%s sweeper: no string field to match the test prefix", rm.Name)
	}
	slices.Sort(d.MatchExprs)
	return d, nil
}

// --- parent-scoped collections ---

// bindSweepParent populates rm.SweepList/rm.SweepParent from a `sweep_parent`
// archetype declaration: the resource's own collection re-resolved with the
// parent-id param allowed, plus the parent collection that supplies those ids.
func (g *generator) bindSweepParent(rm *ResourceModel, ds dsOps, idx sdkIndex, arch archetype) error {
	param := arch.SweepParentParam
	if param == "" {
		return fmt.Errorf("sweep_parent %q needs sweep_parent_param", arch.SweepParent)
	}
	lm, err := resolveListWith(ds.Read, idx, rm.Read, listOpts{ParentParam: param})
	if err != nil {
		return fmt.Errorf("sweep parent-scoped list: %w", err)
	}
	parentDS, ok := g.dataSources[arch.SweepParent]
	if !ok {
		return fmt.Errorf("sweep_parent %q is not a configured data source", arch.SweepParent)
	}
	ps, err := resolveSweepParent(arch.SweepParent, parentDS.Read, idx, lm, param)
	if err != nil {
		return err
	}
	if len(arch.SweepParentMatch) > 0 {
		var plm *listModel
		if plm, err = resolveListWith(parentDS.Read, idx, OpModel{}, listOpts{AcceptElem: func(elem string) bool {
			_, ok := parentIDField(elem, idx)
			return ok
		}}); err != nil {
			return err
		}
		if ps.Match, err = sweepMatchExpr(plm.ElemType, arch.SweepParentMatch, idx); err != nil {
			return fmt.Errorf("sweep_parent_match: %w", err)
		}
	}
	rm.SweepList, rm.SweepParent = lm, ps
	return nil
}

// resolveSweepParent resolves the parent half of a parent-scoped sweeper: the
// parent resource's own collection op (for enumerating parent ids) plus the
// child list param those ids feed. childList must already have been resolved
// with listOpts.ParentParam set to param.
func resolveSweepParent(parent string, ref *opRef, idx sdkIndex, childList *listModel, param string) (*parentSweep, error) {
	if ref == nil {
		return nil, fmt.Errorf("sweep parent %q has no data-source collection op configured", parent)
	}
	// The parent is enumerated for ids only, so any element carrying a usable ID
	// will do; the read payload it must match for a data source is irrelevant.
	plm, err := resolveListWith(ref, idx, OpModel{}, listOpts{AcceptElem: func(elem string) bool {
		_, ok := parentIDField(elem, idx)
		return ok
	}})
	if err != nil {
		return nil, fmt.Errorf("sweep parent %q: %w", parent, err)
	}
	idType, ok := parentIDField(plm.ElemType, idx)
	if !ok {
		return nil, fmt.Errorf("sweep parent %q: list element %s has no usable ID field", parent, plm.ElemType)
	}

	paramType := ""
	if childList.Params != nil {
		for _, f := range childList.Params.Fields {
			if f.GoName == param {
				paramType = f.Type
			}
		}
	}
	if paramType == "" {
		return nil, fmt.Errorf("sweep parent %q: child list op %s has no param %q", parent, childList.Method.Name, param)
	}

	return &parentSweep{
		listAccess: plm.access(),
		Resource:   parent,
		Method:     plm.Method.Name,
		Params:     plm.Method.ParamsType,
		RespType:   plm.RespType,
		PageParam:  plm.PageParam,
		CountParam: plm.CountParam,
		IDGo:       "ID",
		IDOpt:      strings.HasPrefix(idType, "Opt"),
		Param:      param,
		ParamType:  paramType,
	}, nil
}

// sweepMatchExpr renders match ({json field: integer}) as a Go condition on a
// parent list element named item. Every field must be an integer on elem.
func sweepMatchExpr(elem string, match map[string]string, idx sdkIndex) (string, error) {
	st, ok := idx.structs[elem]
	if !ok {
		return "", fmt.Errorf("parent element %s not found in SDK", elem)
	}
	keys := make([]string, 0, len(match))
	for k := range match {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var conds []string
	for _, k := range keys {
		v := match[k]
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			return "", fmt.Errorf("%s: value %q is not an integer", k, v)
		}
		var f *Field
		for i := range st.Fields {
			if st.Fields[i].JSONName == k {
				f = &st.Fields[i]
			}
		}
		if f == nil {
			return "", fmt.Errorf("parent element %s has no field %q", elem, k)
		}
		x := "item." + f.GoName
		switch f.Type {
		case "OptNilUint64", "OptNilInt64":
			conds = append(conds, fmt.Sprintf("%s.Set && !%s.Null && %s.Value == %s", x, x, x, v))
		case "OptUint64", "OptInt64":
			conds = append(conds, fmt.Sprintf("%s.Set && %s.Value == %s", x, x, v))
		case "uint64", "int64":
			conds = append(conds, fmt.Sprintf("%s == %s", x, v))
		default:
			return "", fmt.Errorf("%s: type %s is not an integer", k, f.Type)
		}
	}
	return strings.Join(conds, " && "), nil
}

// parentIDField reports the Go type of elem's ID field when it is one a sweeper
// can convert to an int64 parent id.
func parentIDField(elem string, idx sdkIndex) (string, bool) {
	st, ok := idx.structs[elem]
	if !ok {
		return "", false
	}
	for _, f := range st.Fields {
		if f.GoName != "ID" {
			continue
		}
		if listIDOK(f.Type) || f.Type == "int64" || f.Type == "uint64" {
			return f.Type, true
		}
	}
	return "", false
}

// deleteMethodName is empty for a resource whose only delete is the private
// route, so the sweeper template renders the raw call instead of an SDK one.
func deleteMethodName(rm ResourceModel) string {
	if rm.Delete == nil {
		return ""
	}
	return rm.Delete.Method.Name
}
