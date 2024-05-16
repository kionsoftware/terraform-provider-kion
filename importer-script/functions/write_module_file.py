import os


def write_module_file(file_name, content):
    """
    Write Module File

    Writes the tf files for each module with provided content

    If the file already exists, it will instead print out a
    modulename.tf.example file with the text needed by the
    Kion TF

    Params:
        file_name   (str)   - full file name including path to write
        content     (str)   - content to write into the file
    """

    if os.path.exists(file_name):
        file_name = "%s.example" % file_name

    with open(file_name, 'w', encoding='utf8') as outfile:
        outfile.write(content)

    return True
