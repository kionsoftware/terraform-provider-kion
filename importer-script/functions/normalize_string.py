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

    # A Terraform identifier must start with a letter or underscore. The block
    # label itself is quoted and so tolerates a leading digit, but the reference
    # generated alongside it is not:
    #
    #   resource "kion_cloud_rule" "0_Prod_Baseline" { ... }        # parses
    #   value = kion_cloud_rule.0_Prod_Baseline.id                  # does NOT
    #
    # Terraform rejects the whole file, so a single Kion resource whose name
    # begins with a digit ("800-53 Audit") makes the import unusable. This also
    # affects every name when --prepend-id is set, since that always produces a
    # leading digit.
    if string and string[0].isdigit():
        string = "_" + string

    return string
