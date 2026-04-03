#!/usr/bin/env python3
"""Simulates a training script that reads API keys from the environment."""
import os
import sys
import json

required = ["OPENAI_API_KEY"]
optional = ["OPENAI_ORG_ID", "WANDB_API_KEY"]

result = {"found": {}, "missing": []}

for key in required:
    val = os.environ.get(key)
    if val:
        result["found"][key] = val
    else:
        result["missing"].append(key)

for key in optional:
    val = os.environ.get(key)
    if val:
        result["found"][key] = val

if result["missing"]:
    print(json.dumps(result), file=sys.stderr)
    sys.exit(1)

print(json.dumps(result))
sys.exit(0)
