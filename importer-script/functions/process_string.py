import re


def process_string(input):
    """
    Helper function to handle routine string processing

    Params:
        input   (str)       - the original string
        output  (str)       - the processed string
    """
    # replace newlines with a literal "\n"
    output = input.replace("\r", "\\r")
    output = output.replace("\n", "\\n")
    # replace double quotes with single quotes
    output = output.replace('"', "'")
    output = re.sub(r'\s{2,}', ' ', output)
    output = output.strip()  # strip leading and trailing whitespace
    return output
