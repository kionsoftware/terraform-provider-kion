import re


def process_string(input):
    """
    Helper function to handle routine string processing

    Params:
        input   (str)       - the original string
        output  (str)       - the processed string
    """
    # output = re.sub('\\r', '', input)           # replace windows carriage-returns with a space
    output = input.replace("\r", "\\r")
    # replace newlines with a space
    output = output.replace("\n", "\\n")
    # replace double quotes with single quotes
    output = output.replace('"', "'")
    # replace single backslashes with double backslashes
    output = output.replace('\\', '\\\\')
    # replace multiple spaces with a single space
    output = re.sub('\s{2,}', ' ', output)
    output = output.strip()                     # strip leading and trailing whitespace
    return output
