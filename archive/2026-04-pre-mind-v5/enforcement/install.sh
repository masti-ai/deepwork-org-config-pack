#!/bin/bash
# enforcement/install.sh — Install Deepwork standards enforcement into a repo
#
# Usage: bash install.sh [repo-path]
#   repo-path: Path to the git repo (default: current directory)
#
# Installs:
# 1. Git hooks (commit-msg, pre-push, pre-commit) → .git/hooks/
# 2. commitlint config → repo root
# 3. GitHub Actions workflow → .github/workflows/
# 4. PR template → .github/PULL_REQUEST_TEMPLATE.md
#
# Safe to re-run — overwrites hooks, preserves existing workflows.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="${1:-.}"

cd "$REPO_DIR" || { echo "ERROR: $REPO_DIR not found"; exit 1; }

# Verify it's a git repo
git rev-parse --git-dir >/dev/null 2>&1 || { echo "ERROR: $REPO_DIR is not a git repo"; exit 1; }

REPO_NAME=$(basename "$(git rev-parse --show-toplevel)")
echo "Installing Deepwork standards in: $REPO_NAME"

# 1. Git hooks
echo "  Installing git hooks..."
HOOKS_DIR="$(git rev-parse --git-dir)/hooks"
mkdir -p "$HOOKS_DIR"

for hook in commit-msg pre-push pre-commit; do
  if [ -f "$SCRIPT_DIR/hooks/$hook" ]; then
    cp -f "$SCRIPT_DIR/hooks/$hook" "$HOOKS_DIR/$hook"
    chmod +x "$HOOKS_DIR/$hook"
    echo "    ✓ $hook"
  fi
done

# 2. commitlint config (only if package.json exists — JS/TS repos)
if [ -f "package.json" ]; then
  echo "  Installing commitlint config..."
  cp -f "$SCRIPT_DIR/commitlint.config.js" ./commitlint.config.js
  echo "    ✓ commitlint.config.js"
fi

# 3. GitHub Actions workflow
echo "  Installing CI workflow..."
mkdir -p .github/workflows
if [ ! -f ".github/workflows/pr-standards.yml" ]; then
  cp -f "$SCRIPT_DIR/ci/pr-standards.yml" .github/workflows/pr-standards.yml
  echo "    ✓ .github/workflows/pr-standards.yml"
else
  echo "    ~ pr-standards.yml already exists (skipped)"
fi

# 4. PR template
echo "  Installing PR template..."
mkdir -p .github
if [ ! -f ".github/PULL_REQUEST_TEMPLATE.md" ]; then
  cat > .github/PULL_REQUEST_TEMPLATE.md << 'TMPL'
## Summary
<!-- 1-3 bullet points: what changed and why -->

## Bead
<!-- bead-id — bead title (or N/A) -->

## Changes
<!-- List key file changes -->

## Testing
- [ ] How you verified this works
- [ ] Edge cases tested

## Screenshots (if UI change)
<!-- Before/after -->
TMPL
  echo "    ✓ .github/PULL_REQUEST_TEMPLATE.md"
else
  echo "    ~ PR template already exists (skipped)"
fi

echo ""
echo "Done! Standards enforcement installed in $REPO_NAME"
echo ""
echo "Hooks installed:"
echo "  commit-msg  — Validates conventional commit format"
echo "  pre-push    — Validates branch naming"
echo "  pre-commit  — Blocks secrets, .env files, large files"
echo ""
echo "CI installed:"
echo "  pr-standards.yml — Validates PR title, body, commits, branch on GitHub"
