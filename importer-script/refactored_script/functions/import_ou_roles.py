import os
import api_call
import normalize_string
import get_ou_roles
import process_list
import get_objects_or_ids
import textwrap
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from constants import PROVIDER_TEMPLATE


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
                print("Found %s roles on OU (ID: %s) %s" %
                      (len(roles), ou_id, ou['name']))

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
                        print(
                            "\tDetails for OU role %s weren't found. Data will be incomplete. Skipping" % r['name'])
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
                        short_term_access_keys=str(
                            role['short_term_access_keys']).lower(),
                        long_term_access_keys=str(
                            role['long_term_access_keys']).lower(),
                        web_access=str(role['web_access']).lower(),
                        aws_iam_policies=process_list(
                            role['aws_iam_policies'], "aws_iam_policies"),
                        users=process_list(role['user_ids'], "users"),
                        user_groups=process_list(
                            role['user_group_ids'], "user_groups"),
                    )

                    # write the metadata file
                    if ARGS.prepend_id:
                        base_filename = normalize_string(role['name'], r_id)
                    else:
                        base_filename = normalize_string(role['name'])

                    filename = "%s/ou-cloud-access-role/%s/%s.tf" % (
                        ARGS.import_dir, ou_dir, base_filename)

                    write_file(filename, process_template(content))

                    # add to IMPORTED_RESOURCES
                    resource = "module.ou-cloud-access-role.%s.%s_ou_cloud_access_role.%s %s" % (
                        RESOURCE_PREFIX, ou_dir, normalize_string(role['name']), r_id)
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
