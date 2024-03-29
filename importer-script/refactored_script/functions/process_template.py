import re


def process_template(input):
    """
    Helper function to apply some uniform processing
    to rendered templates

    Params:
        input   (str)       - the original string
        output  (str)       - the processed string
    """
    output = re.sub(r'\s*\w+\s+{\s+}', r'', input)
    # output = re.sub('    $', '', output, re.MULTILINE)
    return output
