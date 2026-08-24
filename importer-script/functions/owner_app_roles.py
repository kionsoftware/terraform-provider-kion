"""Discover which app roles confer ownership on this Kion installation."""

from .api_call import api_call
from constants import BASE_URL

# Resolved once per run; the answer cannot change mid-import.
_CACHE = None

# The UNION of these is used, not the first match. An installation can map some
# projects through a custom "Owner" role and others through the system-managed
# "Admin" — this one had 26 of the former and 14 of the latter, so stopping at
# the first name found owners for only part of the estate.
_OWNER_ROLE_NAMES = ("owner", "admin")


def owner_app_role_ids():
    """
    Owner App Role IDs

    Return the set of app_role_ids that should be read as conferring ownership.

    This used to be hardcoded to app_role_id == 1, on the assumption that role 1
    is always Admin and Admin always means owner. Neither half holds. App roles
    are installation-specific: on a real instance the ownership role was a
    CUSTOM role named "Owner" with id 67, while id 1 was the system-managed
    "Admin". Fourteen project mappings used role 1 and twenty-six used role 67,
    so the hardcoded check silently found no owners for most projects — and
    kion_project requires at least one of owner_user_ids/owner_user_group_ids,
    so those projects imported into configuration Terraform would not accept.

    Matching on the role NAME instead means a stock installation (Admin, id 1)
    and one with a custom Owner role both resolve correctly.

    Return:
        ids (set[int]) - app_role_ids to treat as ownership-granting
    """
    global _CACHE
    if _CACHE is not None:
        return _CACHE

    roles = api_call("%s/v3/app-role" % BASE_URL) or []

    by_name = {}
    for role in roles:
        name = (role.get('name') or '').strip().lower()
        if name:
            by_name.setdefault(name, []).append(role.get('id'))

    found = set()
    matched_names = []
    for wanted in _OWNER_ROLE_NAMES:
        ids = {i for i in by_name.get(wanted, []) if i is not None}
        if ids:
            found |= ids
            matched_names.append(wanted)

    if found:
        _CACHE = found
        print("Treating app role(s) %s (%s) as conferring ownership." % (
            sorted(_CACHE), ", ".join(matched_names)))
        return _CACHE

    # Nothing recognisable. Fall back to the historical assumption rather than
    # finding no owners at all, but say so — a silent empty result is what made
    # the original bug hard to see.
    print("WARNING: no app role named Owner or Admin found; "
          "falling back to app_role_id == 1 for ownership.")
    _CACHE = {1}
    return _CACHE
