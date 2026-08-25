"""Terraform label allocation, unique within a module."""

# scope -> {label: kion_id}
_ALLOCATED = {}


def unique_label(scope, base, r_id):
    """
    Unique Label

    Return a Terraform identifier for this resource that no other resource in
    the same module is already using.

    Kion names are not unique. A single installation can hold two CloudFormation
    templates both called "CIS-EnableVPCFlowLogs" (ids 81 and 182) and two IAM
    policies both called "deny-non-tagged-resources". Both used to normalize to
    the same label, which meant:

      - the second .tf file overwrote the first, so one resource vanished from
        the configuration entirely, and
      - both appeared in import_resource_state.sh at the same address, so the
        second import failed with "Resource already managed by Terraform" while
        the resource it pointed at was never imported at all.

    On a collision the Kion id is appended, which is unique by construction. The
    first resource to claim a name keeps the clean label, so a configuration
    without duplicates is unaffected.

    Callers should use the return value for BOTH the filename and the import
    address. Deriving them separately is what let --prepend-id put the id in the
    filename but not the address, so every address it generated was wrong.

    Params:
        scope   (str)   - module the label lives in (labels only need to be
                          unique within one module)
        base    (str)   - normalized name to use if it is free
        r_id    (int)   - Kion resource id, appended on collision

    Return:
        label   (str)   - a label unique within scope
    """
    seen = _ALLOCATED.setdefault(scope, {})

    if base not in seen:
        seen[base] = r_id
        return base

    # Same resource visited twice (a retry, or a resource reachable by two
    # paths): keep the label it already has rather than inventing a second one.
    if seen[base] == r_id:
        return base

    label = "%s-%s" % (base, r_id)
    seen[label] = r_id
    return label
