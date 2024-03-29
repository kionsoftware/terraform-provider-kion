import sys
from .api_call import api_call


def validate_connection(url):
    """
    Validate Connection

    Make sure that the supplied CT URL can be reached

    Params:
        url (str) - the provided kion_url argument

    Returns:
        success - True
        failure - sys.exit
    """

    if api_call(url, 'get', False, False, 30, True):
        return True
    else:
        sys.exit("Unable to connect to %s" % url)
