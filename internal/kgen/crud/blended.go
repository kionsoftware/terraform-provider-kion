package crud

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
)

//go:embed blendedtest.gtpl
var blendedTestTmpl string

//go:embed blended.gtpl
var blendedTmpl string

// blendedKind is the crud_archetypes.yaml kind for a resource whose CRUD is
// split per-op between the typed public SDK (/v3,/v4,/beta) and raw private
// HTTP (/v1,/v2). The public ops (always create; usually update/delete) are
// generated exactly like the entity archetype; the private ops (always the
// by-id read; sometimes update or delete) are generated over conns.Raw*.
const blendedKind = "blended"

// rawReadField is one flat scalar field of a blended resource's raw-read wire
// struct. Reuses rawField's expressions (ToExpr for flatten, FromExpr for a raw
// write body).
type rawReadData struct {
	Method string     // "RawGet"
	Path   string     // "/v1/app-role/{id}"
	IDGo   string     // model id field GoName
	Fields []rawField // flat scalar wire fields (excludes id). Flat reads only
	// Nested reads (declared read_shape): the generator emits WireStructGo and a
	// flatten(ctx, w, m) diag.Diagnostics body rather than the flat wire machinery.
	Nested       bool
	WireStructGo string // the full `type <pkg>Wire struct {…}` text
	FlattenGo    string // the flatten function body text
	// UsesAttr reports whether the flatten body constructs a nested Value and
	// therefore needs the framework attr package. A read_shape of only scalars
	// assigns plain types.* values and must not import it.
	UsesAttr bool
	// MissingViaList: see rawResourceOps.MissingViaList.
	MissingViaList string
}

// rawWriteData is a blended resource's raw update or delete op.
type rawWriteData struct {
	Method  string // "RawPatch" | "RawDelete"
	Path    string // "/v2/project-note/{id}"
	HasBody bool   // update marshals the model into the wire body; delete does not
	// MergeGo is the []flex.MergeField body of a read-modify-write update
	// (declared write_shape); empty for a body built from the model alone.
	MergeGo string
}

// isRawOp reports whether an op is rendered over raw HTTP rather than the SDK.
//
// A private path is the usual reason, and was once the only one -- hence `spec`,
// which is recorded for documentation and is not used for routing (the raw call
// takes the literal path). But it is not the invariant. kion_billing_rule is
// served entirely by public endpoints whose spec schema for the read is EMPTY:
// the SDK's typed model has no fields, so the typed flatten assigns nothing and
// an import yields a bare id. That read has to be raw too.
//
// A public path can need it too, and says so with `raw: true`. kion_billing_rule
// is served entirely by public endpoints whose spec schema for the read is
// EMPTY: the SDK's typed model has no fields, so the typed flatten assigns
// nothing and an import yields a bare id.
//
// The flag is required rather than inferred from the path, because these
// entries also DOCUMENT the typed public writes that sit beside a private read.
// Treating any declared path as raw turned those into raw writes and broke
// every other blended resource.
func isRawOp(op rawOp) bool {
	return (op.Spec == "private" || op.Raw) && op.Path != ""
}

// generateBlended resolves and writes a blended resource (<name>.go +
// service_package.go, resource-only like the raw archetype).
func (g *generator) generateBlended(dir, name string, ops resOps, idx sdkIndex, pe rawResourceOps, model []ModelField, tvPath string, gated, force bool) (int, error) {
	schemaGen := filepath.Join(dir, name+"_schema_gen.go")
	d, err := g.resolveBlended(name, ops, idx, pe, model, gated, schemaGen)
	if err != nil {
		return 0, err
	}
	resourceGo, err := execGoTemplate("blended", blendedTmpl, d, name+".go")
	if err != nil {
		return 0, err
	}
	// Companions first, then prune: service_package.go registers whatever data
	// source is on disk, so a companion written after it would not be
	// registered, and a scaffolder stub left before it would be registered
	// without ever being generated.
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	if err := g.pruneUnwritten(dir, name); err != nil {
		return 0, err
	}
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newServicePackageData(name, d.Pascal, blendedDataSourceCtor(dir, name, d.Pascal)), "service_package.go")
	if err != nil {
		return 0, err
	}
	files := []genFile{
		{filepath.Join(dir, name+".go"), resourceGo},
		{filepath.Join(dir, "service_package.go"), pkgGo},
	}
	for _, f := range files {
		if err := g.writeFile(f.path, f.data, force); err != nil {
			return 0, err
		}
	}
	tv, hasTV, err := loadTestValues(tvPath, name)
	if err != nil {
		return 0, err
	}
	if hasTV {
		td, err := buildBlendedTestData(d, model, tv)
		if err != nil {
			return 0, err
		}
		test, err := execGoTemplate("blendedtest", blendedTestTmpl, td, name+"_test.go")
		if err != nil {
			return 0, err
		}
		if err := g.writeFile(filepath.Join(dir, name+"_test.go"), test, force); err != nil {
			return 0, err
		}
	} else {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no test_values entry; skipping acceptance tests\n", name)
	}
	return 1, nil
}

// resolveBlended builds the entityData for a blended resource: the typed public
// ops go through the same resolution + binds as the entity archetype (via
// buildEntityData over a partial ResourceModel), and the private ops are
// layered on as raw read/update/delete.
func (g *generator) resolveBlended(name string, ops resOps, idx sdkIndex, pe rawResourceOps, model []ModelField, gated bool, schemaGen string) (entityData, error) {
	if pe.Read.Path == "" || !isRawOp(pe.Read) {
		return entityData{}, fmt.Errorf("%s: blended archetype requires a private read in private_endpoints.yaml", name)
	}

	pascal := pascalCase(name)
	rm := ResourceModel{Name: name, Pascal: pascal, Model: pascal + "Model", Gated: gated,
		AtLeastOneOf:     g.configValidators.For(name),
		RequiredWhen:     g.configValidators.RequiredWhenFor(name),
		RequiredTogether: g.configValidators.RequiredTogetherFor(name),
	}

	// Typed public create (required).
	create, err := resolveOp("create", ops.Create, idx)
	if err != nil {
		return entityData{}, fmt.Errorf("%s create: %w", name, err)
	}
	if create == nil {
		return entityData{}, fmt.Errorf("%s: blended archetype requires a typed public create", name)
	}
	rm.Create = *create

	// Typed public update/delete. Only when that op is NOT private.
	if !isRawOp(pe.Update) {
		if rm.Update, err = resolveOp("update", ops.Update, idx); err != nil {
			return entityData{}, fmt.Errorf("%s update: %w", name, err)
		}
	}
	if !isRawOp(pe.Delete) {
		if rm.Delete, err = resolveOp("delete", ops.Delete, idx); err != nil {
			return entityData{}, fmt.Errorf("%s delete: %w", name, err)
		}
	}

	for _, mf := range model {
		if mf.TFSDK == "id" {
			rm.IDField = mf
			continue
		}
		rm.Fields = append(rm.Fields, mf)
	}
	if rm.IDField.TFSDK == "" {
		return entityData{}, fmt.Errorf("%s: generated model %s has no tfsdk:%q field", name, rm.Model, "id")
	}

	// Nested create/update expand: a blended create/update body may still carry
	// nested objects (permission_scheme roles, billing_source aws_connection).
	byTF := map[string]ModelField{}
	for _, mf := range model {
		byTF[mf.TFSDK] = mf
	}
	if rm.CreateNested, err = resolveNested(g.src, schemaGen, rm.Create.Body, byTF, idx, nestedOpts{NoGuard: pe.NoGuard}); err != nil {
		return entityData{}, fmt.Errorf("%s create nested: %w", name, err)
	}
	if rm.Update != nil {
		// The private read's declared shape doubles as the write shape: a body
		// array that the read explodes into flat model rows has to be regrouped
		// on the way out. See implodeBind.
		opts := nestedOpts{NoGuard: pe.NoGuard, IDAttr: rm.IDField.TFSDK, IDVar: "idInt"}
		if pe.ReadShape != nil {
			opts.Implode = pe.ReadShape.Explode
		}
		if rm.UpdateNested, err = resolveNested(g.src, schemaGen, rm.Update.Body, byTF, idx, opts); err != nil {
			return entityData{}, fmt.Errorf("%s update nested: %w", name, err)
		}
	}

	d, err := buildEntityData(rm)
	if err != nil {
		return entityData{}, fmt.Errorf("%s: %w", name, err)
	}
	d.Blended = true
	if d.IDParamType == "string" {
		return entityData{}, fmt.Errorf("%s: blended archetype does not support a string id", name)
	}

	// buildEntityData derives IDParamType from the read/delete params; a blended
	// resource has no typed read, and delete may be raw, so source the typed
	// id-param Go type from whichever typed op has a params struct.
	if d.IDParamType == "" {
		d.IDParamType = "int64"
		for _, op := range []*OpModel{rm.Update, rm.Delete} {
			if op == nil {
				continue
			}
			if _, t, e := idParamName(op.Params); e == nil {
				d.IDParamType = t
				break
			}
		}
		if d.IDParamType == "string" {
			return entityData{}, fmt.Errorf("%s: blended archetype does not support a string id", name)
		}
		d.IDParseBits = idParseBits(d.IDParamType)
	}

	// Raw private read: a declared nested shape (read_shape) when the wire isn't
	// a flat 1:1 of the model scalars, else the flat wire from the model.
	readMethod, err := rawVerb(pe.Read.Method)
	if err != nil {
		return entityData{}, fmt.Errorf("%s raw read: %w", name, err)
	}
	if pe.ReadShape != nil {
		wireGo, err := buildWireStruct(name, *pe.ReadShape)
		if err != nil {
			return entityData{}, fmt.Errorf("%s read_shape wire: %w", name, err)
		}
		flattenGo, err := buildNestedFlatten(*pe.ReadShape, byTF)
		if err != nil {
			return entityData{}, fmt.Errorf("%s read_shape flatten: %w", name, err)
		}
		subTypes := map[string]map[string]string{}
		for _, o := range pe.ReadShape.Objects {
			subs, serr := g.src.ModelFields(schemaGen, o.ValueType)
			if serr != nil {
				return entityData{}, fmt.Errorf("%s read_shape object %s: %w", name, o.TF, serr)
			}
			subTypes[o.ValueType] = map[string]string{}
			for _, sf := range subs {
				subTypes[o.ValueType][sf.TFSDK] = sf.Type
			}
		}
		if err := checkReadShapeKinds(*pe.ReadShape, byTF, subTypes); err != nil {
			return entityData{}, fmt.Errorf("%s: %w", name, err)
		}
		d.RawRead = &rawReadData{Method: readMethod, Path: pe.Read.Path, IDGo: rm.IDField.GoName, Nested: true, WireStructGo: wireGo, FlattenGo: flattenGo,
			UsesAttr: len(pe.ReadShape.Objects) > 0 || pe.ReadShape.Explode != nil, MissingViaList: pe.MissingViaList}
	} else {
		fields, idGo, ferr := rawModelFields(model, pe.ReadKinds)
		if ferr != nil {
			return entityData{}, fmt.Errorf("%s raw read: %w", name, ferr)
		}
		d.RawRead = &rawReadData{Method: readMethod, Path: pe.Read.Path, IDGo: idGo, Fields: fields, MissingViaList: pe.MissingViaList}
	}

	if isRawOp(pe.Update) {
		m, err := rawVerb(pe.Update.Method)
		if err != nil {
			return entityData{}, fmt.Errorf("%s raw update: %w", name, err)
		}
		d.RawUpdate = &rawWriteData{Method: m, Path: pe.Update.Path, HasBody: true}
		switch {
		case len(pe.WriteShape) > 0:
			if d.RawUpdate.MergeGo, err = buildMergeFields(pe.WriteShape, byTF); err != nil {
				return entityData{}, fmt.Errorf("%s: %w", name, err)
			}
		case pe.ReadShape != nil:
			// A nested read has no flat wire struct to marshal the model into.
			return entityData{}, fmt.Errorf("%s: a raw update beside a read_shape needs a write_shape", name)
		}
	} else if len(pe.WriteShape) > 0 {
		return entityData{}, fmt.Errorf("%s: write_shape declared but the update is not raw", name)
	}
	if isRawOp(pe.Delete) {
		m, err := rawVerb(pe.Delete.Method)
		if err != nil {
			return entityData{}, fmt.Errorf("%s raw delete: %w", name, err)
		}
		d.RawDelete = &rawWriteData{Method: m, Path: pe.Delete.Path, HasBody: false}
	}

	return d, nil
}

// blendedDataSourceCtor names the data-source constructor a blended resource's
// service package must register.
//
// A blended resource may have either kind: a hand-registered companion, or one
// the data-source generator derives from a public list/read op. Only the
// companion case was handled, so a derived data source was generated and
// compiled but never registered -- kion_billing_rule's disappeared from the
// provider the moment it became blended, and its acceptance test failed with
// "Invalid data source" rather than anything pointing at the cause.
func blendedDataSourceCtor(dir, name, pascal string) string {
	if ctor := dataSourceCompanionCtor(name, pascal); ctor != "" {
		return ctor
	}
	// A public read op is not enough: the data-source generator declines some
	// (no usable list shape), and registering a constructor for a file that was
	// never written does not compile. The file on disk is the only honest
	// signal -- provided the caller has already emitted companions and pruned
	// any scaffolder stub, which is neither generated nor wanted here.
	if _, err := os.Stat(filepath.Join(dir, name+"_data_source.go")); err == nil {
		return "New" + pascal + "DataSource"
	}
	return ""
}
