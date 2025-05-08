# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/). This project adheres to [Semantic Versioning](http://semver.org/) with the exception of version 0 as we find our footing. Only changes to the application should be logged here. Repository maintenance, tests, and other non application changes should be excluded.

[Unreleased] - yyyy-mm-dd
Notes for upgrading...

Added
Changed
Deprecated
Removed
Fixed

## [0.3.24] - 2024-05-08

### Fixed

- Updated Funding Sources Example Documentation to use the correct permission_scheme_id

## [0.3.23] - 2024-05-02

### Fixed

- Resolved concurrent OU creation issue causing database constraint violations
- Prevented "Duplicate entry" errors when creating multiple OUs in Terraform
- Added mutex synchronization for OU hierarchy operations to maintain data integrity

## [0.3.22] - 2024-03-13

### Added

- Enhanced account location detection with explicit location handling in state
- Added support for preserving original account IDs during import operations
- Improved logging with additional context for account operations

### Changed

- Refactored account location handling to prioritize explicitly set locations
- Updated account import logic to better handle account_id and account_cache_id prefixes
- Enhanced documentation structure and examples across all resources
- Modernized provider configuration with clearer environment variable guidance

### Fixed

- Improved account lookup logic to prevent unnecessary API calls
- Enhanced error handling for account operations with better context
- Fixed account ID handling during import operations

## [0.3.21] - 2024-28-28

### Added

- New `kion_project_note` resource and data source for managing project notes
- New filtering options for AWS IAM policies:
  - `query` parameter for name matching
  - `policy_type` filter for AWS managed policies
  - Pagination support with `page` and `page_size`
- Validation for `start_datecode` format in AWS accounts (must be YYYY-MM)

### Changed

- Improved account location detection to handle both project and cache locations
- Updated documentation structure:
  - Moved examples from README to dedicated examples directory
  - Added organized list of available examples by resource type
  - Updated AWS IAM policy examples to demonstrate data source usage

### Fixed

- Account reading logic to better handle accounts that move between projects and cache
- Validation for date code formats in AWS accounts

### Documentation

- Added new examples for project notes
- Updated AWS IAM policy documentation with new filtering options
- Improved organization of examples by resource type
- Added missing fields in various resource documentation

## [0.3.20] - 2024-02-20

### Added

* Added support for custom variables with new resources `kion_custom_variable` and `kion_custom_variable_override` and corresponding data sources
* Added v4 API support for IAM policy list operations with enhanced filtering capabilities (query, policy_type, pagination)
* Added concurrent CloudFormation template sync option to cloud rules via `concurrent_cft_sync` parameter to control parallel/sequential template deployment
* Added funding source ID field to funding source data source output for better resource referencing
* Added support for non-allocation mode in funding sources by making `ou_id` optional

### Changed

* Enhanced account management with improved import functionality and validation:
  * Simplified account import process with multiple import methods
  * Added better validation for account moves between projects
  * Improved error messages and logging for account operations
* Improved project budget management:
  * Added proper state tracking for budget changes
  * Enhanced validation for budget periods and date formats
  * Implemented proper ordering of budget operations
* Refactored AWS IAM policy handling:
  * Maintained v3 endpoint compatibility for single resource operations
  * Added v4 API support for enhanced filtering and pagination
* Standardized error handling across resources using consistent error wrapping patterns

### Fixed

* Fixed compliance check removal in compliance standards to properly handle "compliance check not found" errors
* Fixed project budget updates to properly apply changes and prevent overlapping periods
* Fixed permission scheme handling in funding sources to prevent unnecessary updates
* Fixed account cache operations and removed unused testing functions
* Improved handling of budget period modifications to prevent conflicts during splits or updates

## [0.3.19] - 2024-09-26

## What's Changed

### Fixed

* Ensured consistent handling of funding source `amount` fields across API models and Terraform schema definitions to prevent type mismatches.

### Changed

* Updated the `amount` field in the `FundingSource` schema from `TypeInt` to `TypeFloat` for more accurate representation of funding values.
* Modified the `models_funding_source.go` to reflect the new `float64` type for the `amount` field in all funding source-related structures.
* Adjusted the `resourceFundingSource` and `resourceFundingSourceUpdate` methods to use `float64` for the `amount` attribute during creation and updates.

## [0.3.18] - 2024-08-15

## What's Changed

### Fixed

* Corrected a regression in the `NewFilterable` function that caused issues with filter list processing and key-value handling for HCL-defined filters.
* Fixed formatting issues in the `clone_resource` function to prevent unnecessary line breaks in resource names.

## [0.3.17] - 2024-08-13

## What's Changed

### Added

* Introduced permission mapping management for funding sources, OUs, projects, and global permissions.
* Added support for managing webhooks and retrieving user data with new resources and data sources.
* Account alias field added to all account-related resources for better identification.

### Changed

* Refactored helper functions for improved efficiency and code quality.
* Enhanced error handling across account-related resources and project enforcement.

### Removed

* Deprecated `OptionalString` function in favor of the more versatile `OptionalValue`.

### Fixed

* Standardized the handling of optional fields across various resources.
* Resolved issues with redundant error handling by centralizing logic.
* Normalized JSON fields in webhook resources to prevent unnecessary changes.

## [0.3.16] - 2024-06-07

## What's Changed

### Added

* Added GCP IAM Roles to Cloud Access Roles [pull/77](https://github.com/kionsoftware/terraform-provider-kion/pull/77)
* Add Project Enforcement Support [pull/75](https://github.com/kionsoftware/terraform-provider-kion/pull/75)
* Add golangci-lint Workflow [pull/71](https://github.com/kionsoftware/terraform-provider-kion/pull/71)
* Added `.golangci.yml` configuration file for linting and then fixed nearly everything that it found [pull/68](https://github.com/kionsoftware/terraform-provider-kion/pull/68)
* Added New Azure Policies Support [pull/74](https://github.com/kionsoftware/terraform-provider-kion/pull/74)

### Changed

* Update `.gitignore` and enhanced `README.md` for Terraform Importer Script [pull/67](https://github.com/kionsoftware/terraform-provider-kion/pull/67)
* Refactor Kion Client Codebase [pull/69](https://github.com/kionsoftware/terraform-provider-kion/pull/69)

## [0.3.15] - 2024-05-17

* No changes.  This change was needed to get the Terraform Registry back in sync with Github.

## [0.3.14] - 2024-05-15

* Fixed an issue where the provider would attempt to update project labels even when labels were not defined in the HCL. The provider now checks if labels are explicitly set and non-empty before making update requests, preventing unnecessary API calls and avoiding the `405 Method Not Allowed error` from any Kion version before v3.7.7.
* Added examples into example directory that previously only were in the `README.md` file.  This allows each resources to have more example information directly in the resource docs.

## [0.3.13] - 2024-04-09

* Changed `aws_cloudformation_templates` and `azure_arm_template_definitions` from `TypeSet` to `TypeList` in the Terraform provider schema to preserve the order of templates as specified in Terraform configurations.
* Updated the `resourceCloudRuleCreate` function to build lists of template IDs directly from the Terraform state, ensuring order preservation during cloud rule creation.
* Modified the `resourceCloudRuleRead` function to correctly parse and set ordered template lists in the Terraform state, maintaining consistency with the API.
* Bump Go Version to 1.22

## [0.3.12] - 2024-03-29

* Introduced several enhancements and refactoring changes to the Makefile and the Terraform provider. The key changes include adding a development build option, introducing a sync mechanism for AWS account creation, allowing more than one `kion_aws_account` to be created at once.

## [0.3.11] - 2024-03-20

* Introduced retry logic in the AWS account creation process within the Terraform Provider. This improvement is designed to handle scenarios where the account creation is temporarily hindered by ongoing AWS service operations. The logic includes a default retry count of three attempts with a delay of 30 seconds between each attempt, enhancing the robustness of the account creation process in fluctuating AWS service conditions.

## [0.3.10] - 2024-02-29

* Added support for `azure_role_definitions` in OU Cloud Access Roles. This enhancement includes the ability to define and manage Azure role definitions directly within Terraform.

## [0.3.9] - 2024-02-15

* Fixed an issue where users were unable to create Kion Azure Policy resources due to certain attributes (`name`, `description`, `policy`, `parameters`) being set to "Read-Only" in the Terraform Kion Provider. This fix involves changing the `Computed: true` parameter for these attributes to either `Required` or `Optional`, based on the API documentation for azure-policy. This change allows for the proper creation and management of Azure Policy resources within Terraform.

## [0.3.8] - 2024-01-03

* Fixed an issue that prevented importing existing Kion accounts into the terraform state.  When importing existing accounts, the user should specify the account ID using an `account_id=` or `account_cache_id=` prefix to tell the terraform provider whether the provided ID is an account ID or a cached account ID.  See the README for more information.

## [0.3.7] - 2023-12-21

* Added `kion_aws_account`, `kion_gcp_account` and `kion_azure_account` resources ([#32](https://github.com/kionsoftware/terraform-provider-kion/pull/32))
* Added `kion_account` and `kion_cached_account` data sources
* Upgrade terraform-plugin-sdk from v2.10.0 to v2.30.0

## [0.3.6] - 2023-11-20

* Added a `kion_label` resource ([#31](https://github.com/kionsoftware/terraform-provider-kion/pull/31))
* Added a `labels` attribute to OUs, Projects, Funding sources and cloud rules

## [0.3.5] - 2023-09-22

* Added Funding Source resource ([#25](https://github.com/kionsoftware/terraform-provider-kion/pull/25))

## [0.3.4] - 2023-03-27

* Added Tags support to Cloudformation template resource (#24)
* Fix KeyError bug in import script (#7)

## [0.3.3] - 2023-03-10

* Added Better documentation around datasource filtering.
* Fixed a bug where Kion's provider would panic when applying some resources.

## [0.3.2] - 2023-03-09

* Added documentation that clarifies that either an owner user or owner group must be defined for some resources.
* Added better error handling when a user attempts to create a Kion resource without an owner user or owner group.
* Fixed a bug where Terraform expected resources to be in a specific order on unordered resources.

## [0.3.1] - 2023-02-27

* Fix the description for IAM policies in the documentation to be more accurate.
* Added clarity around creating compliance standards.
* Allow project creation with budget when enabled in Kion.

## [0.3.0] - 2022-02-25

* Added Support for creating, updating, and deleting resources for: AWS Service Control Policies.
* Added Support for adding and removing AWS Service Control Policies on Project and OU Cloud Rules.
* Added Support for creating, updating, and deleting resources for: Azure ARM Templates.
* Added Support for adding and removing Azure ARM Templates on Project and OU Cloud Rules.
* Added Support for creating, updating, and deleting resources for: Azure Role Definitions.
* Added Support for adding and removing Azure Role Definitions on Project and OU Cloud Rules.
* Rebrand from cloudtamer.io to Kion.
* Change provider name: `cloudtamer-io/cloudtamerio` to `kionsoftware/kion`.
* Change resource and data source names prefix: `cloudtamer_` to `kion_`.
* Change environment variables: `CLOUDTAMERIO_URL` and `CLOUDTAMERIO_APIKEY` to `KION_URL` and `KION_APIKEY`, respectively.
* Made the `created_by_user_id` field for Compliance Checks optional. This field will default to the requesting user's ID if not specified.

## [0.2.0] - 2021-11-19

* Added Support for creating, updating, and deleting resources for: user groups.
* Added Support for creating, updating, and deleting resources for: SAML IDMS user group associations.
* Added Support for creating, updating, and deleting resources for: Projects
* Added Support for creating, updating, and deleting resources for: Google Cloud IAM Roles.
* Added Support for adding and removing Google Cloud IAM Roles on Project and OU Cloud Rules.
* Fix several requests that use the wrong user & user group IDs to remove owners from a resource.

## [0.1.4] - 2021-08-09

* Added Support for creating, updating, and deleting resources for: OUs. (Requires Kion v2.31.0 or newer)

## [0.1.3] - 2021-06-29

* Fix bug on project cloud access role creation so 'apply_to_all_accounts' and 'accounts' fields are mutually exclusive.
* Remove unused errors throughout the code.

## [0.1.2] - 2021-04-01

* Added Support for creating, updating, and deleting resources for: OU cloud access roles and project cloud access roles.
* Fix bug on compliance standard creation so compliance checks are attached during creation instead of requiring another `terraform apply`.
* Fix bug on cloud rule creation so associated items are attached during creation instead of requiring another `terraform apply`.

## [0.1.1] - 2021-03-30

* Added Ability to import resources using `terraform import`.

## [0.1.0] - 2021-02-08

* Initial release of the provider.
* Added Support for creating, updating, and deleting resources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
* Added Support for querying data sources for: AWS CloudFormation templates, AWS IAM policies, Azure policies, cloud rules, compliance checks, and compliance standards.
