# Changelog

All notable changes to this project will be documented in this file.

## [0.3.14] - 2024-05-15

- Fixed an issue where the provider would attempt to update project labels even when labels were not defined in the HCL. The provider now checks if labels are explicitly set and non-empty before making update requests, preventing unnecessary API calls and avoiding the `405 Method Not Allowed error` from any Kion version before v3.7.7.
- Added examples into example directory that previously only were in the `README.md` file.  This allows each resources to have more example information directly in the resource docs.

## [0.3.13] - 2024-04-09

- Changed `aws_cloudformation_templates` and `azure_arm_template_definitions` from `TypeSet` to `TypeList` in the Terraform provider schema to preserve the order of templates as specified in Terraform configurations.
- Updated the `resourceCloudRuleCreate` function to build lists of template IDs directly from the Terraform state, ensuring order preservation during cloud rule creation.
- Modified the `resourceCloudRuleRead` function to correctly parse and set ordered template lists in the Terraform state, maintaining consistency with the API.
- Bump Go Version to 1.22

## [0.3.12] - 2024-03-29

- Introduced several enhancements and refactoring changes to the Makefile and the Terraform provider. The key changes include adding a development build option, introducing a sync mechanism for AWS account creation, allowing more than one `kion_aws_account` to be created at once.

## [0.3.11] - 2024-03-20

- Introduced retry logic in the AWS account creation process within the Terraform Provider. This improvement is designed to handle scenarios where the account creation is temporarily hindered by ongoing AWS service operations. The logic includes a default retry count of three attempts with a delay of 30 seconds between each attempt, enhancing the robustness of the account creation process in fluctuating AWS service conditions.

## [0.3.10] - 2024-02-29

- Added support for `azure_role_definitions` in OU Cloud Access Roles. This enhancement includes the ability to define and manage Azure role definitions directly within Terraform.

## [0.3.9] - 2024-02-15

- Fixed an issue where users were unable to create Kion Azure Policy resources due to certain attributes (`name`, `description`, `policy`, `parameters`) being set to "Read-Only" in the Terraform Kion Provider. This fix involves changing the `Computed: true` parameter for these attributes to either `Required` or `Optional`, based on the API documentation for azure-policy. This change allows for the proper creation and management of Azure Policy resources within Terraform.

## [0.3.8] - 2024-01-03

- Fixed an issue that prevented importing existing Kion accounts into the terraform state.  When importing existing accounts, the user should specify the account ID using an `account_id=` or `account_cache_id=` prefix to tell the terraform provider whether the provided ID is an account ID or a cached account ID.  See the README for more information.

## [0.3.7] - 2023-12-21

- Added `kion_aws_account`, `kion_gcp_account` and `kion_azure_account` resources ([#32](https://github.com/kionsoftware/terraform-provider-kion/pull/32))
- Added `kion_account` and `kion_cached_account` data sources
- Upgrade terraform-plugin-sdk from v2.10.0 to v2.30.0

## [0.3.6] - 2023-11-20

- Added a `kion_label` resource ([#31](https://github.com/kionsoftware/terraform-provider-kion/pull/31))
- Added a `labels` attribute to OUs, Projects, Funding sources and cloud rules

## [0.3.5] - 2023-09-22

- Added Funding Source resource ([#25](https://github.com/kionsoftware/terraform-provider-kion/pull/25))

## [0.3.4] - 2023-03-27

- Added Tags support to Cloudformation template resource (#24)
- Fix KeyError bug in import script (#7)

## [0.3.3] - 2023-03-10

- Added Better documentation around datasource filtering.
- Fixed a bug where Kion's provider would panic when applying some resources.

## [0.3.2] - 2023-03-09

- Added documentation that clarifies that either an owner user or owner group must be defined for some resources.
- Added better error handling when a user attempts to create a Kion resource without an owner user or owner group.
- Fixed a bug where Terraform expected resources to be in a specific order on unordered resources.

## [0.3.1] - 2023-02-27

- Fix the description for IAM policies in the documentation to be more accurate.
- Added clarity around creating compliance standards.
- Allow project creation with budget when enabled in Kion.

## [0.3.0] - 2022-02-25

- Added Support for creating, updating, and deleting resources for: AWS Service Control Policies.
- Added Support for adding and removing AWS Service Control Policies on Project and OU Cloud Rules.
- Added Support for creating, updating, and deleting resources for: Azure ARM Templates.
- Added Support for adding and removing Azure ARM Templates on Project and OU Cloud Rules.
- Added Support for creating, updating, and deleting resources for: Azure Role Definitions.
- Added Support for adding and removing Azure Role Definitions on Project and OU Cloud Rules.
- Rebrand from cloudtamer.io to Kion.
- Change provider name: `cloudtamer-io/cloudtamerio` to `kionsoftware/kion`.
- Change resource and data source names prefix: `cloudtamer_` to `kion_`.
- Change environment variables: `CLOUDTAMERIO_URL` and `CLOUDTAMERIO_APIKEY` to `KION_URL` and `KION_APIKEY`, respectively.
- Made the `created_by_user_id` field for Compliance Checks optional. This field will default to the requesting user's ID if not specified.

## [0.2.0] - 2021-11-19

- Added Support for creating, updating, and deleting resources for: user groups.
- Added Support for creating, updating, and deleting resources for: SAML IDMS user group associations.
- Added Support for creating, updating, and deleting resources for: Projects
- Added Support for creating, updating, and deleting resources for: Google Cloud IAM Roles.
- Added Support for adding and removing Google Cloud IAM Roles on Project and OU Cloud Rules.
- Fix several requests that use the wrong user & user group IDs to remove owners from a resource.

## [0.1.4] - 2021-08-09

- Added Support for creating, updating, and deleting resources for: OUs. (Requires Kion v2.31.0 or newer)

## [0.1.3] - 2021-06-29

- Fix bug on project cloud access role creation so 'apply_to_all_accounts' and 'accounts' fields are mutually exclusive.
- Remove unused errors throughout the code.

## [0.1.2] - 2021-04-01

- Added Support for creating, updating, and deleting resources for: OU cloud access roles and project cloud access roles.
- Fix bug on compliance standard creation so compliance checks are attached during creation instead of requiring another `terraform apply`.
- Fix bug on cloud rule creation so associated items are attached during creation instead of requiring another `terraform apply`.

## [0.1.1] - 2021-03-30

- Added Ability to import resources using `terraform import`.

## [0.1.0] - 2021-02-08

- Initial release of the provider.
- Added Support for creating, updating, and deleting resources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
- Added Support for querying data sources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
