import sys
import textwrap
from functions.validate_connection import validate_connection
from functions.import_iams import import_iams
from functions.validate_import_dir import validate_import_dir
from functions.import_azure_policies import import_azure_policies
from functions.import_azure_roles import import_azure_roles
from functions.import_project_roles import import_project_roles
from functions.import_ou_roles import import_ou_roles
from functions.import_cfts import import_cfts
from functions.import_cloud_rules import import_cloud_rules
from functions.import_compliance_checks import import_compliance_checks
from functions.import_compliance_standards import import_compliance_standards
from functions.write_resource_import_script import write_resource_import_script
from functions.write_provider_file import write_provider_file
from functions.write_module_file import write_module_file
from constants import IMPORTED_MODULES
from constants import IMPORTED_RESOURCES
from functions.templates import PROVIDER_TEMPLATE
from config import ARGS

"""
Kion Terraform Provider Importer

This script imports existing cloud resources into a source control repository
for management by the Terraform Provider.

See the README for usage and optional flags.
"""


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

    # now out of the loop, write the provider.tf file
    provider_filename = f"{ARGS.import_dir}/providers.tf"
    provider_content = PROVIDER_TEMPLATE
    write_provider_file(provider_filename, provider_content)

    # Loop over each module and create a .tf file for it
    for module in IMPORTED_MODULES:
        module_filename = f"{ARGS.import_dir}/{module}.tf"
        module_content = textwrap.dedent(f'''\
            module "{module}" {{
                source = "./{module}"
            }}
        ''')

        write_module_file(module_filename, module_content)

    # Placeholder for additional scripting
    write_resource_import_script(ARGS, IMPORTED_RESOURCES)
    print("\nIf you need to refresh the terraform state of the imported resources, run:\n")
    print(f"cd {ARGS.import_dir} ; bash import_resource_state.sh")
    print("\nImport finished.")
    sys.exit()


if __name__ == "__main__":
    main()
