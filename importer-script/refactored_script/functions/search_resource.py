import api_call
from constants import BASE_URL


def search_resource(type, terms, match_key='name'):
    """
    Helper function to search Kion for objects of type using provided search terms

    Params:
        type:       (str) - the type of resource to search for
        terms:      (str) - the search terms
        match_key:  (str) - the key to match the terms against. defaults to 'name'

    Return:
        If found:
            item            (dict) - dict of the matching object
        If not found:
            empty list      (list)
        If error:
            False           (bool)
    """

    # maps the type that we receive to the type as it will show up in the search results
    type_map = {
        'aws_iam_policies': 'iam',
        'aws_cloudformation_templates': 'cft',
        'cloud_rules': 'cloud_rule',
        'compliance_checks': 'compliance_check',
        'compliance_standards': 'compliance_standard',
        'azure_role_definitions': 'azure_role',
        'azure_policy_definitions': 'azure_policy',
        'azure_arm_template_definitions': 'arm_template'
    }

    if type not in type_map.keys():
        print("Received unmapped type: %s" % type)
        return False
    else:
        type_match = type_map[type]

    url = "%s/v1/search" % BASE_URL
    payload = {"query": terms}
    response = api_call(url, 'post', payload)

    # print("search response: %s" % json.dumps(response))

    if response == []:
        # an empty list means 0 search results
        return []
    elif not response:
        # a False response means some sort of error
        return False
    elif len(response) > 0:
        # here we have some matches
        # loop over them and compare item[match_key] to search terms
        # return empty list if nothing matches (should find a match)
        for item in response:
            if item['type'] == type_match:
                if item[match_key] == terms:
                    return item
        return []
    else:
        # default to a False return - something went wrong
        return False
