#!/usr/bin/env python3
"""Discover the KION_ACC_* values an installation can supply.

Read-only: every call is a GET. Prints an env block for the eleven variables the
acceptance tests skip without, and says plainly which ones this install cannot
provide -- a test that skips for a missing id is not a passing test, and the
difference matters when reading a run.

    discover_acc.py <url> <api-key> [api-prefix]
"""
import json
import sys
import urllib.error
import urllib.request

BASE = sys.argv[1].rstrip("/")
KEY = sys.argv[2]
PREFIX = sys.argv[3] if len(sys.argv) > 3 else "/api"


ERRORS = []


def get(path):
    """GET path, recording why a call failed rather than swallowing it.

    Returning a bare None on error made a revoked API key look identical to an
    install with no billing sources: the first run of this script reported
    "0/10 discovered, no billing source at all" against an install holding 17,
    because every call was 401ing.
    """
    req = urllib.request.Request(BASE + PREFIX + path,
                                 headers={"Authorization": "Bearer " + KEY})
    try:
        with urllib.request.urlopen(req, timeout=60) as r:
            return json.loads(r.read())
    except urllib.error.HTTPError as e:
        ERRORS.append(f"{path}: HTTP {e.code}")
    except Exception as e:
        ERRORS.append(f"{path}: {type(e).__name__}")
    return None


def items(payload):
    if not payload:
        return []
    d = payload.get("data")
    if isinstance(d, list):
        return d
    if isinstance(d, dict):
        for k in ("items", "records"):
            if isinstance(d.get(k), list):
                return d[k]
    return []


found, missing = {}, []


def record(name, value, why):
    if value in (None, "", 0):
        missing.append((name, why))
    else:
        found[name] = str(value)


# --- billing sources: the outer id is the payer id every endpoint addresses ---
bs = items(get("/v4/billing-source"))
aws_bs = next((b for b in bs if "aws_payer" in b), None)
azure_bs = next((b for b in bs if "azure_payer" in b), None)
record("KION_ACC_PAYER_ID", (aws_bs or {}).get("id"), "no AWS billing source on this install")
record("KION_ACC_BILLING_SOURCE_ID", (bs[0] if bs else {}).get("id"), "no billing source at all")
record("KION_ACC_AZURE_PAYER_ID", (azure_bs or {}).get("id"), "no Azure billing source on this install")

# kion_custom_account needs a payer of the CUSTOM type, not any payer. Creating
# a custom account under an AWS one is refused -- "The account is of a different
# type than the payer" -- which reads as a provider fault and is not one.
custom_bs = next((b for b in bs if any("custom" in k for k in b)), None)
record("KION_ACC_CUSTOM_PAYER_ID", (custom_bs or {}).get("id"),
       "no custom (other cloud) billing source on this install")

# --- accounts ---
accts = items(get("/v3/account?count=500"))


def inner(a):
    return a.get("account", a) if isinstance(a, dict) else {}


aws_acct = next((inner(a) for a in accts
                 if inner(a).get("account_number") and inner(a).get("account_type_id") == 1), None)
any_acct = next((inner(a) for a in accts if inner(a).get("account_number")), None)
record("KION_ACC_AWS_ACCOUNT_NUMBER", (aws_acct or {}).get("account_number"),
       "no AWS account with an account_number")
record("KION_ACC_CUSTOM_ACCOUNT_NUMBER", (any_acct or {}).get("account_number"),
       "no account with an account_number")

# kion_ami attaches to an account by Kion's own id, not the cloud account
# number: "Bad Request: account not found" is what a stale hardcoded id gives.
record("KION_ACC_ACCOUNT_ID", (aws_acct or any_acct or {}).get("id"),
       "no account Kion can attach an AMI to")

# kion_ami registers an image that must actually exist: Kion assumes its service
# role in the owner account and calls ec2:DescribeImages, so a made-up id fails
# with "Could not validate the presence of this AMI". An AMI already registered
# here is known-good; otherwise it has to be supplied, since Kion has no endpoint
# that enumerates what AWS holds.
amis = items(get("/v3/ami"))
record("KION_ACC_AWS_AMI_ID", next((a.get("aws_ami_id") for a in amis if a.get("aws_ami_id")), None),
       "no AMI registered to copy an id from; set it to any AMI id the install's "
       "accounts can describe (a public Amazon image works)")

azure_acct = next((inner(a) for a in accts if inner(a).get("subscription_uuid")), None)
record("KION_ACC_AZURE_SUBSCRIPTION_UUID", (azure_acct or {}).get("subscription_uuid"),
       "no Azure account carrying a subscription_uuid")

# --- IDMS: the tests want a SAML one specifically ---
idms = items(get("/v3/idms"))
saml = next((i for i in idms if str(i.get("idms_type_id")) == "2"
             or "saml" in json.dumps(i).lower()), None)
record("KION_ACC_SAML_IDMS_ID", (saml or {}).get("id"), "no SAML IDMS configured")

# --- app role ---
roles = items(get("/v3/app-role"))
record("KION_ACC_APP_ROLE_ID", (roles[0] if roles else {}).get("id"), "no app roles")

# --- azure region: Kion's own database id, not the Azure name ---
regions = items(get("/v3/azure-region")) or items(get("/v3/region"))
record("KION_ACC_AZURE_REGION_ID", (regions[0] if regions else {}).get("id"),
       "no azure region endpoint answered (/v3/azure-region, /v3/region)")

if ERRORS:
    print(f"# REQUESTS FAILED against {BASE} -- results below are not trustworthy:")
    for e in ERRORS:
        print(f"#   {e}")
    if all("HTTP 401" in e for e in ERRORS):
        print("# every call was 401: the API key is revoked or rotated, not an empty install")
    print()

print(f"# discovered against {BASE}\n")
for k, v in found.items():
    print(f'export {k}="{v}"')
# Opt-in rather than discovered: app config is global, so mutating it affects
# everyone on the install.
print('# export KION_ACC_APP_CONFIG_MUTATION_OK="1"   # opt-in: app config is global state')

if missing:
    print(f"\n# NOT AVAILABLE on this install ({len(missing)}):")
    for k, why in missing:
        print(f"#   {k}: {why}")
print(f"\n# {len(found)}/{len(found) + len(missing)} discovered")
