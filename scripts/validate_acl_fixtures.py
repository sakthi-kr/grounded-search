#!/usr/bin/env python3
"""Validate deterministic ACL fixture structure and expected reasons."""

from __future__ import annotations

import json
from pathlib import Path
import sys

ALLOWED_REASONS = {
    "document_deleted",
    "document_disabled",
    "unknown_user",
    "user_disabled",
    "explicit_user_deny",
    "matching_group_deny",
    "public_document",
    "explicit_user_allow",
    "matching_group_allow",
    "default_deny",
}

ALLOWED_STATUS = {"active", "disabled", "deleted"}
ALLOWED_DECISIONS = {"allow", "deny"}


def main() -> int:
    root = Path(__file__).resolve().parents[1]
    path = root / "data" / "fixtures" / "acl_cases.json"
    cases = json.loads(path.read_text(encoding="utf-8"))

    if not isinstance(cases, list) or not cases:
        raise ValueError("ACL fixture file must contain a non-empty list")

    names: set[str] = set()

    for index, case in enumerate(cases):
        name = case["name"]
        if name in names:
            raise ValueError(f"duplicate ACL fixture name: {name}")
        names.add(name)

        document = case["document"]
        if document["status"] not in ALLOWED_STATUS:
            raise ValueError(f"case {index}: invalid document status")

        for collection_name in ("user_rules", "group_rules"):
            rules = document[collection_name]
            invalid = set(rules.values()) - ALLOWED_DECISIONS
            if invalid:
                raise ValueError(
                    f"case {index}: invalid {collection_name}: {sorted(invalid)}"
                )

        expected = case["expected"]
        if not isinstance(expected["allowed"], bool):
            raise ValueError(f"case {index}: expected.allowed must be boolean")
        if expected["reason"] not in ALLOWED_REASONS:
            raise ValueError(f"case {index}: invalid expected reason")

    print(f"ACL fixture validation passed: {len(cases)} cases")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except (KeyError, TypeError, ValueError, json.JSONDecodeError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        raise SystemExit(1) from error
