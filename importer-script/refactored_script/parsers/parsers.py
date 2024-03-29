import argparse
import os
import sys
import re


def parse_args():

    parser = argparse.ArgumentParser(
        description='Import Cloud Resources into the Repo Module')
    parser.add_argument('--kion-url', type=str, required=True,
                        help='URL to Kion, without trailing slash. Example: https://kion.example.com')
    parser.add_argument('--kion-api-key', type=str,
                        help='Kion API key. Can be set via env variable KION_APIKEY instead (preferred).')
    parser.add_argument('--import-dir', type=str, required=True,
                        help='Path to the root of the target import directory, without trailing slash.')
    parser.add_argument('--skip-cfts', action='store_true',
                        help='Skip importing AWS CloudFormation templates.')
    parser.add_argument('--skip-iams', action='store_true',
                        help='Skip importing AWS IAM policies.')
    # parser.add_argument('--skip-arms', action='store_true', help='Skip importing Azure ARM templates.')
    parser.add_argument('--skip-azure-policies', action='store_true',
                        help='Skip importing Azure Policies.')
    parser.add_argument('--skip-azure-roles', action='store_true',
                        help='Skip importing Azure Roles.')
    parser.add_argument('--skip-project-roles', action='store_true',
                        help='Skip importing Project Cloud Access Roles.')
    parser.add_argument('--skip-ou-roles', action='store_true',
                        help='Skip importing OU Cloud Access Roles.')
    parser.add_argument('--skip-cloud-rules', action='store_true',
                        help='Skip importing Cloud Rules.')
    parser.add_argument('--skip-checks', action='store_true',
                        help='Skip importing Compliance Checks.')
    parser.add_argument('--skip-standards', action='store_true',
                        help='Skip importing Compliance Standards.')
    parser.add_argument('--skip-ssl-verify', action='store_true',
                        help='Skip SSL verification. Use if Kion does not have a valid SSL certificate.')
    parser.add_argument('--overwrite', action='store_true',
                        help='Overwrite existing files during import.')
    parser.add_argument('--import-aws-managed', action='store_true',
                        help='Import AWS-managed resources (only those that were already imported into Kion).')
    parser.add_argument('--prepend-id', action='store_true',
                        help='Prepend each resource\'s ID to its filenames. Useful for easily correlating IDs to resources')
    parser.add_argument('--clone-system-managed', action='store_true',
                        help='Clone system-managed resources. Names of clones will be prefixed using --clone-prefix argument. Ownership of clones will be set with --clone-user-ids and/or --clone-user-group-ids')
    parser.add_argument('--clone-prefix', type=str,
                        help='A prefix for the name of cloned system-managed resources. Use with --clone-system-managed.')
    parser.add_argument('--clone-user-ids', nargs='+', type=int,
                        help='Space separated user IDs to set as owner users for cloned resources')
    parser.add_argument('--clone-user-group-ids', nargs='+', type=int,
                        help='Space separated user group IDs to set as owner user groups for cloned resources')

    # parser.add_argument('--dry-run', action='store_true', help='Perform a dry run without writing any files.')
    # parser.add_argument('--sync', action='store_true',help='Sync repository resources into Kion.')
    return parser.parse_args()


ARGS = parse_args()

# validate kion_url
if not ARGS.kion_url:
    sys.exit(
        "Please provide the URL to Kion. Example: --kion-url https://kion.example.com")
# remove trailing slash if found
elif re.compile(".+/$").match(ARGS.kion_url):
    ARGS.kion_url = re.sub(r'/$', '', ARGS.kion_url)

# validate import_dir
if not ARGS.import_dir:
    sys.exit("Please provide the path to the directory in which to import. Example: --import-dir /Users/me/code/repo-module-dir")
# remove trailing slash if found
elif re.compile(".+/$").match(ARGS.import_dir):
    ARGS.import_dir = re.sub(r'/$', '', ARGS.import_dir)

# validate API key
if not ARGS.kion_api_key:
    if os.environ.get('KION_APIKEY'):
        ARGS.kion_api_key = os.environ['KION_APIKEY']
    else:
        sys.exit(
            "Did not find a Kion API key supplied via CLI argument or environment variable (KION_APIKEY).")

# validate flags related to cloning
if not ARGS.clone_prefix:
    ARGS.clone_prefix = ""
if ARGS.clone_system_managed:

    # validate clone prefix
    if not ARGS.clone_prefix:
        sys.exit(
            "You did not provide a clone prefix value using the --clone-prefix flag.")
    else:
        if not ARGS.clone_prefix.endswith('-') and not ARGS.clone_prefix.endswith("_"):
            sys.exit("Did not find a _ or - in clone prefix.")

    # validate clone-user-ids and clone-user-group-ids
    if not ARGS.clone_user_ids and not ARGS.clone_user_group_ids:
        sys.exit("You must provide at least one of --clone-user-ids or --clone-user-group-ids in order to import system-managed resources.")
    else:
        if not ARGS.clone_user_ids:
            ARGS.clone_user_ids = []
        if not ARGS.clone_user_group_ids:
            ARGS.clone_user_group_ids = []
