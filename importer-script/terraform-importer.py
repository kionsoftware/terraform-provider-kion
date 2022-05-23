"""
Kion Terraform Provider Importer

This script imports existing cloud resources into a source control repository
for management by the Terraform Provider.

See the README for usage and optional flags.
"""

import os
import re
import sys
import json
import textwrap
import argparse
import requests
from json.decoder import JSONDecodeError

PARSER = argparse.ArgumentParser(description='Import Cloud Resources into the Repo Module')
PARSER.add_argument('--kion-url', type=str, required=True, help='URL to Kion, without trailing slash. Example: https://kion.example.com')
PARSER.add_argument('--kion-api-key', type=str, help='Kion API key. Can be set via env variable KION_APIKEY or KION_API_KEY instead (preferred).')
PARSER.add_argument('--import-dir', type=str, required=True, help='Path to the root of the target import directory, without trailing slash.')
PARSER.add_argument('--skip-cfts', action='store_true', help='Skip importing AWS CloudFormation templates.')
PARSER.add_argument('--skip-iams', action='store_true', help='Skip importing AWS IAM policies.')
# PARSER.add_argument('--skip-arms', action='store_true', help='Skip importing Azure ARM templates.')
PARSER.add_argument('--skip-azure-policies', action='store_true', help='Skip importing Azure Policies.')
PARSER.add_argument('--skip-azure-roles', action='store_true', help='Skip importing Azure Roles.')
PARSER.add_argument('--skip-project-roles', action='store_true', help='Skip importing Project Cloud Access Roles.')
PARSER.add_argument('--skip-ou-roles', action='store_true', help='Skip importing OU Cloud Access Roles.')
PARSER.add_argument('--skip-cloud-rules', action='store_true', help='Skip importing Cloud Rules.')
PARSER.add_argument('--skip-checks', action='store_true', help='Skip importing Compliance Checks.')
PARSER.add_argument('--skip-standards', action='store_true', help='Skip importing Compliance Standards.')
PARSER.add_argument('--skip-ssl-verify', action='store_true', help='Skip SSL verification. Use if Kion does not have a valid SSL certificate.')
PARSER.add_argument('--overwrite', action='store_true', help='Overwrite existing files during import.')
PARSER.add_argument('--import-aws-managed', action='store_true', help='Import AWS-managed resources (only those that were already imported into Kion).')
PARSER.add_argument('--prepend-id', action='store_true', help='Prepend each resource\'s ID to its filenames. Useful for easily correlating IDs to resources')
PARSER.add_argument('--clone-system-managed', action='store_true', help='Clone system-managed resources. Names of clones will be prefixed using --clone-prefix argument. Ownership of clones will be set with --clone-user-ids and/or --clone-user-group-ids')
PARSER.add_argument('--clone-prefix', type=str, help='A prefix for the name of cloned system-managed resources. Use with --clone-system-managed.')
PARSER.add_argument('--clone-user-ids', nargs='+', type=int, help='Space separated user IDs to set as owner users for cloned resources')
PARSER.add_argument('--clone-user-group-ids', nargs='+', type=int, help='Space separated user group IDs to set as owner user groups for cloned resources')
# PARSER.add_argument('--dry-run', action='store_true', help='Perform a dry run without writing any files.')
# PARSER.add_argument('--sync', action='store_true',help='Sync repository resources into Kion.')
ARGS = PARSER.parse_args()

# validate kion_url
if not ARGS.kion_url:
    sys.exit("Please provide the URL to Kion. Example: --kion-url https://kion.example.com")
# remove trailing slash if found
elif re.compile(".+/$").match(ARGS.kion_url):
    ARGS.kion_url = re.sub(r'/$', '', ARGS.kion_url)

# validate import_dir
if not ARGS.import_dir:
    sys.exit("Please provide the path to the directory in which to import. Example: --import-dir /Users/me/code/repo-module-dir")
# remove trailing slash if found
elif re.compile(".+/$").match(ARGS.import_dir):
    ARGS.import_dir = re.sub(r'/$', '', ARGS.import_dir)

# validate API key
if not ARGS.kion_api_key:
    if os.environ.get('KION_APIKEY'):
        ARGS.kion_api_key = os.environ['KION_APIKEY']
    elif os.environ.get('KION_API_KEY'):
        ARGS.kion_api_key = os.environ['KION_API_KEY']
    else:
        sys.exit("Did not find a Kion API key supplied via CLI argument or environment variable (KION_APIKEY or KION_API_KEY).")

# validate flags related to cloning
if not ARGS.clone_prefix:
    ARGS.clone_prefix = ""
if ARGS.clone_system_managed:

    # validate clone prefix
    if not ARGS.clone_prefix:
        sys.exit("You did not provide a clone prefix value using the --clone-prefix flag.")
    else:
        if not ARGS.clone_prefix.endswith('-') and not ARGS.clone_prefix.endswith("_"):
            sys.exit("Did not find a _ or - in clone prefix.")

    # validate clone-user-ids and clone-user-group-ids
    if not ARGS.clone_user_ids and not ARGS.clone_user_group_ids:
        sys.exit("You must provide at least one of --clone-user-ids or --clone-user-group-ids in order to import system-managed resources.")
    else:
        if not ARGS.clone_user_ids:
            ARGS.clone_user_ids = []
        if not ARGS.clone_user_group_ids:
            ARGS.clone_user_group_ids = []

BASE_URL = "%s/api" % ARGS.kion_url
HEADERS = {"accept": "application/json", "Authorization": "Bearer " + ARGS.kion_api_key}

MAX_UNAUTH_RETRIES = 15
UNAUTH_RETRY_COUNTER = 0

RESOURCE_PREFIX     = 'kion'
IMPORTED_MODULES    = []
IMPORTED_RESOURCES  = []

PROVIDER_TEMPLATE = textwrap.dedent('''\
    terraform {
        required_providers {
            kion = {
                source  = "kionsoftware/kion"
                version = "0.3.0"
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
OBJECT_API_MAP = {
    'aws_cloudformation_templates': {
        'GET': 'v3/cft',
        'POST': 'v3/cft'
    },
    'aws_iam_policies': {
        'GET': 'v3/iam-policy',
        'POST': 'v3/iam-policy'
    },
    'azure_arm_template_definitions': {
        'GET': 'v4/azure-arm-template',
        'POST': 'v3/azure-arm-template'
    },
    'azure_policy_definitions': {
        'GET': 'v3/azure-policy',
        'POST': 'v3/azure-policy'
    },
    'azure_role_definitions': {
        'GET': 'v3/azure-role',
        'POST': 'v3/azure-role'
    },
    'compliance_standards': {
        'GET': 'v3/compliance/standard',
        'POST': 'v3/compliance/standard'
    },
    'compliance_checks': {
        'GET': 'v3/compliance/check',
        'POST': 'v3/compliance/check'
    },
    'internal_aws_amis': {
        'GET': 'v3/ami',
        'POST': 'v3/ami'
    },
    'internal_aws_service_catalog_portfolios': {
        'GET': 'v3/service-catalog',
        'POST': 'v3/service-catalog'
    },
    'ous': {
        'GET': 'v3/ou',
        'POST': 'v3/ou'
    },
    'owner_user_groups': {
        'GET': 'v3/user-group',
        'POST': 'v3/user-group'
    },
    'owner_users': {
        'GET': 'v3/user',
        'POST': 'v3/user'
    },
    'service_control_policies': {
        'GET': 'v3/service-control-policy',
        'POST': 'v3/service-control-policy'
    },
    'cloud_rules': {
        'GET': 'v3/cloud-rule',
        'POST': 'v3/cloud-rule'
    }
}

def main():
    """
    Main Function

    All processing occurs here.
    """

    # Run some validations prior to starting
    validate_connection(ARGS.kion_url)
    validate_import_dir(ARGS.import_dir)

    print("\nBeginning import from %s" % ARGS.kion_url)

    if not ARGS.skip_cfts:
        import_cfts()
    else:
        print("\nSkipping AWS CloudFormation Templates")

    if not ARGS.skip_iams:
        import_iams()
    else:
        print("\nSkipping AWS IAM Policies")

    # ARMs cannot be cloned. Creating a new ARM requires setting a Resource Group
    # which we won't know
    # if not ARGS.skip_arms:
    #     import_arms()
    # else:
    #     print("\nSkipping Azure ARM Templates")

    if not ARGS.skip_azure_policies:
        import_azure_policies()
    else:
        print("\nSkipping Azure Policies")

    if not ARGS.skip_azure_roles:
        import_azure_roles()
    else:
        print("\nSkipping Azure Roles")

    if not ARGS.skip_project_roles:
        import_project_roles()
    else:
        print("\nSkipping Project Cloud Access Roles")

    if not ARGS.skip_ou_roles:
        import_ou_roles()
    else:
        print("\nSkipping OU Cloud Access Roles")

    if not ARGS.skip_cloud_rules:
        import_cloud_rules()
    else:
        print("\nSkipping Cloud Rules")

    if not ARGS.skip_checks:
        import_compliance_checks()
    else:
        print("\nSkipping Compliance Checks")

    if not ARGS.skip_standards:
        import_compliance_standards()
    else:
        print("\nSkipping Compliance Standards")

    # now out of the loop, write the main provider.tf file
    # to pull in the modules
    provider_filename = "%s/main.tf" % ARGS.import_dir
    content = PROVIDER_TEMPLATE

    # loop over all the modules that were imported and
    # add those to main.tf
    for module in IMPORTED_MODULES:
        module_template = textwrap.dedent('''\

            module "{module_name}" {{
                source = "./{module_name}"
            }}
            ''')

        module_content = module_template.format(
            module_name=module
        )
        content += module_content

    write_provider_file(provider_filename, content)
    write_resource_import_script(ARGS, IMPORTED_RESOURCES)
    print("\nIf you need to refresh the terraform state of the imported resources, run:\n")
    print("cd %s ; bash import_resource_state.sh" % ARGS.import_dir)
    print("\nImport finished.")
    sys.exit()


def import_cfts():
    """
    Import CloudFormation Templates

    Handles full process to import CFTs

    Returns:
        success - True
        failure - False
    """
    CFTs = get_objects_or_ids('aws_cloudformation_templates')

    if CFTs:
        print("\nImporting AWS CloudFormation Templates\n--------------------------")
        print("Found %s CFTs" % len(CFTs))
        IMPORTED_MODULES.append("aws-cloudformation-template")

        for c in CFTs:
            # init new cft object
            cft = {}
            c_id                            = c['cft']['id']
            cft['name']                     = process_string(c['cft']['name'])
            cft['description']              = process_string(c['cft']['description'])
            cft['regions']                  = json.dumps(c['cft']['regions'])
            cft['region']                   = c['cft']['region']
            cft['sns_arns']                 = process_string(c['cft']['sns_arns'])
            cft['template_parameters']      = c['cft']['template_parameters'].rstrip()
            cft['termination_protection']   = c['cft']['termination_protection']
            cft['owner_user_ids']           = []
            cft['owner_user_group_ids']     = []
            cft['policy']                   = c['cft']['policy'].rstrip()

            print("Importing CFT - %s" % cft['name'])

            # get owner user and group IDs formatted into required format
            owner_users     = process_owners(c['owner_users'], 'owner_users')
            owner_groups    = process_owners(c['owner_user_groups'], 'owner_user_groups')

            # pre-process some of the data to fit the required format
            cft['sns_params'] = '\n'.join(cft['sns_arns'])

            # double all single dollar signs to be valid for TF format
            cft['policy'] = re.sub(r'\${1}\{', r'$${', cft['policy'])

            if not cft['region']:
                cft['region'] = 'null'

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                      = {id}
                    name                    = "{resource_name}"
                    description             = "{description}"
                    regions                 = {regions}
                    region                  = "{region}"
                    sns_arns                = "{sns_arns}"
                    termination_protection  = {termination_protection}
                    {owner_users}
                    {owner_groups}

                    template_parameters = <<-EOT
                {template_params}
                EOT

                    policy = <<-EOT
                {policy}
                EOT
                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_aws_cloudformation_template" % RESOURCE_PREFIX,
                resource_id=normalize_string(cft['name']),
                id=c_id,
                resource_name=cft['name'],
                description=cft['description'],
                regions=cft['regions'],
                region=cft['region'],
                sns_arns=cft['sns_arns'],
                termination_protection=str(cft['termination_protection']).lower(),
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
                template_params=cft['template_parameters'],
                policy=cft['policy']
            )

            # do some post-processing of the rendered template prior to
            # writing it out
            if not c['cft']['template_parameters']:
                content = re.sub('\s*template_parameters = <<-EOT\n\nEOT', '', content)

            if cft['region'] == "null":
                content = re.sub('\s*region\s*= "null"', '', content)

            # build the file name
            if ARGS.prepend_id:
                base_filename = normalize_string(cft['name'], c_id)
            else:
                base_filename = normalize_string(cft['name'])

            filename = "%s/aws-cloudformation-template/%s.tf" % (ARGS.import_dir, base_filename)

            write_file(filename, process_template(content))

            # add to IMPORTED_RESOURCES
            resource = "module.aws-cloudformation-template.%s_aws_cloudformation_template.%s %s" % (RESOURCE_PREFIX, normalize_string(cft['name']), c_id)
            IMPORTED_RESOURCES.append(resource)

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/aws-cloudformation-template/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing CFTs.")
        return False


def import_iams():
    """
    Import IAM Policies

    Handles full process to import IAM Policies

    Returns:
        success - True
        failure - False
    """
    IAMs = get_objects_or_ids('aws_iam_policies')

    if IAMs:
        print("\nImporting AWS IAM Policies\n--------------------------")
        print("Found %s IAM Policies" % len(IAMs))
        IMPORTED_MODULES.append("aws-iam-policy")

        for i in IAMs:
            aws_managed     = False

            if i['iam_policy']['aws_managed_policy']:
                if not ARGS.import_aws_managed:
                    print("Skipping AWS-managed IAM Policy: %s" % i['iam_policy']['name'])
                    continue
                else:
                    aws_managed = True

            if i['iam_policy']['system_managed_policy']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed IAM Policy: %s" % i['iam_policy']['name'])
                    continue
                else:
                    original_name = i['iam_policy']['name']       # save original name
                    i = i['iam_policy']                           # reset i to the lower-level object key

                    # remove unnecessary fields
                    i.pop('id')
                    i.pop('aws_managed_policy')
                    i.pop('system_managed_policy')

                    # the clone_resource function checks if this object with the updated
                    # name already exists and won't create a clone if it does
                    result, clone = clone_resource('aws_iam_policies', i)
                    if clone:
                        print("Cloning System-managed IAM Policy: %s -> %s" % (original_name, i['name']))
                        i = clone # reset the i object to the new clone
                        owner_users     = process_owners(ARGS.clone_user_ids, "owner_users")
                        owner_groups    = process_owners(ARGS.clone_user_group_ids, "owner_user_groups")
                    else:
                        if result:
                            print("Already found a clone of %s. Skipping." % original_name)
                            continue
                        else:
                            print("An error occurred cloning %s" % original_name)
                            continue
            else:
                print("Importing IAM Policy - %s" % i['iam_policy']['name'])
                # get owner user and group IDs formatted into required format
                owner_users     = process_owners(i['owner_users'], 'owner_users')
                owner_groups    = process_owners(i['owner_user_groups'], 'owner_user_groups')

            # init new IAM object
            iam = {}
            i_id                        = i['iam_policy']['id']
            iam['name']                 = process_string(i['iam_policy']['name'])
            iam['description']          = process_string(i['iam_policy']['description'])
            iam['owner_user_ids']       = []
            iam['owner_user_group_ids'] = []
            iam['policy']               = i['iam_policy']['policy'].rstrip()

            # check for IAM path - requires Kion > 2.23
            if 'aws_iam_path' in i:
                iam['aws_iam_path'] = i['aws_iam_path'].strip()
            else:
                iam['aws_iam_path'] = ''

            # double all single dollar signs to be valid for TF format
            iam['policy'] = re.sub(r'\${1}\{', r'$${', iam['policy'])

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id            = {id}
                    name            = "{resource_name}"
                    description     = "{description}"
                    aws_iam_path    = "{aws_iam_path}"
                    {owner_users}
                    {owner_groups}
                    policy = <<-EOT
                {policy}
                EOT

                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_aws_iam_policy" % RESOURCE_PREFIX,
                resource_id=normalize_string(iam['name']),
                id=i_id,
                resource_name=iam['name'],
                description=iam['description'],
                aws_iam_path=iam['aws_iam_path'],
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
                policy=iam['policy']
            )

            # build the base file name
            base_filename = build_filename(iam['name'], aws_managed, ARGS.prepend_id, i_id)

            # if it is not an AWS managed, then set a standard filename
            # and add it to the list of imported resources. Otherwise, add .skip to the filename and
            # don't add it to the list of imported resources
            if not aws_managed:
                filename = "%s/aws-iam-policy/%s.tf" % (ARGS.import_dir, base_filename)

                # add to IMPORTED_RESOURCES
                resource = "module.aws-iam-policy.%s_aws_iam_policy.%s %s" % (RESOURCE_PREFIX, normalize_string(iam['name']), i_id)
                IMPORTED_RESOURCES.append(resource)
            else:
                filename = "%s/aws-iam-policy/%s.tf.skip" % (ARGS.import_dir, base_filename)

            # write the file
            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/aws-iam-policy/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing IAM policies.")
        return False


def import_project_roles():
    """
    Import Project Cloud Access Roles

    Handles full process to import Project Cloud Access Roles

    Returns:
        success - True
    """
    print("\nImporting Project Cloud Access Roles\n--------------------------")

    base_path = ARGS.import_dir + "/project-cloud-access-role/"

    # first we need a list of all projects in Kion
    all_projects = get_projects()

    if all_projects:
        IMPORTED_MODULES.append("project-cloud-access-role")

        incomplete_projs = []

        for proj in all_projects:
            proj_id = proj['id']

            # get the normalized name for this project
            proj_name = normalize_string(proj['name'])

            # create the directory name for this project
            # it will be of the format {ID}-{normalized_name}
            # Ex: 13-Test_Project
            proj_dir = "%s-%s" % (str(proj_id), proj_name)

            # get the list of roles for this project
            url = "%s/v3/project/%s/project-cloud-access-role" % (BASE_URL, proj_id)
            roles = api_call(url)

            if roles:
                print("Found %s roles on Project (ID: %s) %s" % (len(roles), proj_id, proj['name']))

                # create a folder for this project to keep all roles together
                if not os.path.isdir(base_path+proj_dir):
                    os.mkdir(base_path+proj_dir)

                for r in roles:
                    # pull out the id for this role
                    r_id = r['id']

                    # init new role object
                    role = {}
                    role['name'] = r['name']
                    role['aws_iam_role_name'] = r['aws_iam_role_name']
                    role['project_id'] = proj_id
                    role['aws_iam_policies'] = []
                    role['user_ids'] = []
                    role['user_group_ids'] = []
                    role['account_ids'] = []
                    role['future_accounts'] = r['future_accounts']
                    role['long_term_access_keys'] = r['long_term_access_keys']
                    role['short_term_access_keys'] = r['short_term_access_keys']
                    role['web_access'] = r['web_access']

                    # get extra metadata that we need
                    url = "%s/v3/project-cloud-access-role/%s" % (BASE_URL, r_id)
                    proj_details = api_call(url)
                    if proj_details:

                      if 'aws_iam_policies' in proj_details and proj_details['aws_iam_policies'] is not None:
                        for i in proj_details['aws_iam_policies']:
                            role['aws_iam_policies'].append(i['id'])

                      if 'users' in proj_details and proj_details['users'] is not None:
                        for u in proj_details['users']:
                            role['user_ids'].append(u['id'])

                      if 'user_groups' in proj_details and proj_details['user_groups'] is not None:
                        for g in proj_details['user_groups']:
                            role['user_group_ids'].append(g['id'])

                      if 'accounts' in proj_details and proj_details['accounts'] is not None:
                        for a in proj_details['accounts']:
                            role['account_ids'].append(a['id'])

                      if 'aws_iam_permissions_boundary' in proj_details:
                        # sys.exit(json.dumps(proj_details))
                        if proj_details['aws_iam_permissions_boundary'] is not None:
                          role['aws_iam_permissions_boundary'] = proj_details['aws_iam_permissions_boundary']['id']
                        else:
                          role['aws_iam_permissions_boundary'] = 'null'
                    else:
                        print("\tDetails for Project role %s weren't found. Data will be incomplete. Skipping" % r['name'])
                        print("\tReceived data: %s" % proj_details)
                        incomplete_projs.append(r['name'])
                        continue

                    # check for iam path
                    if 'aws_iam_path' in r:
                      role['aws_iam_path'] = r['aws_iam_path']
                    else:
                      role['aws_iam_path'] = ''

                    template = textwrap.dedent('''\
                        resource "{resource_type}" "{resource_id}" {{
                            # id                          = {id}
                            name                        = "{resource_name}"
                            project_id                  = {project_id}
                            aws_iam_role_name           = "{aws_iam_role_name}"
                            aws_iam_path                = "{aws_iam_path}"
                            aws_permissions_boundary_id = {aws_perm_boundary}
                            short_term_access_keys      = {short_term_access_keys}
                            long_term_access_keys       = {long_term_access_keys}
                            web_access                  = {web_access}
                            future_accounts             = {future_accounts}
                            {aws_iam_policies}
                            {accounts}
                            {users}
                            {user_groups}
                        }}

                        output "{resource_id}" {{
                            value = {resource_type}.{resource_id}.id
                        }}''')

                    content = template.format(
                        resource_type="%s_project_cloud_access_role" % RESOURCE_PREFIX,
                        resource_id=normalize_string(role['name']),
                        resource_name=role['name'],
                        id=r_id,
                        project_id=proj_id,
                        aws_iam_role_name=role['aws_iam_role_name'],
                        aws_iam_path=role['aws_iam_path'],
                        aws_perm_boundary=role['aws_iam_permissions_boundary'],
                        short_term_access_keys=str(role['short_term_access_keys']).lower(),
                        long_term_access_keys=str(role['long_term_access_keys']).lower(),
                        web_access=str(role['web_access']).lower(),
                        future_accounts=str(role['future_accounts']).lower(),
                        aws_iam_policies=process_list(role['aws_iam_policies'], "aws_iam_policies"),
                        accounts=process_list(role['account_ids'], "accounts"),
                        users=process_list(role['user_ids'], "users"),
                        user_groups=process_list(role['user_group_ids'], "user_groups"),
                    )

                    # write the metadata file
                    if ARGS.prepend_id:
                        base_filename = normalize_string(role['name'], r_id)
                    else:
                        base_filename = normalize_string(role['name'])

                    filename = "%s/project-cloud-access-role/%s/%s.tf" % (ARGS.import_dir, proj_dir, base_filename)

                    write_file(filename, process_template(content))

                    # add to IMPORTED_RESOURCES
                    resource = "module.project-cloud-access-role.%s_project_cloud_access_role.%s %s" % (RESOURCE_PREFIX, normalize_string(role['name']), r_id)
                    IMPORTED_RESOURCES.append(resource)

            else:
                print("Found 0 roles on Project (ID: %s) %s" % (proj_id, proj['name']))

    if incomplete_projs != []:
        print("Project roles that failed to return full details:")
        for p in incomplete_projs:
            print(p)

    # now out of the loop, write the provider.tf file
    provider_filename = "%s/project-cloud-access-role/provider.tf" % ARGS.import_dir
    write_provider_file(provider_filename, PROVIDER_TEMPLATE)

    print("Done.")
    return True


def import_ou_roles():
    """
    Import OU Cloud Access Roles

    Handles full process to import OU Access Roles

    Returns:
        True
    """
    print("\nImporting OU Cloud Access Roles\n--------------------------")

    base_path = ARGS.import_dir + "/ou-cloud-access-role/"

    # first we need a list of all ous in Kion
    all_ous = get_objects_or_ids('ous')

    if all_ous:
        IMPORTED_MODULES.append("ou-cloud-access-role")

        incomplete_ous = []

        for ou in all_ous:
            ou_id = ou['id']

            # get the normalized name for this OU
            ou_name = normalize_string(ou['name'])

            # create the directory name for this OU
            # it will be of the format {ID}-{normalized_name}
            # Ex: 13-Test_OU
            ou_dir = "%s-%s" % (str(ou_id), ou_name)

            # get the list of roles for this ou
            roles = get_ou_roles(ou_id)

            if roles:
                print("Found %s roles on OU (ID: %s) %s" % (len(roles), ou_id, ou['name']))

                # create a folder for this ou to keep all roles together
                if not os.path.isdir(base_path+ou_dir):
                    os.mkdir(base_path+ou_dir)

                for r in roles:
                    # pull out the id for this role
                    r_id = r['id']

                    # init new role object
                    role = {}
                    role['name'] = r['name']
                    role['aws_iam_role_name'] = r['aws_iam_role_name']
                    role['ou_id'] = ou_id
                    role['aws_iam_policies'] = []
                    role['user_ids'] = []
                    role['user_group_ids'] = []

                    # get extra metadata that we need
                    url = "%s/v3/ou-cloud-access-role/%s" % (BASE_URL, r_id)
                    ou_details = api_call(url)
                    if ou_details:
                        role['long_term_access_keys'] = ou_details['ou_cloud_access_role']['long_term_access_keys']
                        role['short_term_access_keys'] = ou_details['ou_cloud_access_role']['short_term_access_keys']
                        role['web_access'] = ou_details['ou_cloud_access_role']['web_access']

                        if ou_details['aws_iam_policies'] is not None:
                          for i in ou_details['aws_iam_policies']:
                              role['aws_iam_policies'].append(i['id'])

                        if ou_details['users'] is not None:
                          for u in ou_details['users']:
                              role['user_ids'].append(u['id'])

                        if ou_details['user_groups'] is not None:
                          for g in ou_details['user_groups']:
                              role['user_group_ids'].append(g['id'])

                        if 'aws_iam_permissions_boundary' in ou_details:
                          if ou_details['aws_iam_permissions_boundary'] is not None:
                            role['aws_iam_permissions_boundary'] = ou_details['aws_iam_permissions_boundary']['id']
                          else:
                            role['aws_iam_permissions_boundary'] = 'null'
                    else:
                        print("\tDetails for OU role %s weren't found. Data will be incomplete. Skipping" % r['name'])
                        print("\tReceived data: %s" % ou_details)
                        incomplete_ous.append(r['name'])
                        continue

                    # check for iam path
                    if 'aws_iam_path' in r:
                      role['aws_iam_path'] = r['aws_iam_path']
                    else:
                      role['aws_iam_path'] = ''

                    template = textwrap.dedent('''\
                        resource "{resource_type}" "{resource_id}" {{
                            # id                          = {id}
                            name                        = "{resource_name}"
                            ou_id                       = {ou_id}
                            aws_iam_role_name           = "{aws_iam_role_name}"
                            aws_iam_path                = "{aws_iam_path}"
                            aws_permissions_boundary_id = {aws_perm_boundary}
                            short_term_access_keys      = {short_term_access_keys}
                            long_term_access_keys       = {long_term_access_keys}
                            web_access                  = {web_access}
                            {aws_iam_policies}
                            {users}
                            {user_groups}
                        }}

                        output "{resource_id}" {{
                            value = {resource_type}.{resource_id}.id
                        }}''')

                    content = template.format(
                        resource_type="%s_ou_cloud_access_role" % RESOURCE_PREFIX,
                        resource_id=normalize_string(role['name']),
                        resource_name=role['name'],
                        id=r_id,
                        ou_id=ou_id,
                        aws_iam_role_name=role['aws_iam_role_name'],
                        aws_iam_path=role['aws_iam_path'],
                        aws_perm_boundary=role['aws_iam_permissions_boundary'],
                        short_term_access_keys=str(role['short_term_access_keys']).lower(),
                        long_term_access_keys=str(role['long_term_access_keys']).lower(),
                        web_access=str(role['web_access']).lower(),
                        aws_iam_policies=process_list(role['aws_iam_policies'], "aws_iam_policies"),
                        users=process_list(role['user_ids'], "users"),
                        user_groups=process_list(role['user_group_ids'], "user_groups"),
                    )

                    # write the metadata file
                    if ARGS.prepend_id:
                        base_filename = normalize_string(role['name'], r_id)
                    else:
                        base_filename = normalize_string(role['name'])

                    filename = "%s/ou-cloud-access-role/%s/%s.tf" % (ARGS.import_dir, ou_dir, base_filename)

                    write_file(filename, process_template(content))

                    # add to IMPORTED_RESOURCES
                    resource = "module.ou-cloud-access-role.%s.%s_ou_cloud_access_role.%s %s" % (RESOURCE_PREFIX, ou_dir, normalize_string(role['name']), r_id)
                    IMPORTED_RESOURCES.append(resource)

            else:
                print("Found 0 roles on OU (ID: %s) %s" % (ou_id, ou['name']))

    if incomplete_ous != []:
        print("OU roles that failed to return full details:")
        for p in incomplete_ous:
            print(p)

    # now out of the loop, write the provider.tf file
    provider_filename = "%s/ou-cloud-access-role/provider.tf" % ARGS.import_dir
    write_provider_file(provider_filename, PROVIDER_TEMPLATE)

    print("Done.")
    return True


def import_cloud_rules():
    """
    Import Cloud Rules

    Handles full process to import Cloud Rules

    Returns:
        True
    """
    print("\nImporting Cloud Rules\n--------------------------")
    cloud_rules = get_objects_or_ids('cloud_rules')

    if cloud_rules:
        print("Found %s Cloud Rules" % len(cloud_rules))
        IMPORTED_MODULES.append("cloud-rule")

        # now loop over them and get the CFT and IAM policy associations
        for c in cloud_rules:
            system_managed = False

            # skip the built_in rules unless toggled on
            if c['built_in']:
                if not ARGS.clone_system_managed:
                    print("Skipping built-in Cloud Rule: %s" % c['name'])
                    continue
                else:
                    system_managed = True

            # get the the cloud rule's metadata
            cloud_rule = get_objects_or_ids("cloud_rules", False, c['id'])

            if cloud_rule:
                c['azure_arm_template_definition_ids']      = get_objects_or_ids('azure_arm_template_definitions', cloud_rule)
                c['azure_policy_definition_ids']            = get_objects_or_ids('azure_policy_definitions', cloud_rule)
                c['azure_role_definition_ids']              = get_objects_or_ids('azure_role_definitions', cloud_rule)
                c['cft_ids']                                = get_objects_or_ids('aws_cloudformation_templates', cloud_rule)
                c['compliance_standard_ids']                = get_objects_or_ids('compliance_standards', cloud_rule)
                c['iam_policy_ids']                         = get_objects_or_ids('aws_iam_policies', cloud_rule)
                c['internal_ami_ids']                       = get_objects_or_ids('internal_aws_amis', cloud_rule)
                c['ou_ids']                                 = get_objects_or_ids('ous', cloud_rule)
                c['internal_portfolio_ids']                 = get_objects_or_ids('internal_aws_service_catalog_portfolios', cloud_rule)
                c['project_ids']                            = get_projects(cloud_rule)
                c['service_control_policy_ids']             = get_objects_or_ids('service_control_policies', cloud_rule)
            else:
                print("Failed getting Cloud Rule details.")

            # now that we have all these details, go through cloning process if this is a built-in cloud rule
            if system_managed:
                original_name = c['name']                       # save original name
                # the clone_resource function checks if this object with the updated
                # name already exists and won't create a clone if it does
                result, clone = clone_resource('cloud_rules', c)
                if clone:
                    print("Cloning System-managed Cloud Rule: %s -> %s" % (original_name, clone['cloud_rule']['name']))
                    c['id']         = clone['cloud_rule']['id']
                    c['name']       = clone['cloud_rule']['name']
                    owner_users     = process_owners(ARGS.clone_user_ids, "owner_users")
                    owner_groups    = process_owners(ARGS.clone_user_group_ids, "owner_user_groups")
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." % original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
            else:
                print("Importing Cloud Rule - %s" % c['name'])
                # get owner user and group IDs formatted into required format
                owner_users     = process_owners(cloud_rule['owner_users'], 'owner_users')
                owner_groups    = process_owners(cloud_rule['owner_user_groups'], 'owner_user_groups')

            for i in ["pre_webhook_id", "post_webhook_id"]:
                if c[i] is None:
                    c[i] = 'null'

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                                    = {id}
                    name                                    = "{resource_name}"
                    description                             = "{description}"
                    pre_webhook_id                          = {pre_webhook_id}
                    post_webhook_id                         = {post_webhook_id}
                    {aws_iam_policies}
                    {cfts}
                    {azure_arm_template_definitions}
                    {azure_policy_definitions}
                    {azure_role_definitions}
                    {compliance_standards}
                    {amis}
                    {portfolios}
                    {scps}
                    {ous}
                    {projects}
                    {owner_users}
                    {owner_groups}
                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_cloud_rule" % RESOURCE_PREFIX,
                resource_id=normalize_string(c['name']),
                resource_name=c['name'],
                id=c['id'],
                description=c['description'],
                pre_webhook_id=c['pre_webhook_id'],
                post_webhook_id=c['post_webhook_id'],
                aws_iam_policies=process_list(c['iam_policy_ids'], "aws_iam_policies"),
                cfts=process_list(c['cft_ids'], "aws_cloudformation_templates"),
                azure_arm_template_definitions=process_list(c['azure_arm_template_definition_ids'], "azure_arm_template_definitions"),
                azure_policy_definitions=process_list(c['azure_policy_definition_ids'], "azure_policy_definitions"),
                azure_role_definitions=process_list(c['azure_role_definition_ids'], "azure_role_definitions"),
                compliance_standards=process_list(c['compliance_standard_ids'], "compliance_standards"),
                amis=process_list(c['internal_ami_ids'], "internal_aws_amis"),
                portfolios=process_list(c['internal_portfolio_ids'], "internal_aws_service_catalog_portfolios"),
                scps=process_list(c['service_control_policy_ids'], "service_control_policies"),
                ous=process_list(c['ou_ids'], "ous"),
                projects=process_list(c['project_ids'], "projects"),
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
            )

            # build the base file name
            base_filename = build_filename(c['name'], False, ARGS.prepend_id, c['id'])
            filename = "%s/cloud-rule/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.cloud-rule.%s_cloud_rule.%s %s" % (RESOURCE_PREFIX, normalize_string(c['name']), c['id'])
            IMPORTED_RESOURCES.append(resource)

            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/cloud-rule/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Cloud Rules.")
        return False


def import_compliance_checks():
    """
    Import Compliance Checks

    Handles full process to import compliance checks

    Returns:
        success - True
        failure - False
    """
    CHECKS = get_comp_checks()

    if CHECKS:
        print("\nImporting Compliance Checks\n--------------------------")
        print("Found %s Compliance Checks" % len(CHECKS))
        IMPORTED_MODULES.append("compliance-check")

        for c in CHECKS:
            system_managed = False

            # skip Kion managed checks unless toggled on
            if c['ct_managed']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed Compliance Check - %s" % c['name'])
                    continue
                else:
                    system_managed = True

            # init new check object
            check = {}
            check['name']                       = process_string(c['name'])
            check['description']                = process_string(c['description'])
            check['regions']                    = c['regions']
            check['azure_policy_id']            = c['azure_policy_id']
            check['cloud_provider_id']          = c['cloud_provider_id']
            check['compliance_check_type_id']   = c['compliance_check_type_id']
            check['severity_type_id']           = c['severity_type_id']
            check['frequency_minutes']          = c['frequency_minutes']
            check['frequency_type_id']          = c['frequency_type_id']
            check['is_all_regions']             = c['is_all_regions']
            check['is_auto_archived']           = c['is_auto_archived']
            check['body']                       = c['body'].rstrip()
            check['owner_user_ids']             = []
            check['owner_user_group_ids']       = []

            # we need to make an additional call to get owner users and groups
            url = "%s/v3/compliance/check/%s" % (BASE_URL, c['id'])
            details = api_call(url)

            # get owner user and group IDs formatted into required format
            if details:
                owner_users     = process_owners(details['owner_users'], 'owner_users')
                owner_groups    = process_owners(details['owner_user_groups'], 'owner_user_groups')
            else:
                print("Failed to get details for check %s" % check['name'])
                print(json.dumps(details))

            # now attempt to clone if importing system-managed resources
            if system_managed:
                original_name = check['name']

                # remove these fields before cloning
                c.pop('id')
                c.pop('ct_managed')

                result, clone = clone_resource('compliance_checks', c)
                if clone:
                    print("Cloning System-managed Compliance Check: %s -> %s" % (original_name, c['name']))
                    c               = clone['compliance_check']
                    check['name']   = process_string(c['name'])     # override this to maintain refs to it later
                    owner_users     = process_owners(ARGS.clone_user_ids, 'owner_users')
                    owner_groups    = process_owners(ARGS.clone_user_group_ids, 'owner_user_groups')
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." % original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
                        continue
            else:
                print("Importing Compliance Check - %s" % c['name'])

            # properly format regions based on contents
            if check['regions'][0] == '':
                check['regions'] = []
            else:
                check['regions'] = json.dumps(check['regions'])

            # calculate minutes based on frequency type
            if check['frequency_type_id'] == int(3):
                check['frequency_minutes'] = check['frequency_minutes'] // 60   # hourly, divide minutes by 60
            elif check['frequency_type_id'] == int(4):
                check['frequency_minutes'] = check['frequency_minutes'] // 1440 # daily, divide minutes by 1440

            # build template based on cloud provider
            # AWS = 1
            # Azure = 2
            # GCP = 3
            if check['cloud_provider_id'] == 1:
                template = textwrap.dedent('''\
                    resource "{resource_type}" "{resource_id}" {{
                        # id                        = {id}
                        name                        = "{resource_name}"
                        description                 = "{description}"
                        created_by_user_id          = {created_by_user_id}
                        regions                     = {regions}
                        cloud_provider_id           = {cloud_provider_id} # 1 = AWS, 2 = Azure, 3 = GCP
                        compliance_check_type_id    = {compliance_check_type_id} # 1 = external, 2 = c7n, 3 = azure, 4 = tenable
                        severity_type_id            = {severity_type_id} # 5 = critical, 4 = high, 3 = medium, 2 = low, 1 = info
                        frequency_minutes           = {frequency_min}
                        frequency_type_id           = {frequency_type_id} # 2 = minutes, 3 = hours, 4 = days
                        is_all_regions              = {is_all_regions}
                        is_auto_archived            = {is_auto_archived}
                        {owner_users}
                        {owner_groups}

                        body = <<-EOT
                    {body}
                    EOT
                    }}

                    output "{resource_id}" {{
                        value = {resource_type}.{resource_id}.id
                    }}''')

                content = template.format(
                    resource_type="%s_compliance_check" % RESOURCE_PREFIX,
                    resource_id=normalize_string(check['name']),
                    resource_name=check['name'],
                    id=c['id'],
                    description=check['description'],
                    created_by_user_id=c['created_by_user_id'],
                    regions=check['regions'],
                    cloud_provider_id=check['cloud_provider_id'],
                    compliance_check_type_id=check['compliance_check_type_id'],
                    severity_type_id=check['severity_type_id'],
                    frequency_min=check['frequency_minutes'],
                    frequency_type_id=check['frequency_type_id'],
                    is_all_regions=str(check['is_all_regions']).lower(),
                    is_auto_archived=str(check['is_auto_archived']).lower(),
                    owner_users='\n    '.join(owner_users),
                    owner_groups='\n    '.join(owner_groups),
                    body=check['body']
                )
            elif check['cloud_provider_id'] == 2:
                if check['compliance_check_type_id'] == 1:
                    template = textwrap.dedent('''\
                        resource "{resource_type}" "{resource_id}" {{
                            # id                        = {id}
                            name                        = "{resource_name}"
                            description                 = "{description}"
                            created_by_user_id          = {created_by_user_id}
                            regions                     = {regions}
                            cloud_provider_id           = {cloud_provider_id} # 1 = AWS, 2 = Azure, 3 = GCP
                            compliance_check_type_id    = {compliance_check_type_id} # 1 = external, 2 = c7n, 3 = azure, 4 = tenable
                            severity_type_id            = {severity_type_id} # 5 = critical, 4 = high, 3 = medium, 2 = low, 1 = info
                            frequency_minutes           = {frequency_min}
                            frequency_type_id           = {frequency_type_id} # 2 = minutes, 3 = hours, 4 = days
                            is_all_regions              = {is_all_regions}
                            is_auto_archived            = {is_auto_archived}
                            {owner_users}
                            {owner_groups}
                        }}

                        output "{resource_id}" {{
                            value = {resource_type}.{resource_id}.id
                        }}''')

                    content = template.format(
                        resource_type="%s_compliance_check" % RESOURCE_PREFIX,
                        resource_id=normalize_string(check['name']),
                        resource_name=check['name'],
                        id=c['id'],
                        description=check['description'],
                        created_by_user_id=c['created_by_user_id'],
                        regions=check['regions'],
                        cloud_provider_id=check['cloud_provider_id'],
                        compliance_check_type_id=check['compliance_check_type_id'],
                        severity_type_id=check['severity_type_id'],
                        frequency_min=check['frequency_minutes'],
                        frequency_type_id=check['frequency_type_id'],
                        is_all_regions=str(check['is_all_regions']).lower(),
                        is_auto_archived=str(check['is_auto_archived']).lower(),
                        owner_users='\n    '.join(owner_users),
                        owner_groups='\n    '.join(owner_groups)
                    )
                elif check['compliance_check_type_id'] == 2:
                    template = textwrap.dedent('''\
                        resource "{resource_type}" "{resource_id}" {{
                            # id                        = {id}
                            name                        = "{resource_name}"
                            description                 = "{description}"
                            created_by_user_id          = {created_by_user_id}
                            regions                     = {regions}
                            cloud_provider_id           = {cloud_provider_id} # 1 = AWS, 2 = Azure, 3 = GCP
                            compliance_check_type_id    = {compliance_check_type_id} # 1 = external, 2 = c7n, 3 = azure, 4 = tenable
                            severity_type_id            = {severity_type_id} # 5 = critical, 4 = high, 3 = medium, 2 = low, 1 = info
                            frequency_minutes           = {frequency_min}
                            frequency_type_id           = {frequency_type_id} # 2 = minutes, 3 = hours, 4 = days
                            is_all_regions              = {is_all_regions}
                            is_auto_archived            = {is_auto_archived}
                            {owner_users}
                            {owner_groups}

                            body = <<-EOT
                        {body}
                        EOT
                        }}

                        output "{resource_id}" {{
                            value = {resource_type}.{resource_id}.id
                        }}''')

                    content = template.format(
                        resource_type="%s_compliance_check" % RESOURCE_PREFIX,
                        resource_id=normalize_string(check['name']),
                        resource_name=check['name'],
                        id=c['id'],
                        description=check['description'],
                        created_by_user_id=c['created_by_user_id'],
                        regions=check['regions'],
                        cloud_provider_id=check['cloud_provider_id'],
                        compliance_check_type_id=check['compliance_check_type_id'],
                        severity_type_id=check['severity_type_id'],
                        frequency_min=check['frequency_minutes'],
                        frequency_type_id=check['frequency_type_id'],
                        is_all_regions=str(check['is_all_regions']).lower(),
                        is_auto_archived=str(check['is_auto_archived']).lower(),
                        owner_users='\n    '.join(owner_users),
                        owner_groups='\n    '.join(owner_groups),
                        body=check['body']
                    )
                elif check['compliance_check_type_id'] == 3:
                    template = textwrap.dedent('''\
                        resource "{resource_type}" "{resource_id}" {{
                            # id                        = {id}
                            name                        = "{resource_name}"
                            description                 = "{description}"
                            created_by_user_id          = {created_by_user_id}
                            regions                     = {regions}
                            azure_policy_id             = {azure_policy_id}
                            cloud_provider_id           = {cloud_provider_id} # 1 = AWS, 2 = Azure, 3 = GCP
                            compliance_check_type_id    = {compliance_check_type_id} # 1 = external, 2 = c7n, 3 = azure, 4 = tenable
                            severity_type_id            = {severity_type_id} # 5 = critical, 4 = high, 3 = medium, 2 = low, 1 = info
                            frequency_minutes           = {frequency_min}
                            frequency_type_id           = {frequency_type_id} # 2 = minutes, 3 = hours, 4 = days
                            is_all_regions              = {is_all_regions}
                            is_auto_archived            = {is_auto_archived}
                            {owner_users}
                            {owner_groups}
                        }}

                        output "{resource_id}" {{
                            value = {resource_type}.{resource_id}.id
                        }}''')

                    content = template.format(
                        resource_type="%s_compliance_check" % RESOURCE_PREFIX,
                        resource_id=normalize_string(check['name']),
                        resource_name=check['name'],
                        id=c['id'],
                        description=check['description'],
                        created_by_user_id=c['created_by_user_id'],
                        regions=check['regions'],
                        azure_policy_id=check['azure_policy_id'],
                        cloud_provider_id=check['cloud_provider_id'],
                        compliance_check_type_id=check['compliance_check_type_id'],
                        severity_type_id=check['severity_type_id'],
                        frequency_min=check['frequency_minutes'],
                        frequency_type_id=check['frequency_type_id'],
                        is_all_regions=str(check['is_all_regions']).lower(),
                        is_auto_archived=str(check['is_auto_archived']).lower(),
                        owner_users='\n    '.join(owner_users),
                        owner_groups='\n    '.join(owner_groups)
                    )
                else:
                    print("Unhandled compliance_check_type_id: %s" % check['compliance_check_type_id'])
            elif check['cloud_provider_id'] == 3:
                template = textwrap.dedent('''\
                    resource "{resource_type}" "{resource_id}" {{
                        # id                        = {id}
                        name                        = "{resource_name}"
                        description                 = "{description}"
                        created_by_user_id          = {created_by_user_id}
                        regions                     = {regions}
                        cloud_provider_id           = {cloud_provider_id} # 1 = AWS, 2 = Azure, 3 = GCP
                        compliance_check_type_id    = {compliance_check_type_id} # 1 = external, 2 = c7n, 3 = azure, 4 = tenable
                        severity_type_id            = {severity_type_id} # 5 = critical, 4 = high, 3 = medium, 2 = low, 1 = info
                        frequency_minutes           = {frequency_min}
                        frequency_type_id           = {frequency_type_id} # 2 = minutes, 3 = hours, 4 = days
                        is_all_regions              = {is_all_regions}
                        is_auto_archived            = {is_auto_archived}
                        {owner_users}
                        {owner_groups}

                        body = <<-EOT
                    {body}
                    EOT
                    }}

                    output "{resource_id}" {{
                        value = {resource_type}.{resource_id}.id
                    }}''')

                content = template.format(
                    resource_type="%s_compliance_check" % RESOURCE_PREFIX,
                    resource_id=normalize_string(check['name']),
                    resource_name=check['name'],
                    id=c['id'],
                    description=check['description'],
                    created_by_user_id=c['created_by_user_id'],
                    regions=check['regions'], ensure_ascii=True,
                    cloud_provider_id=check['cloud_provider_id'],
                    compliance_check_type_id=check['compliance_check_type_id'],
                    severity_type_id=check['severity_type_id'],
                    frequency_min=check['frequency_minutes'],
                    frequency_type_id=check['frequency_type_id'],
                    is_all_regions=str(check['is_all_regions']).lower(),
                    is_auto_archived=str(check['is_auto_archived']).lower(),
                    owner_users='\n    '.join(owner_users),
                    owner_groups='\n    '.join(owner_groups),
                    body=check['body']
                )
            else:
                print("Skipping. Unhandled cloud_provider_id found: %s" % check['cloud_provider_id'])
                continue

            # for external and tenable checks, remove the empty body block
            if check['compliance_check_type_id'] == 1 or check['compliance_check_type_id'] == 4:
                content = re.sub(r'\s*body = <<-EOT\n\nEOT', '', content, re.MULTILINE)

            # build the base file name
            base_filename = build_filename(check['name'], False, ARGS.prepend_id, c['id'])
            filename = "%s/compliance-check/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.compliance-check.%s_compliance_check.%s %s" % (RESOURCE_PREFIX, normalize_string(check['name']), c['id'])
            IMPORTED_RESOURCES.append(resource)

            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/compliance-check/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Compliance Checks.")
        return False


def import_compliance_standards():
    """
    Import Compliance Standards

    Handles full process to import compliance standards

    Returns:
        success - True
        failure - False
    """
    STANDARDS = get_objects_or_ids('compliance_standards')

    if STANDARDS:
        print("\nImporting Compliance Standards\n--------------------------")
        print("Found %s Compliance Standards" % len(STANDARDS))
        IMPORTED_MODULES.append("compliance-standard")

        for s in STANDARDS:
            system_managed = False

            # skip system managed standards unless toggled on
            if (s['ct_managed'] or s['created_by_user_id'] == 0) and not s['name'].startswith(ARGS.clone_prefix):
                if not ARGS.clone_system_managed:
                    print("Skipping built-in Compliance Standard - %s" % s['name'])
                    continue
                else:
                    system_managed = True

            # init new object
            standard = {}
            standard['name']                    = process_string(s['name'])
            standard['checks']                  = []
            standard['owner_user_ids']          = []
            standard['owner_user_group_ids']    = []
            standard['description']             = ''
            standard['created_by_user_id']      = ''

            # we need to make an additional call to get attached checks, owner users and groups
            url = "%s/v3/compliance/standard/%s" % (BASE_URL, s['id'])
            details = api_call(url)
            if details:
                if 'description' in details['compliance_standard']:
                    standard['description'] = details['compliance_standard']['description']
                if 'created_by_user_id' in details['compliance_standard']:
                    standard['created_by_user_id'] = details['compliance_standard']['created_by_user_id']
                for c in details['compliance_checks']:
                    standard['checks'].append(c['id'])

            if system_managed:
                original_name = standard['name']

                result, clone = clone_resource('compliance_standards', standard)
                if clone:
                    print("Cloning System-managed Compliance Standard: %s -> %s" % (original_name, clone['compliance_standard']['name']))
                    standard['name']                = clone['compliance_standard']['name']
                    standard['created_by_user_id']  = clone['compliance_standard']['created_by_user_id']
                    s['id']                         = clone['compliance_standard']['id']
                    owner_users                     = process_owners(ARGS.clone_user_ids, 'owner_users')
                    owner_groups                    = process_owners(ARGS.clone_user_group_ids, 'owner_user_groups')
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." % original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
                        continue
            else:
                print("Importing Compliance Standard - %s" % s['name'])
                owner_users     = process_owners(details['owner_users'], 'owner_users')
                owner_groups    = process_owners(details['owner_user_groups'], 'owner_user_groups')

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                        = {id}
                    name                        = "{resource_name}"
                    description                 = "{description}"
                    created_by_user_id          = {created_by_user_id}
                    {owner_users}
                    {owner_groups}
                    {compliance_checks}
                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_compliance_standard" % RESOURCE_PREFIX,
                resource_id=normalize_string(standard['name']),
                id=s['id'],
                resource_name=standard['name'],
                description=standard['description'],
                created_by_user_id=standard['created_by_user_id'],
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
                compliance_checks=process_list(standard['checks'], "compliance_checks")
            )

            # build the file name
            base_filename = build_filename(standard['name'], False, ARGS.prepend_id, s['id'])
            filename = "%s/compliance-standard/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.compliance-standard.%s_compliance_standard.%s %s" % (RESOURCE_PREFIX, normalize_string(standard['name']), s['id'])
            IMPORTED_RESOURCES.append(resource)

            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/compliance-standard/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Compliance Standards.")
        return False


def import_arms():
    """
    Import Azure ARM Templates

    Handles full process to import Azure ARM Templates

    Returns:
        success - True
        failure - False
    """
    ARMs = get_objects_or_ids('azure_arm_template_definitions')

    if ARMs:
        ARMs = ARMs['items']

        print("\nImporting Azure ARM Templates\n--------------------------")
        print("Found %s Azure ARM Templates" % len(ARMs))
        IMPORTED_MODULES.append("azure-arm-template")

        for a in ARMs:
            system_managed = False

            if a['azure_arm_template']['ct_managed']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed Azure ARM Template: %s" % a['azure_arm_template']['name'])
                    continue
                else:
                    system_managed = True

            # init new IAM object
            arm = {}
            a_id                                = a['azure_arm_template']['id']
            arm['name']                         = process_string(a['azure_arm_template']['name'])
            arm['description']                  = process_string(a['azure_arm_template']['description'])
            arm['deployment_mode']              = a['azure_arm_template']['deployment_mode']
            arm['resource_group_name']          = process_string(a['azure_arm_template']['resource_group_name'])
            arm['resource_group_region_id']     = a['azure_arm_template']['resource_group_region_id']
            arm['owner_user_ids']               = []
            arm['owner_user_group_ids']         = []
            arm['template']                     = a['azure_arm_template']['template'].rstrip()
            arm['template_parameters']          = a['azure_arm_template']['template_parameters'].rstrip()
            arm['version']                      = a['azure_arm_template']['version']

            # double all single dollar signs to be valid for TF format
            arm['template'] = re.sub(r'\${1}\{', r'$${', arm['template'])

            if system_managed:
                original_name = arm['name']

                a = a['azure_arm_template']
                a.pop('version')

                a['name']                   = "\"%s\"" % a['name']
                a['description']            = "\"%s\"" % a['description']
                a['resource_group_name']    = "\"%s\"" % a['resource_group_name']

                print("clone1: %s" % json.dumps(a))
                result, clone = clone_resource('azure_arm_template_definitions', a)
                if clone:
                    print("clone: %s" % json.dumps(clone))
                    a_id            = clone['azure_arm_template']['id']
                    arm['name']     = process_string(clone['azure_arm_template']['name'])
                    owner_users     = process_owners(ARGS.clone_user_ids, "owner_users")
                    owner_groups    = process_owners(ARGS.clone_user_group_ids, "owner_user_groups")
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." % original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
            else:
                print("Importing Azure ARM Template - %s" % arm['name'])
                owner_users     = process_owners(a['owner_users'], 'owner_users')
                owner_groups    = process_owners(a['owner_user_groups'], 'owner_user_groups')

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                        = {id}
                    name                        = "{resource_name}"
                    description                 = "{description}"
                    deployment_mode             = {deployment_mode} # 1 = incremental, 2 = complete
                    resource_group_name         = "{resource_group_name}"
                    resource_group_region_id    = {resource_group_region_id}
                    version                     = {version}
                    {owner_users}
                    {owner_groups}
                    template = <<-EOT
                {template}
                EOT

                    template_parameters = <<-EOT
                {template_parameters}
                EOT

                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_azure_arm_template" % RESOURCE_PREFIX,
                resource_id=normalize_string(arm['name']),
                id=a_id,
                resource_name=arm['name'],
                description=arm['description'],
                deployment_mode=arm['deployment_mode'],
                resource_group_name=arm['resource_group_name'],
                resource_group_region_id=arm['resource_group_region_id'],
                version=arm['version'],
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
                template=arm['template'],
                template_parameters=arm['template_parameters']
            )

            # build the base file name
            base_filename = build_filename(arm['name'], False, ARGS.prepend_id, a_id)
            filename = "%s/azure-arm-template/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.azure-arm-template.%s_azure_arm_template.%s %s" % (RESOURCE_PREFIX, normalize_string(arm['name']), a_id)
            IMPORTED_RESOURCES.append(resource)

            # write the file
            write_file(filename, process_template(content))
        # now out of the loop, write the provider.tf file
        provider_filename = "%s/azure-arm-template/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Azure ARM Templates.")
        return False


def import_azure_policies():
    """
    Import Azure Policies

    Handles full process to import Azure Policies

    Returns:
        success - True
        failure - False
    """
    POLICIES = get_objects_or_ids('azure_policy_definitions')

    if POLICIES:

        print("\nImporting Azure Policies\n--------------------------")
        print("Found %s Azure Policies" % len(POLICIES))
        IMPORTED_MODULES.append("azure-policy")

        for p in POLICIES:
            system_managed      = False

            if p['azure_policy']['ct_managed']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed Azure Policy: %s" % p['azure_policy']['name'])
                    continue
                else:
                    system_managed = True

            # init new IAM object
            policy = {}
            p_id                                    = p['azure_policy']['id']
            policy['name']                          = p['azure_policy']['name']
            policy['description']                   = process_string(p['azure_policy']['description'])
            policy['azure_managed_policy_def_id']   = p['azure_policy']['azure_managed_policy_def_id']
            policy['owner_user_ids']                = []
            policy['owner_user_group_ids']          = []
            policy['policy']                        = p['azure_policy']['policy'].rstrip()
            policy['parameters']                    = p['azure_policy']['parameters'].rstrip()

            if system_managed:
                original_name = p['azure_policy']['name']

                # # remove unnecessary fields
                p['azure_policy'].pop('azure_managed_policy_def_id', None)

                # set policyType to Custom
                P = json.loads(p['azure_policy']['policy'])
                P['policyType'] = "Custom"
                p['azure_policy']['policy'] = json.dumps(P)

                result, clone = clone_resource('azure_policy_definitions', p)
                if clone:
                    print("Cloning System-managed Azure Policy: %s -> %s" % (original_name, clone['azure_policy']['name']))
                    p_id            = clone['azure_policy']['id']
                    policy['name']  = clone['azure_policy']['name']
                    owner_users     = process_owners(ARGS.clone_user_ids, "owner_users")
                    owner_groups    = process_owners(ARGS.clone_user_group_ids, "owner_user_groups")
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." % original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
                        continue
            else:
                print("Importing Azure Policy - %s" % policy['name'])
                # get owner user and group IDs formatted into required format
                owner_users     = process_owners(p['owner_users'], 'owner_users')
                owner_groups    = process_owners(p['owner_user_groups'], 'owner_user_groups')

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                        = {id}
                    name                        = "{resource_name}"
                    description                 = "{description}"
                    azure_managed_policy_def_id = "{azure_managed_policy_def_id}"
                    {owner_users}
                    {owner_groups}
                    policy = <<-EOT
                {policy}
                EOT

                    parameters = <<-EOT
                {parameters}
                EOT

                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_azure_policy" % RESOURCE_PREFIX,
                resource_id=normalize_string(policy['name']),
                id=p_id,
                resource_name=policy['name'],
                description=policy['description'],
                azure_managed_policy_def_id=policy['azure_managed_policy_def_id'],
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
                policy=policy['policy'],
                parameters=policy['parameters']
            )

            # build the base file name
            base_filename = build_filename(policy['name'], False, ARGS.prepend_id, p_id)
            filename = "%s/azure-policy/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.azure-policy.%s_azure_policy.%s %s" % (RESOURCE_PREFIX, normalize_string(policy['name']), p_id)
            IMPORTED_RESOURCES.append(resource)

            # write the file
            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/azure-policy/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Azure Policies.")
        return False


def import_azure_roles():
    """
    Import Azure Roles

    Handles full process to import Azure Roles

    Returns:
        success - True
        failure - False
    """
    ROLES = get_objects_or_ids('azure_role_definitions')

    if ROLES:

        print("\nImporting Azure Roles\n--------------------------")
        print("Found %s Azure Roles" % len(ROLES))
        IMPORTED_MODULES.append("azure-role")

        for r in ROLES:
            system_managed      = False

            if r['azure_role']['azure_managed_policy']:
                print("Skipping Azure-managed Azure Role: %s" % r['azure_role']['name'])
                continue

            if r['azure_role']['system_managed_policy']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed Azure Role: %s" % r['azure_role']['name'])
                    continue
                else:
                    system_managed = True

            # init new object
            role = {}
            r_id                            = r['azure_role']['id']
            role['name']                    = process_string(r['azure_role']['name'])
            role['description']             = process_string(r['azure_role']['description'])
            role['role_permissions']        = r['azure_role']['role_permissions'].rstrip()
            role['owner_user_ids']          = []
            role['owner_user_group_ids']    = []


            if system_managed:
                original_name = role['name']

                r = r['azure_role']
                result, clone = clone_resource('azure_role_definitions', r)
                if clone:
                    print("Cloning System-managed Azure Role: %s -> %s" % (original_name, clone['azure_role']['name']))
                    role['name']    = clone['azure_role']['name']
                    r_id            = clone['azure_role']['id']
                    owner_users     = process_owners(ARGS.clone_user_ids, "owner_users")
                    owner_groups    = process_owners(ARGS.clone_user_group_ids, "owner_user_groups")
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." % original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
                        continue
            else:
                print("Importing Azure Role - %s" % role['name'])
                owner_users     = process_owners(r['owner_users'], 'owner_users')
                owner_groups    = process_owners(r['owner_user_groups'], 'owner_user_groups')

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                        = {id}
                    name                        = "{resource_name}"
                    description                 = "{description}"
                    {owner_users}
                    {owner_groups}
                    role_permissions = <<-EOT
                {role_permissions}
                EOT

                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_azure_policy" % RESOURCE_PREFIX,
                resource_id=normalize_string(role['name']),
                id=r_id,
                resource_name=role['name'],
                description=role['description'],
                role_permissions=role['role_permissions'],
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
            )

            # build the base file name
            base_filename = build_filename(role['name'], False, ARGS.prepend_id, r_id)
            filename = "%s/azure-role/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.azure-role.%s_azure_role.%s %s" % (RESOURCE_PREFIX, normalize_string(role['name']), r_id)
            IMPORTED_RESOURCES.append(resource)

            # write the file
            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/azure-role/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Azure Roles.")
        return False


def get_projects(cloud_rule=False):
    """
    Get Projects

    Params:
        cloud_rule (dict) - cloud rule object for which to return project IDs where the cloud rule is applied locally
                            if not set, it will return a list of all projects from Kion
    Return:
        success - list of project IDs or project objects (based on cloud_rule param)
        failure - False
    """
    if cloud_rule:
        ids = []

        # pull out the ID of the provided cloud rule
        c_id = cloud_rule['cloud_rule']['id']

        # for each project, call the v3/project/{id}/cloud-rule endpoint
        # to get it's locally applied cloud rules. Then check if
        # c_id is in that list of cloud rules
        for p in cloud_rule['projects']:
            p_id = p['id']
            url = "%s/v3/project/%s/cloud-rule" % (BASE_URL, p_id)
            project_rules = api_call(url)
            if project_rules:
                for rule in project_rules:
                    if rule['id'] == c_id:
                        if p_id not in ids:
                            ids.append(p_id)
        return ids
    else:
        url = '%s/v3/project' % BASE_URL
        projects = api_call(url)
        if projects:
            return projects
        else:
            print("Could not get projects from Kion.")
            return False


def get_ou_roles(ou_id):
    """
    Get OU Roles

    Returns a list of role objects that are assigned locally at OU with the given ID.

    This function is needed because the endpoint "v3/ou/{id}/ou-cloud-access-role"
    returns inherited roles too, which we don't want, so we have to do some extra work
    to only get the local roles.

    Params:
        ou_id (int) - ID of the OU for which to return locally applied roles

    Return:
        success - a list of role objects
        failure - False
    """
    url = "%s/v3/ou/%s/ou-cloud-access-role" % (BASE_URL, ou_id)
    roles = api_call(url)
    ROLES = []
    if roles:
        for role in roles:
            if role['ou_id'] == ou_id:
                ROLES.append(role)
        return ROLES
    else:
        return False


def get_comp_checks(comp_standard=False):
    """
    Get Compliance Checks

    Params:
        comp_standard (dict) -  compliance_standard object for which to return
                                associated compliance standard IDs
                                if not set, it will return a list of all compliance check
                                objects from Kion
    Return:
        ids (list) - list of compliance standard IDs
    """
    if comp_standard:
        ids = []
        for i in comp_standard['compliance_checks']:
            ids.append(i['id'])
        return ids
    else:
        url = "%s/v3/compliance/check" % BASE_URL
        checks = api_call(url)
        if checks:
            return checks
        else:
            print("Could not get Compliance Checks from Kion.")
            return False


def get_objects_or_ids(object_type, cloud_rule=False, object_id=False):
    """
    Generic helper function to either return all objects of object_type from Kion

    If cloud_rule is set, return a list of IDs of the associated object_type in cloud_rule

    If object_id is set, return only the object of object_type with that ID

    Params:
        object_type     (str)   -   the type of object to get from the cloud rule, or out of Kion
                                    must be one of the keys of OBJECT_API_MAP
        cloud_rule      (dict)  -   the cloud_rule object to return IDs of object_type. If not set, this
                                    function will return all objects of object_type
        object_id       (int)   -   the ID of the individual object to return

    Return:
        Success:
            if cloud_rule:  ids     (list)  - list of IDs of object_type found in cloud_rule
            else:           objects (list)  - list of objects of object_type
        Failure:
            False   (bool)
    """

    if cloud_rule:
        ids = []
        for i in cloud_rule[object_type]:
            ids.append(i['id'])
        return ids
    else:
        api_endpoint = get_api_endpoint(object_type, 'GET')

        if object_id:
          url = "%s/%s/%s" % (BASE_URL, api_endpoint, object_id)
        else:
          url = "%s/%s" % (BASE_URL, api_endpoint)

        objects = api_call(url)
        if objects:
            return objects
        else:
            print("Could not get return from %s endpoint from Kion." % url)
            return False


def clone_resource(resource_type, resource):
    """
    Clones the resource of provided type.
    Makes use of the OBJECT_API_MAP for mapping type -> API endpoint

    Params:
        resource_type        (str)  - the type of resource being cloned
        resource            (dict)  - a dict of the resource's attributes

    Returns:
        If clone was successful:
            True        (bool)
            resource    (dict)  - A dict of the newly cloned resource
        If matching cloned resource was already found:
            True        (bool)
            False       (bool)
        If failure:
            False       (bool)
            False       (bool)
    """

    # first do some preparation for cloning

    # find the name key, ensure its prepended with the clone prefix
    # and save it to a temp variable

    if 'name' in resource:
        if not resource['name'].startswith(ARGS.clone_prefix):
            resource['name'] = f"{ARGS.clone_prefix}{resource['name']}"
            name = resource['name']
    else:

        # some types of resources are structured differently when it comes to creating them.
        # azure_policies for example needs to have a nested key called 'azure_policy' and under
        # that is the name key. most others have the name key at the root level
        other_structures = ['azure_policy']
        for s in other_structures:
            if s in resource:
                if 'name' in resource[s]:
                    if not resource[s]['name'].startswith(ARGS.clone_prefix):
                        resource[s]['name'] = f"{ARGS.clone_prefix}{resource[s]['name']}"
                        name = resource[s]['name']
                    else:
                        name = False

                # remove some fields while were in here
                resource[s].pop('ct_managed', None)
                resource[s].pop('built_in', None)
                resource[s].pop('id', None)

    # validate that we found the name
    if not name:
        print("Couldn't find the name key in %s" % json.dumps(resource))
        return False, False

    # remove some fields
    resource.pop('ct_managed', None)
    resource.pop('built_in', None)
    resource.pop('id', None)

    # at this point, we found the name and made sure it's prefixed with the clone prefix
    # now lets search for a matching resource
    search = search_resource(resource_type, name)

    if search == []:

        # an empty list means there were no resources matching the clone were found
        # so attempt to create it

        # set owner users and groups
        # these keys are inconsistent across the different resource types
        # so just set both
        for key in ['owner_users', 'owner_user_ids']:
            resource[key] = ARGS.clone_user_ids
        for key in ['owner_user_groups', 'owner_user_group_ids']:
            resource[key] = ARGS.clone_user_group_ids

        new_resource = create_resource(resource_type, resource)
        if new_resource:
            return True, new_resource
        else:
            return False, False

    elif search is False:
        # False means an error occurred
        return False, False
    elif isinstance(search, dict):
        # if we got a dict back, it means a matching resource was found
        # we dont need to return this back as the caller already has it
        return True, False


def create_resource(resource_type, resource):
    """
    Creates a new resource of resource_type

    Params:
        resource_type       (str)   - the type of resource to create
        resource            (dict)  - the complete resource object to be created

    Returns:
        If success:
            resource        (dict)  - the newly created resource object
        If failure:
            False           (bool)
    """


    # set up the API URL to hit and make the call
    # this post should create the new cloned resource
    api_endpoint = get_api_endpoint(resource_type, 'POST')
    url = "%s/%s" % (BASE_URL, api_endpoint)
    response = api_call(url, 'post', resource)

    # print("post payload: %s" % json.dumps(resource, indent=2))
    # print("post response: %s" % json.dumps(response))

    if response:
        if 'status' in response:
            if 'record_id' in response:
                resource = get_objects_or_ids(resource_type, False, response['record_id'])
                if resource:
                    return resource
                else:
                    return False
            else:
                print("Didn't receive a record ID when creating resource: %s" % json.dumps(response))
                return False
        else:
            print("Received bad response while creating resource: %s" % json.dumps(response))
            return False
    else:
        print("Failed creating resource: %s" % json.dumps(resource))
        return False


def search_resource(type, terms, match_key='name'):
    """
    Helper function to search Kion for objects of type using provided search terms

    Params:
        type:       (str) - the type of resource to search for
        terms:      (str) - the search terms
        match_key:  (str) - the key to match the terms against. defaults to 'name'

    Return:
        If found:
            item            (dict) - dict of the matching object
        If not found:
            empty list      (list)
        If error:
            False           (bool)
    """

    # maps the type that we receive to the type as it will show up in the search results
    type_map = {
        'aws_iam_policies': 'iam',
        'aws_cloudformation_templates': 'cft',
        'cloud_rules': 'cloud_rule',
        'compliance_checks': 'compliance_check',
        'compliance_standards': 'compliance_standard',
        'azure_role_definitions': 'azure_role',
        'azure_policy_definitions': 'azure_policy',
        'azure_arm_template_definitions': 'arm_template'
    }

    if type not in type_map.keys():
        print("Received unmapped type: %s" % type)
        return False
    else:
        type_match = type_map[type]

    url = "%s/v1/search" % BASE_URL
    payload = {"query": terms}
    response = api_call(url, 'post', payload)

    # print("search response: %s" % json.dumps(response))

    if response == []:
        # an empty list means 0 search results
        return []
    elif not response:
        # a False response means some sort of error
        return False
    elif len(response) > 0:
        # here we have some matches
        # loop over them and compare item[match_key] to search terms
        # return empty list if nothing matches (should find a match)
        for item in response:
            if item['type'] == type_match:
                if item[match_key] == terms:
                    return item
        return []
    else:
        # default to a False return - something went wrong
        return False


def normalize_string(string, id_ = False):
    """
    Normalize String

    Receives a string and normalizes it for proper source control handling

    Params:
        name (str)  - original string
        id_ (int)   - id of the resource. If set, it will be prepended to filename

    Return:
        string (str) - normalized string
    """
    string = re.sub(r'\s', '_', string)                 # replace spaces with underscores
    string = re.sub(r'[^A-Za-z0-9_-]', '', string)      # remove all non alphanumeric characters

    # prepend {ID} if id_ is set
    if id_:
        string = "%s-%s" % (id_, string)

    return string


def write_file(file_name, content):
    """
    Write File

    Writes the given file_name with given content.
    Based on file_type, it will determine how to write the file,
    either as-is or using json.dump.

    Params:
        file_name   (str)           - name of the file to write. Expecting absolute path.
        content     (str)           - content to write to file_name

    Return:
        success - True
    """

    # only proceed if filename doesn't exist yet or the overwrite flag was set
    if not os.path.exists(file_name) or ARGS.overwrite:
        with open(file_name, 'w', encoding='utf8') as outfile:
            outfile.write(content)
    #else:
        #print("Found %s already. Will not overwrite." % file_name)

    return True


def read_file(file, content_type):
    """
    Read File

    Reads in the provided file and returns the content

    Params:
        file (str)          - name of the file to write. Expecting full path.
        content_type (str)  - type of file content (json, yaml, yml)

    Return:
        success - content of file
        failure - False
      """
    with open(file, "r", encoding='utf8') as f:
        if content_type == "json":
            # validate that the file contains json
            try:
                data = json.load(f)
            except:
                print("Failed to load json file %s." % file)
                return False
            else:
                return data


def write_provider_file(file_name, content):
    """
    Write Provider File

    Writes the provider.tf file with provided content

    If the file already exists, it will instead print out a
    provider.tf.example file with the text needed by the
    Kion TF provider

    Params:
        file_name   (str)   - full file name including path to write
        content     (str)   - content to write into the file
    """

    if os.path.exists(file_name):
        file_name = "%s.example" % file_name

    with open(file_name, 'w', encoding='utf8') as outfile:
        outfile.write(content)

    return True


def write_resource_import_script(args, imported_resources):
    """
    Write Resource Import Script

    Writes a bash script to automate imported the current state of
    all resources that were just imported by the script

    Params:
        args                (dict)      - CLI ARGS
        imported_resources  (list)      - the IMPORTED_RESOURCES list
    """

    file_name = "%s/import_resource_state.sh" % args.import_dir
    with open(file_name, 'w', encoding='utf8') as outfile:
        outfile.write("#!/bin/bash\n")
        for line in IMPORTED_RESOURCES:
            outfile.write("terraform import %s\n" % line)

    return True


def api_call(url, method='get', payload=None, headers=None, timeout=30, test=False):
    """
    API Call

    Common helper function for making the API calls needed for this script.

    Params:
        url         (str)   - full URL to call
        method      (str)   - API method - GET or POST
        payload     (dict)  - payload for POST requests
        headers     (dict)  - different headers to use
        timeout     (int)   - timeout for the call, defaults to 10
        test        (bool)  - if true, just test success of response and return
                              True / False accordingly, rather than returning the response data

    Return:
        success - response['data']
        failure - False
    """
    # check for the skip_ssl_verify flag
    if ARGS.skip_ssl_verify:
      verify = False
    else:
      verify = True

    # override headers if set
    if headers:
      _headers = headers
    else:
      _headers = HEADERS

    # make the API call without JSON decoding
    try:
      if method.lower() == 'get':
        response = requests.get(url, headers=_headers, timeout=timeout, verify=verify)
      elif method.lower() == 'post':
        if payload:
            response = requests.post(url, headers=_headers, json=payload, timeout=timeout, verify=verify)
        else:
          response = requests.post(url, headers=_headers, timeout=timeout, verify=verify)
      else:
        print("Unhandled method supplied to api_call function: %s" % method.lower())
        return False
    except (requests.ConnectionError, requests.exceptions.ReadTimeout, requests.exceptions.Timeout) as e:
      print("Request to %s timed out. Error: %s" % (url, e))
      return False
    except requests.exceptions.TooManyRedirects as e:
      print("Connection to %s returned Too Many Redirects error: %s" % (url, e))
      return False
    except requests.exceptions.RequestException as e:
      print("Connection to %s resulted in error: %s" % (url, e))
      return False
    except Exception as e:
      print("Exception occurred during connection to %s: %s" % (url, e))
      return False
    else:

        # at this point, no exceptions were thrown so the
        # the request succeeded

        # check if test is True, if so return True
        if test:
            return True

        # test for valid json response
        try:
            response.json()
        except JSONDecodeError as e:
            print("JSON decode error on response: %s, %s" % (response, e))
            return False
        else:
            response = response.json()

        if response['status'] == 200:
            # reset the unauth retry counter
            global UNAUTH_RETRY_COUNTER
            UNAUTH_RETRY_COUNTER = 0
            return response['data']
        elif response['status'] == 201:
            # 201's are the return code for resource creations
            # and the response object can vary, so just return the whole thing
            # and make the calling function deal with it
            return response
        elif response['status'] == 401:
                # retry up to MAX_UNAUTH_RETRIES
            if UNAUTH_RETRY_COUNTER < MAX_UNAUTH_RETRIES:
                retries = MAX_UNAUTH_RETRIES - UNAUTH_RETRY_COUNTER
                print("Received unauthorized response. Will retry %s more times." % retries)
                UNAUTH_RETRY_COUNTER += 1
                api_call(url)
            else:
                print("Hit max unauth retries.")
                return False
        else:
            print(response['status'])
            print("Error calling API: %s\n%s" % (url, response))
        return False


def get_api_endpoint(resource, method):
    """
    Helper function to return the proper API endpoint for the
    provided resource and method

    Params:
        resource    (str)   - the resource that we are interacting with. Must be defined in OBJECT_API_MAP
        method      (str)   - the method being used to interact with the resource's API

    Returns:
        endpoint    (str)   - the corresponding endpoint
    """
    if resource in OBJECT_API_MAP.keys():
        if method in OBJECT_API_MAP[resource].keys():
            return OBJECT_API_MAP[resource][method]
        else:
            print("Didn't find %s defined for %s in the map." % (method, resource))
            return False
    else:
        print("Didn't find %s defined in the map." % resource)
        return False


def validate_connection(url):
    """
    Validate Connection

    Make sure that the supplied CT URL can be reached

    Params:
        url (str) - the provided kion_url argument

    Returns:
        success - True
        failure - sys.exit
    """

    if api_call(url, 'get', False, False, 30, True):
      return True
    else:
      sys.exit("Unable to connect to %s" % url)


def validate_import_dir(path):
    """
    Validate Import Directory

    Make sure the import directory entered has the proper sub-directories
    for the resources being imported.

    Params:
        path (str) - value of ARGS.import_dir

    Returns:
        success - True
        failure - sys.exit with message
    """
    missing_dirs = []
    dir_map = {
        'aws-cloudformation-template': ARGS.skip_cfts,
        'aws-iam-policy': ARGS.skip_iams,
        'cloud-rule': ARGS.skip_cloud_rules,
        'ou-cloud-access-role': ARGS.skip_ou_roles,
        'project-cloud-access-role': ARGS.skip_project_roles,
        'compliance-check': ARGS.skip_checks,
        'compliance-standard': ARGS.skip_standards,
        # 'azure-arm-template': ARGS.skip_arms,
        'azure-policy': ARGS.skip_azure_policies,
        'azure-role': ARGS.skip_azure_roles
    }

    if os.path.isdir(path):
        for folder, flag in dir_map.items():
            if not flag:
                if not os.path.isdir("%s/%s" % (path, folder)):
                    missing_dirs.append(folder)

        if missing_dirs != []:
            print("%s is missing the following sub-directories:" % path)
            for d in missing_dirs:
                print("- %s" % d)

            create = input("Create them? (y/N) ")
            if create == "y":
                for d in missing_dirs:
                    dir = "%s/%s" % (path,d)
                    os.mkdir(dir)
                print("Done. You can now re-run the previous command.")
                sys.exit()
            else:
                sys.exit()
    else:
        sys.exit("Did not find import directory: %s" % path)

    return True


def process_owners(input, text):
    """
    Process the passed list of owner_users or owner_user_groups into the format
    required for the TF config files

    Param:
        input   (list)  - list of owner user objects returned from Kion, or just IDs
        text    (str)   - text to prepend to each line in output
                            ('owner_users' or 'owner_user_groups')

    Return:
        output (list)       - processed list of owner user IDs
    """
    ids = []
    output = []

    # get all of the user IDs, store in ids
    for i in input:
        if isinstance(i, dict):
            if 'id' in i:
                ids.append(i['id'])
        elif isinstance(i, int):
            ids.append(i)

    if len(ids):
        for i in ids:
            line = "%s { id = %s }" % (text, i)
            output.append(line)
    else:
        output.append("%s { }" % text)
    return output


def process_list(input, text):
    """
    Process a list of IDs into a multi-line list of
    objects as required by TF

    Param:
        input   (list)      - list of IDs to process
        text    (str)       - text to prepend to each line

    Return:
        output  (str)      - multi-line string list of objects
    """

    output = []

    if len(input):
        for i in input:
            line = "%s { id = %s }" % (text, i)
            output.append(line)
    else:
        line = "%s { }" % text
        output.append(line)

    return '\n    '.join(output)


def process_string(input):
    """
    Helper function to handle routine string processing

    Params:
        input   (str)       - the original string
        output  (str)       - the processed string
    """
    # output = re.sub('\\r', '', input)           # replace windows carriage-returns with a space
    output = input.replace("\r", "\\r")
    output = output.replace("\n", "\\n")          # replace newlines with a space
    output = output.replace('"', "'")           # replace double quotes with single quotes
    output = output.replace('\\', '\\\\')       # replace single backslashes with double backslashes
    output = re.sub('\s{2,}', ' ', output)      # replace multiple spaces with a single space
    output = output.strip()                     # strip leading and trailing whitespace
    return output


def process_template(input):
    """
    Helper function to apply some uniform processing
    to rendered templates

    Params:
        input   (str)       - the original string
        output  (str)       - the processed string
    """
    output = re.sub(r'\s*\w+\s+{\s+}', r'', input)
    # output = re.sub('    $', '', output, re.MULTILINE)
    return output


def build_filename(base, aws_managed=False, prepend_id=False, r_id=False):
    """
    Helper funcion to build the filename based on provided parameters

    Params:
        base        (str)   - base name of the file
        aws_managed (bool)  - whether or not this is an AWS-managed resource
        prepend_id  (bool)  - whether or not to prepend the resource ID to the name
        r_id        (str)   - the resource's ID to prepend, if prepend_id is True

    Returns:
        base_filename   (str)   - formatted base filename
    """
    base_filename = base

    if aws_managed:
        base_filename = "AWS_Managed_%s" % base_filename

    if prepend_id:
        if r_id:
            base_filename = normalize_string(base_filename, r_id)
        else:
            print("Error - prepend ID was set to true but the ID was not provided. Will return without the ID prepended.")
            base_filename = normalize_string(base_filename)
    else:
        base_filename = normalize_string(base_filename)

    return base_filename


if __name__ == "__main__":
    main()
