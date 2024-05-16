from .normalize_string import normalize_string


def build_filename(base, aws_managed=False, prepend_id=False, r_id=False):
    """
    Helper funcion to build the filename based on provided parameters

    Params:
        base        (str)   - base name of the file
        aws_managed (bool)  - whether or not this is an AWS-managed resource
        prepend_id  (bool)  - whether or not to prepend the resource ID to the name
        r_id        (str)   - the resource's ID to prepend, if prepend_id is True

    Returns:
        base_filename   (str)   - formatted base filename
    """
    base_filename = base

    if aws_managed:
        base_filename = "AWS_Managed_%s" % base_filename

    if prepend_id:
        if r_id:
            base_filename = normalize_string(base_filename, r_id)
        else:
            print("Error - prepend ID was set to true but the ID was not provided. Will return without the ID prepended.")
            base_filename = normalize_string(base_filename)
    else:
        base_filename = normalize_string(base_filename)

    return base_filename
