def get_api_endpoint(resource, method):
    """
    Helper function to return the proper API endpoint for the
    provided resource and method

    Params:
        resource    (str)   - the resource that we are interacting with. Must be defined in OBJECT_API_MAP
        method      (str)   - the method being used to interact with the resource's API

    Returns:
        endpoint    (str)   - the corresponding endpoint
    """
    if resource in OBJECT_API_MAP.keys():
        if method in OBJECT_API_MAP[resource].keys():
            return OBJECT_API_MAP[resource][method]
        else:
            print("Didn't find %s defined for %s in the map." %
                  (method, resource))
            return False
    else:
        print("Didn't find %s defined in the map." % resource)
        return False


