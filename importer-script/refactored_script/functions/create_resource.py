import json
import get_api_endpoint
import api_call
import get_objects_or_ids
from constants import BASE_URL


def create_resource(resource_type, resource):
    """
    Creates a new resource of resource_type

    Params:
        resource_type       (str)   - the type of resource to create
        resource            (dict)  - the complete resource object to be created

    Returns:
        If success:
            resource        (dict)  - the newly created resource object
        If failure:
            False           (bool)
    """

    # set up the API URL to hit and make the call
    # this post should create the new cloned resource
    api_endpoint = get_api_endpoint(resource_type, 'POST')
    url = "%s/%s" % (BASE_URL, api_endpoint)
    response = api_call(url, 'post', resource)

    # print("post payload: %s" % json.dumps(resource, indent=2))
    # print("post response: %s" % json.dumps(response))

    if response:
        if 'status' in response:
            if 'record_id' in response:
                resource = get_objects_or_ids(
                    resource_type, False, response['record_id'])
                if resource:
                    return resource
                else:
                    return False
            else:
                print("Didn't receive a record ID when creating resource: %s" %
                      json.dumps(response))
                return False
        else:
            print("Received bad response while creating resource: %s" %
                  json.dumps(response))
            return False
    else:
        print("Failed creating resource: %s" % json.dumps(resource))
        return False
