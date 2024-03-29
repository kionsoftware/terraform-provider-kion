import json


def read_file(file, content_type):
    """
    Read File

    Reads in the provided file and returns the content

    Params:
        file (str)          - name of the file to write. Expecting full path.
        content_type (str)  - type of file content (json, yaml, yml)

    Return:
        success - content of file
        failure - False
      """
    with open(file, "r", encoding='utf8') as f:
        if content_type == "json":
            # validate that the file contains json
            try:
                data = json.load(f)
            except:
                print("Failed to load json file %s." % file)
                return False
            else:
                return data
