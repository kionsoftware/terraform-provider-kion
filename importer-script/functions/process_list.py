def process_list(input, text):
    """
    Process a list of IDs into a multi-line list of
    objects as required by TF

    Param:
        input   (list)      - list of IDs to process
        text    (str)       - text to prepend to each line

    Return:
        output  (str)      - multi-line string list of objects
    """

    output = []

    if len(input):
        for i in input:
            line = "%s { id = %s }" % (text, i)
            output.append(line)
    else:
        line = "%s { }" % text
        output.append(line)

    return '\n    '.join(output)


