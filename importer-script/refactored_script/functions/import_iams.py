import re
import build_filename
import normalize_string
import process_owners
import clone_resource
import get_objects_or_ids
import textwrap
import process_string
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from templates import PROVIDER_TEMPLATE
from parsers import ARGS


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
            aws_managed = False

            if i['iam_policy']['aws_managed_policy']:
                if not ARGS.import_aws_managed:
                    print("Skipping AWS-managed IAM Policy: %s" %
                          i['iam_policy']['name'])
                    continue
                else:
                    aws_managed = True

            if i['iam_policy']['system_managed_policy']:
                if not ARGS.clone_system_managed:
                    print("Skipping System-managed IAM Policy: %s" %
                          i['iam_policy']['name'])
                    continue
                else:
                    # save original name
                    original_name = i['iam_policy']['name']
                    # reset i to the lower-level object key
                    i = i['iam_policy']

                    # remove unnecessary fields
                    i.pop('id')
                    i.pop('aws_managed_policy')
                    i.pop('system_managed_policy')

                    # the clone_resource function checks if this object with the updated
                    # name already exists and won't create a clone if it does
                    result, clone = clone_resource('aws_iam_policies', i)
                    if clone:
                        print("Cloning System-managed IAM Policy: %s -> %s" %
                              (original_name, i['name']))
                        i = clone  # reset the i object to the new clone
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
                            print("An error occurred cloning %s" %
                                  original_name)
                            continue
            else:
                print("Importing IAM Policy - %s" % i['iam_policy']['name'])
                # get owner user and group IDs formatted into required format
                owner_users = process_owners(i['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    i['owner_user_groups'], 'owner_user_groups')

            # init new IAM object
            iam = {}
            i_id = i['iam_policy']['id']
            iam['name'] = process_string(i['iam_policy']['name'])
            iam['description'] = process_string(i['iam_policy']['description'])
            iam['owner_user_ids'] = []
            iam['owner_user_group_ids'] = []
            iam['policy'] = i['iam_policy']['policy'].rstrip()

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
            base_filename = build_filename(
                iam['name'], aws_managed, ARGS.prepend_id, i_id)

            # if it is not an AWS managed, then set a standard filename
            # and add it to the list of imported resources. Otherwise, add .skip to the filename and
            # don't add it to the list of imported resources
            if not aws_managed:
                filename = "%s/aws-iam-policy/%s.tf" % (
                    ARGS.import_dir, base_filename)

                # add to IMPORTED_RESOURCES
                resource = "module.aws-iam-policy.%s_aws_iam_policy.%s %s" % (
                    RESOURCE_PREFIX, normalize_string(iam['name']), i_id)
                IMPORTED_RESOURCES.append(resource)
            else:
                filename = "%s/aws-iam-policy/%s.tf.skip" % (
                    ARGS.import_dir, base_filename)

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
