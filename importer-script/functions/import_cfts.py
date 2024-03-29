import re
import json
import textwrap
from .normalize_string import normalize_string
from .process_string import process_string
from .process_owners import process_owners
from .get_objects_or_ids import get_objects_or_ids
from .write_file import write_file
from .write_provider_file import write_provider_file
from .process_template import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from .templates import PROVIDER_TEMPLATE
from config import ARGS


def import_cfts():
    """
    Import CloudFormation Templates

    Handles full process to import CFTs

    Returns:
        success - True
        failure - False
    """
    CFTs = get_objects_or_ids('aws_cloudformation_templates')

    if CFTs:
        print("\nImporting AWS CloudFormation Templates\n--------------------------")
        print("Found %s CFTs" % len(CFTs))
        IMPORTED_MODULES.append("aws-cloudformation-template")

        for c in CFTs:
            # init new cft object
            cft = {}
            c_id = c['cft']['id']
            cft['name'] = process_string(c['cft']['name'])
            cft['description'] = process_string(c['cft']['description'])
            cft['regions'] = json.dumps(c['cft']['regions'])
            cft['region'] = c['cft']['region']
            cft['sns_arns'] = process_string(c['cft']['sns_arns'])
            cft['template_parameters'] = c['cft']['template_parameters'].rstrip()
            cft['termination_protection'] = c['cft']['termination_protection']
            cft['owner_user_ids'] = []
            cft['owner_user_group_ids'] = []
            cft['policy'] = c['cft']['policy'].rstrip()

            print("Importing CFT - %s" % cft['name'])

            # get owner user and group IDs formatted into required format
            owner_users = process_owners(c['owner_users'], 'owner_users')
            owner_groups = process_owners(
                c['owner_user_groups'], 'owner_user_groups')

            # pre-process some of the data to fit the required format
            cft['sns_params'] = '\n'.join(cft['sns_arns'])

            # double all single dollar signs to be valid for TF format
            cft['policy'] = re.sub(r'\${1}\{', r'$${', cft['policy'])

            if not cft['region']:
                cft['region'] = 'null'

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                      = {id}
                    name                    = "{resource_name}"
                    description             = "{description}"
                    regions                 = {regions}
                    region                  = "{region}"
                    sns_arns                = "{sns_arns}"
                    termination_protection  = {termination_protection}
                    {owner_users}
                    {owner_groups}

                    template_parameters = <<-EOT
                {template_params}
                EOT

                    policy = <<-EOT
                {policy}
                EOT
                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_aws_cloudformation_template" % RESOURCE_PREFIX,
                resource_id=normalize_string(cft['name']),
                id=c_id,
                resource_name=cft['name'],
                description=cft['description'],
                regions=cft['regions'],
                region=cft['region'],
                sns_arns=cft['sns_arns'],
                termination_protection=str(
                    cft['termination_protection']).lower(),
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
                template_params=cft['template_parameters'],
                policy=cft['policy']
            )

            # do some post-processing of the rendered template prior to
            # writing it out
            if not c['cft']['template_parameters']:
                content = re.sub(
                    '\s*template_parameters = <<-EOT\n\nEOT', '', content)

            if cft['region'] == "null":
                content = re.sub('\s*region\s*= "null"', '', content)

            # build the file name
            if ARGS.prepend_id:
                base_filename = normalize_string(cft['name'], c_id)
            else:
                base_filename = normalize_string(cft['name'])

            filename = "%s/aws-cloudformation-template/%s.tf" % (
                ARGS.import_dir, base_filename)

            write_file(filename, process_template(content))

            # add to IMPORTED_RESOURCES
            resource = "module.aws-cloudformation-template.%s_aws_cloudformation_template.%s %s" % (
                RESOURCE_PREFIX, normalize_string(cft['name']), c_id)
            IMPORTED_RESOURCES.append(resource)

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/aws-cloudformation-template/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing CFTs.")
        return False
