# terraform-provider-kion <!-- omit in toc -->

The Terraform provider for Kion allows you interact with the Kion API using the Terraform HCL language. Our provider supports creating, updating, reading, and deleting resources. You can also use it to query for resources using filters even if a resource is not created through Terraform.

- [Getting Started](#getting-started)
  - [Importing Resource State](#importing-resource-state)
- [Migrating from the cloudtamerio Provider](#migrating-from-the-cloudtamerio-provider)
- [Sample Commands](#sample-commands)
  - [Resources](#resources)
  - [Data Sources](#data-sources)
  - [Locals](#locals)
- [Repository Maintainer: Push to Terraform Registry](#repository-maintainer-push-to-terraform-registry)

## Getting Started

Below is sample code on how to create an IAM policy in Kion using Terraform.

First, set your environment variables:

```bash
export KION_URL=https://kion.example.com
export KION_APIKEY=API-KEY-HERE
```

Next, paste this code into a `main.tf` file:

```hcl
terraform {
  required_providers {
    kion = {
      source  = "kionsoftware/kion"
      version = "0.3.0"
    }
  }
}

provider "kion" {
  # If these are commented out, they will be loaded from environment variables.
  # url = "https://kion.example.com"
  # apikey = "key here"
}

# Create an IAM policy.
resource "kion_aws_iam_policy" "p1" {
  name         = "sample-resource"
  description  = "Provides AdministratorAccess to all AWS Services"
  aws_iam_path = ""
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "*",
            "Resource": "*"
        }
    ]
}
EOF
}

# Output the ID of the resource created.
output "policy_id" {
  value = kion_aws_iam_policy.p1.id
}
```

Then, run these commands:

```bash
# Initialize the project.
terraform init

# Show the plan.
terraform plan

# Apply the changes.
terraform apply --auto-approve
```

You can now make changes to the `main.tf` file and then re-run the `apply` command to push the changes to Kion.

### Importing Resource State

This provider does support [importing state for resources](https://www.terraform.io/docs/cli/import/index.html). You will need to create the Terraform files and then you can run commands like this to generate the `terraform.tfstate` so you don't have to delete all your resources and then recreate them to work with Terraform:

```bash
# Initialize the project.
terraform init

# Import the resource from your environment - this assumes you have a module called
# 'aws-cloudformation-template' and you are importing into a resource you defined as 'AuditLogging'.
# The '20' at the end is the ID of the resource in Kion.
terraform import module.aws-cloudformation-template.kion_aws_cloudformation_template.AuditLogging 20

# Verify the state is correct - there shouldn't be any changes listed.
terraform plan
```

## Migrating from the cloudtamerio Provider

See the instructions [here](docs/provider-migration.md) to migrate existing managed resources from the cloudtamerio provider
to the new Kion provider.

## Sample Commands

Below is a collection of sample commands when working with the Terraform provider.

### Resources

A few of the optional fields are commented out.

```hcl
# Create an IAM policy.
resource "kion_aws_iam_policy" "p1" {
  name = "sample-resource"
  # description  = "Provides AdministratorAccess to all AWS Services"
  # aws_iam_path = ""
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "*",
            "Resource": "*"
        }
    ]
}
EOF
}

# Output the ID of the resource created.
output "policy_id" {
  value = kion_aws_iam_policy.p1.id
}
```

```hcl
# Create a CloudFormation template.
resource "kion_aws_cloudformation_template" "t1" {
  name    = "sample-resource"
  regions = ["us-east-1"]
  # description = "Creates a test IAM role."
  # region                 = ""
  # sns_arns               = ""
  # template_parameters    = ""
  # termination_protection = false
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  policy = <<EOF
{
    "AWSTemplateFormatVersion": "2010-09-09",
    "Description": "Creates a test IAM role.",
    "Metadata": {
        "VersionDate": {
            "Value": "20180718"
        },
        "Identifier": {
            "Value": "blank-role.json"
        }
    },
    "Resources": {
        "EnvTestRole": {
            "Type": "AWS::IAM::Role",
            "Properties": {
                "RoleName": "env-test-role",
                "Path": "/",
                "Policies": [],
                "AssumeRolePolicyDocument": {
                    "Statement": [
                        {
                            "Effect": "Allow",
                            "Principal": {
                                "Service": [
                                    "ec2.amazonaws.com"
                                ]
                            },
                            "Action": [
                                "sts:AssumeRole"
                            ]
                        }
                    ]
                }
            }
        }
    }
}
EOF
}

# Output the ID of the resource created.
output "template_id" {
  value = kion_aws_cloudformation_template.t1.id
}
```

```hcl
# Create an external compliance check.
resource "kion_compliance_check" "c1" {
  name                     = "sample-resource"
  cloud_provider_id        = 1
  compliance_check_type_id = 1
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
  #   body = <<EOF
  # {
  #     "Version": "2012-10-17",
  #     "Statement": [
  #         {
  #             "Effect": "Allow",
  #             "Action": "*",
  #             "Resource": "*"
  #         }
  #     ]
  # }
  # EOF
}

# Output the ID of the resource created.
output "check_id" {
  value = kion_compliance_check.c1.id
}
```

```hcl
# Create a compliance standard.
resource "kion_compliance_standard" "s1" {
  name               = "sample-resource"
  created_by_user_id = 1
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "standard_id" {
  value = kion_compliance_standard.s1.id
}
```

```hcl
# Create a cloud rule.
resource "kion_cloud_rule" "cr1" {
  name        = "sample-resource"
  description = "Sample cloud rule."
  aws_iam_policies { id = 1 }
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "rule_id" {
  value = kion_cloud_rule.cr1.id
}
```

```hcl
# Create a cloud access role on a project.
resource "kion_project_cloud_access_role" "carp1" {
  name                   = "sample-car"
  project_id             = 1
  aws_iam_role_name      = "sample-car"
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = true
  aws_iam_policies { id = 1 }
  aws_iam_permissions_boundary = 1
  future_accounts              = true
  #accounts { id = 1 }
  users { id = 1 }
  user_groups { id = 1 }
}

# Output the ID of the resource created.
output "project_car_id" {
  value = kion_project_cloud_access_role.carp1.id
}
```

```hcl
# Create a cloud access role on an OU.
resource "kion_ou_cloud_access_role" "carou1" {
  name                   = "sample-car"
  ou_id                  = 3
  aws_iam_role_name      = "sample-car"
  web_access             = true
  short_term_access_keys = true
  long_term_access_keys  = true
  aws_iam_policies { id = 628 }
  #aws_iam_permissions_boundary = 1
  users { id = 1 }
  user_groups { id = 1 }
}

# Output the ID of the resource created.
output "ou_car_id" {
  value = kion_ou_cloud_access_role.carou1.id
}
```

```hcl
# Create an OU.
resource "kion_ou" "ou1" {
  name         = "sample-ou"
  description  = "Sample OU."
  parent_ou_id = 0
  permission_scheme_id = 2
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "ou_id" {
  value = kion_ou.ou1.id
}
```

```hcl
# Create a User Group.
resource "kion_user_group" "ug1" {
  name        = "sample-user-group2"
  description = "This is a sample user group."
  idms_id     = 1
  owner_groups { id = 1 }
  owner_users { id = 1 }
  users { id = 1 }
}

# Output the ID of the resource created.
output "user_group_id" {
  value = kion_user_group.ug1.id
}
```

```hcl
# Create an association between a User Group and a SAML IDMS.
resource "kion_saml_group_association" "sa1" {
  assertion_name         = "memberOf"
  assertion_regex        = "^kion-admins"
  idms_id                = 5
  update_on_login        = true
  user_group_id          = 1
}

# Output the ID of the resource created.
output "saml_group_assocation_id" {
  value = kion_saml_group_association.sa1.id
}
```

```hcl
# Create a Project.
resource "kion_project" "p1" {
  ou_id = 1
  name = "Tech Project I"
  description = "This is a sample project."
  permission_scheme_id = 3
  owner_user_ids { id = 1 }
  project_funding {
    amount = 1000
    funding_order = 1
    funding_source_id = 1
    start_datecode = "2021-01"
    end_datecode = "2022-01"
  }
}

# Output the ID of the resource created.
output "project_id" {
  value = kion_project.p1.id
}
```

```hcl
# Create a GCP IAM role.
resource "kion_gcp_iam_role" "gr1" {
  name = "Read permissions"
  description = "Allow user to list & get IAM roles."
  role_permissions = ["iam.roles.get", "iam.roles.list"]
  gcp_role_launch_stage = 4
  owner_users { id = 1 }
}

# Output the ID of the resource created.
output "gcp_role" {
  value = kion_gcp_iam_role.gr1.id
}
```

```hcl
# Create an AWS Service Control Policy.
resource "kion_service_control_policy" "scp1" {
  name = "Test SCP"
  description  = "This is a sample SCP."
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Deny",
      "Action": [
        "config:StopConfigurationRecorder"
      ],
      "Resource": "*"
    }
  ]
}
EOF
  owner_users { id = 1 }
  owner_user_groups  { id = 1 }
}

# Output the ID of the resource created.
output "scp_id" {
  value = kion_service_control_policy.scp1.id
}
```

```hcl
# Create an Azure ARM template.
resource "kion_azure_arm_template" "arm1" {
  name = "tf test"
  description  = "A test Azure ARM template created via our Terraform provider."
  resource_group_name = "3797RGI"
  resource_group_region_id = 41

  # Valid values are either 1 ("incremental") or 2 ("complete")
  # deployment_mode = 1
  deployment_mode = 2
  owner_users { id = 1 }

  template = <<EOF
{
    "$schema": "https://schema.management.azure.com/schemas/2019-04-01/deploymentTemplate.json#",
    "contentVersion": "1.0.0.0",
    "parameters": {},
    "variables": {},
    "resources": [],
    "outputs": {}
}
EOF
}

# Output the ID of the resource created.
output "arm_template_id" {
  value = kion_azure_arm_template.arm1.id
}
```

```hcl
# Create an Azure Role Definition.
resource "kion_azure_role" "ar1" {
  name = "Test Role"
  description = "A test role created by our Terraform provider."
  role_permissions = <<EOF
{
    "actions": [
        "Microsoft.Authorization/roleDefinitions/read"
    ],
    "notActions": [],
    "dataActions": [],
    "notDataActions": []
}
EOF
  owner_users { id = 1 }
  owner_user_groups { id = 1 }
}

# Output the ID of the resource created.
output "ar_id" {
  value = kion_azure_role.ar1.id
}
```

### Data Sources

```hcl
# Declare a data source to get all IAM policies.
data "kion_aws_iam_policy" "p1" {}

# Output the list of all policies.
output "policies" {
  value = data.kion_aws_iam_policy.p1.list
}

# Output the first policy (which by default is the newest policy).
output "first" {
  value = data.kion_aws_iam_policy.p1.list[0]
}

# Output the first policy name.
output "policy_name" {
  value = data.kion_aws_iam_policy.p1.list[0].name
}

# Output a list of all policy names.
output "policy_names" {
  value = data.kion_aws_iam_policy.p1.list.*.name
}

# Output a list of all owner users for all policies.
output "policy_owner_users" {
  value = data.kion_aws_iam_policy.p1.list.*.owner_users
}
```

```hcl
# Declare a data source to get 1 IAM policy that matches the name filter.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "name"
    values = ["SystemReadOnlyAccess"]
  }
}

# Declare a data source to get 2 IAM policies that matches the name filter.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "name"
    values = ["SystemReadOnlyAccess", "test-policy"]
  }
}

# Declare a data source to get 1 IAM policy that matches both of the filters.
# SystemReadOnlyAccess has the id of 1 so only that policy matches all of the filters.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "name"
    values = ["SystemReadOnlyAccess", "test-policy"]
  }

  filter {
    name   = "id"
    values = [1]
  }
}

# Declare a data source to get all IAM policies that matches the owner filter.
# Syntax to filter on an array.
data "kion_aws_iam_policy" "p1" {
  filter {
    name   = "owner_users.id"
    values = ["20"]
  }
}

# Declare a data source to get all IAM policies that matches the id filter.
# Notice that terraform will convert these to strings even though you
# passed in an integer.
data "kion_aws_iam_policy" "p1" {
  filter {
    name = "id"
    values = [1, "3"]
    # Terraform will convert all of these to strings.
    # + values = [
    #     + "1",
    #     + "3",
    #   ]
  }
}

# Declare a data source to get all IAM policies that matches the query.
output "policy_access" {
  value = {
    # Loop through each policy
    for c in data.kion_aws_iam_policy.p1.list :
    # Create a map with a key of: id
    c.id => c
    # Filter out an names that don't match the passed in variable
    if c.name == "SystemReadOnlyAccess"
  }
}
```

### Locals

```hcl
# Declare a data source to get all IAM policies.
data "kion_aws_iam_policy" "p1" {}

# Declare a local variable: local.owners
locals {
  # Owners for multiple policies merged together.
  owners = concat(data.kion_aws_iam_policy.p1.list[0].owner_users, data.kion_aws_iam_policy.p1.list[1].owner_users)
}

# Output the local variable.
output "policy_owner_users" {
  value = local.owners
}
```

## Repository Maintainer: Push to Terraform Registry

To push a new version of this provider to the Terraform Registry:

- In the main branch, update the Makefile with the correct version
- Use the commands below to create a new tag (ensure you change the version number and description):

```bash
git tag -a v0.2.0 -m "Add your description here"
git push origin v0.2.0
```
