import api_call
from constants import BASE_URL


def get_ou_roles(ou_id):
    """
    Get OU Roles

    Returns a list of role objects that are assigned locally at OU with the given ID.

    This function is needed because the endpoint "v3/ou/{id}/ou-cloud-access-role"
    returns inherited roles too, which we don't want, so we have to do some extra work
    to only get the local roles.

    Params:
        ou_id (int) - ID of the OU for which to return locally applied roles

    Return:
        success - a list of role objects
        failure - False
    """
    url = "%s/v3/ou/%s/ou-cloud-access-role" % (BASE_URL, ou_id)
    roles = api_call(url)
    ROLES = []
    if roles:
        for role in roles:
            if role['ou_id'] == ou_id:
                ROLES.append(role)
        return ROLES
    else:
        return False
