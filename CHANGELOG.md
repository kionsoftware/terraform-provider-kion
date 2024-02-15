# Changelog

All notable changes to this project will be documented in this file.

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

### Added
- Added Funding Source resource ([#25](https://github.com/kionsoftware/terraform-provider-kion/pull/25))

## [0.3.4] - 2023-03-27

### Added
- Added tags support to Cloudformation template resource (#24)

### Changed
- Fix KeyError bug in import script (#7)


## [0.3.3] - 2023-03-10

### Added
- Added better documentation around datasource filtering.

### Changed
- Fixed a bug where Kion's provider would panic when applying some resources.

## [0.3.2] - 2023-03-09
### Added
- Added documentation that clarifies that either an owner user or owner group must be defined for some resources.
- Added better error handling when a user attempts to create a Kion resource without an owner user or owner group.

### Changed
- Fixed a bug where Terraform expected resources to be in a specific order on unordered resources.

## [0.3.1] - 2023-02-27
### Changed
- Fix the description for IAM policies in the documentation to be more accurate.
- Added clarity around creating compliance standards.
- Allow project creation with budget when enabled in Kion.

## [0.3.0] - 2022-02-25
### Added
- Support creating, updating, and deleting resources for: AWS Service Control Policies.
- Support adding and removing AWS Service Control Policies on Project and OU Cloud Rules.
- Support creating, updating, and deleting resources for: Azure ARM Templates.
- Support adding and removing Azure ARM Templates on Project and OU Cloud Rules.
- Support creating, updating, and deleting resources for: Azure Role Definitions.
- Support adding and removing Azure Role Definitions on Project and OU Cloud Rules.

### Changed
- Rebrand from cloudtamer.io to Kion.
- Change provider name: `cloudtamer-io/cloudtamerio` to `kionsoftware/kion`.
- Change resource and data source names prefix: `cloudtamer_` to `kion_`.
- Change environment variables: `CLOUDTAMERIO_URL` and `CLOUDTAMERIO_APIKEY` to `KION_URL` and `KION_APIKEY`, respectively.
- Made the `created_by_user_id` field for Compliance Checks optional. This field will default to the requesting user's ID if not specified.

## [0.2.0] - 2021-11-19
### Added
- Support creating, updating, and deleting resources for: user groups.
- Support creating, updating, and deleting resources for: SAML IDMS user group associations.
- Support creating, updating, and deleting resources for: Projects
- Support creating, updating, and deleting resources for: Google Cloud IAM Roles.
- Support adding and removing Google Cloud IAM Roles on Project and OU Cloud Rules.

### Changed
- Fix several requests that use the wrong user & user group IDs to remove owners from a resource.

## [0.1.4] - 2021-08-09
### Added
- Support creating, updating, and deleting resources for: OUs. (Requires Kion v2.31.0 or newer)

## [0.1.3] - 2021-06-29
### Changed
- Fix bug on project cloud access role creation so 'apply_to_all_accounts' and 'accounts' fields are mutually exclusive.
- Remove unused errors throughout the code.

## [0.1.2] - 2021-04-01
### Added
- Support creating, updating, and deleting resources for: OU cloud access roles and project cloud access roles.

### Changed
- Fix bug on compliance standard creation so compliance checks are attached during creation instead of requiring another `terraform apply`.
- Fix bug on cloud rule creation so associated items are attached during creation instead of requiring another `terraform apply`.

## [0.1.1] - 2021-03-30
### Added
- Ability to import resources using `terraform import`.

## [0.1.0] - 2021-02-08
### Added
- Initial release of the provider.
- Support creating, updating, and deleting resources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
- Support querying data sources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
