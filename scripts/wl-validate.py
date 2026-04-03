#!/usr/bin/env python3
"""wl-validate.py — Validate a wasteland item before posting.

Can validate from:
  1. Command-line args: wl-validate.py --title "..." --project "..." --type "..." --description "..."
  2. YAML template file: wl-validate.py --from item.yaml
  3. Stdin JSON: echo '{"title":"..."}' | wl-validate.py --stdin

Exits 0 if valid, 1 if invalid (with errors on stderr).
If valid, prints the gt wl post command to stdout.
"""

import argparse
import json
import re
import sys

VALID_TYPES = {"feature", "bug", "design", "rfc", "docs"}
VALID_EFFORTS = {"trivial", "small", "medium", "large", "epic"}
VALID_PRIORITIES = {0, 1, 2, 3, 4}

# Internal/private terms that must NEVER appear in public wasteland items
PRIVATE_PATTERNS = re.compile(
    r"villa_ai_planogram|villa_alc_ai|alc-ai-villa|pratham|freebird|"
    r"gasclaw|Gas City|deepwork_site|officeworld|command_center|"
    r"content_studio|media_studio|products|3300|3307|"
    r"d43a23e8|16a77301|gt-local|localhost:3300|172\.17\.0\.1",
    re.IGNORECASE,
)

REQUIRED_DESC_SECTIONS = [
    (r"##\s*(Context|What|Background)", "## Context section"),
    (r"##\s*(Acceptance Criteria|How to Test|Checklist)", "## Acceptance Criteria section"),
]


def validate(item: dict) -> list[str]:
    """Return list of validation errors. Empty = valid."""
    errors = []

    # Required fields
    if not item.get("title", "").strip():
        errors.append("title is required")
    if not item.get("project", "").strip():
        errors.append("project is required (e.g., gt-monitor, ai-planogram)")
    if not item.get("type", "").strip():
        errors.append("type is required (feature, bug, design, rfc, docs)")

    # Type validation
    t = item.get("type", "")
    if t and t not in VALID_TYPES:
        errors.append(f"type '{t}' invalid — must be one of: {', '.join(sorted(VALID_TYPES))}")

    # Priority validation
    p = item.get("priority")
    if p is not None and int(p) not in VALID_PRIORITIES:
        errors.append(f"priority {p} invalid — must be 0-4")

    # Effort validation
    e = item.get("effort", "")
    if e and e not in VALID_EFFORTS:
        errors.append(f"effort '{e}' invalid — must be one of: {', '.join(sorted(VALID_EFFORTS))}")

    # Description validation
    desc = item.get("description", "")
    if t in ("feature", "bug") and not desc.strip():
        errors.append("description required for features and bugs")

    if desc.strip():
        for pattern, name in REQUIRED_DESC_SECTIONS:
            if not re.search(pattern, desc, re.IGNORECASE):
                errors.append(f"description must include {name}")

        # Check for repo link in features
        if t == "feature" and "github.com" not in desc and "## Repo" not in desc:
            errors.append("features should include a ## Repo section with GitHub URL")

    # Private info check (HARD BLOCK)
    for field_name in ("title", "description", "tags"):
        val = item.get(field_name, "")
        if val and PRIVATE_PATTERNS.search(val):
            match = PRIVATE_PATTERNS.search(val)
            errors.append(
                f"BLOCKED: {field_name} contains private info: '{match.group()}'. "
                "Use generic names for public wasteland items."
            )

    return errors


def build_command(item: dict) -> str:
    """Build the gt wl post command string."""
    parts = ["gt wl post"]
    parts.append(f'--title "{item["title"]}"')
    parts.append(f'--project "{item["project"]}"')
    parts.append(f'--type "{item["type"]}"')
    parts.append(f'--priority {item.get("priority", 2)}')
    parts.append(f'--effort "{item.get("effort", "medium")}"')
    if item.get("tags"):
        parts.append(f'--tags "{item["tags"]}"')
    if item.get("description"):
        parts.append(f'--description "{item["description"]}"')
    return " \\\n  ".join(parts)


def main():
    parser = argparse.ArgumentParser(description="Validate wasteland items")
    parser.add_argument("--title", default="")
    parser.add_argument("--project", default="")
    parser.add_argument("--type", default="")
    parser.add_argument("--priority", type=int, default=2)
    parser.add_argument("--effort", default="medium")
    parser.add_argument("--tags", default="")
    parser.add_argument("--description", "-d", default="")
    parser.add_argument("--from-file", default="")
    parser.add_argument("--stdin", action="store_true")
    parser.add_argument("--json-out", action="store_true", help="Output validation result as JSON")
    args = parser.parse_args()

    item = {}

    if args.stdin:
        item = json.load(sys.stdin)
    elif args.from_file:
        # Simple key: value parser
        with open(args.from_file) as f:
            content = f.read()
        for line in content.split("\n"):
            if ":" in line and not line.startswith(" ") and not line.startswith("#"):
                key, val = line.split(":", 1)
                key = key.strip().lower()
                val = val.strip()
                if key in ("title", "project", "type", "priority", "effort", "tags"):
                    item[key] = int(val) if key == "priority" else val
        # Description is multi-line after "description:"
        if "description:" in content:
            desc_start = content.index("description:") + len("description:")
            # Find next top-level key or end
            remaining = content[desc_start:].strip()
            item["description"] = remaining
    else:
        item = {
            "title": args.title,
            "project": args.project,
            "type": args.type,
            "priority": args.priority,
            "effort": args.effort,
            "tags": args.tags,
            "description": args.description,
        }

    errors = validate(item)

    if args.json_out:
        print(json.dumps({"valid": len(errors) == 0, "errors": errors, "item": item}))
        sys.exit(0 if not errors else 1)

    if errors:
        print("VALIDATION FAILED:", file=sys.stderr)
        for e in errors:
            print(f"  - {e}", file=sys.stderr)
        sys.exit(1)
    else:
        print(build_command(item))
        sys.exit(0)


if __name__ == "__main__":
    main()
