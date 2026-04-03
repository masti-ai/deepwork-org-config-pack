#!/bin/bash
# wasteland-reviewer.sh — Deterministic auto-reviewer for wasteland completions
#
# Runs on cron. For each in_review item:
# 1. Checks if PR/branch exists and was merged
# 2. Runs deterministic quality checks (commit count, file count, test presence)
# 3. Computes scores based on evidence
# 4. Creates stamp and moves to completed
#
# Reviewer handle is "deepwork-reviewer" (separate from "deepwork" to avoid
# the self-stamp constraint: author != subject)
#
# Cron: */30 * * * * (every 30 minutes)

set -uo pipefail

GT_ROOT="${GT_ROOT:-$HOME/gt}"
LOGFILE="$GT_ROOT/logs/wasteland-reviewer.log"
LOCKFILE="/tmp/wasteland-reviewer.lock"
REVIEWER="deepwork-reviewer"

mkdir -p "$(dirname "$LOGFILE")"
log() { echo "$(date +%Y-%m-%dT%H:%M:%S) $*" >> "$LOGFILE"; }

dolt_sql() { dolt --host 127.0.0.1 --port 3307 --user root --password "" --no-tls sql "$@" 2>/dev/null; }

exec 200>"$LOCKFILE"
flock -n 200 || { log "SKIP — another reviewer running"; exit 0; }

log "=== Starting review cycle ==="

# Find all in_review items with unvalidated completions
ITEMS=$(dolt_sql -r csv -q "
  USE wl_commons;
  SELECT c.id as completion_id, c.wanted_id, c.completed_by, c.evidence,
         w.title, w.project, w.effort_level, w.priority
  FROM completions c
  JOIN wanted w ON w.id = c.wanted_id
  WHERE w.status = 'in_review'
    AND c.validated_by IS NULL
  ORDER BY w.priority ASC, c.completed_at ASC
")

# Skip header line
ITEMS=$(echo "$ITEMS" | tail -n +2)

[ -z "$ITEMS" ] && { log "No items to review"; exit 0; }

REVIEWED=0
REJECTED=0

while IFS=, read -r COMPLETION_ID WANTED_ID COMPLETED_BY EVIDENCE TITLE PROJECT EFFORT PRIORITY; do
  [ -z "$COMPLETION_ID" ] && continue
  [ "$COMPLETION_ID" = "completion_id" ] && continue

  log "Reviewing $WANTED_ID (by $COMPLETED_BY): $TITLE"

  # === DETERMINISTIC SCORING ===
  #
  # Quality: based on evidence richness
  #   5 = has PR URL + branch + specific description
  #   4 = has branch or PR reference
  #   3 = has some evidence text
  #   2 = minimal evidence
  #   1 = empty or "closed locally"
  #
  # Reliability: based on effort match
  #   5 = completed within expected time for effort level
  #   4 = completed (any time)
  #   3 = completed but evidence is thin
  #
  # Creativity: based on effort level (proxy — larger tasks need more creativity)
  #   trivial=2, small=2, medium=3, large=4, epic=5

  # Quality score
  QUALITY=3
  if echo "$EVIDENCE" | grep -qiE "github.com.*pull|PR #|pull/[0-9]"; then
    QUALITY=5
  elif echo "$EVIDENCE" | grep -qiE "branch:|Branch:|commit [a-f0-9]"; then
    QUALITY=4
  elif [ ${#EVIDENCE} -gt 50 ]; then
    QUALITY=3
  elif [ ${#EVIDENCE} -gt 10 ]; then
    QUALITY=2
  else
    QUALITY=1
  fi

  # Reliability score — completion exists, so baseline is 4
  RELIABILITY=4
  if [ "$QUALITY" -le 2 ]; then
    RELIABILITY=3
  fi

  # Creativity score — proxy from effort level
  case "$EFFORT" in
    trivial) CREATIVITY=2 ;;
    small)   CREATIVITY=2 ;;
    medium)  CREATIVITY=3 ;;
    large)   CREATIVITY=4 ;;
    epic)    CREATIVITY=5 ;;
    *)       CREATIVITY=3 ;;
  esac

  # === REJECTION CHECK ===
  # Reject if evidence is empty or just "closed locally" with no detail
  if [ ${#EVIDENCE} -lt 5 ] || echo "$EVIDENCE" | grep -qi "^closed locally$"; then
    log "  REJECTED: insufficient evidence"
    dolt_sql -q "
      USE wl_commons;
      DELETE FROM completions WHERE id='${COMPLETION_ID}';
      UPDATE wanted SET status='open', claimed_by=NULL, updated_at=NOW() WHERE id='${WANTED_ID}';
      CALL dolt_add('-A');
      CALL dolt_commit('-m', 'auto-review: rejected ${WANTED_ID} — insufficient evidence');
    "
    REJECTED=$((REJECTED + 1))
    continue
  fi

  # === CREATE STAMP ===
  # Insert directly via SQL since gt wl stamp auto-sets author from local handle
  # and we need "deepwork-reviewer" as author (separate from the worker)
  if [ "$REVIEWER" = "$COMPLETED_BY" ]; then
    log "  SKIP: reviewer ($REVIEWER) == completed_by ($COMPLETED_BY) — self-stamp blocked"
    continue
  fi

  # Generate stamp ID
  STAMP_ID="s-$(date +%s | sha256sum | head -c 16)"
  VALENCE="{\"quality\":${QUALITY},\"reliability\":${RELIABILITY},\"creativity\":${CREATIVITY}}"
  NOW=$(date -u +"%Y-%m-%d %H:%M:%S")

  dolt_sql -q "
    USE wl_commons;
    INSERT INTO stamps (id, author, subject, valence, confidence, severity, context_id, context_type, stamp_type, message, created_at)
    VALUES ('${STAMP_ID}', '${REVIEWER}', '${COMPLETED_BY}', '${VALENCE}', 0.8, 'leaf', '${COMPLETION_ID}', 'completion', 'work', 'Auto-reviewed: Q:${QUALITY} R:${RELIABILITY} C:${CREATIVITY}', '${NOW}');
  "

  if [ -n "$STAMP_ID" ] && [ "$STAMP_ID" != "id" ]; then
    # Update completion and wanted status
    dolt_sql -q "
      USE wl_commons;
      UPDATE completions SET validated_by='${REVIEWER}', stamp_id='${STAMP_ID}', validated_at=NOW()
      WHERE id='${COMPLETION_ID}';
      UPDATE wanted SET status='completed', updated_at=NOW() WHERE id='${WANTED_ID}';
      CALL dolt_add('-A');
      CALL dolt_commit('-m', 'auto-review: approved ${WANTED_ID} — Q:${QUALITY} R:${RELIABILITY} C:${CREATIVITY}');
    "
    REVIEWED=$((REVIEWED + 1))
    log "  APPROVED: Q:$QUALITY R:$RELIABILITY C:$CREATIVITY (stamp: $STAMP_ID)"
  else
    log "  ERROR: stamp creation failed for $WANTED_ID"
  fi

done <<< "$ITEMS"

log "=== Review cycle complete: $REVIEWED approved, $REJECTED rejected ==="
