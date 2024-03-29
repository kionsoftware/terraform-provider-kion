import os


def write_file(file_name, content):
    """
    Write File

    Writes the given file_name with given content.
    Based on file_type, it will determine how to write the file,
    either as-is or using json.dump.

    Params:
        file_name   (str)           - name of the file to write. Expecting absolute path.
        content     (str)           - content to write to file_name

    Return:
        success - True
    """

    # only proceed if filename doesn't exist yet or the overwrite flag was set
    if not os.path.exists(file_name) or ARGS.overwrite:
        with open(file_name, 'w', encoding='utf8') as outfile:
            outfile.write(content)
    # else:
        # print("Found %s already. Will not overwrite." % file_name)

    return True
