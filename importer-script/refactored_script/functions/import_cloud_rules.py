import build_filename
import normalize_string
import process_owners
import get_objects_or_ids
import process_list
import clone_resource
import get_projects
import textwrap
import write_file
import write_provider_file
import process_template
from constants import RESOURCE_PREFIX
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from templates import PROVIDER_TEMPLATE
from parsers import ARGS


def import_cloud_rules():
    """
    Import Cloud Rules

    Handles full process to import Cloud Rules

    Returns:
        True
    """
    print("\nImporting Cloud Rules\n--------------------------")
    cloud_rules = get_objects_or_ids('cloud_rules')

    if cloud_rules:
        print("Found %s Cloud Rules" % len(cloud_rules))
        IMPORTED_MODULES.append("cloud-rule")

        # now loop over them and get the CFT and IAM policy associations
        for c in cloud_rules:
            system_managed = False

            # skip the built_in rules unless toggled on
            if c['built_in']:
                if not ARGS.clone_system_managed:
                    print("Skipping built-in Cloud Rule: %s" % c['name'])
                    continue
                else:
                    system_managed = True

            # get the the cloud rule's metadata
            cloud_rule = get_objects_or_ids("cloud_rules", False, c['id'])

            if cloud_rule:
                c['azure_arm_template_definition_ids'] = get_objects_or_ids(
                    'azure_arm_template_definitions', cloud_rule)
                c['azure_policy_definition_ids'] = get_objects_or_ids(
                    'azure_policy_definitions', cloud_rule)
                c['azure_role_definition_ids'] = get_objects_or_ids(
                    'azure_role_definitions', cloud_rule)
                c['cft_ids'] = get_objects_or_ids(
                    'aws_cloudformation_templates', cloud_rule)
                c['compliance_standard_ids'] = get_objects_or_ids(
                    'compliance_standards', cloud_rule)
                c['iam_policy_ids'] = get_objects_or_ids(
                    'aws_iam_policies', cloud_rule)
                c['internal_ami_ids'] = get_objects_or_ids(
                    'internal_aws_amis', cloud_rule)
                c['ou_ids'] = get_objects_or_ids('ous', cloud_rule)
                c['internal_portfolio_ids'] = get_objects_or_ids(
                    'internal_aws_service_catalog_portfolios', cloud_rule)
                c['project_ids'] = get_projects(cloud_rule)
                c['service_control_policy_ids'] = get_objects_or_ids(
                    'service_control_policies', cloud_rule)
            else:
                print("Failed getting Cloud Rule details.")

            # now that we have all these details, go through cloning process if this is a built-in cloud rule
            if system_managed:
                # save original name
                original_name = c['name']
                # the clone_resource function checks if this object with the updated
                # name already exists and won't create a clone if it does
                result, clone = clone_resource('cloud_rules', c)
                if clone:
                    print("Cloning System-managed Cloud Rule: %s -> %s" %
                          (original_name, clone['cloud_rule']['name']))
                    c['id'] = clone['cloud_rule']['id']
                    c['name'] = clone['cloud_rule']['name']
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
            else:
                print("Importing Cloud Rule - %s" % c['name'])
                # get owner user and group IDs formatted into required format
                owner_users = process_owners(
                    cloud_rule['owner_users'], 'owner_users')
                owner_groups = process_owners(
                    cloud_rule['owner_user_groups'], 'owner_user_groups')

            for i in ["pre_webhook_id", "post_webhook_id"]:
                if c[i] is None:
                    c[i] = 'null'

            template = textwrap.dedent('''\
                resource "{resource_type}" "{resource_id}" {{
                    # id                                    = {id}
                    name                                    = "{resource_name}"
                    description                             = "{description}"
                    pre_webhook_id                          = {pre_webhook_id}
                    post_webhook_id                         = {post_webhook_id}
                    {aws_iam_policies}
                    {cfts}
                    {azure_arm_template_definitions}
                    {azure_policy_definitions}
                    {azure_role_definitions}
                    {compliance_standards}
                    {amis}
                    {portfolios}
                    {scps}
                    {ous}
                    {projects}
                    {owner_users}
                    {owner_groups}
                }}

                output "{resource_id}" {{
                    value = {resource_type}.{resource_id}.id
                }}''')

            content = template.format(
                resource_type="%s_cloud_rule" % RESOURCE_PREFIX,
                resource_id=normalize_string(c['name']),
                resource_name=c['name'],
                id=c['id'],
                description=c['description'],
                pre_webhook_id=c['pre_webhook_id'],
                post_webhook_id=c['post_webhook_id'],
                aws_iam_policies=process_list(
                    c['iam_policy_ids'], "aws_iam_policies"),
                cfts=process_list(
                    c['cft_ids'], "aws_cloudformation_templates"),
                azure_arm_template_definitions=process_list(
                    c['azure_arm_template_definition_ids'], "azure_arm_template_definitions"),
                azure_policy_definitions=process_list(
                    c['azure_policy_definition_ids'], "azure_policy_definitions"),
                azure_role_definitions=process_list(
                    c['azure_role_definition_ids'], "azure_role_definitions"),
                compliance_standards=process_list(
                    c['compliance_standard_ids'], "compliance_standards"),
                amis=process_list(c['internal_ami_ids'], "internal_aws_amis"),
                portfolios=process_list(
                    c['internal_portfolio_ids'], "internal_aws_service_catalog_portfolios"),
                scps=process_list(
                    c['service_control_policy_ids'], "service_control_policies"),
                ous=process_list(c['ou_ids'], "ous"),
                projects=process_list(c['project_ids'], "projects"),
                owner_users='\n    '.join(owner_users),
                owner_groups='\n    '.join(owner_groups),
            )

            # build the base file name
            base_filename = build_filename(
                c['name'], False, ARGS.prepend_id, c['id'])
            filename = "%s/cloud-rule/%s.tf" % (ARGS.import_dir, base_filename)

            # add to IMPORTED_RESOURCES
            resource = "module.cloud-rule.%s_cloud_rule.%s %s" % (
                RESOURCE_PREFIX, normalize_string(c['name']), c['id'])
            IMPORTED_RESOURCES.append(resource)

            write_file(filename, process_template(content))

        # now out of the loop, write the provider.tf file
        provider_filename = "%s/cloud-rule/provider.tf" % ARGS.import_dir
        write_provider_file(provider_filename, PROVIDER_TEMPLATE)

        print("Done.")
        return True
    else:
        print("Error while importing Cloud Rules.")
        return False
