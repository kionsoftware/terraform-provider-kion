# Terraform Importer Script

The `terraform-importer.py` script was built to quickly import existing resources from a Kion installation and begin managing them (and creating new ones) using the Kion Terraform Provider.

## Prerequisites

### python3 and requests

The system running the script will need to have python3 installed, and the requests library.

```bash
$ pip3 install requests
```

## Required Terraform Files

This script will check your import directory for a `main.tf` file. If it doesn't find one, it will create one that is
configured with the modules included based on which resources you imported.

If `main.tf` already exists, the script will create a `main.tf.example` file instead. You may have to copy some lines from that file into the existing `main.tf`, or just overwrite the existing file entirely with the new one.

For all imported resources, the script will follow a similar process for creating a `provider.tf` file in each module directory.

## Workflow

The general workflow when using this script for the first time, or on subsequent runs when new resource types are imported is:

1. Import all resources
2. Change directories into the terraform-managed directory
3. Run `terraform init`
4. Run `bash import_resource_state.sh` (imports resources into the terraform state - can take a long time)
5. Run `terraform plan`
6. Review proposed changes
7. Run `terraform apply`

## Importing Current Resource State

After importing resources from your installation of Kion, but before running your first `terraform plan`, you'll need to generate the Terraform state. After running the importer script, it will generate a bash script for you that you can run to easily do this. The script contains the `terraform import` command for each individual resource that was imported during the run.

At the end of the run, you'll see this in the output:
`cd <import directory> ; bash import_resource_state.sh`

Just copy and paste that command to run the script and generate the initial Terraform state file.

## Usage

Here is an example of usage and output:

```
$ python3 terraform-importer.py --import-dir ./kion/roles --kion-url https://kion.example.com --skip-cfts --skip-cloud-rules --skip-iams --prepend-id

Beginning import from https://kion.example.com

Skipping AWS CloudFormation Templates

Skipping AWS IAM Policies

Importing Project Cloud Access Roles
--------------------------
Found 1 roles on Project AD
Found 1 roles on Project Logging
Done.

Importing OU Cloud Access Roles
--------------------------
Found 8 roles on OU Central Services
Found 1 roles on OU Networking
Found 1 roles on OU Security
Done.

Skipping Cloud Rules

Import finished.
```

The example above used a few flags to skip certain resources, and change naming of the imported files. See the description of those flags lower on the page for more info on their usage.

## Name Normalization

This script will normalize the names of your resources for better management by source control.

It will replace all spaces with underscores, and remove all non-alphanumeric characters except for dashes and underscores.

This only applies to the names of the files kept in the import directory, and the values of some fields. It will not rename anything in Kion.

## Handling of Cloud Access Roles

This script will create sub-directories in the `cloud-access-role-ou` and `cloud-access-role-project` directories for each of the OUs and projects found in Kion that have locally applied Cloud Access Roles.

This way, roles will be organized by OU / Project for easier management.

The sub-directories will follow the naming scheme `{id}-{normalized name}`.

For example, an OU with ID `5` and name `Central Services` will have a sub-directory named `5-Central_Services`.

Here is an example of what the folder structure will look like in the import directory for OU Cloud Access Roles.

```
├── 5-Central_Services
│   ├── Account_Managers.metadata.json
│   ├── Engineers.metadata.json
│   ├── Engineers_Read_Only.metadata.json
│   ├── Finance.metadata.json
│   ├── Security_Analyst.metadata.json
├── 33-Security
│   └── Security_Analyst.metadata.json
└── 35-Splunk
    └── Splunk_Developers.metadata.json
```

### Import Directory

Before running this script, you will need to create a root directory for the import.

The required sub-folders in the root depend on which resources you are importing.

Here is a mapping.

- Cloud Rules           -> `cloud-rule`
- CFTs                  -> `aws-cloudformation-template`
- IAM policies          -> `aws-iam-policy`
- Project Roles         -> `cloud-access-role-project`
- OU Roles              -> `cloud-access-role-ou`
- Compliance Standards  -> `compliance-standard`
- Compliance Checks     -> `compliance-checks`

The script will tell you if any required folder is missing, and exit. For example:

```
<import_dir> is missing the following sub-directories:
- aws-cloudformation-template
- aws-iam-policy
- cloud-rule
- cloud-access-role-ou
- cloud-access-role-project
```

### Kion API key & user with permissions

You will need a Kion app API key in order for the script to authenticate with Kion. The user for which this key is generated will need at least `Browse` permissions on:
- Cloud Rules
- AWS CloudFormation templates
- AWS IAM policies
- OU Cloud Access Roles
- Project Cloud Access Roles
- Compliance

It is recommended to export this to the environment rather than providing the key as a CLI argument. For example:

```bash
$ export KION_API_KEY='app_1_ksjdlZUisjdlFKSJdlfskj'
```

## Required Arguments

### `--kion-url`

The URL to Kion.

Example usage:

```bash
$ python3 terraform-importer --kion-url https://kion.example.com
```

### `--import-dir`

The path to the root of the directory for the import. All resources will be kept in sub-directories of this directory.

Example usage:

```bash
$ python3 terraform-importer --import-dir /Users/me/Code/kion/terraform/import
```

## Optional Arguments

### `--kion-api-key`

The Kion app API key mentioned in the `Prerequisites` section.
This is actually required if the environment variable hasn't been set

Example usage:

```bash
$ python3 terraform-importer --kion-api-key app_thisshouldreallybeanenvironmentvariable
```

This can also be set as an environment variable called `KION_API_KEY`. (preferred)

Here is an example of doing that:

```bash
$ export KION_API_KEY='app_1_ksjdlZUisjdlFKSJdlfskj'
```

### `--skip-cfts`

Skip importing AWS CloudFormation templates.

Example usage:

```bash
$ python3 terraform-importer --skip-cfts
```

### `--skip-iams`

Skip importing AWS IAM policies.

Example usage:

```bash
$ python3 terraform-importer --skip-iams
```

### `--skip-project-roles`

Skip importing Project Cloud Access Roles.

Example usage:

```bash
$ python3 terraform-importer --skip-project-roles
```

### `--skip-ou-roles`

Skip importing OU Cloud Access Roles.

Example usage:

```bash
$ python3 terraform-importer --skip-ou-roles
```

### `--skip-cloud-rules`

Skip importing Cloud Rules.

Example usage:

```bash
$ python3 terraform-importer --skip-cloud-rules
```

### `--skip-checks`

Skip importing Compliance Checks.

Example usage:

```bash
$ python3 terraform-importer --skip-checks
```

### `--skip-standards`

Skip importing Compliance Standards.

Example usage:

```bash
$ python3 terraform-importer --skip-standards
```

### `--skip-azure-policies`

Skip importing Azure Policies.

Example usage:

```bash
$ python3 terraform-importer --skip-azure-policies
```

### `--skip-azure-roles`

Skip importing Azure Roles.

Example usage:

```bash
$ python3 terraform-importer --skip-azure-roles
```

### `--skip-ssl-verify`

Skip verification of the Kion SSL certificate.

Use this if Kion does not have a valid certificate.

Using this will output a warning message during the import and may affect functionality overall.

Example usage:

```bash
$ python3 terraform-importer --skip-ssl-verify
```

### `--clone-system-managed`

Clone system-managed resources.

When this flag is used, the script will make a clone of all system-managed resources and then import those clones into the repository.

Keep in mind that doing this will effectively double the number of resources in your instance of Kion, which can lead to a lot of clutter.

Not all resources are compatible with being cloned and some errors may occur during the cloning process.

Some errors that occurred while developing this function were all Azure related:

- Azure compliance checks not having a valid Azure Policy Definition ID
- Azure Policies failing due to an invalid policy
- Azure Roles failing due to an invalid policy

This flag requires 2 to 3 other flags to be used along with it:

#### `--clone-prefix`

A string to prepend to each cloned resource's name. It must end with a dash or underscore.

#### `--clone-user-ids`

A space separated list of user IDs to be set as the owners of the cloned resources.

You must provide this flag or `--clone-user-group-ids` or both when cloning.

#### `--clone-user-group-ids`

A space separated list of user group IDs to be set as the owners of the cloned resources.

You must provide this flag or `--clone-user-ids` or both when cloning.

Example of cloning and providing all flags:

```bash
$ python3 terraform-importer --clone-system-managed --clone-prefix MYCLONE_ --clone-user-ids 1 2 --clone-user-group-ids 3 4
```

### `--import-aws-managed`

Import AWS-managed resources that are already present in Kion.

You may want to import then to easily find their IDs
for referencing in other resources, or just for reviewing their content.

These resources cannot be changed or deleted, and so they will have `.skip` appended to their filenames
to make Terraform ignore them. This is done so that Terraform will not try to manage them.

They will also not be listed in the `import_resource_state.sh` bash script as there is no
purpose in importing them into the Terraform state since they cannot be managed.

```bash
$ python3 terraform-importer --import-aws-managed
```

### `--overwrite`

Overwrite existing files during the import.

The default behavior is to not overwrite existing files so that the copy in source control stays as
the authoritative source.

You may want to use this flag on subsequent imports to ensure your copies in source control are
up to date prior to using source control exclusively for making changes.

Example usage:

```bash
$ python3 terraform-importer --overwrite
```

### `--prepend-id`

Prepend each resource's ID to it's associated files.

Since associations between IAM policies, CFTs, etc to their Cloud Rules and Roles are made through IDs,
having each resource's ID prepended to it's filename makes it much easier to find.

For example, the following Cloud Rule metadata file contains IAM policy `4`:

```json
{
    "name": "US Regions Only",
    "description": "Deny access to regions outside of the US.",
    "pre_webhook_id": null,
    "post_webhook_id": null,
    "project_ids": [],
    "iam_policy_ids": [
        4
    ],
}
```

By setting this flag, the resource and metadata files that manage that IAM policy will have `4-` prepended to their names:

```
4-AWSServicesOnlyInUSA.json
4-AWSServicesOnlyInUSA.metadata.json
```

One caveat to this is that you will need to manually maintain this naming scheme after running the import. Since new resources will not have an ID yet in Kion that means you will have to rename the files with the ID after its been created. Depending on your source control workflow, that may require opening merge / pull requests just for file name changes.

Example usage:

```bash
$ python3 terraform-importer --prepend-id
```
