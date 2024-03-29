RESOURCE_PREFIX = 'kion'
BASE_URL = "%s/api" % ARGS.kion_url
HEADERS = {"accept": "application/json",
           "Authorization": "Bearer " + ARGS.kion_api_key}

MAX_UNAUTH_RETRIES = 15
UNAUTH_RETRY_COUNTER = 0

IMPORTED_MODULES = []
IMPORTED_RESOURCES = []
