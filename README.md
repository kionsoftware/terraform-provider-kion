# terraform-provider-kion <!-- omit in toc -->

The Terraform provider for Kion allows you interact with the Kion API using the Terraform HCL language. Our provider supports creating, updating, reading, and deleting resources. You can also use it to query for resources using filters even if a resource is not created through Terraform.

- [Getting Started](#getting-started)
  - [Importing Resource State](#importing-resource-state)
- [Examples](#examples)
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
      version = "0.3.21"
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

## Examples

For examples of how to use each resource and data source, please see the [examples](examples) directory. The examples are organized by resource type and include:

- Resources:
  - AWS Accounts
  - AWS CloudFormation Templates
  - AWS IAM Policies
  - Azure Accounts
  - Azure ARM Templates
  - Azure Roles
  - Cloud Access Roles
  - Cloud Rules
  - Compliance Checks
  - Compliance Standards
  - Funding Sources
  - GCP Accounts
  - GCP IAM Roles
  - Labels
  - OUs
  - Projects
  - SAML Group Associations
  - Service Control Policies
  - User Groups

- Data Sources:
  - AWS IAM Policies
  - Labels

## Repository Maintainer: Push to Terraform Registry

To push a new version of this provider to the Terraform Registry:

- In the main branch, update the Makefile with the correct version
- Use the commands below to create a new tag (ensure you change the version number and description):

```bash
git tag -a v0.2.0 -m "Add your description here"
git push origin v0.2.0
```
