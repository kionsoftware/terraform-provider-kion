import json
import re
import api_call
import build_filename
import normalize_string
import process_owners
import clone_resource
import get_comp_checks
import textwrap
import process_string
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from constants import PROVIDER_TEMPLATE


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
                    print("Skipping System-managed Compliance Check - %s" %
                          c['name'])
                    continue
                else:
                    system_managed = True

            # init new check object
            check = {}
            check['name'] = process_string(c['name'])
            check['description'] = process_string(c['description'])
            check['regions'] = c['regions']
            check['azure_policy_id'] = c['azure_policy_id']
            check['cloud_provider_id'] = c['cloud_provider_id']
            check['compliance_check_type_id'] = c['compliance_check_type_id']
            check['severity_type_id'] = c['severity_type_id']
            check['frequency_minutes'] = c['frequency_minutes']
            check['frequency_type_id'] = c['frequency_type_id']
            check['is_all_regions'] = c['is_all_regions']
            check['is_auto_archived'] = c['is_auto_archived']
            check['body'] = c['body'].rstrip()
            check['owner_user_ids'] = []
            check['owner_user_group_ids'] = []

            # we need to make an additional call to get owner users and groups
            url = "%s/v3/compliance/check/%s" % (BASE_URL, c['id'])
            details = api_call(url)

            # get owner user and group IDs formatted into required format
            if details:
                owner_users = process_owners(
                    details['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    details['owner_user_groups'], 'owner_user_groups')
            else:
                print("Failed to get details for check %s" % check['name'])
                print(json.dumps(details))

            # now attempt to clone if importing system-managed resources
            if system_managed:
                original_name = check['name']

                # remove these fields before cloning
                k.pop('id')
                k.pop('ct_managed')

                result, clone = clone_resource('compliance_checks', c)
                if clone:
                    print("Cloning System-managed Compliance Check: %s -> %s" %
                          (original_name, c['name']))
                    c = clone['compliance_check']
                    # override this to maintain refs to it later
                    check['name'] = process_string(c['name'])
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
                print("Importing Compliance Check - %s" % c['name'])

            # properly format regions based on contents
            if check['regions'][0] == '':
                check['regions'] = []
            else:
                check['regions'] = json.dumps(check['regions'])

            # calculate minutes based on frequency type
            if check['frequency_type_id'] == int(3):
                # hourly, divide minutes by 60
                check['frequency_minutes'] = check['frequency_minutes'] // 60
            elif check['frequency_type_id'] == int(4):
                # daily, divide minutes by 1440
                check['frequency_minutes'] = check['frequency_minutes'] // 1440

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
                        is_auto_archived=str(
                            check['is_auto_archived']).lower(),
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
                        is_auto_archived=str(
                            check['is_auto_archived']).lower(),
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
                        is_auto_archived=str(
                            check['is_auto_archived']).lower(),
                        owner_users='\n    '.join(owner_users),
                        owner_groups='\n    '.join(owner_groups)
                    )
                else:
                    print("Unhandled compliance_check_type_id: %s" %
                          check['compliance_check_type_id'])
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
                print("Skipping. Unhandled cloud_provider_id found: %s" %
                      check['cloud_provider_id'])
                continue

            # for external and tenable checks, remove the empty body block
            if check['compliance_check_type_id'] == 1 or check['compliance_check_type_id'] == 4:
                content = re.sub(r'\s*body = <<-EOT\n\nEOT',
                                 '', content, re.MULTILINE)

            # build the base file name
            base_filename = build_filename(
                check['name'], False, ARGS.prepend_id, c['id'])
            filename = "%s/compliance-check/%s.tf" % (
                ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.compliance-check.%s_compliance_check.%s %s" % (
                RESOURCE_PREFIX, normalize_string(check['name']), c['id'])
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
