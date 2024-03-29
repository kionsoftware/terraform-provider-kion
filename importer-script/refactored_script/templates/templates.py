import textwrap

PROVIDER_TEMPLATE = textwrap.dedent('''\
    terraform {
        required_providers {
            kion = {
                source  = "kionsoftware/kion"
                version = "0.13.12"
            }
        }
    }

    # provider "kion" {
        # Configuration options
    # }
    ''')

MAIN_PROVIDER_TEMPLATE = textwrap.dedent('''\

    module "aws-cloudformation-template" {
        source = "./aws-cloudformation-template"
    }

    module "aws-iam-policy" {
        source = "./aws-iam-policy"
    }

    module "cloud-rule" {
        source = "./cloud-rule"
    }

    module "compliance-check" {
        source = "./compliance-check"
    }

    module "compliance-standard" {
        source = "./compliance-standard"
    }
    ''')

OWNERS_TEMPLATE = textwrap.dedent('''\
    {owner_users}
    ''')

OWNER_GROUPS_TEMPLATE = textwrap.dedent('''\
    {owner_user_groups}
    ''')

OUTPUT_TEMPLATE = textwrap.dedent('''\
    output "{resource_id}" {{
        value = {resource_type}.{resource_id}.id
    }}''')

# this maps the various object types that can be attached to cloud rules
# to the API endpoint for that resource type
