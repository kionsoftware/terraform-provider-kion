import re
from .build_filename import build_filename
from .normalize_string import normalize_string
from .process_string import process_string
from .process_owners import process_owners
from .get_objects_or_ids import get_objects_or_ids
from .clone_resource import clone_resource
import json
import textwrap
from .write_file import write_file
from .write_provider_file import write_provider_file
from .process_template import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from .templates import PROVIDER_TEMPLATE
from config import ARGS


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
                    print("Skipping System-managed Azure ARM Template: %s" %
                          a['azure_arm_template']['name'])
                    continue
                else:
                    system_managed = True

            # init new IAM object
            arm = {}
            a_id = a['azure_arm_template']['id']
            arm['name'] = process_string(a['azure_arm_template']['name'])
            arm['description'] = process_string(
                a['azure_arm_template']['description'])
            arm['deployment_mode'] = a['azure_arm_template']['deployment_mode']
            arm['resource_group_name'] = process_string(
                a['azure_arm_template']['resource_group_name'])
            arm['resource_group_region_id'] = a['azure_arm_template']['resource_group_region_id']
            arm['owner_user_ids'] = []
            arm['owner_user_group_ids'] = []
            arm['template'] = a['azure_arm_template']['template'].rstrip()
            arm['template_parameters'] = a['azure_arm_template']['template_parameters'].rstrip()
            arm['version'] = a['azure_arm_template']['version']

            # double all single dollar signs to be valid for TF format
            arm['template'] = re.sub(r'\${1}\{', r'$${', arm['template'])

            if system_managed:
                original_name = arm['name']

                a = a['azure_arm_template']
                a.pop('version')

                a['name'] = "\"%s\"" % a['name']
                a['description'] = "\"%s\"" % a['description']
                a['resource_group_name'] = "\"%s\"" % a['resource_group_name']

                print("clone1: %s" % json.dumps(a))
                result, clone = clone_resource(
                    'azure_arm_template_definitions', a)
                if clone:
                    print("clone: %s" % json.dumps(clone))
                    a_id = clone['azure_arm_template']['id']
                    arm['name'] = process_string(
                        clone['azure_arm_template']['name'])
                    owner_users = process_owners(
                        ARGS.clone_user_ids, "owner_users")
                    owner_groups = process_owners(
                        ARGS.clone_user_group_ids, "owner_user_groups")
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." %
                              original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
            else:
                print("Importing Azure ARM Template - %s" % arm['name'])
                owner_users = process_owners(a['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    a['owner_user_groups'], 'owner_user_groups')

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
            base_filename = build_filename(
                arm['name'], False, ARGS.prepend_id, a_id)
            filename = "%s/azure-arm-template/%s.tf" % (
                ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.azure-arm-template.%s_azure_arm_template.%s %s" % (
                RESOURCE_PREFIX, normalize_string(arm['name']), a_id)
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
