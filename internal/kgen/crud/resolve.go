package crud

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// opRef is one (method, path) op reference from generator_config.yaml.
type opRef struct {
	Method string
	Path   string
}

// resOps is the resource op-set for one resource.
type resOps struct {
	Create, Read, Update, Delete *opRef
}

// dsOps is the data-source op-set for one resource.
type dsOps struct {
	Read *opRef
}

// --- config loading (codegen/generator_config.yaml, the merged committed op-set) ---

type cfgOp struct {
	Path   string `yaml:"path"`
	Method string `yaml:"method"`
}

func (o *cfgOp) toRef() *opRef {
	if o == nil || o.Path == "" || o.Method == "" {
		return nil
	}
	return &opRef{Method: o.Method, Path: o.Path}
}

type cfgEntry struct {
	Create *cfgOp `yaml:"create"`
	Read   *cfgOp `yaml:"read"`
	Update *cfgOp `yaml:"update"`
	Delete *cfgOp `yaml:"delete"`
}

type cfgFile struct {
	Resources   map[string]cfgEntry `yaml:"resources"`
	DataSources map[string]cfgEntry `yaml:"data_sources"`
}

// heuristicLine matches a resource key annotated for review, e.g.
// "  guessed:  # heuristic, verify". Such resources are not safe to generate
// until a human verifies their ops, so the loader flags them for skipping.
var heuristicLine = regexp.MustCompile(`(?m)^\s{2}([A-Za-z0-9_]+):\s*#\s*(heuristic|INCOMPLETE)`)

// loadConfig parses the merged generator config. INCOMPLETE resources are
// comments (never unmarshalled); resources annotated "# heuristic" are returned
// in flagged so the caller can refuse them per the op-mapping precondition.
func loadConfig(path string) (resources map[string]resOps, dataSources map[string]dsOps, flagged map[string]bool, err error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("reading config %s: %w", path, err)
	}
	var cf cfgFile
	if err := yaml.Unmarshal(raw, &cf); err != nil {
		return nil, nil, nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	resources = make(map[string]resOps, len(cf.Resources))
	for name, e := range cf.Resources {
		resources[name] = resOps{
			Create: e.Create.toRef(),
			Read:   e.Read.toRef(),
			Update: e.Update.toRef(),
			Delete: e.Delete.toRef(),
		}
	}
	dataSources = make(map[string]dsOps, len(cf.DataSources))
	for name, e := range cf.DataSources {
		dataSources[name] = dsOps{Read: e.Read.toRef()}
	}
	flagged = map[string]bool{}
	for _, m := range heuristicLine.FindAllStringSubmatch(string(raw), -1) {
		flagged[m[1]] = true
	}
	return resources, dataSources, flagged, nil
}

// versionRange is one version_support.yaml entry.
type versionRange struct {
	Min string `yaml:"min"`
	Max string `yaml:"max"`
}

// versionSupport is codegen/version_support.yaml: the gated resources and data
// sources, each with its derived range.
type versionSupport struct {
	Resources   map[string]versionRange `yaml:"resources"`
	DataSources map[string]versionRange `yaml:"data_sources"`
}

func (vs versionSupport) resourceGated(name string) bool {
	_, ok := vs.Resources[name]
	return ok
}

// dataSourceGated reports whether the data source is gated. The package's one
// minKionVersion/maxKionVersion pair serves both, so a data source whose range
// differs from its resource's cannot be gated correctly and is an error.
func (vs versionSupport) dataSourceGated(name string) (bool, error) {
	ds, ok := vs.DataSources[name]
	if !ok {
		return false, nil
	}
	if res, ok := vs.Resources[name]; ok && res != ds {
		return false, fmt.Errorf("%s: data source range %+v differs from the resource's %+v; one version gate cannot enforce both", name, ds, res)
	}
	return true, nil
}

// loadVersionSupport reads codegen/version_support.yaml. A missing file means
// nothing is gated.
func loadVersionSupport(path string) (versionSupport, error) {
	var vs versionSupport
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return vs, nil
		}
		return vs, fmt.Errorf("reading version support %s: %w", path, err)
	}
	if err := yaml.Unmarshal(raw, &vs); err != nil {
		return vs, fmt.Errorf("parsing version support %s: %w", path, err)
	}
	return vs, nil
}

// --- SDK index ---

type sdkIndex struct {
	methods     map[string]ClientMethod // keyed by "METHOD path"
	structs     map[string]Struct       // schemas + parameters merged
	markerImpls map[string][]string
}

// buildIndex parses the SDK's generated Go for one version into an sdkIndex.
func buildIndex(src Source, sdkDir, version string) (sdkIndex, error) {
	if version == "" {
		return sdkIndex{}, errors.New("sdk version is required (the caller supplies it; see --sdk-version)")
	}
	gen := filepath.Join(sdkDir, "generated", version)
	cms, err := src.ClientMethods(filepath.Join(gen, "oas_client_gen.go"))
	if err != nil {
		return sdkIndex{}, err
	}
	methods := make(map[string]ClientMethod, len(cms))
	for _, m := range cms {
		methods[m.HTTPMethod+" "+m.Path] = m
	}
	schemas, err := src.Structs(filepath.Join(gen, "oas_schemas_gen.go"))
	if err != nil {
		return sdkIndex{}, err
	}
	params, err := src.Structs(filepath.Join(gen, "oas_parameters_gen.go"))
	if err != nil {
		return sdkIndex{}, err
	}
	structs := maps.Clone(schemas)
	maps.Copy(structs, params)
	markers, err := src.MarkerImpls(filepath.Join(gen, "oas_schemas_gen.go"))
	if err != nil {
		return sdkIndex{}, err
	}
	return sdkIndex{methods: methods, structs: structs, markerImpls: markers}, nil
}

// sharedResponse names the ogen result-union members that are NOT the resource
// payload envelope (shared error/status responses). Resolving a read's success
// envelope is "the one impl that is not in this set".
var sharedResponse = map[string]bool{
	"BadRequestResponse":          true,
	"UnauthorizedResponse":        true,
	"ForbiddenResponse":           true,
	"NotFoundResponse":            true,
	"ConflictResponse":            true,
	"UnprocessableEntityResponse": true,
	"TooManyRequestsResponse":     true,
	"InternalServerErrorResponse": true,
	"CreatedResponse":             true,
	"NoContentResponse":           true,
}

// noSDKOpError signals a mapped op that the SDK does not cover (private endpoint).
// The entity generator refuses these; raw-HTTP support is a follow-up.
type noSDKOpError struct {
	verb, method, path string
}

func (e noSDKOpError) Error() string {
	return fmt.Sprintf("%s op %s %s not found in SDK (private endpoint? raw-HTTP not yet supported)", e.verb, e.method, e.path)
}

// --- list envelope resolution ---

// sliceElem returns the element of a rendered slice type ("[]Label" -> "Label").
func sliceElem(typ string) (string, bool) {
	return strings.CutPrefix(typ, "[]")
}

// wrapperSliceElem resolves a field type that wraps a slice (e.g.
// OptNilLabelArray -> Value []Label -> Label) and reports whether the wrapper
// carries a .Null field. Returns ok=false when the type is not a slice wrapper.
func wrapperSliceElem(typeName string, idx sdkIndex) (elem string, nilAware, ok bool) {
	if e, isSlice := sliceElem(typeName); isSlice {
		return e, false, true
	}
	st, found := idx.structs[typeName]
	if !found {
		return "", false, false
	}
	for _, f := range st.Fields {
		if f.GoName != "Value" {
			continue
		}
		e, isSlice := sliceElem(f.Type)
		if !isSlice {
			return "", false, false
		}
		for _, g := range st.Fields {
			if g.GoName == "Null" {
				nilAware = true
			}
		}
		return e, nilAware, true
	}
	return "", false, false
}

// resolveList resolves a collection op (the data-source read) into a listModel:
// the success response, its Data envelope, the items slice, the Total field,
// and the pagination params. Any gap returns an error, and the caller then
// DOWNGRADES the data source to id-only.
//
// The items slice may sit one level down (shape A: Data wraps a struct holding
// Items) or be the Data field itself (shape B: Data is []Elem). Its element
// must be a type the single read also yields. Either the read payload or the
// payload's record-wrapper sub-object, so one field mapping serves both the
// by-id and the by-filter mode.
func resolveList(ref *opRef, idx sdkIndex, read OpModel) (*listModel, error) {
	return resolveListWith(ref, idx, read, listOpts{})
}

// listOpts relaxes resolveList for callers that are not the data source.
type listOpts struct {
	// AcceptElem overrides the "element must be a type the read yields" rule.
	AcceptElem func(elem string) bool
	// ParentParam names one required (non-Opt) param the caller supplies itself,
	// e.g. the parent id of a parent-scoped collection. The data source has no
	// value for such a param, but a sweeper enumerating parents does.
	ParentParam string
}

func resolveListWith(ref *opRef, idx sdkIndex, read OpModel, opts listOpts) (*listModel, error) {
	if ref == nil {
		return nil, fmt.Errorf("no data-source list op configured")
	}
	m, ok := idx.methods[ref.Method+" "+ref.Path]
	if !ok {
		return nil, noSDKOpError{verb: "list", method: ref.Method, path: ref.Path}
	}

	marker := lowerFirst(m.ResultType)
	var payload []string
	for _, t := range idx.markerImpls[marker] {
		if !sharedResponse[t] {
			payload = append(payload, t)
		}
	}
	if len(payload) != 1 {
		return nil, fmt.Errorf("list envelope: expected one payload impl for %s, got %v", marker, payload)
	}
	respType := payload[0]

	respStruct := idx.structs[respType]
	var data Field
	found := false
	for _, f := range respStruct.Fields {
		if f.GoName == "Data" || f.JSONName == "data" {
			data, found = f, true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("list envelope %s has no Data field", respType)
	}

	// The element type the read also produces. wrapperType is "" when the read
	// payload carries the record flat.
	wrapperType := ""
	if read.RespWrapperGo != "" {
		for _, f := range read.RespFields {
			if f.GoName == read.RespWrapperGo {
				wrapperType = strings.TrimPrefix(f.Type, "Opt")
			}
		}
	}
	accept := opts.AcceptElem
	if accept == nil {
		accept = func(elem string) bool { return elem == read.RespPayload || elem == wrapperType }
	}

	lm := &listModel{Method: m, RespType: respType, DataOpt: strings.HasPrefix(data.Type, "Opt")}
	if m.ParamsType != "" {
		if p, ok := idx.structs[m.ParamsType]; ok {
			lm.Params = &p
		}
	}

	if innerStruct, isEnvelope := idx.structs[strings.TrimPrefix(data.Type, "Opt")]; isEnvelope {
		// Shape A: Data wraps a struct holding the items slice (and often Total).
		lm.DataInner = strings.TrimPrefix(data.Type, "Opt")
		for _, f := range innerStruct.Fields {
			if f.GoName == "Total" {
				lm.TotalGo = "Total"
			}
			elem, nilAware, ok := wrapperSliceElem(f.Type, idx)
			if !ok {
				continue
			}
			if !accept(elem) {
				continue
			}
			lm.ItemsGo, lm.ItemsNil, lm.ElemType = f.GoName, nilAware, elem
			lm.ItemsOpt = !strings.HasPrefix(f.Type, "[]")
			lm.ElemIsWrapper = elem != read.RespPayload
		}
		if lm.ItemsGo == "" {
			return nil, fmt.Errorf("list envelope %s holds no []%s items field (read payload)", lm.DataInner, read.RespPayload)
		}
	} else {
		// Shape B: the Data field IS the items slice.
		elem, nilAware, ok := wrapperSliceElem(data.Type, idx)
		if !ok {
			return nil, fmt.Errorf("list envelope %s: Data field type %s is neither a known struct nor a slice", respType, data.Type)
		}
		if !accept(elem) {
			return nil, fmt.Errorf("list envelope %s: Data is []%s but the read yields %s", respType, elem, read.RespPayload)
		}
		lm.DataDirect = true
		lm.ItemsGo, lm.ItemsNil, lm.ElemType = data.GoName, nilAware, elem
		lm.ItemsOpt = !strings.HasPrefix(data.Type, "[]")
		lm.ElemIsWrapper = elem != read.RespPayload
		// A shape-B envelope carries its total (when it has one) at the top level.
		for _, f := range respStruct.Fields {
			if f.GoName == "Total" {
				lm.TotalGo = "Total"
			}
		}
	}

	// Pagination is optional: many Kion index endpoints return the whole
	// collection in one response and take no Page/Count at all. What is NOT
	// optional is being able to call the op blind, a required param the data
	// source has no value for (e.g. the parent id of a nested collection) makes
	// the endpoint unusable here.
	if lm.Params != nil {
		for _, f := range lm.Params.Fields {
			switch f.GoName {
			case "Page":
				if f.Type == "OptInt64" {
					lm.PageParam = "Page"
				}
			case "Count":
				if f.Type == "OptInt64" {
					lm.CountParam = "Count"
				}
			case opts.ParentParam:
				// Supplied by the caller (a sweeper that enumerates parents first).
			default:
				// A slice param (an ogen repeated query param) is omitted when nil,
				// so it is optional despite carrying no Opt wrapper.
				if !f.Optional && !strings.HasPrefix(f.Type, "[]") {
					return nil, fmt.Errorf("list op %s requires param %s (%s) that a whole-collection read cannot supply", m.Name, f.GoName, f.Type)
				}
			}
		}
	}
	lm.Paginated = lm.PageParam != "" && lm.CountParam != ""
	return lm, nil
}

// --- resolution ---

// resolveOp resolves one verb's op-ref against the SDK index into an OpModel.
// Returns (nil, nil) when ref is nil (verb absent).
func resolveOp(verb string, ref *opRef, idx sdkIndex) (*OpModel, error) {
	if ref == nil {
		return nil, nil
	}
	m, ok := idx.methods[ref.Method+" "+ref.Path]
	if !ok {
		return nil, noSDKOpError{verb: verb, method: ref.Method, path: ref.Path}
	}
	op := &OpModel{Verb: verb, Method: m}
	if m.BodyType != "" {
		if b, ok := idx.structs[strings.TrimPrefix(m.BodyType, "Opt")]; ok {
			op.Body = &b
		}
	}
	if m.ParamsType != "" {
		if p, ok := idx.structs[m.ParamsType]; ok {
			op.Params = &p
		}
	}
	if verb == "read" {
		respType, payload, dataPtr, respFields, err := resolveEnvelope(m.ResultType, idx)
		if err != nil {
			return nil, err
		}
		op.RespType = respType
		op.RespPayload = payload
		op.RespDataPtr = dataPtr
		op.RespFields = respFields
	}
	if verb == "create" {
		// A create that returns the record envelope (e.g. *ScopeResponse) rather
		// than a bare *CreatedResponse lets the id be read from the response.
		// Best-effort: on failure the entity template falls back to errs.CreatedID.
		if respType, payload, dataPtr, respFields, err := resolveEnvelope(m.ResultType, idx); err == nil {
			op.RespType = respType
			op.RespPayload = payload
			op.RespDataPtr = dataPtr
			op.RespFields = respFields
		}
	}
	return op, nil
}

// resolveEnvelope finds the read success response type (the non-error impl of
// the result marker), the domain payload type inside its Data envelope, the
// payload's fields, and whether the Data field is a pointer (*Payload) rather
// than an Opt-wrapper.
func resolveEnvelope(resultType string, idx sdkIndex) (respType, payload string, dataPtr bool, fields []Field, err error) {
	marker := lowerFirst(resultType)
	var candidates []string
	for _, t := range idx.markerImpls[marker] {
		if !sharedResponse[t] {
			candidates = append(candidates, t)
		}
	}
	if len(candidates) != 1 {
		return "", "", false, nil, fmt.Errorf("read envelope: expected exactly one payload impl for %s, got %v", marker, candidates)
	}
	respType = candidates[0]
	rs := idx.structs[respType]
	for _, f := range rs.Fields {
		if f.GoName == "Data" || f.JSONName == "data" {
			inner := strings.TrimPrefix(f.Type, "Opt")
			if p, ok := idx.structs[inner]; ok {
				return respType, inner, f.Ptr, p.Fields, nil
			}
		}
	}
	// Flat response with no Data envelope: the payload is the response itself.
	return respType, respType, false, rs.Fields, nil
}

// resolveResource assembles a ResourceModel for the entity archetype. ds is the
// data-source op-set; when its list op resolves, rm.List is populated for the
// dual-mode data source and real sweeper (nil otherwise, a logged degradation).
// rawDeleteOK says a private raw delete is declared for this resource, so a
// delete the SDK does not cover is expected rather than fatal.
func resolveResource(name string, ops resOps, ds dsOps, idx sdkIndex, model []ModelField, rawDeleteOK bool) (ResourceModel, error) {
	pascal := pascalCase(name)
	rm := ResourceModel{Name: name, Pascal: pascal, Model: pascal + "Model"}

	create, err := resolveOp("create", ops.Create, idx)
	if err != nil {
		return rm, err
	}
	read, err := resolveOp("read", ops.Read, idx)
	if err != nil {
		return rm, err
	}
	if create == nil || read == nil {
		return rm, fmt.Errorf("%s: entity archetype requires both create and read ops", name)
	}
	rm.Create, rm.Read = *create, *read

	if rm.Update, err = resolveOp("update", ops.Update, idx); err != nil {
		return rm, err
	}
	if rm.Delete, err = resolveOp("delete", ops.Delete, idx); err != nil {
		// A delete the SDK has no typed method for is fatal unless the resource
		// declares a private route to delete through; RawDeletePath then carries
		// it. Without this the whole resource is skipped and its committed files
		// become stale output nothing regenerates.
		var noOp noSDKOpError
		if !rawDeleteOK || !errors.As(err, &noOp) {
			return rm, err
		}
		rm.Delete = nil
	}

	for _, mf := range model {
		if mf.TFSDK == "id" {
			rm.IDField = mf
			continue
		}
		rm.Fields = append(rm.Fields, mf)
	}
	if rm.IDField.TFSDK == "" {
		return rm, fmt.Errorf("%s: generated model %s has no tfsdk:\"id\" field", name, rm.Model)
	}

	// Detect a record wrapper in the read payload (the flat model's fields nested
	// under a sub-object, e.g. CFTWithOwnersAndTags.cft).
	detectRecordWrapper(&rm.Read, rm.IDField, rm.Fields, idx)

	// The list element must be a type the single read also yields, so the
	// dual-mode DS and sweeper reuse the same field mapping.
	if lm, err := resolveList(ds.Read, idx, rm.Read); err != nil {
		rm.ListDowngrade = err.Error()
	} else {
		rm.List = lm
	}
	return rm, nil
}

// detectRecordWrapper finds a read-payload field that holds the flat model's
// fields as a nested sub-object (a "record wrapper"), setting read.RespWrapper*.
// The distinguishing test: the field's own json-name is NOT a model attribute
// (so it isn't a nested Value attribute like azure_policy's azure_policy), but
// several of the sub-object's fields ARE model attributes. Only the first such
// wrapper is used.
func detectRecordWrapper(read *OpModel, idField ModelField, fields []ModelField, idx sdkIndex) {
	byTF := map[string]ModelField{idField.TFSDK: idField}
	for _, f := range fields {
		byTF[f.TFSDK] = f
	}
	// How many model attributes the TOP-LEVEL payload already provides. A wrapper
	// is only the record holder when it covers MORE than the top level does,
	// otherwise a mere sub-record (e.g. scope's active_criteria_record, which has
	// id+criteria) would be mistaken for the wrapper and duplicate the id.
	topMatched := 0
	for _, f := range read.RespFields {
		if _, ok := byTF[f.JSONName]; ok {
			topMatched++
		}
	}
	bestGo, bestOpt, bestMatched := "", false, topMatched
	var bestFields []Field
	for _, f := range read.RespFields {
		if _, isModelAttr := byTF[f.JSONName]; isModelAttr {
			continue // a nested Value attribute, not a wrapper
		}
		base := strings.TrimPrefix(f.Type, "Opt")
		st, ok := idx.structs[base]
		if !ok || len(st.Fields) == 0 {
			continue
		}
		matched := 0
		for _, sf := range st.Fields {
			if _, ok := byTF[sf.JSONName]; ok {
				matched++
			}
		}
		if matched >= 2 && matched > bestMatched {
			bestGo, bestOpt, bestMatched, bestFields = f.GoName, strings.HasPrefix(f.Type, "Opt"), matched, st.Fields
		}
	}
	// The field-count test above only sees a FLAT model. When the model nests the
	// record instead, fall back to the shape test.
	if bestGo == "" {
		if wrapGo, opt, flds, ok := nestedRecordWrapper(read, byTF, idx); ok {
			bestGo, bestOpt, bestFields = wrapGo, opt, flds
		}
	}
	read.RespWrapperGo = bestGo
	read.RespWrapperOpt = bestOpt
	read.RespWrapperFields = bestFields
}

// nestedRecordWrapper finds the record wrapper when the MODEL nests the record
// rather than flattening it: the read payload carries no top-level id, and
// exactly one of its fields is a model OBJECT attribute (a tfplugingen
// <Pascal>Value type) whose SDK struct does carry the id, e.g.
// AzurePolicyAugmented.azure_policy -> AzurePolicyDefinition.id.
//
// detectRecordWrapper's field-count test is blind to this case twice over: the
// wrapper's json-name IS a model attribute (so the "nested Value attribute"
// guard skips it), and none of the record's fields are TOP-LEVEL model
// attributes either. They are sub-attributes of that object attribute, so the
// count is 0. Without this the payload looks id-less and the data source
// degrades to id-only.
//
// Deliberately narrow: it needs a genuinely missing top-level id, a model object
// attribute, an SDK struct behind it, an id inside that struct, and no second
// candidate. Anything ambiguous is left to degrade loudly rather than guessed at.
func nestedRecordWrapper(read *OpModel, byTF map[string]ModelField, idx sdkIndex) (goName string, opt bool, fields []Field, ok bool) {
	for _, f := range read.RespFields {
		if f.JSONName == "id" {
			return "", false, nil, false // the payload already yields the id
		}
	}
	for _, f := range read.RespFields {
		mf, isModelAttr := byTF[f.JSONName]
		if !isModelAttr || !strings.HasSuffix(mf.Type, "Value") {
			continue // not a nested-object attribute of the model
		}
		st, isStruct := idx.structs[strings.TrimPrefix(strings.TrimPrefix(f.Type, "Opt"), "Nil")]
		if !isStruct {
			continue
		}
		hasID := false
		for _, sf := range st.Fields {
			// json-name, not GoName: the record view keys its fields by json name,
			// so only a `json:"id"` field actually makes the id reachable.
			if sf.JSONName == "id" {
				hasID = true
				break
			}
		}
		if !hasID {
			continue
		}
		if goName != "" {
			return "", false, nil, false // two candidate records: too ambiguous to pick
		}
		goName, opt, fields = f.GoName, strings.HasPrefix(f.Type, "Opt"), st.Fields
	}
	return goName, opt, fields, goName != ""
}

// --- small helpers ---

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

// pascalCase title-cases each snake segment WITHOUT acronym expansion, matching
// tfplugingen's generated type names (ou_note -> OuNote, gcp_account ->
// GcpAccount), so <Pascal>Model / <Pascal>ResourceSchema line up.
func pascalCase(snake string) string {
	parts := strings.Split(snake, "_")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

// lowerCamelCase turns a snake_case name into lowerCamelCase
// ("aws_account" -> "awsAccount", "account" -> "account").
func lowerCamelCase(snake string) string {
	p := pascalCase(snake)
	if p == "" {
		return p
	}
	return strings.ToLower(p[:1]) + p[1:]
}

// spaceBeforeCaps inserts a space before each interior uppercase letter
// ("AwsAccount" -> "Aws Account", "Account" -> "Account").
func spaceBeforeCaps(pascal string) string {
	var b strings.Builder
	for i, r := range pascal {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
