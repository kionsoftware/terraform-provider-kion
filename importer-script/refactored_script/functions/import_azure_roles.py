import build_filename
import normalize_string
import process_string
import process_owners
import get_objects_or_ids
import clone_resource
import textwrap
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from constants import PROVIDER_TEMPLATE


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
            system_managed = False

            if r['azure_role']['azure_managed_policy']:
                print("Skipping Azure-managed Azure Role: %s" %
                      r['azure_role']['name'])
                continue

            if r['azure_role']['system_managed_policy']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed Azure Role: %s" %
                          r['azure_role']['name'])
                    continue
                else:
                    system_managed = True

            # init new object
            role = {}
            r_id = r['azure_role']['id']
            role['name'] = process_string(r['azure_role']['name'])
            role['description'] = process_string(
                r['azure_role']['description'])
            role['role_permissions'] = r['azure_role']['role_permissions'].rstrip()
            role['owner_user_ids'] = []
            role['owner_user_group_ids'] = []

            if system_managed:
                original_name = role['name']

                r = r['azure_role']
                result, clone = clone_resource('azure_role_definitions', r)
                if clone:
                    print("Cloning System-managed Azure Role: %s -> %s" %
                          (original_name, clone['azure_role']['name']))
                    role['name'] = clone['azure_role']['name']
                    r_id = clone['azure_role']['id']
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
                print("Importing Azure Role - %s" % role['name'])
                owner_users = process_owners(r['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    r['owner_user_groups'], 'owner_user_groups')

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
            base_filename = build_filename(
                role['name'], False, ARGS.prepend_id, r_id)
            filename = "%s/azure-role/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.azure-role.%s_azure_role.%s %s" % (
                RESOURCE_PREFIX, normalize_string(role['name']), r_id)
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
