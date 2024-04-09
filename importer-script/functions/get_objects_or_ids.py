import json
from .get_api_endpoint import get_api_endpoint
from .api_call import api_call
from constants import BASE_URL


def get_objects_or_ids(object_type, cloud_rule=False, object_id=False):
    """
    Generic helper function to either return all objects of object_type from Kion

    If cloud_rule is set, return a list of IDs of the associated object_type in cloud_rule

    If object_id is set, return only the object of object_type with that ID

    Params:
        object_type     (str)   -   the type of object to get from the cloud rule, or out of Kion
                                    must be one of the keys of OBJECT_API_MAP
        cloud_rule      (dict)  -   the cloud_rule object to return IDs of object_type. If not set, this
                                    function will return all objects of object_type
        object_id       (int)   -   the ID of the individual object to return

    Return:
        Success:
            if cloud_rule:  ids     (list)  - list of IDs of object_type found in cloud_rule
            else:           objects (list)  - list of objects of object_type
        Failure:
            False   (bool)
    """

    if cloud_rule:
        ids = []
        for i in cloud_rule[object_type]:
            ids.append(i['id'])
        return ids
    else:
        api_endpoint = get_api_endpoint(object_type, 'GET')

        if object_id:
            url = "%s/%s/%s" % (BASE_URL, api_endpoint, object_id)
        else:
            url = "%s/%s" % (BASE_URL, api_endpoint)

        objects = api_call(url)
        if objects:
            return objects
        else:
            print("Could not get return from %s endpoint from Kion." % url)
            return False
