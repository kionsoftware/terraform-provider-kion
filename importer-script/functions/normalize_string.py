import re


def normalize_string(string, id_=False):
    """
    Normalize String

    Receives a string and normalizes it for proper source control handling

    Params:
        name (str)  - original string
        id_ (int)   - id of the resource. If set, it will be prepended to filename

    Return:
        string (str) - normalized string
    """
    string = re.sub(
        r'\s', '_', string)                 # replace spaces with underscores
    # remove all non alphanumeric characters
    string = re.sub(r'[^A-Za-z0-9_-]', '', string)

    # prepend {ID} if id_ is set
    if id_:
        string = "%s-%s" % (id_, string)

    return string
