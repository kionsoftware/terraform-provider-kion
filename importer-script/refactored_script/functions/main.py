import sys
import textwrap
import write_provider_file
import validate_connection
import import_cfts
import import_iams
import validate_import_dir
import import_azure_policies
import import_azure_roles
import import_project_roles
import import_ou_roles
import import_cloud_rules
import import_compliance_checks
import import_compliance_standards
import write_resource_import_script
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from templates import PROVIDER_TEMPLATE
from parsers import ARGS


def main():
    """
    Main Function

    All processing occurs here.
    """

    # Run some validations prior to starting
    validate_connection(ARGS.kion_url)
    validate_import_dir(ARGS.import_dir)

    print("\nBeginning import from %s" % ARGS.kion_url)

    if not ARGS.skip_cfts:
        import_cfts()
    else:
        print("\nSkipping AWS CloudFormation Templates")

    if not ARGS.skip_iams:
        import_iams()
    else:
        print("\nSkipping AWS IAM Policies")

    # ARMs cannot be cloned. Creating a new ARM requires setting a Resource Group
    # which we won't know
    # if not ARGS.skip_arms:
    #     import_arms()
    # else:
    #     print("\nSkipping Azure ARM Templates")

    if not ARGS.skip_azure_policies:
        import_azure_policies()
    else:
        print("\nSkipping Azure Policies")

    if not ARGS.skip_azure_roles:
        import_azure_roles()
    else:
        print("\nSkipping Azure Roles")

    if not ARGS.skip_project_roles:
        import_project_roles()
    else:
        print("\nSkipping Project Cloud Access Roles")

    if not ARGS.skip_ou_roles:
        import_ou_roles()
    else:
        print("\nSkipping OU Cloud Access Roles")

    if not ARGS.skip_cloud_rules:
        import_cloud_rules()
    else:
        print("\nSkipping Cloud Rules")

    if not ARGS.skip_checks:
        import_compliance_checks()
    else:
        print("\nSkipping Compliance Checks")

    if not ARGS.skip_standards:
        import_compliance_standards()
    else:
        print("\nSkipping Compliance Standards")

    # now out of the loop, write the main provider.tf file
    # to pull in the modules
    provider_filename = "%s/main.tf" % ARGS.import_dir
    content = PROVIDER_TEMPLATE

    # loop over all the modules that were imported and
    # add those to main.tf
    for module in IMPORTED_MODULES:
        module_template = textwrap.dedent('''\

            module "{module_name}" {{
                source = "./{module_name}"
            }}
            ''')

        module_content = module_template.format(
            module_name=module
        )
        content += module_content

    write_provider_file(provider_filename, content)
    write_resource_import_script(ARGS, IMPORTED_RESOURCES)
    print("\nIf you need to refresh the terraform state of the imported resources, run:\n")
    print("cd %s ; bash import_resource_state.sh" % ARGS.import_dir)
    print("\nImport finished.")
    sys.exit()
