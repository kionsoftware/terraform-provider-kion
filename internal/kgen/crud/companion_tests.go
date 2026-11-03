package crud

import _ "embed"

//go:embed servicepackage.gtpl
var servicePackageTmpl string

//go:embed companion_account_linkage_data_source_test.gtpl
var accountLinkageDataSourceTestTmpl string

//go:embed companion_account_test.gtpl
var accountTestTmpl string

//go:embed companion_account_data_source_test.gtpl
var accountDataSourceTestTmpl string

//go:embed companion_app_config_test.gtpl
var appConfigTestTmpl string

//go:embed companion_app_config_data_source_test.gtpl
var appConfigDataSourceTestTmpl string

//go:embed companion_automation_policy_test.gtpl
var automationPolicyTestTmpl string

//go:embed companion_automation_policy_data_source_test.gtpl
var automationPolicyDataSourceTestTmpl string

//go:embed companion_aws_account_test.gtpl
var awsAccountTestTmpl string

//go:embed companion_aws_account_data_source_test.gtpl
var awsAccountDataSourceTestTmpl string

//go:embed companion_billing_rule_data_source_test.gtpl
var billingRuleDataSourceTestTmpl string

//go:embed companion_custom_variable_override_test.gtpl
var customVariableOverrideTestTmpl string

//go:embed companion_custom_variable_override_data_source_test.gtpl
var customVariableOverrideDataSourceTestTmpl string

//go:embed companion_gcp_regions_data_source_test.gtpl
var gcpRegionsDataSourceTestTmpl string

//go:embed companion_ou_permission_mapping_data_source_test.gtpl
var ouPermissionMappingDataSourceTestTmpl string

//go:embed companion_project_enforcement_data_source_test.gtpl
var projectEnforcementDataSourceTestTmpl string

//go:embed companion_project_permission_mapping_data_source_test.gtpl
var projectPermissionMappingDataSourceTestTmpl string

//go:embed companion_scope_criteria_test.gtpl
var scopeCriteriaTestTmpl string

//go:embed companion_user_test.gtpl
var userTestTmpl string

//go:embed companion_user_data_source_test.gtpl
var userDataSourceTestTmpl string

//go:embed companion_webhook_test.gtpl
var webhookTestTmpl string

//go:embed companion_webhook_data_source_test.gtpl
var webhookDataSourceTestTmpl string

//go:embed companion_account_linkage_data_source_decode_test.gtpl
var accountLinkageDataSourceDecodeTestTmpl string

//go:embed companion_account_linkage_data_source.gtpl
var accountLinkageDataSourceTmpl string

//go:embed companion_account_linkage_sweep.gtpl
var accountLinkageSweepTmpl string

//go:embed companion_billing_rule_data_source.gtpl
var billingRuleDataSourceTmpl string

//go:embed companion_billing_rule_sweep.gtpl
var billingRuleSweepTmpl string

//go:embed companion_saml_group_association_data_source.gtpl
var samlGroupAssociationDataSourceTmpl string

//go:embed companion_saml_group_association_data_source_test.gtpl
var samlGroupAssociationDataSourceTestTmpl string

//go:embed companion_saml_group_association_data_source_list_test.gtpl
var samlGroupAssociationListTestTmpl string

// companionBlendedByName are the data source and sweeper a blended resource
// ships but its archetype does not derive: it emits only the resource and the
// service package. Without these the committed files claimed "Code generated"
// while nothing regenerated them.
var companionBlendedByName = map[string][]bespokeFile{
	"account_linkage": {
		{accountLinkageDataSourceTmpl, "account_linkage_data_source.go"},
		{accountLinkageSweepTmpl, "sweep.go"},
	},
	"billing_rule": {
		{billingRuleDataSourceTmpl, "billing_rule_data_source.go"},
		{billingRuleSweepTmpl, "sweep.go"},
	},
	// An entity, but Kion lists its records only under their IDMS, which the
	// derived data source cannot do; this keeps the old provider's list shape.
	"saml_group_association": {
		{samlGroupAssociationDataSourceTmpl, "saml_group_association_data_source.go"},
	},
}

//go:embed companion_automation_policy_pagination_test.gtpl
var automationPolicyPaginationTestTmpl string

//go:embed companion_aws_account_upgrade_test.gtpl
var awsAccountUpgradeTestTmpl string

//go:embed companion_cft_upgrade_test.gtpl
var cftUpgradeTestTmpl string

//go:embed companion_aws_iam_policy_alias_upgrade_test.gtpl
var awsIamPolicyAliasUpgradeTestTmpl string

//go:embed companion_move_ou_settings_test.gtpl
var moveOuSettingsTestTmpl string

//go:embed companion_user_group_upgrade_test.gtpl
var userGroupUpgradeTestTmpl string

// companionTestsByName are acceptance tests kept verbatim for resources whose
// archetype derives no test. Registered separately from companionsByName so
// neither map depends on package init order.
var companionTestsByName = map[string][]bespokeFile{
	"user_group": {
		{userGroupUpgradeTestTmpl, "user_group_upgrade_test.go"},
	},
	"project": {
		{moveOuSettingsTestTmpl, "move_ou_settings_test.go"},
	},
	"iam_policy": {
		{awsIamPolicyAliasUpgradeTestTmpl, "aws_iam_policy_alias_upgrade_test.go"},
	},
	"cft": {
		{cftUpgradeTestTmpl, "cft_upgrade_test.go"},
	},
	"account": {
		{accountDataSourceTestTmpl, "account_data_source_test.go"},
		{accountTestTmpl, "account_test.go"},
	},
	"account_linkage": {
		{accountLinkageDataSourceTestTmpl, "account_linkage_data_source_test.go"},
		{accountLinkageDataSourceDecodeTestTmpl, "account_linkage_data_source_decode_test.go"},
	},
	"app_config": {
		{appConfigDataSourceTestTmpl, "app_config_data_source_test.go"},
		{appConfigTestTmpl, "app_config_test.go"},
	},
	"automation_policy": {
		{automationPolicyPaginationTestTmpl, "automation_policy_pagination_test.go"},
		{automationPolicyDataSourceTestTmpl, "automation_policy_data_source_test.go"},
		{automationPolicyTestTmpl, "automation_policy_test.go"},
	},
	"aws_account": {
		{awsAccountUpgradeTestTmpl, "aws_account_upgrade_test.go"},
		{awsAccountDataSourceTestTmpl, "aws_account_data_source_test.go"},
		{awsAccountTestTmpl, "aws_account_test.go"},
	},
	"billing_rule": {
		{billingRuleDataSourceTestTmpl, "billing_rule_data_source_test.go"},
	},
	"saml_group_association": {
		{samlGroupAssociationDataSourceTestTmpl, "saml_group_association_data_source_test.go"},
		{samlGroupAssociationListTestTmpl, "saml_group_association_data_source_list_test.go"},
	},
	"custom_variable_override": {
		{customVariableOverrideDataSourceTestTmpl, "custom_variable_override_data_source_test.go"},
		{customVariableOverrideTestTmpl, "custom_variable_override_test.go"},
	},
	"gcp_regions": {
		{gcpRegionsDataSourceTestTmpl, "gcp_regions_data_source_test.go"},
	},
	"ou_permission_mapping": {
		{ouPermissionMappingDataSourceTestTmpl, "ou_permission_mapping_data_source_test.go"},
	},
	"project_enforcement": {
		{projectEnforcementDataSourceTestTmpl, "project_enforcement_data_source_test.go"},
	},
	"project_permission_mapping": {
		{projectPermissionMappingDataSourceTestTmpl, "project_permission_mapping_data_source_test.go"},
	},
	"scope_criteria": {
		{scopeCriteriaTestTmpl, "scope_criteria_test.go"},
	},
	"user": {
		{userDataSourceTestTmpl, "user_data_source_test.go"},
		{userTestTmpl, "user_test.go"},
	},
	"webhook": {
		{webhookDataSourceTestTmpl, "webhook_data_source_test.go"},
		{webhookTestTmpl, "webhook_test.go"},
	},
}

// aliasFactory is an extra constructor a service package registers beside the
// derived one, for a resource kept under a second Terraform type name.
type aliasFactory struct {
	Factory string
	Comment string
}

// aliasFactoriesByName are the backwards-compatible type aliases three
// resources carry from the previous provider. The alias bodies themselves are
// companion files; this is what registers them.
var aliasFactoriesByName = map[string]struct {
	Resources   []aliasFactory
	DataSources []aliasFactory
}{
	"account_cache": {
		DataSources: []aliasFactory{{"NewCachedAccountDataSource", "backwards-compat alias: kion_cached_account"}},
	},
	"cft": {
		Resources:   []aliasFactory{{"NewAwsCftResource", "backwards-compat alias: kion_aws_cloudformation_template"}},
		DataSources: []aliasFactory{{"NewAwsCftDataSource", "backwards-compat alias: kion_aws_cloudformation_template"}},
	},
	"iam_policy": {
		Resources:   []aliasFactory{{"NewAwsIamPolicyResource", "backwards-compat alias: kion_aws_iam_policy"}},
		DataSources: []aliasFactory{{"NewAwsIamPolicyDataSource", "backwards-compat alias: kion_aws_iam_policy"}},
	},
}

// servicePackageData is the payload for servicepackage.gtpl.
type servicePackageData struct {
	Pkg, Pascal       string
	DataSourceCtor    string
	ResourceAliases   []aliasFactory
	DataSourceAliases []aliasFactory
	// NoResource marks a package that registers only a data source.
	NoResource bool
}

// newServicePackageData builds the registration payload. dsCtor is empty when
// the package ships no data source.
func newServicePackageData(name, pascal, dsCtor string) servicePackageData {
	a := aliasFactoriesByName[name]
	return servicePackageData{
		Pkg: name, Pascal: pascal, DataSourceCtor: dsCtor,
		ResourceAliases: a.Resources, DataSourceAliases: a.DataSources,
	}
}

// newDataSourceOnlyPackageData is the registration for a package with a data
// source and no resource.
func newDataSourceOnlyPackageData(name, pascal string) servicePackageData {
	d := newServicePackageData(name, pascal, "New"+pascal+"DataSource")
	d.NoResource = true
	return d
}
