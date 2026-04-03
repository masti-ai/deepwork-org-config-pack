#!/bin/bash
# wl-post.sh — Validated wasteland post with template enforcement
#
# Wraps `gt wl post` with validation. Rejects items missing required fields.
# Use this instead of `gt wl post` directly.
#
# Usage:
#   wl-post.sh --title "..." --project "..." --type "..." --priority N \
#              --effort "..." --tags "..." --description "..."
#
# Or with a YAML/TOML config file:
#   wl-post.sh --from template.yaml

set -euo pipefail

# Parse args
TITLE="" PROJECT="" TYPE="" PRIORITY="" EFFORT="" TAGS="" DESCRIPTION="" FROM_FILE=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --title) TITLE="$2"; shift 2 ;;
    --project) PROJECT="$2"; shift 2 ;;
    --type) TYPE="$2"; shift 2 ;;
    --priority) PRIORITY="$2"; shift 2 ;;
    --effort) EFFORT="$2"; shift 2 ;;
    --tags) TAGS="$2"; shift 2 ;;
    --description|-d) DESCRIPTION="$2"; shift 2 ;;
    --from) FROM_FILE="$2"; shift 2 ;;
    *) echo "Unknown flag: $1"; exit 1 ;;
  esac
done

# If --from file, parse it
if [ -n "$FROM_FILE" ]; then
  if [ ! -f "$FROM_FILE" ]; then
    echo "ERROR: File not found: $FROM_FILE"
    exit 1
  fi
  # Support simple key: value format
  TITLE=$(grep -oP '^title:\s*\K.*' "$FROM_FILE" || echo "")
  PROJECT=$(grep -oP '^project:\s*\K.*' "$FROM_FILE" || echo "")
  TYPE=$(grep -oP '^type:\s*\K.*' "$FROM_FILE" || echo "")
  PRIORITY=$(grep -oP '^priority:\s*\K.*' "$FROM_FILE" || echo "")
  EFFORT=$(grep -oP '^effort:\s*\K.*' "$FROM_FILE" || echo "")
  TAGS=$(grep -oP '^tags:\s*\K.*' "$FROM_FILE" || echo "")
  # Description is everything after "description:" line
  DESCRIPTION=$(sed -n '/^description:/,$ { /^description:/d; p }' "$FROM_FILE" || echo "")
fi

# === VALIDATION ===
ERRORS=()

# Required fields
[ -z "$TITLE" ] && ERRORS+=("title is required")
[ -z "$PROJECT" ] && ERRORS+=("project is required (e.g., gt-monitor, ai-planogram)")
[ -z "$TYPE" ] && ERRORS+=("type is required (feature, bug, design, rfc, docs)")

# Type must be valid
if [ -n "$TYPE" ]; then
  case "$TYPE" in
    feature|bug|design|rfc|docs) ;;
    *) ERRORS+=("type '$TYPE' invalid — must be: feature, bug, design, rfc, docs") ;;
  esac
fi

# Priority must be 0-4
if [ -n "$PRIORITY" ]; then
  case "$PRIORITY" in
    0|1|2|3|4) ;;
    *) ERRORS+=("priority '$PRIORITY' invalid — must be 0-4") ;;
  esac
fi

# Effort must be valid
if [ -n "$EFFORT" ]; then
  case "$EFFORT" in
    trivial|small|medium|large|epic) ;;
    *) ERRORS+=("effort '$EFFORT' invalid — must be: trivial, small, medium, large, epic") ;;
  esac
fi

# Description must have required sections
if [ -n "$DESCRIPTION" ]; then
  if ! echo "$DESCRIPTION" | grep -qi "## Context\|## Repo\|## What"; then
    ERRORS+=("description must include at least a '## Context' or '## Repo' section")
  fi
  if ! echo "$DESCRIPTION" | grep -qi "## Acceptance Criteria\|## How to Test\|- \["; then
    ERRORS+=("description must include '## Acceptance Criteria' with checkboxes")
  fi
fi

# Description required for features and bugs
if [[ "$TYPE" == "feature" || "$TYPE" == "bug" ]] && [ -z "$DESCRIPTION" ]; then
  ERRORS+=("description required for features and bugs")
fi

# No private info check
PRIVATE_PATTERNS="villa_ai_planogram|villa_alc_ai|alc-ai-villa|pratham|freebird|gasclaw|3300|3307|Gas City|deepwork_site|officeworld"
if [ -n "$DESCRIPTION" ] && echo "$DESCRIPTION" | grep -qEi "$PRIVATE_PATTERNS"; then
  ERRORS+=("BLOCKED: description contains private/internal info. Use generic names.")
fi
if echo "$TITLE" | grep -qEi "$PRIVATE_PATTERNS"; then
  ERRORS+=("BLOCKED: title contains private/internal info. Use generic names.")
fi

# Print errors and exit
if [ ${#ERRORS[@]} -gt 0 ]; then
  echo ""
  echo "ERROR: Wasteland post validation failed:"
  echo ""
  for err in "${ERRORS[@]}"; do
    echo "  - $err"
  done
  echo ""
  echo "Required format:"
  echo "  --title \"Clear, outsider-readable title\""
  echo "  --project \"project-name\""
  echo "  --type \"feature|bug|design|rfc|docs\""
  echo "  --priority N  (0=critical, 1=high, 2=medium)"
  echo "  --effort \"trivial|small|medium|large|epic\""
  echo "  --description \"Must include ## Context and ## Acceptance Criteria\""
  echo ""
  exit 1
fi

# Defaults
[ -z "$PRIORITY" ] && PRIORITY="2"
[ -z "$EFFORT" ] && EFFORT="medium"

# === POST ===
CMD="gt wl post --title \"$TITLE\" --project \"$PROJECT\" --type \"$TYPE\" --priority $PRIORITY --effort \"$EFFORT\""
[ -n "$TAGS" ] && CMD="$CMD --tags \"$TAGS\""
[ -n "$DESCRIPTION" ] && CMD="$CMD --description \"$DESCRIPTION\""

eval "$CMD" 2>&1
EXIT=$?

if [ $EXIT -eq 0 ]; then
  echo "Validated and posted: $TITLE"
else
  echo "ERROR: gt wl post failed (exit $EXIT)"
  exit $EXIT
fi
