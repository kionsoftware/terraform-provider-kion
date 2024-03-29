def process_owners(input, text):
    """
    Process the passed list of owner_users or owner_user_groups into the format
    required for the TF config files

    Param:
        input   (list)  - list of owner user objects returned from Kion, or just IDs
        text    (str)   - text to prepend to each line in output
                            ('owner_users' or 'owner_user_groups')

    Return:
        output (list)       - processed list of owner user IDs
    """
    ids = []
    output = []

    # get all of the user IDs, store in ids
    for i in input:
        if isinstance(i, dict):
            if 'id' in i:
                ids.append(i['id'])
        elif isinstance(i, int):
            ids.append(i)

    if len(ids):
        for i in ids:
            line = "%s { id = %s }" % (text, i)
            output.append(line)
    else:
        output.append("%s { }" % text)
    return output


