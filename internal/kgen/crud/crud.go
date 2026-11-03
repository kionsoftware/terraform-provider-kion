// Package crud generates CRUD methods, data sources, acceptance tests, and
// sweepers for entity-archetype service packages by AST-introspecting the
// ogen SDK. It imports neither the SDK nor internal/service, so it runs
// mid-development even when the provider does not compile.
package crud

import (
	"bytes"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"terraform-provider-kion/internal/kgen/kfs"
	"terraform-provider-kion/internal/kgen/migrate"
)

// Options configures a crud generation run.
type Options struct {
	ProjectRoot     string // resolved via findProjectRoot() when empty
	Config          string // codegen/generator_config.yaml
	ConfigOverrides string // codegen/config_overrides.yaml (reserved; the merged config is read directly)
	SDKDir          string // ../kion-sdk-go
	SDKVersion      string // the generated/<version> sub-package, e.g. "v3_NN"
	CrudOverrides   string // codegen/crud_overrides.yaml
	Archetypes      string // codegen/crud_archetypes.yaml (compound-key/parent-read declarations)
	TestValues      string // codegen/test_values.yaml
	VersionSupport  string // codegen/version_support.yaml
	Force           bool
	OnlyResource    string // "" = all
	// Strict fails the run when any data source was downgraded from
	// list+filter to id-only. Off by default because a handful of resources
	// legitimately have no callable collection endpoint; on in a gate that wants
	// the set frozen.
	Strict bool
}

// downgrade records a data source that fell back from the list+filter shape to
// id-only. Losing the `filter` block is a silent, config-breaking regression for
// anyone migrating from the SDKv2 provider, so these are reported loudly and in
// aggregate at the end of a run rather than as a lone line mid-stream.
type downgrade struct {
	Resource string
	Reason   string
}

// droppedField records a field a dual-mode data source built without. It is not
// an error -- a slice or nested object has no place in a `list` object -- but it
// narrows what a `filter` block can match, so the set is printed rather than
// absorbed by the bare `continue` that used to swallow it.
type droppedField struct {
	Resource string
	Field    string
	Reason   string
}

type generator struct {
	fs          kfs.FS
	dsGated     map[string]bool // data sources gated in version_support.yaml
	src         Source
	privEnds    map[string]rawResourceOps    // codegen/private_endpoints.yaml (raw_http archetype)
	memberships map[string]membershipConfig  // codegen/memberships.yaml (owner sync on update)
	upgrades    map[string]migrate.Transform // codegen/state_upgrades.yaml (state upgraders)
	oldSchema   map[string]migrate.Resource  // codegen/schema_snapshots/old.json (lazy)
	newSchema   map[string]migrate.Resource  // codegen/schema_snapshots/new.json (lazy)
	root        string
	// written records every path this pass wrote, so pruneUnwritten can tell a
	// generated file from one the scaffolder left behind.
	written map[string]bool

	// fieldPolicy is codegen/unexposed_fields.yaml: the response fields a list
	// data source must withhold. Loaded once per run alongside the other
	// codegen inputs above.
	fieldPolicy FieldPolicy

	// configValidators is codegen/config_validators.yaml: cross-field API
	// constraints an attribute-level schema cannot express.
	configValidators ConfigValidatorPolicy

	// renames is codegen/renames.yaml, resource -> {api_attribute: tf_attribute}.
	// The schemas generator applies it; crud needs it to bind, because a request
	// body field is matched to a model attribute by the API's own name.
	renames    map[string]map[string]string
	downgrades []downgrade    // data sources that lost their filter block this run
	dropped    []droppedField // fields a list data source could not project this run
	// dataSources is the generator_config `data_sources` op-set, needed to reach
	// a resource OTHER than the one being generated: a parent-scoped sweeper
	// enumerates its parent's collection.
	dataSources map[string]dsOps

	// idx is this run's SDK index, which companion templates resolve routes in.
	idx     *sdkIndex
	unswept []downgrade // resources that got no sweeper this run, with the reason
	failed  []string    // per-resource generation failures this run
}

// recordFailure notes a resource whose generation failed; failureError turns
// the set into the run's error so a failed resource cannot exit 0.
func (g *generator) recordFailure(name string, err error) {
	fmt.Fprintf(os.Stderr, "kgen crud: FAILED %s: %v\n", name, err)
	g.failed = append(g.failed, fmt.Sprintf("%s: %v", name, err))
}

// errExists marks a file left alone because --force was not given.
var errExists = errors.New("already exists (use --force to overwrite)")

// recordUnlessExists records err as a failure unless it is the deliberate
// non-force skip, and reports whether it was recorded.
func (g *generator) recordUnlessExists(name string, err error) bool {
	if errors.Is(err, errExists) {
		return false
	}
	g.recordFailure(name, err)
	return true
}

func (g *generator) failureError() error {
	if len(g.failed) == 0 {
		return nil
	}
	return fmt.Errorf("%d resource(s) failed to generate:\n  %s", len(g.failed), strings.Join(g.failed, "\n  "))
}

// Generate is the stage entry point.
func Generate(opts Options) (int, error) {
	g := &generator{fs: kfs.OS{}, src: NewFileSource()}
	return g.generate(opts)
}

func (g *generator) generate(opts Options) (int, error) {
	root := opts.ProjectRoot
	if root == "" {
		r, err := findProjectRoot()
		if err != nil {
			return 0, err
		}
		root = r
	}

	policy, err := LoadFieldPolicy(root)
	if err != nil {
		return 0, err
	}
	g.fieldPolicy = policy

	cv, err := LoadConfigValidators(root)
	if err != nil {
		return 0, err
	}
	g.configValidators = cv

	rn, err := loadRenames(root)
	if err != nil {
		return 0, err
	}
	g.renames = rn

	cfgPath := opts.Config
	if cfgPath == "" {
		cfgPath = filepath.Join("codegen", "generator_config.yaml")
	}
	cfgPath = resolveAgainst(root, cfgPath)

	resources, dataSources, flagged, err := loadConfig(cfgPath)
	if err != nil {
		return 0, err
	}
	g.dataSources = dataSources

	sdkDir := opts.SDKDir
	if sdkDir == "" {
		sdkDir = "../kion-sdk-go"
	}
	sdkDir = resolveAgainst(root, sdkDir)

	sdkVersion = opts.SDKVersion
	idx, err := buildIndex(g.src, sdkDir, opts.SDKVersion)
	if err != nil {
		return 0, err
	}
	g.idx = &idx

	tvPath := opts.TestValues
	if tvPath == "" {
		tvPath = filepath.Join("codegen", "test_values.yaml")
	}
	tvPath = resolveAgainst(root, tvPath)

	vsPath := opts.VersionSupport
	if vsPath == "" {
		vsPath = filepath.Join("codegen", "version_support.yaml")
	}
	vs, err := loadVersionSupport(resolveAgainst(root, vsPath))
	if err != nil {
		return 0, err
	}

	archPath := opts.Archetypes
	if archPath == "" {
		archPath = filepath.Join("codegen", "crud_archetypes.yaml")
	}
	archetypes, err := loadArchetypes(resolveAgainst(root, archPath))
	if err != nil {
		return 0, err
	}

	g.privEnds, err = loadPrivateEndpoints(resolveAgainst(root, filepath.Join("codegen", "private_endpoints.yaml")))
	if err != nil {
		return 0, err
	}

	g.memberships, err = loadMemberships(resolveAgainst(root, filepath.Join("codegen", "memberships.yaml")))
	if err != nil {
		return 0, err
	}

	// State upgraders (old SDKv2 → new schema). Optional file; absent = no migration.
	g.root = root
	if ups, uerr := migrate.LoadUpgrades(resolveAgainst(root, filepath.Join("codegen", "state_upgrades.yaml"))); uerr == nil {
		g.upgrades = ups
	}

	// Process every configured resource plus any bespoke-archetype resource that
	// is declared only in crud_archetypes.yaml (bespoke resources carry their
	// schema inline in a template, so they are intentionally absent from the
	// generator_config `resources` list that drives the schema stage).
	g.dsGated = map[string]bool{}
	names := map[string]bool{}
	for name := range resources {
		names[name] = true
	}
	for name, a := range archetypes {
		if isBespokeKind(a.Kind) {
			names[name] = true
		}
	}

	written := 0
	for _, name := range sortedKeys(names) {
		if opts.OnlyResource != "" && name != opts.OnlyResource {
			continue
		}
		if flagged[name] {
			fmt.Fprintf(os.Stderr, "kgen crud: skip %s, flagged for review in generator_config.yaml\n", name)
			continue
		}
		var arch *archetype
		if a, ok := archetypes[name]; ok {
			arch = &a
		}
		dsGated, err := vs.dataSourceGated(name)
		if err != nil {
			g.recordUnlessExists(name, err)
			continue
		}
		g.dsGated[name] = dsGated
		n, err := g.generateResource(root, name, resources[name], dataSources[name], idx, tvPath, vs.resourceGated(name), arch, opts.Force)
		if err != nil {
			g.recordUnlessExists(name, err)
			continue
		}
		if err := g.checkDataSourceGate(root, name); err != nil {
			g.recordUnlessExists(name, err)
			continue
		}
		// Emit the state upgrader for any resource that migrates old SDKv2 state.
		if err := g.emitUpgrade(name, opts.Force); err != nil {
			g.recordUnlessExists(name+" upgrader", err)
		}
		written += n
	}
	g.reportUnswept()
	g.reportDropped()
	dErr := g.reportDowngrades(opts.Strict)
	if fErr := g.failureError(); fErr != nil {
		return written, fErr
	}
	return written, dErr
}

// reportUnswept prints the aggregated no-sweeper report. Each entry is a
// resource whose orphaned test-acc records `make sweep` will NOT clean up. It is
// not an error: several are structurally unsweepable (no delete endpoint at
// all). It is printed so the set is visible rather than silently assumed empty.
func (g *generator) reportUnswept() {
	if len(g.unswept) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "\n-- kgen crud: %d resource(s) got NO sweeper; `make sweep` leaves their\n", len(g.unswept))
	fmt.Fprintf(os.Stderr, "-- orphaned test-acc records behind. Each sweep.go says the same.\n")
	for _, u := range g.unswept {
		fmt.Fprintf(os.Stderr, "--   kion_%s: %s\n", u.Resource, u.Reason)
	}
	fmt.Fprint(os.Stderr, "\n")
}

// reportDropped prints the fields no `filter` block can match because the
// dual-mode build could not project them into the `list` objects.
func (g *generator) reportDropped() {
	if len(g.dropped) == 0 {
		return
	}
	fmt.Fprintf(os.Stderr, "\n-- kgen crud: %d field(s) dropped from a data source's `list` objects;\n", len(g.dropped))
	fmt.Fprintf(os.Stderr, "-- no `filter` block can match on them.\n")
	for _, d := range g.dropped {
		fmt.Fprintf(os.Stderr, "--   kion_%s.%s: %s\n", d.Resource, d.Field, d.Reason)
	}
	fmt.Fprint(os.Stderr, "\n")
}

// reportDowngrades prints the aggregated downgrade report. Every entry means a
// data source that SHOULD accept `filter` blocks (the config maps a collection
// read for it) instead demands a required `id`, which silently breaks
// configurations carried over from the SDKv2 provider. Under strict it is an
// error, so the set cannot grow unnoticed.
func (g *generator) reportDowngrades(strict bool) error {
	if len(g.downgrades) == 0 {
		return nil
	}
	fmt.Fprintf(os.Stderr, "\n!! kgen crud: %d DATA SOURCE(S) DOWNGRADED to id-only. Each lost its `filter` block\n", len(g.downgrades))
	fmt.Fprintf(os.Stderr, "!! and now REQUIRES `id`, which breaks configurations written for the old provider.\n")
	for _, d := range g.downgrades {
		fmt.Fprintf(os.Stderr, "!!   kion_%s: %s\n", d.Resource, d.Reason)
	}
	fmt.Fprintf(os.Stderr, "!! Fix the generator (internal/kgen/crud) or accept each case deliberately.\n\n")
	if strict {
		return fmt.Errorf("%d data source(s) downgraded to id-only (--strict)", len(g.downgrades))
	}
	return nil
}

// generateResource resolves one resource and writes its generated files:
// <name>.go, <name>_data_source.go, and (when test values exist) the
// acceptance-test files.
func (g *generator) generateResource(root, name string, ops resOps, ds dsOps, idx sdkIndex, tvPath string, gated bool, arch *archetype, force bool) (int, error) {
	pascal := pascalCase(name)
	dir := filepath.Join(root, "internal", "service", name)

	// Bespoke archetypes emit verbatim resource/data-source bodies and need
	// neither the schema_gen model nor the resolved op-set, so dispatch before
	// reading either (these resources have no *Model in a schema_gen file).
	if arch != nil {
		switch arch.Kind {
		case singletonKind:
			return g.generateSingleton(dir, name, force)
		case cvOverrideKind:
			return g.generateCVOverride(dir, name, force)
		case cloudAccountKind:
			return g.generateCloudAccount(dir, name, force)
		case datasourceOnlyKind:
			return g.generateDatasourceOnly(dir, name, force)
		case automationPolicyKind:
			return g.generateAutomationPolicy(dir, name, force)
		case payerKind:
			return g.generatePayer(dir, name, tvPath, gated, force)
		}
	}

	model, err := g.src.ModelFields(filepath.Join(dir, name+"_schema_gen.go"), pascal+"Model")
	if err != nil {
		return 0, err
	}

	var entityArch *archetype
	if arch != nil && arch.Kind != entityKind && len(arch.CreateOnly) > 0 {
		return 0, fmt.Errorf("%s: create_only is supported only by kind: entity, not %s", name, arch.Kind)
	}
	if arch != nil && arch.Kind != entityKind && arch.Kind != rawKind {
		if where := stringIDParam(ops, idx); where != "" {
			return 0, fmt.Errorf("%s: the %s archetype does not support a string id (%s); only entity does", name, arch.Kind, where)
		}
	}
	if arch != nil {
		switch arch.Kind {
		case noReadKind:
			return g.generateNoRead(dir, name, ops, idx, model, tvPath, gated, force)
		case assocKind:
			return g.generateAssoc(dir, name, ops, idx, *arch, model, tvPath, gated, force)
		case rawKind:
			pe, ok := g.privEnds[name]
			if !ok {
				return 0, fmt.Errorf("%s: raw_http archetype but no codegen/private_endpoints.yaml entry", name)
			}
			return g.generateRaw(dir, name, pe, model, tvPath, gated, force)
		case blendedKind:
			pe, ok := g.privEnds[name]
			if !ok {
				return 0, fmt.Errorf("%s: blended archetype but no codegen/private_endpoints.yaml entry", name)
			}
			return g.generateBlended(dir, name, ops, idx, pe, model, tvPath, gated, force)
		case parentListKind:
			return g.generateParentList(dir, name, ops, idx, *arch, model, tvPath, gated, force)
		case entityKind:
			entityArch = arch // normal entity path, with archetype tweaks applied below
		default:
			return g.generateCompound(dir, name, ops, idx, *arch, model, gated, force)
		}
	}

	pe, hasPriv := g.privEnds[name]
	rm, err := resolveResource(name, ops, ds, idx, model, hasPriv && isRawOp(pe.Delete))
	if err != nil {
		return 0, err
	}
	rm.Gated = gated
	rm.DSGated = g.dsGated[name]
	if _, ok := g.upgrades["kion_"+name]; ok {
		rm.SchemaVersion = 1
	}
	if entityArch != nil {
		rm.ReadCompanion = entityArch.ReadCompanion
		if rc := entityArch.RawCreate; rc != nil {
			verb, err := rawVerb(rc.Method)
			if err != nil {
				return 0, fmt.Errorf("%s raw_create: %w", name, err)
			}
			// Nested attributes are declared, not mapped: hide them from
			// rawModelFields so its refusal still covers everything else.
			declared := map[string]bool{}
			var objs []rawCreateObject
			for _, o := range rc.Objects {
				declared[o.TF] = true
				var goType string
				for _, mf := range model {
					if mf.TFSDK == o.TF {
						goType = mf.Type
						break
					}
				}
				if goType == "" {
					return 0, fmt.Errorf("%s raw_create: object %q is not in the model", name, o.TF)
				}
				obj := rawCreateObject{ModelGo: pascalCase(o.TF), JSON: o.TF, GoType: goType}
				for _, s := range o.Subs {
					obj.Subs = append(obj.Subs, rawCreateObjectSub{ModelGo: pascalCase(s), JSON: s})
				}
				objs = append(objs, obj)
			}
			scalars := make([]ModelField, 0, len(model))
			for _, mf := range model {
				if !declared[mf.TFSDK] {
					scalars = append(scalars, mf)
				}
			}
			fields, idGo, err := rawModelFields(scalars, nil)
			if err != nil {
				return 0, fmt.Errorf("%s raw_create: %w", name, err)
			}
			rm.RawCreate = &rawCreateData{Method: verb, Path: rc.Path, Fields: fields, Objects: objs, IDGo: idGo}
		}
		for _, tfName := range entityArch.Rewritten {
			var goName string
			for _, mf := range model {
				if mf.TFSDK == tfName {
					goName = mf.GoName
					break
				}
			}
			if goName == "" {
				fmt.Fprintf(os.Stderr, "kgen crud: %s: rewritten attribute %q is not in the model\n", name, tfName)
				continue
			}
			rm.Rewritten = append(rm.Rewritten, goName)
		}
		for _, tfName := range entityArch.CreateViaUpdate {
			var goName string
			for _, mf := range model {
				if mf.TFSDK == tfName {
					goName = mf.GoName
					break
				}
			}
			if goName == "" {
				return 0, fmt.Errorf("%s: create_via_update attribute %q is not in the model", name, tfName)
			}
			rm.CreateViaUpdate = append(rm.CreateViaUpdate, goName)
		}
		if rm.CreateOnly, err = resolveCreateOnly(name, *entityArch, model); err != nil {
			return 0, err
		}
		rm.DeleteRecordParam = entityArch.DeleteRecordParam
		rm.DeleteExtraParam = entityArch.DeleteExtraParam
		rm.DeleteExtraField = entityArch.DeleteExtraField
		if entityArch.SweepParent != "" {
			if err := g.bindSweepParent(&rm, ds, idx, *entityArch); err != nil {
				fmt.Fprintf(os.Stderr, "kgen crud: %s: %v\n", name, err)
			}
		}
	}

	// Resolve nested-object body/response fields (needs the schema_gen Value types).
	schemaGen := filepath.Join(dir, name+"_schema_gen.go")
	byTF := map[string]ModelField{}
	for _, mf := range model {
		byTF[mf.TFSDK] = mf
	}

	// A model that nests the whole record under one object attribute exposes none
	// of the record's fields at the top level; promote them so the data source and
	// sweeper can still project the record.
	rm.RecordSubFields = recordSubFields(g.src, schemaGen, rm.Read.RespWrapperGo, byTF)

	// Create-under-parent: the create op's id param is the parent id from a model
	// attribute (not a literal discriminator). Overrides the literal create-param
	// path resolved above.
	if entityArch != nil && entityArch.CreateParentParam != "" {
		mf, ok := byTF[entityArch.CreateParentField]
		if !ok {
			return 0, fmt.Errorf("%s: create_parent_field %q not in model", name, entityArch.CreateParentField)
		}
		rm.CreateParentParam = entityArch.CreateParentParam
		rm.CreateParentField = mf.GoName
		rm.ConfirmListed = entityArch.ConfirmListed
		rm.IDIsParent = entityArch.IDIsParent
		rm.ConfirmListedTF = entityArch.CreateParentField
		rm.ConfirmParents = entityArch.ConfirmParents
		if rm.Create.Params != nil {
			for _, f := range rm.Create.Params.Fields {
				if f.GoName == entityArch.CreateParentParam && f.Type != "int64" && f.Type != "" {
					rm.CreateParentCast = f.Type
				}
			}
		}
	}
	if rm.CreateNested, err = resolveNested(g.src, schemaGen, rm.Create.Body, byTF, idx, nestedOpts{}); err != nil {
		return 0, fmt.Errorf("%s create nested: %w", name, err)
	}
	if rm.Update != nil {
		if rm.UpdateNested, err = resolveNested(g.src, schemaGen, rm.Update.Body, byTF, idx, updateIDOpts(rm)); err != nil {
			return 0, fmt.Errorf("%s update nested: %w", name, err)
		}
	}
	readPrefix := "v.Data.Value."
	if rm.Read.RespDataPtr {
		readPrefix = "v.Data."
	}
	if rm.ReadNested, err = resolveNestedFlatten(g.src, schemaGen, rm.Read.RespFields, byTF, idx, readPrefix); err != nil {
		return 0, fmt.Errorf("%s read nested: %w", name, err)
	}
	if entityArch != nil && len(entityArch.SumFrom) > 0 {
		if rm.ReadSums, err = resolveSumFlats(entityArch.SumFrom, rm.Read.RespFields, byTF, idx, readPrefix); err != nil {
			return 0, fmt.Errorf("%s read sums: %w", name, err)
		}
	}

	// Owner association synced on Update via paired add/remove endpoints.
	if mc, ok := g.memberships[name]; ok {
		if mc.Labels != nil {
			if rm.Labels, err = resolveLabelSync(*mc.Labels, byTF); err != nil {
				return 0, fmt.Errorf("%s labels: %w", name, err)
			}
		}
		if mc.Owners != nil {
			if rm.Owners, err = resolveOwnerMembership(*mc.Owners, byTF); err != nil {
				return 0, fmt.Errorf("%s owners: %w", name, err)
			}
		}
		for i := range mc.Associations {
			a, aerr := resolveAssocMembership(mc.Associations[i], byTF, idx)
			if aerr != nil {
				return 0, fmt.Errorf("%s associations: %w", name, aerr)
			}
			rm.Assocs = append(rm.Assocs, a)
		}
		for i := range mc.SliceMembers {
			s, serr := resolveSliceMember(mc.SliceMembers[i], byTF)
			if serr != nil {
				return 0, fmt.Errorf("%s slice_members: %w", name, serr)
			}
			rm.SliceMembers = append(rm.SliceMembers, s)
		}
	}

	rm.AtLeastOneOf = g.configValidators.For(name)
	rm.RequiredWhen = g.configValidators.RequiredWhenFor(name)
	rm.RequiredTogether = g.configValidators.RequiredTogetherFor(name)
	if entityArch != nil {
		rm.EmptyCollections = entityArch.EmptyCollections
		rm.ListOrderKeys = entityArch.ListOrderKeys
		rm.MissingViaList = entityArch.MissingViaList
		// An entity whose delete needs a parent the read does not return must
		// carry that parent in its import id.
		if entityArch.DeleteExtraField != "" {
			rm.ImportParentTF = entityArch.DeleteExtraField
		}
	}
	if entityArch != nil {
		for _, m := range entityArch.Moves {
			mf, ok := byTF[m.Attr]
			if !ok {
				return 0, fmt.Errorf("%s: move attr %q not in model", name, m.Attr)
			}
			if (m.RawPost == "") == !m.Account || mf.Type != "types.Int64" {
				return 0, fmt.Errorf("%s: move %q needs exactly one of raw_post/account and an int64 attribute", name, m.Attr)
			}
			rm.Moves = append(rm.Moves, moveBind{ModelGo: mf.GoName, TF: mf.TFSDK, RawPost: m.RawPost, AccountMove: m.Account})
		}
	}
	rm.Renames = g.renames[name]
	// A private delete only applies when the public spec published none; if the
	// SDK has a typed delete the generator must keep using it.
	if rm.Delete == nil {
		if pe, ok := g.privEnds[name]; ok && isRawOp(pe.Delete) {
			rm.RawDeletePath = pe.Delete.Path
		}
	}

	resourceGo, err := renderEntity(rm)
	if err != nil {
		return 0, err
	}
	// A verbatim resource body replaces the derived one; everything below it
	// (data source, sweeper, tests) still comes from the resolved op-set.
	if entityArch != nil && entityArch.ResourceTemplate != "" {
		tmpl, ok := resourceTemplates[entityArch.ResourceTemplate]
		if !ok {
			return 0, fmt.Errorf("%s: resource_template %q is not registered in resourceTemplates", name, entityArch.ResourceTemplate)
		}
		// A verbatim body still needs the cross-field constraints: they are
		// authored per resource, not derived, so a hand-written template would
		// otherwise silently drop the validator the declaration promises.
		tmplData := struct {
			AtLeastOneOf   []string
			RequiredWhen   []RequiredWhen
			SDKAlias       string
			SDKImport      string
			ResConst       string
			Labels         *labelSyncBind
			LabelsRespType string
		}{
			AtLeastOneOf:   rm.AtLeastOneOf,
			RequiredWhen:   rm.RequiredWhen,
			SDKAlias:       "generated",
			SDKImport:      sdkImport(),
			ResConst:       "ResName" + pascalCase(name),
			Labels:         rm.Labels,
			LabelsRespType: labelsRespType(rm.Labels),
		}
		if resourceGo, err = execGoTemplate(name+":"+name+".go", tmpl, tmplData, name+".go"); err != nil {
			return 0, err
		}
	}
	dataSourceGo, dsDowngrade, dsDrops, dsAttrs, err := renderDataSourceAttrs(rm, g.fieldPolicy)
	if err != nil {
		return 0, err
	}
	if dsDowngrade != "" && !companionSupplies(name, name+"_data_source.go") {
		g.downgrades = append(g.downgrades, downgrade{Resource: name, Reason: dsDowngrade})
	}
	for _, d := range dsDrops {
		g.dropped = append(g.dropped, droppedField{Resource: name, Field: d.Field, Reason: d.Reason})
	}
	sweepGo, sweepReason, err := renderSweep(rm)
	if err != nil {
		return 0, err
	}
	if sweepReason != "" {
		g.unswept = append(g.unswept, downgrade{Resource: name, Reason: sweepReason})
	}

	files := []genFile{
		{filepath.Join(dir, name+".go"), resourceGo},
		{filepath.Join(dir, name+"_data_source.go"), dataSourceGo},
		{filepath.Join(dir, "sweep.go"), sweepGo},
	}
	// An entity always ships a data source, so its registration is derivable;
	// the alias registry supplies any second type name the resource answers to.
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newServicePackageData(name, rm.Pascal, "New"+rm.Pascal+"DataSource"), "service_package.go")
	if err != nil {
		return 0, err
	}
	files = append(files, genFile{filepath.Join(dir, "service_package.go"), pkgGo})

	tv, hasTV, err := loadTestValues(tvPath, name)
	if err != nil {
		return 0, err
	}
	if hasTV {
		resTest, err := renderResourceTest(rm, tv)
		if err != nil {
			return 0, err
		}
		dsTest, err := renderDataSourceTest(rm, tv, dsAttrs)
		if err != nil {
			return 0, err
		}
		files = append(files,
			genFile{filepath.Join(dir, name+"_test.go"), resTest},
			genFile{filepath.Join(dir, name+"_data_source_test.go"), dsTest},
		)
	} else {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: no test_values entry; skipping acceptance tests\n", name)
	}

	for _, f := range withoutCompanionFiles(name, files) {
		if err := g.writeFile(f.path, f.data, force); err != nil {
			return 0, err
		}
	}
	// Emit any hand-authored companion files kept beside this resource (e.g. the
	// aws_cloudformation_template / aws_iam_policy alias registrations).
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	return 1, nil
}

// updateIDOpts sources a wrapped update object's id sub-field from the local
// the entity template parses the record id into.
func updateIDOpts(rm ResourceModel) nestedOpts {
	o := nestedOpts{IDAttr: rm.IDField.TFSDK, IDVar: "idInt"}
	if _, t, err := idParamName(rm.Read.Params); err == nil {
		o.IDType = t
	}
	if o.IDType == "string" {
		o.IDVar = "id"
	}
	return o
}

// stringIDParam names the first op whose params carry a string id, or "". The
// archetypes other than entity parse ids as integers, so they must refuse one.
func stringIDParam(ops resOps, idx sdkIndex) string {
	for _, ref := range []*opRef{ops.Create, ops.Read, ops.Update, ops.Delete} {
		if ref == nil {
			continue
		}
		m, ok := idx.methods[ref.Method+" "+ref.Path]
		if !ok {
			continue
		}
		for _, f := range idx.structs[m.ParamsType].Fields {
			if f.Type == "string" && strings.HasSuffix(f.GoName, "ID") {
				return fmt.Sprintf("%s %s param %s", ref.Method, ref.Path, f.GoName)
			}
		}
	}
	return ""
}

// generateCompound writes the files for a compound-key/parent-read resource.
// Acceptance tests are deferred (they need a parent FK fixture), so only the
// resource, data source, and sweeper stub are emitted.
func (g *generator) generateCompound(dir, name string, ops resOps, idx sdkIndex, arch archetype, model []ModelField, gated, force bool) (int, error) {
	if arch.Kind != compoundKindParentRead {
		return 0, fmt.Errorf("%s: unknown archetype kind %q", name, arch.Kind)
	}
	d, err := resolveCompound(name, ops, idx, arch, model, gated)
	if err != nil {
		return 0, err
	}
	d.DSGated = g.dsGated[name]
	resourceGo, err := renderCompoundEntity(d)
	if err != nil {
		return 0, err
	}
	dataSourceGo, err := renderCompoundDataSource(d)
	if err != nil {
		return 0, err
	}
	// A compound-key sub-resource has no independent collection: records are
	// enumerated per parent, which needs the FK-fixture sweeper follow-up.
	sweepReason := "a compound-key sub-resource has no independent collection endpoint"
	sweepGo, err := renderNoSweep(name, d.TypeName, sweepReason)
	if err != nil {
		return 0, err
	}
	g.unswept = append(g.unswept, downgrade{Resource: name, Reason: sweepReason})
	files := []genFile{
		{filepath.Join(dir, name+".go"), resourceGo},
		{filepath.Join(dir, name+"_data_source.go"), dataSourceGo},
		{filepath.Join(dir, "sweep.go"), sweepGo},
	}
	for _, f := range files {
		if err := g.writeFile(f.path, f.data, force); err != nil {
			return 0, err
		}
	}
	// A compound resource derives no test; scope_criteria keeps one as a
	// companion, emitted here so a wipe reproduces it.
	if _, ok := companionTestsByName[name]; !ok {
		fmt.Fprintf(os.Stderr, "kgen crud: %s: compound archetype; no acceptance test derived or registered\n", name)
	}
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	return 1, nil
}

// withoutCompanionFiles drops the derived files a companion replaces. Writing
// both would leave the derived one in place on a non-force wipe, since the
// companion's write then fails as already existing.
func withoutCompanionFiles(name string, files []genFile) []genFile {
	out := files[:0:0]
	for _, f := range files {
		if !companionSupplies(name, filepath.Base(f.path)) {
			out = append(out, f)
		}
	}
	return out
}

// companionSupplies reports whether a companion template writes outName for name.
func companionSupplies(name, outName string) bool {
	for _, m := range []map[string][]bespokeFile{companionsByName, companionBlendedByName, companionTestsByName} {
		for _, f := range m[name] {
			if f.outName == outName {
				return true
			}
		}
	}
	return false
}

// genFile is a rendered file destined for a service package.
type genFile struct {
	path string
	data []byte
}

// emitUpgrade writes <name>_upgrade_gen.go for a resource that migrates old
// SDKv2 state (has a codegen/state_upgrades.yaml entry). No-op otherwise. The
// resource's schema.Version bump is emitted separately by the archetype template
// (gated on ResourceModel.SchemaVersion).
func (g *generator) emitUpgrade(name string, force bool) error {
	tfType := "kion_" + name
	t, ok := g.upgrades[tfType]
	if !ok {
		return nil
	}
	if g.oldSchema == nil {
		old, err := migrate.LoadSchema(resolveAgainst(g.root, filepath.Join("codegen", "schema_snapshots", "old.json")))
		if err != nil {
			return err
		}
		nw, err := migrate.LoadSchema(resolveAgainst(g.root, filepath.Join("codegen", "schema_snapshots", "new.json")))
		if err != nil {
			return err
		}
		g.oldSchema, g.newSchema = old, nw
	}
	// Alias resources carry old state under a different type name; use it for the
	// old-schema (attribute passthrough) lookup.
	oldType := tfType
	if t.OldType != "" {
		oldType = t.OldType
	}
	out, err := migrate.GenerateUpgrade(g.root, tfType, t, g.oldSchema[oldType], g.newSchema[tfType])
	if err != nil {
		return err
	}
	path := filepath.Join(g.root, "internal", "service", name, name+"_upgrade_gen.go")
	if err := g.writeFile(path, out, force); err != nil {
		return err
	}

	// Emit the decode-clean golden test alongside the upgrader.
	testOut, err := migrate.GenerateUpgradeTest(g.root, tfType, t, g.oldSchema[oldType], g.newSchema[tfType])
	if err != nil {
		return err
	}
	testPath := filepath.Join(g.root, "internal", "service", name, name+"_upgrade_gen_test.go")
	return g.writeFile(testPath, testOut, force)
}

// writeFile writes data to path, refusing to overwrite an existing file unless
// force is set. Every written path is recorded so a generate pass can tell what
// it owns from what was left behind.
func (g *generator) writeFile(path string, data []byte, force bool) error {
	if !force {
		if _, err := g.fs.Stat(path); err == nil {
			return fmt.Errorf("%s %w", path, errExists)
		}
	}
	if g.written == nil {
		g.written = map[string]bool{}
	}
	g.written[path] = true
	// The package directory normally exists because `kgen service` scaffolded
	// it, but two are created by generation itself: the shared accounthelper,
	// and any bespoke package absent from generator_config. Without this a wipe
	// fails those with a bare "no such file or directory" naming the file.
	if err := g.fs.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("creating %s: %w", filepath.Dir(path), err)
	}
	return g.fs.WriteFile(path, data, 0o600)
}

// pruneUnwritten removes a file this pass did not write. `kgen service`
// scaffolds a data source for every new package, but a resource-only archetype
// never emits one, so the stub would survive a wipe-and-regenerate and get
// registered in service_package.go -- output no input asked for.
func (g *generator) pruneUnwritten(dir, name string) error {
	path := filepath.Join(dir, name+"_data_source.go")
	if g.written[path] {
		return nil
	}
	if _, err := g.fs.Stat(path); err != nil {
		return nil // nothing there
	}
	return g.fs.RemoveAll(path)
}

// findProjectRoot walks up from the working directory to the module root.
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found walking up from working directory")
		}
		dir = parent
	}
}

func resolveAgainst(root, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(root, p)
}

func sortedKeys[V any](m map[string]V) []string {
	ks := slices.Collect(maps.Keys(m))
	slices.Sort(ks)
	return ks
}

// sdkVersion is the SDK sub-package templates import, set once per run from
// Options.SDKVersion (tests set it directly).
var sdkVersion string

// sdkImport is the import path of the SDK sub-package generated code compiles against.
func sdkImport() string {
	return "github.com/kionsoftware/kion-sdk-go/generated/" + sdkVersion
}

// checkDataSourceGate fails when version_support.yaml gates a data source but
// the generated Read never checks the gate: the template for that archetype
// does not support it, and the floor would go unenforced.
func (g *generator) checkDataSourceGate(root, name string) error {
	if !g.dsGated[name] {
		return nil
	}
	path := filepath.Join(root, "internal", "service", name, name+"_data_source.go")
	b, err := g.fs.ReadFile(path)
	if err != nil {
		return fmt.Errorf("data source is version-gated but %s is unreadable: %w", path, err)
	}
	if !bytes.Contains(b, []byte("RequireKionVersionInRange(d.Meta()")) {
		return fmt.Errorf("data source is version-gated but %s never checks the gate; its template does not support one", path)
	}
	return nil
}
