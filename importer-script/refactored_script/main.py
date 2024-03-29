"""
Kion Terraform Provider Importer

This script imports existing cloud resources into a source control repository
for management by the Terraform Provider.

See the README for usage and optional flags.
"""

import os
import re
import sys
import json
import textwrap
import argparse
import requests
from json.decoder import JSONDecodeError



# Import statements for refactored structure
from parsers import PARSER
BASE_URL = "%s/api" % ARGS.kion_url
HEADERS = {"accept": "application/json",
           "Authorization": "Bearer " + ARGS.kion_api_key}

MAX_UNAUTH_RETRIES = 15
UNAUTH_RETRY_COUNTER = 0

RESOURCE_PREFIX = 'kion'
IMPORTED_MODULES = []
IMPORTED_RESOURCES = []

