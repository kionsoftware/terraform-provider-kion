from constants import IMPORTED_RESOURCES


def write_resource_import_script(args, imported_resources):
    """
    Write Resource Import Script

    Writes a bash script to automate imported the current state of
    all resources that were just imported by the script

    Params:
        args                (dict)      - CLI ARGS
        imported_resources  (list)      - the IMPORTED_RESOURCES list
    """

    file_name = "%s/import_resource_state.sh" % args.import_dir
    with open(file_name, 'w', encoding='utf8') as outfile:
        outfile.write("#!/bin/bash\n")
        for line in IMPORTED_RESOURCES:
            outfile.write("terraform import %s\n" % line)

    return True
