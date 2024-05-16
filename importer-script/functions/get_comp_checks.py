from .api_call import api_call
from constants import BASE_URL


def get_comp_checks(comp_standard=False):
    """
    Get Compliance Checks

    Params:
        comp_standard (dict) -  compliance_standard object for which to return
                                associated compliance standard IDs
                                if not set, it will return a list of all compliance check
                                objects from Kion
    Return:
        ids (list) - list of compliance standard IDs
    """
    if comp_standard:
        ids = []
        for i in comp_standard['compliance_checks']:
            ids.append(i['id'])
        return ids
    else:
        url = "%s/v3/compliance/check" % BASE_URL
        checks = api_call(url)
        if checks:
            return checks
        else:
            print("Could not get Compliance Checks from Kion.")
            return False
