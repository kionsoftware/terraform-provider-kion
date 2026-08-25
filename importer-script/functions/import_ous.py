import os
import json
from .api_call import api_call
from .normalize_string import normalize_string
from .owner_app_roles import owner_app_role_ids
from .process_list import process_list
from .get_objects_or_ids import get_objects_or_ids
import textwrap
from .write_file import write_file
from .write_provider_file import write_provider_file
from .process_template import process_template
from constants import BASE_URL
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from .templates import PROVIDER_TEMPLATE
from config import ARGS


def import_ous():
    """
    Import OUs

    Handles full process to import Kion OUs as kion_ou resources.

    Note: This imports the OUs themselves (kion_ou), which is different from
    import_ou_roles (which imports OU Cloud Access Roles).

    The root OU (parent_ou_id of 0/None) is skipped because it cannot be
    managed as a Terraform child resource.

    Returns:
        True
    """
    print("\nImporting OUs\n--------------------------")

    base_path = "%s/ou" % ARGS.import_dir

    # first we need a list of all ous in Kion
    all_ous = get_objects_or_ids('ous')

    if all_ous:
        # track whether we've registered the module / created the dir yet
        module_registered = False

        for ou in all_ous:
            ou_id = ou['id']
            ou_name = ou['name']

            # skip the root OU - it has no valid parent so it cannot be
            # managed as a Terraform child resource
            parent_ou_id = ou.get('parent_ou_id')
            if not parent_ou_id:
                print("Skipping root OU (ID: %s) %s - cannot be managed via Terraform" % (
                    ou_id, ou_name))
                continue

            # get detail for this OU to pull out owners
            owner_user_ids = []
            owner_user_group_ids = []

            url = "%s/v3/ou/%s" % (BASE_URL, ou_id)
            details = api_call(url)
            if details:
                if details.get('owner_users') is not None:
                    for u in details['owner_users']:
                        owner_user_ids.append(u['id'])

                if details.get('owner_user_groups') is not None:
                    for g in details['owner_user_groups']:
                        owner_user_group_ids.append(g['id'])
            else:
                print(
                    "\tDetails for OU %s weren't found. Owner data will be incomplete." % ou_name)
                print("\tReceived data: %s" % details)

            # /v3/ou/{id} returns owner_users: null for OUs whose ownership is
            # expressed as a permission mapping instead — which kion_ou requires
            # (one of owner_users/owner_user_groups), so without this fallback
            # those OUs import into configuration Terraform rejects.
            if not owner_user_ids and not owner_user_group_ids:
                mapping = api_call("%s/v3/ou/%s/permission-mapping" % (BASE_URL, ou_id))
                if mapping:
                    owner_roles = owner_app_role_ids()
                    for entry in mapping:
                        if entry.get('app_role_id') not in owner_roles:
                            continue
                        if entry.get('user_ids'):
                            owner_user_ids.extend(
                                i for i in entry['user_ids'] if i not in owner_user_ids)
                        if entry.get('user_groups_ids'):
                            owner_user_group_ids.extend(
                                i for i in entry['user_groups_ids']
                                if i not in owner_user_group_ids)

            # register the module and create the directory once, only if we
            # actually have an OU to write
            if not module_registered:
                IMPORTED_MODULES.append("ou")
                os.makedirs(base_path, exist_ok=True)
                module_registered = True

            resource_id = normalize_string(ou_name)

            # guard: kion_ou requires at least one of owner_users /
            # owner_user_groups. This shouldn't happen for a non-root OU,
            # but be safe and flag it.
            owners_todo = ""
            if not owner_user_ids and not owner_user_group_ids:
                owners_todo = "\n    # TODO: no owners returned by API - kion_ou requires at least one of owner_users/owner_user_groups"
                print(
                    "\tWARNING: OU (ID: %s) %s returned no owners - kion_ou requires at least one of owner_users/owner_user_groups" % (ou_id, ou_name))

            template = textwrap.dedent('''\
                resource "kion_ou" "{resource_id}" {{
                    # id                   = {id}{owners_todo}
                    name                 = {name}
                    description          = {description}
                    parent_ou_id         = {parent_ou_id}
                    permission_scheme_id = {permission_scheme_id}
                    {owner_users}
                    {owner_user_groups}
                }}

                output "{resource_id}" {{
                    value = kion_ou.{resource_id}.id
                }}''')

            content = template.format(
                resource_id=resource_id,
                id=ou_id,
                owners_todo=owners_todo,
                name=json.dumps(ou_name),
                description=json.dumps(ou.get('description') or ""),
                parent_ou_id=parent_ou_id,
                permission_scheme_id=ou['permission_scheme_id'],
                owner_users=process_list(owner_user_ids, "owner_users"),
                owner_user_groups=process_list(
                    owner_user_group_ids, "owner_user_groups"),
            )

            # determine the filename
            if ARGS.prepend_id:
                base_filename = normalize_string(ou_name, ou_id)
            else:
                base_filename = normalize_string(ou_name)

            filename = "%s/%s.tf" % (base_path, base_filename)

            write_file(filename, process_template(content))

            # add to IMPORTED_RESOURCES
            resource = "module.ou.kion_ou.%s %s" % (resource_id, ou_id)
            IMPORTED_RESOURCES.append(resource)

            print("Imported OU (ID: %s) %s" % (ou_id, ou_name))

        # only write the provider file if we wrote at least one OU
        if module_registered:
            write_provider_file(
                "%s/ou/provider.tf" % ARGS.import_dir, PROVIDER_TEMPLATE)

    print("Done.")
    return True
