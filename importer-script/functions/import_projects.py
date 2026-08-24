import os
import json
import textwrap
from .api_call import api_call
from .normalize_string import normalize_string
from .owner_app_roles import owner_app_role_ids
from .process_list import process_list
from .get_projects import get_projects
from .write_file import write_file
from .write_provider_file import write_provider_file
from .process_template import process_template
from constants import BASE_URL
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from .templates import PROVIDER_TEMPLATE
from config import ARGS


def import_projects():
    """
    Import Projects

    Handles full process to import Kion Projects as kion_project resources.

    NOTE: This is a best-effort importer. Some required fields (notably
    permission_scheme_id) are not returned by the Kion API, so placeholders
    with TODO comments are emitted and must be corrected before apply.

    Returns:
        success - True
    """
    print("\nImporting Projects\n--------------------------")

    base_path = "%s/project" % ARGS.import_dir

    # first we need a list of all projects in Kion
    all_projects = get_projects()

    if all_projects:
        # make sure the module directory exists
        os.makedirs(base_path, exist_ok=True)

        # only append the module once, and only if we have projects to write
        IMPORTED_MODULES.append("project")

        for item in all_projects:
            proj_id = item['id']

            # get the normalized name used for both the resource id and filename
            normalized_name = normalize_string(item['name'])

            # look up the owners via the project's permission mapping
            owner_user_ids = []
            owner_user_group_ids = []
            url = "%s/v3/project/%s/permission-mapping" % (BASE_URL, proj_id)
            mapping = api_call(url)
            if mapping:
                # Which app role means "owner" is installation-specific — see
                # owner_app_role_ids(). Accumulate across every matching role
                # rather than stopping at the first: an installation may map
                # owners through more than one.
                owner_roles = owner_app_role_ids()
                for entry in mapping:
                    if entry.get('app_role_id') not in owner_roles:
                        continue
                    if entry.get('user_ids'):
                        owner_user_ids.extend(
                            i for i in entry['user_ids'] if i not in owner_user_ids)
                    # NOTE: the API key is spelled 'user_groups_ids'
                    if entry.get('user_groups_ids'):
                        owner_user_group_ids.extend(
                            i for i in entry['user_groups_ids']
                            if i not in owner_user_group_ids)

            # if we could not infer any owners, warn and add a TODO comment
            # so the operator knows kion_project's AtLeastOneOf requirement
            # is not satisfied
            owners_todo = ""
            if not owner_user_ids and not owner_user_group_ids:
                owners_todo = "\n    # TODO: owners could not be inferred from permission mapping - kion_project requires at least one of owner_user_ids/owner_user_group_ids"
                print("\tWARNING: Could not infer owners for Project (ID: %s) %s" % (
                    proj_id, item['name']))

            template = textwrap.dedent('''\
                resource "kion_project" "{resource_id}" {{
                    # id                   = {id}
                    name                 = {name}
                    description          = {description}
                    ou_id                = {ou_id}
                    auto_pay             = {auto_pay}
                    default_aws_region   = {default_aws_region}
                    permission_scheme_id = 3  # TODO: not returned by Kion API - set the correct project permission scheme ID before apply
                    # NOTE: budget/funding not imported - add manually if needed{owners_todo}
                    {owner_user_ids}
                    {owner_user_group_ids}
                }}

                output "{resource_id}" {{
                    value = kion_project.{resource_id}.id
                }}''')

            content = template.format(
                resource_id=normalized_name,
                id=proj_id,
                name=json.dumps(item['name']),
                description=json.dumps(item.get('description') or ""),
                ou_id=item['ou_id'],
                auto_pay=str(bool(item['auto_pay'])).lower(),
                default_aws_region=json.dumps(
                    item.get('default_aws_region') or ""),
                owners_todo=owners_todo,
                owner_user_ids=process_list(owner_user_ids, "owner_user_ids"),
                owner_user_group_ids=process_list(
                    owner_user_group_ids, "owner_user_group_ids"),
            )

            # determine the filename
            if ARGS.prepend_id:
                base_filename = normalize_string(item['name'], proj_id)
            else:
                base_filename = normalize_string(item['name'])

            filename = "%s/%s.tf" % (base_path, base_filename)

            write_file(filename, process_template(content))

            # add to IMPORTED_RESOURCES
            resource = "module.project.kion_project.%s %s" % (
                normalized_name, proj_id)
            IMPORTED_RESOURCES.append(resource)

            print("Imported Project (ID: %s) %s" % (proj_id, item['name']))

        # now out of the loop, write the provider.tf file
        write_provider_file("%s/project/provider.tf" %
                            ARGS.import_dir, PROVIDER_TEMPLATE)

    print("Done.")
    return True
