#!/usr/bin/env python3
"""Simulates 'jupyter notebook' startup — checks that secrets are in the env
and exits immediately (instead of actually starting a server)."""
import os
import sys
import json

keys_to_check = [
    "OPENAI_API_KEY",
    "ANTHROPIC_API_KEY",
    "HF_TOKEN",
]

result = {"found": {}, "missing": []}

for key in keys_to_check:
    val = os.environ.get(key)
    if val:
        result["found"][key] = val
    else:
        result["missing"].append(key)

print(json.dumps(result))
sys.exit(0)
