from config import ARGS

RESOURCE_PREFIX = 'kion'
BASE_URL = "%s/api" % ARGS.kion_url
HEADERS = {"accept": "application/json",
           "Authorization": "Bearer " + ARGS.kion_api_key}

MAX_UNAUTH_RETRIES = 15
UNAUTH_RETRY_COUNTER = 0

IMPORTED_MODULES = []
IMPORTED_RESOURCES = []

# Nested modules: resources that hang off a parent (OU and project cloud access
# roles) are written into a per-parent subdirectory. Terraform only loads .tf
# files directly inside a module's source directory, never subdirectories, so
# each of those has to be declared as a module of its own or every resource in
# it is invisible — which is what "resource address ... does not exist in the
# configuration" meant on import.
#
# Entries are (parent_module, module_label, directory_name). module_label is the
# Terraform identifier and may differ from directory_name, because a directory
# is free to start with a digit and an identifier is not.
IMPORTED_SUBMODULES = []
