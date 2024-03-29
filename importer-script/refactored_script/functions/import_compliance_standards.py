import api_call
import build_filename
import normalize_string
import process_owners
import clone_resource
import process_list
import get_objects_or_ids
import textwrap
import process_string
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import BASE_URL
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from templates import PROVIDER_TEMPLATE
from parsers import ARGS


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
                    print("Skipping built-in Compliance Standard - %s" %
                          s['name'])
                    continue
                else:
                    system_managed = True

            # init new object
            standard = {}
            standard['name'] = process_string(s['name'])
            standard['checks'] = []
            standard['owner_user_ids'] = []
            standard['owner_user_group_ids'] = []
            standard['description'] = ''
            standard['created_by_user_id'] = ''

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

                result, clone = clone_resource(
                    'compliance_standards', standard)
                if clone:
                    print("Cloning System-managed Compliance Standard: %s -> %s" %
                          (original_name, clone['compliance_standard']['name']))
                    standard['name'] = clone['compliance_standard']['name']
                    standard['created_by_user_id'] = clone['compliance_standard']['created_by_user_id']
                    s['id'] = clone['compliance_standard']['id']
                    owner_users = process_owners(
                        ARGS.clone_user_ids, 'owner_users')
                    owner_groups = process_owners(
                        ARGS.clone_user_group_ids, 'owner_user_groups')
                else:
                    if result:
                        print("Already found a clone of %s. Skipping." %
                              original_name)
                        continue
                    else:
                        print("An error occurred cloning %s" % original_name)
                        continue
            else:
                print("Importing Compliance Standard - %s" % s['name'])
                owner_users = process_owners(
                    details['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    details['owner_user_groups'], 'owner_user_groups')

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
                compliance_checks=process_list(
                    standard['checks'], "compliance_checks")
            )

            # build the file name
            base_filename = build_filename(
                standard['name'], False, ARGS.prepend_id, s['id'])
            filename = "%s/compliance-standard/%s.tf" % (
                ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.compliance-standard.%s_compliance_standard.%s %s" % (
                RESOURCE_PREFIX, normalize_string(standard['name']), s['id'])
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
