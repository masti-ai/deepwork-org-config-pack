#!/usr/bin/env bash
# scripts/pack_auto_refresh.sh — regenerate onboarding knowledge from Mind.
#
# Invoked by crons/pack-auto-refresh.yaml. Safe to run manually:
#   scripts/pack_auto_refresh.sh --dry-run   # print diff, don't write
#   scripts/pack_auto_refresh.sh             # refresh + commit if changed
#
# Uses the DI MCP tool surface — never greps files directly. Three
# regenerations: patterns, decisions, conventions. Each pulls top-K high-
# importance memories of the matching kind and hands them to MiniMax via
# docs_generate to produce a clean markdown doc.
set -euo pipefail

DRY_RUN="${1:-}"
PACK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHECKPOINT="$PACK_DIR/.last-refresh"

log() { printf '[pack-refresh] %s\n' "$*"; }

# One generate-and-diff cycle per knowledge file. The tool call is
# deliberately stateless — we never mutate memory state from here.
refresh() {
  local kind="$1" target="$2"
  local tmp; tmp=$(mktemp)
  log "regenerating $target from kind=$kind memories"
  # TODO: wire DI MCP — the stub below is the shape of the call. Until
  # the CLI entrypoint for MCP lands (tracked in di-ew89.4), this is a
  # no-op that preserves the existing file.
  #
  #   mind docs generate \
  #     --kind "$kind" \
  #     --top-k "${PACK_REFRESH_TOP_K:-30}" \
  #     --min-importance "${PACK_REFRESH_MIN_IMPORTANCE:-0.6}" \
  #     --scope "${PACK_REFRESH_SCOPE:-org}" \
  #     --out "$tmp"
  cat "$target" > "$tmp"

  if ! diff -q "$target" "$tmp" >/dev/null 2>&1; then
    log "$target has changes"
    if [[ "$DRY_RUN" == "--dry-run" ]]; then
      diff -u "$target" "$tmp" || true
    else
      cp "$tmp" "$target"
    fi
  else
    log "$target unchanged"
  fi
  rm -f "$tmp"
}

refresh "pattern"    "$PACK_DIR/knowledge/patterns.md"
refresh "decision"   "$PACK_DIR/knowledge/decisions.md"
refresh "convention" "$PACK_DIR/knowledge/conventions.md"

date -Iseconds > "$CHECKPOINT"

if [[ "$DRY_RUN" != "--dry-run" ]]; then
  cd "$PACK_DIR"
  if ! git diff --quiet knowledge/; then
    git add knowledge/ .last-refresh
    git -c user.email=pack-refresh@deepwork.ai -c user.name="pack-refresh" \
      commit -m "chore(pack): auto-refresh knowledge from Mind memories" >/dev/null
    git push origin HEAD 2>&1 || log "push skipped (no remote auth)"
    log "committed knowledge refresh"
  fi
fi
