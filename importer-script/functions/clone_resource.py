import json
from config import ARGS
from .search_resource import search_resource
from .create_resource import create_resource


def clone_resource(resource_type, resource):
    """
    Clones the resource of provided type.
    Makes use of the OBJECT_API_MAP for mapping type -> API endpoint

    Params:
        resource_type        (str)  - the type of resource being cloned
        resource            (dict)  - a dict of the resource's attributes

    Returns:
        If clone was successful:
            True        (bool)
            resource    (dict)  - A dict of the newly cloned resource
        If matching cloned resource was already found:
            True        (bool)
            False       (bool)
        If failure:
            False       (bool)
            False       (bool)
    """

    # first do some preparation for cloning

    # find the name key, ensure its prepended with the clone prefix
    # and save it to a temp variable

    if 'name' in resource:
        if not resource['name'].startswith(ARGS.clone_prefix):
            resource['name'] = f"{ARGS.clone_prefix}{resource['name']}"
            name = resource['name']
    else:

        # some types of resources are structured differently when it comes to creating them.
        # azure_policies for example needs to have a nested key called 'azure_policy' and under
        # that is the name key. most others have the name key at the root level
        other_structures = ['azure_policy']
        for s in other_structures:
            if s in resource:
                if 'name' in resource[s]:
                    if not resource[s]['name'].startswith(ARGS.clone_prefix):
                        resource[s]['name'] = f"{ARGS.clone_prefix}{resource[s]['name']}"
                        name = resource[s]['name']
                    else:
                        name = False

                # remove some fields while were in here
                resource[s].pop('ct_managed', None)
                resource[s].pop('built_in', None)
                resource[s].pop('id', None)

    # validate that we found the name
    if not name:
        print("Couldn't find the name key in %s" % json.dumps(resource))
        return False, False

    # remove some fields
    resource.pop('ct_managed', None)
    resource.pop('built_in', None)
    resource.pop('id', None)

    # at this point, we found the name and made sure it's prefixed with the clone prefix
    # now lets search for a matching resource
    search = search_resource(resource_type, name)

    if search == []:

        # an empty list means there were no resources matching the clone were found
        # so attempt to create it

        # set owner users and groups
        # these keys are inconsistent across the different resource types
        # so just set both
        for key in ['owner_users', 'owner_user_ids']:
            resource[key] = ARGS.clone_user_ids
        for key in ['owner_user_groups', 'owner_user_group_ids']:
            resource[key] = ARGS.clone_user_group_ids

        new_resource = create_resource(resource_type, resource)
        if new_resource:
            return True, new_resource
        else:
            return False, False

    elif search is False:
        # False means an error occurred
        return False, False
    elif isinstance(search, dict):
        # if we got a dict back, it means a matching resource was found
        # we dont need to return this back as the caller already has it
        return True, False
