import os
import sys
from config import ARGS


def validate_import_dir(path):
    """
    Validate Import Directory

    Make sure the import directory entered has the proper sub-directories
    for the resources being imported.

    Params:
        path (str) - value of ARGS.import_dir

    Returns:
        success - True
        failure - sys.exit with message
    """
    missing_dirs = []
    dir_map = {
        'aws-cloudformation-template': ARGS.skip_cfts,
        'aws-iam-policy': ARGS.skip_iams,
        'cloud-rule': ARGS.skip_cloud_rules,
        'ou-cloud-access-role': ARGS.skip_ou_roles,
        'project-cloud-access-role': ARGS.skip_project_roles,
        'compliance-check': ARGS.skip_checks,
        'compliance-standard': ARGS.skip_standards,
        # 'azure-arm-template': ARGS.skip_arms,
        'azure-policy': ARGS.skip_azure_policies,
        'azure-role': ARGS.skip_azure_roles
    }

    if os.path.isdir(path):
        for folder, flag in dir_map.items():
            if not flag:
                if not os.path.isdir("%s/%s" % (path, folder)):
                    missing_dirs.append(folder)

        if missing_dirs != []:
            print("%s is missing the following sub-directories:" % path)
            for d in missing_dirs:
                print("- %s" % d)

            create = input("Create them? (y/N) ")
            if create == "y":
                for d in missing_dirs:
                    dir = "%s/%s" % (path, d)
                    os.mkdir(dir)
                print("Done. You can now re-run the previous command.")
                sys.exit()
            else:
                sys.exit()
    else:
        sys.exit("Did not find import directory: %s" % path)

    return True
