package crud

import (
	_ "embed"
	"fmt"
	"path/filepath"
)

// Bespoke archetypes cover resources whose CRUD control flow fits none of the
// derived archetypes (entity/association/parent_list/…): a settings singleton,
// the cloud-account convert/move state machine, and the entity-polymorphic
// custom-variable override. Their resource + data source bodies are emitted
// verbatim from per-resource templates so a wipe + regen reproduces them, while
// their `*_schema_gen.go` / `*_version_gen.go` still come from the schema and
// version stages like every other resource. service_package.go is owned by
// kgen service (it registers the hand-kept data source alongside the resource),
// so the bespoke paths do not emit it.

//go:embed singleton.gtpl
var singletonResourceTmpl string

//go:embed singleton_ds.gtpl
var singletonDataSourceTmpl string

//go:embed cvoverride.gtpl
var cvOverrideResourceTmpl string

//go:embed cvoverride_ds.gtpl
var cvOverrideDataSourceTmpl string

//go:embed cloud_account.gtpl
var cloudAccountResourceTmpl string

//go:embed cloud_account_create_test.gtpl
var cloudAccountCreateTestTmpl string

//go:embed automation_policy.gtpl
var automationPolicyResourceTmpl string

//go:embed automation_policy_ds.gtpl
var automationPolicyDataSourceTmpl string

//go:embed payer.gtpl
var payerResourceTmpl string

//go:embed cloud_account_ds_account.gtpl
var cloudAccountDSAccountTmpl string

//go:embed cloud_account_ds_aws.gtpl
var cloudAccountDSAwsTmpl string

// accounthelper is a shared support package (convert/move/read/delete/status)
// imported by the cloud_account resources. It is emitted verbatim alongside them
// so a wipe of internal/service/* reproduces it.
var (
	//go:embed accounthelper/accounthelper.gtpl
	accountHelperMainTmpl string
	//go:embed accounthelper/convert.gtpl
	accountHelperConvertTmpl string
	//go:embed accounthelper/delete.gtpl
	accountHelperDeleteTmpl string
	//go:embed accounthelper/read.gtpl
	accountHelperReadTmpl string
	//go:embed accounthelper/status.gtpl
	accountHelperStatusTmpl string
)

// Hand-kept companion files (data sources for resource-only archetypes, their
// helpers, and alias registrations) emitted verbatim next to a generated
// resource. The data-source-only gcp_regions package is handled separately.
var (
	//go:embed gcp_regions_ds.gtpl
	gcpRegionsDSTmpl string
	//go:embed ou_permission_mapping_ds.gtpl
	ouPermissionMappingDSTmpl string
	//go:embed ou_permission_mapping_helpers.gtpl
	ouPermissionMappingHelpersTmpl string
	//go:embed project_permission_mapping_ds.gtpl
	projectPermissionMappingDSTmpl string
	//go:embed project_permission_mapping_helpers.gtpl
	projectPermissionMappingHelpersTmpl string
	//go:embed funding_source_permission_mapping_ds.gtpl
	fundingSourcePermissionMappingDSTmpl string
	//go:embed funding_source_permission_mapping_helpers.gtpl
	fundingSourcePermissionMappingHelpersTmpl string
	//go:embed global_permission_mapping_ds.gtpl
	globalPermissionMappingDSTmpl string
	//go:embed global_permission_mapping_helpers.gtpl
	globalPermissionMappingHelpersTmpl string
	//go:embed project_enforcement_ds.gtpl
	projectEnforcementDSTmpl string
	//go:embed project_note_ds.gtpl
	projectNoteDSTmpl string
	//go:embed funding_source_read.gtpl
	fundingSourceReadTmpl string
	//go:embed cft_alias.gtpl
	cftAliasTmpl string
	//go:embed iam_policy_alias.gtpl
	iamPolicyAliasTmpl string
	//go:embed cached_account_alias.gtpl
	cachedAccountAliasTmpl string
)

// companionsByName maps a generated resource to the hand-authored companion
// files kept beside it (its archetype does not derive them). emitCompanions
// writes these verbatim after the resource is generated, so a wipe reproduces
// them. The map is populated in an init to reference the embedded templates.
var companionsByName map[string][]bespokeFile

func init() {
	companionsByName = map[string][]bespokeFile{
		"ou_permission_mapping": {
			{ouPermissionMappingDSTmpl, "ou_permission_mapping_data_source.go"},
			{ouPermissionMappingHelpersTmpl, "data_source_helpers.go"},
		},
		"project_permission_mapping": {
			{projectPermissionMappingDSTmpl, "project_permission_mapping_data_source.go"},
			{projectPermissionMappingHelpersTmpl, "data_source_helpers.go"},
		},
		"funding_source_permission_mapping": {
			{fundingSourcePermissionMappingDSTmpl, "funding_source_permission_mapping_data_source.go"},
			{fundingSourcePermissionMappingHelpersTmpl, "data_source_helpers.go"},
		},
		"global_permission_mapping": {
			{globalPermissionMappingDSTmpl, "global_permission_mapping_data_source.go"},
			{globalPermissionMappingHelpersTmpl, "data_source_helpers.go"},
		},
		"project_enforcement": {
			{projectEnforcementDSTmpl, "project_enforcement_data_source.go"},
		},
		// The read_companion the funding_source archetype names; see #68.
		"funding_source": {
			{fundingSourceReadTmpl, "funding_source_read.go"},
		},
		"project_note": {
			{projectNoteDSTmpl, "project_note_data_source.go"},
		},
		"cft": {
			{cftAliasTmpl, "aws_cloudformation_template_alias.go"},
		},
		"iam_policy": {
			{iamPolicyAliasTmpl, "aws_iam_policy_alias.go"},
		},
		"account_cache": {
			{cachedAccountAliasTmpl, "cached_account_alias.go"},
		},
	}
}

const (
	// singletonKind is a PATCH-only settings singleton (app_config): no create/
	// delete endpoint, a single record patched in place, static id, all
	// attributes Optional+Computed.
	singletonKind = "singleton"

	// cloudAccountKind is the account/aws_account convert-and-move state machine
	// (create-new-in-cache, cache<->project convert, project->project move),
	// backed by the shared accounthelper package.
	cloudAccountKind = "cloud_account"

	// cvOverrideKind is the entity-polymorphic custom_variable_override
	// (account/ou/project parent, jx.Raw value polymorphism).
	cvOverrideKind = "cv_override"

	// datasourceOnlyKind is a package with only a hand-kept data source and no
	// resource (gcp_regions). Its data source body is emitted verbatim; its
	// schema still comes from the schema stage.
	datasourceOnlyKind = "datasource_only"

	// automationPolicyKind is the private /v1 automation policy. raw_http cannot
	// express it: that archetype supports scalar attributes only, and this body
	// is an envelope whose siblings include a nested object (scheduled_frequency)
	// and a list of objects (cloud_provider_policies) that itself carries two
	// string lists, each an array of one-key objects on the wire.
	automationPolicyKind = "automation_policy"

	// payerKind is a billing source type with no public surface (#53): created
	// through the shared private POST /v1/payer/standalone, its fields nested
	// under a per-type key, read and updated (read-modify-write) at
	// /v1/payer/{id}. No derived archetype has a private create.
	payerKind = "payer"
)

// datasourceOnlyTemplates maps a datasource_only package to its verbatim data
// source template.
var datasourceOnlyTemplates = map[string]string{
	"gcp_regions": gcpRegionsDSTmpl,
}

//go:embed project.gtpl
var projectResourceTmpl string

// resourceTemplates maps the `resource_template` name on an entity archetype to
// the verbatim body that replaces the derived <name>.go. Unlike a bespoke kind
// this keeps the resource in generator_config's `resources` list, so its schema,
// data source, sweeper and acceptance tests are still derived; only the CRUD
// body is hand-authored. See archetype.ResourceTemplate.
var resourceTemplates = map[string]string{
	"project": projectResourceTmpl,
}

// isBespokeKind reports whether a kind is emitted from verbatim per-resource
// templates rather than derived from the SDK op-set. Bespoke resources may be
// declared only in crud_archetypes.yaml (absent from generator_config
// `resources`), so the generate loop merges them in explicitly.
func isBespokeKind(kind string) bool {
	switch kind {
	case singletonKind, cloudAccountKind, cvOverrideKind, datasourceOnlyKind, automationPolicyKind, payerKind:
		return true
	default:
		return false
	}
}

func (g *generator) generateSingleton(dir, name string, force bool) (int, error) {
	n, err := g.emitBespoke(dir, name, nil, []bespokeFile{
		{singletonResourceTmpl, name + ".go"},
		{singletonDataSourceTmpl, name + "_data_source.go"},
	}, force)
	if err != nil {
		return n, err
	}
	return n, g.emitCompanions(dir, name, force)
}

func (g *generator) generateCVOverride(dir, name string, force bool) (int, error) {
	n, err := g.emitBespoke(dir, name, nil, []bespokeFile{
		{cvOverrideResourceTmpl, name + ".go"},
		{cvOverrideDataSourceTmpl, name + "_data_source.go"},
	}, force)
	if err != nil {
		return n, err
	}
	return n, g.emitCompanions(dir, name, force)
}

// generateAutomationPolicy emits the resource and its data source. The data
// source pages the index itself rather than deriving from the list templates,
// which all speak to the SDK; this endpoint is private and serves ten records
// unless told otherwise.
func (g *generator) generateAutomationPolicy(dir, name string, force bool) (int, error) {
	n, err := g.emitBespoke(dir, name, nil, []bespokeFile{
		{automationPolicyResourceTmpl, name + ".go"},
		{automationPolicyDataSourceTmpl, name + "_data_source.go"},
	}, force)
	if err != nil {
		return n, err
	}
	return n, g.emitCompanions(dir, name, force)
}

// cloudAccountData parameterizes the shared account/aws_account resource
// template. Recv is the lowerCamel receiver base ("account"/"awsAccount"),
// Pascal the exported base ("Account"/"AwsAccount"), ResName the human name.
type cloudAccountData struct {
	Pkg, Recv, Pascal, ResName string
	SchemaVersion              int
	SDKAlias                   string
	SDKImport                  string
	ResConst                   string
	Labels                     *labelSyncBind
	LabelsRespType             string
}

// generateCloudAccount emits the shared convert-and-move resource plus the
// resource-specific data source (account = list+filter, aws_account = by-id).
func (g *generator) generateCloudAccount(dir, name string, force bool) (int, error) {
	pascal := pascalCase(name)
	data := cloudAccountData{
		Pkg:       name,
		Recv:      lowerCamelCase(name),
		Pascal:    pascal,
		ResName:   spaceBeforeCaps(pascal),
		SDKAlias:  "generated",
		SDKImport: sdkImport(),
		ResConst:  "ResName" + pascal,
	}
	if mc, ok := g.memberships[name]; ok && mc.Labels != nil {
		lb, err := resolveLabelSync(*mc.Labels, nil)
		if err != nil {
			return 0, fmt.Errorf("%s labels: %w", name, err)
		}
		data.Labels = lb
		data.LabelsRespType = labelsRespType(lb)
	}
	if _, ok := g.upgrades["kion_"+name]; ok {
		data.SchemaVersion = 1
	}
	dsTmpl := cloudAccountDSAwsTmpl
	if name == "account" {
		dsTmpl = cloudAccountDSAccountTmpl
	}
	if _, err := g.emitBespoke(dir, name, data, []bespokeFile{
		{cloudAccountResourceTmpl, name + ".go"},
		{dsTmpl, name + "_data_source.go"},
		{cloudAccountCreateTestTmpl, name + "_create_test.go"},
	}, force); err != nil {
		return 0, err
	}
	if err := g.emitCompanions(dir, name, force); err != nil {
		return 0, err
	}
	// Emit the shared accounthelper package once (when generating account).
	if name == "account" {
		if err := g.emitAccountHelper(filepath.Join(filepath.Dir(dir), "accounthelper"), force); err != nil {
			return 0, err
		}
	}
	return 1, nil
}

// emitAccountHelper writes the shared accounthelper support package verbatim.
func (g *generator) emitAccountHelper(dir string, force bool) error {
	_, err := g.emitBespoke(dir, "accounthelper", nil, []bespokeFile{
		{accountHelperMainTmpl, "accounthelper.go"},
		{accountHelperConvertTmpl, "convert.go"},
		{accountHelperDeleteTmpl, "delete.go"},
		{accountHelperReadTmpl, "read.go"},
		{accountHelperStatusTmpl, "status.go"},
	}, force)
	return err
}

// generateDatasourceOnly emits a package that has only a hand-kept data source
// (gcp_regions) verbatim.
func (g *generator) generateDatasourceOnly(dir, name string, force bool) (int, error) {
	tmpl, ok := datasourceOnlyTemplates[name]
	if !ok {
		return 0, fmt.Errorf("%s: datasource_only archetype but no template registered", name)
	}
	pascal := pascalCase(name)
	pkgGo, err := execGoTemplate("servicepackage", servicePackageTmpl,
		newDataSourceOnlyPackageData(name, pascal), "service_package.go")
	if err != nil {
		return 0, err
	}
	n, err := g.emitBespoke(dir, name, nil, []bespokeFile{
		{tmpl, name + "_data_source.go"},
	}, force)
	if err != nil {
		return n, err
	}
	if err := g.writeFile(filepath.Join(dir, "service_package.go"), pkgGo, force); err != nil {
		return n, err
	}
	return n, g.emitCompanions(dir, name, force)
}

// emitCompanions writes the hand-authored companion files kept beside a
// generated resource (data source, its helpers, alias registrations), if any
// are registered for name.
func (g *generator) emitCompanions(dir, name string, force bool) error {
	files := append([]bespokeFile(nil), companionsByName[name]...)
	files = append(files, companionBlendedByName[name]...)
	files = append(files, companionTestsByName[name]...)
	if len(files) == 0 {
		return nil
	}
	_, err := g.emitBespoke(dir, name, nil, files, force)
	return err
}

// bespokeFile pairs an embedded template with its output file name.
type bespokeFile struct {
	tmpl    string
	outName string
}

// companionData feeds companion templates that take no per-resource data. The
// empty EnvArgs keeps `range .EnvArgs` in the data-source tests a no-op.
type companionData struct {
	SDKImport string
	EnvArgs   []struct{ Param, Env, Describes string }
	idx       *sdkIndex
}

// sdkOpRef is an SDK operation a companion template calls, resolved by route.
type sdkOpRef struct {
	Name, Params, Resp string
	// DataPtr: the response's Data is *T (nil-checked) rather than an OptT (.Set).
	DataPtr bool
}

// Op resolves "METHOD /path" in this run's SDK, so a companion template names a
// route rather than an ogen method name, which a version bump can reassign.
func (c companionData) Op(route string) (sdkOpRef, error) {
	if c.idx == nil {
		return sdkOpRef{}, fmt.Errorf("op %s: no SDK index for this run", route)
	}
	m, ok := c.idx.methods[route]
	if !ok {
		return sdkOpRef{}, fmt.Errorf("op %s: not in this SDK", route)
	}
	resp, _, dataPtr, _, err := resolveEnvelope(m.ResultType, *c.idx)
	if err != nil {
		return sdkOpRef{}, fmt.Errorf("op %s: %w", route, err)
	}
	return sdkOpRef{Name: m.Name, Params: m.ParamsType, Resp: resp, DataPtr: dataPtr}, nil
}

// emitBespoke renders each template with data and writes it into dir.
func (g *generator) emitBespoke(dir, name string, data any, files []bespokeFile, force bool) (int, error) {
	if data == nil {
		data = companionData{SDKImport: sdkImport(), idx: g.idx}
	}
	for _, f := range files {
		out, err := execGoTemplate(name+":"+f.outName, f.tmpl, data, f.outName)
		if err != nil {
			return 0, err
		}
		if err := g.writeFile(filepath.Join(dir, f.outName), out, force); err != nil {
			return 0, err
		}
	}
	return 1, nil
}
