def process_string(input):
    """
    Helper function to handle routine string processing

    Escapes a value for emission inside a double-quoted HCL string. The result
    must round-trip: what Terraform reads back has to equal what the API
    returned, or every plan reports a difference on an attribute nobody changed.

    Two previous behaviours broke that:

      - `re.sub(r'\\s{2,}', ' ', output)` collapsed every run of whitespace to a
        single space. Kion descriptions are markdown and carry meaningful
        indentation ("    The ARN of the Central Auditing Event Bus."), so the
        imported configuration disagreed with the API on every indented line and
        `terraform plan` showed a permanent description diff.
      - Double quotes were replaced with single quotes to avoid terminating the
        HCL string. That silently rewrote the value; escaping preserves it.

    Backslashes are escaped first, so the escapes added afterwards are not
    themselves re-escaped.

    Params:
        input   (str)       - the original string

    Return:
        output  (str)       - the escaped string, safe inside "..." in HCL
    """
    if input is None:
        return input

    output = input.replace("\\", "\\\\")
    output = output.replace("\r", "\\r")
    output = output.replace("\n", "\\n")
    output = output.replace("\t", "\\t")
    output = output.replace('"', '\\"')
    # HCL interpolates ${...} inside a quoted string; $${ is the literal form.
    output = output.replace("${", "$${")
    return output
