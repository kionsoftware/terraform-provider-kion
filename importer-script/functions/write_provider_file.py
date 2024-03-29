import os


def write_provider_file(file_name, content):
    """
    Write Provider File

    Writes the provider.tf file with provided content

    If the file already exists, it will instead print out a
    provider.tf.example file with the text needed by the
    Kion TF provider

    Params:
        file_name   (str)   - full file name including path to write
        content     (str)   - content to write into the file
    """

    if os.path.exists(file_name):
        file_name = "%s.example" % file_name

    with open(file_name, 'w', encoding='utf8') as outfile:
        outfile.write(content)

    return True
