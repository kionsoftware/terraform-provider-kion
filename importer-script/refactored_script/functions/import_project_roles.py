import os
import api_call
import normalize_string
import process_list
import textwrap
import get_projects
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from constants import PROVIDER_TEMPLATE
from constants import BASE_URL


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
            url = "%s/v3/project/%s/project-cloud-access-role" % (
                BASE_URL, proj_id)
            roles = api_call(url)

            if roles:
                print("Found %s roles on Project (ID: %s) %s" %
                      (len(roles), proj_id, proj['name']))

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
                    url = "%s/v3/project-cloud-access-role/%s" % (
                        BASE_URL, r_id)
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
                        print(
                            "\tDetails for Project role %s weren't found. Data will be incomplete. Skipping" % r['name'])
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
                        short_term_access_keys=str(
                            role['short_term_access_keys']).lower(),
                        long_term_access_keys=str(
                            role['long_term_access_keys']).lower(),
                        web_access=str(role['web_access']).lower(),
                        future_accounts=str(role['future_accounts']).lower(),
                        aws_iam_policies=process_list(
                            role['aws_iam_policies'], "aws_iam_policies"),
                        accounts=process_list(role['account_ids'], "accounts"),
                        users=process_list(role['user_ids'], "users"),
                        user_groups=process_list(
                            role['user_group_ids'], "user_groups"),
                    )

                    # write the metadata file
                    if ARGS.prepend_id:
                        base_filename = normalize_string(role['name'], r_id)
                    else:
                        base_filename = normalize_string(role['name'])

                    filename = "%s/project-cloud-access-role/%s/%s.tf" % (
                        ARGS.import_dir, proj_dir, base_filename)

                    write_file(filename, process_template(content))

                    # add to IMPORTED_RESOURCES
                    resource = "module.project-cloud-access-role.%s_project_cloud_access_role.%s %s" % (
                        RESOURCE_PREFIX, normalize_string(role['name']), r_id)
                    IMPORTED_RESOURCES.append(resource)

            else:
                print("Found 0 roles on Project (ID: %s) %s" %
                      (proj_id, proj['name']))

    if incomplete_projs != []:
        print("Project roles that failed to return full details:")
        for p in incomplete_projs:
            print(p)

    # now out of the loop, write the provider.tf file
    provider_filename = "%s/project-cloud-access-role/provider.tf" % ARGS.import_dir
    write_provider_file(provider_filename, PROVIDER_TEMPLATE)

    print("Done.")
    return True
