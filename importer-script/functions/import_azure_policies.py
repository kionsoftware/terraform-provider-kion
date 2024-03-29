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
            system_managed = False

            if p['azure_policy']['ct_managed']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed Azure Policy: %s" %
                          p['azure_policy']['name'])
                    continue
                else:
                    system_managed = True

            # init new IAM object
            policy = {}
            p_id = p['azure_policy']['id']
            policy['name'] = p['azure_policy']['name']
            policy['description'] = process_string(
                p['azure_policy']['description'])
            policy['azure_managed_policy_def_id'] = p['azure_policy']['azure_managed_policy_def_id']
            policy['owner_user_ids'] = []
            policy['owner_user_group_ids'] = []
            policy['policy'] = p['azure_policy']['policy'].rstrip()
            policy['parameters'] = p['azure_policy']['parameters'].rstrip()

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
                    print("Cloning System-managed Azure Policy: %s -> %s" %
                          (original_name, clone['azure_policy']['name']))
                    p_id = clone['azure_policy']['id']
                    policy['name'] = clone['azure_policy']['name']
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
                        continue
            else:
                print("Importing Azure Policy - %s" % policy['name'])
                # get owner user and group IDs formatted into required format
                owner_users = process_owners(p['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    p['owner_user_groups'], 'owner_user_groups')

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
            base_filename = build_filename(
                policy['name'], False, ARGS.prepend_id, p_id)
            filename = "%s/azure-policy/%s.tf" % (
                ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.azure-policy.%s_azure_policy.%s %s" % (
                RESOURCE_PREFIX, normalize_string(policy['name']), p_id)
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
