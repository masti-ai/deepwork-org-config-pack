# Bead Audit Plugin — Agentic Review on Every Close

Shipped 2026-04-18 as `di_core/mind_events/plugins/bead_audit.py`. This is
the portable, gt_collab-free replacement for wasteland auditing that every
Deepwork Mind install gets by default.

## What it does

Fires on every `bead.closed` and `epic.closed` event emitted by the
`post-close-emit.sh` hook (see `hooks-reference.md`). For each close:

1. Pulls the bead's title + description + close_reason
2. Calls the configured LLM (MiniMax M2.5 by default, honours
   `MIND_AUDIT_MODEL`) with a fixed rubric
3. Scores three axes 0.0-1.0: **Quality · Resolution · Completeness**
4. Generates a 2-3 sentence summary and a 1-sentence next-step suggestion
5. Writes a row to a local `bead_audits` Dolt table (auto-bootstrapped
   via `CREATE TABLE IF NOT EXISTS` on first run)

Same rubric as Gas Town's wasteland stamps — different storage. No
federation DB required, so an external org's Mind install gets agentic
auditing out of the box.

## Config

- `MIND_ENABLE_BEAD_AUDIT=0` to turn off
- `MIND_AUDIT_MODEL=...` to override the model
- `DI_LLM_API_KEY` + `MINIMAX_API_URL` (or `OPENAI_API_KEY` +
  `OPENAI_BASE_URL`) for the HTTP call

If no LLM endpoint is configured, the plugin logs a diagnostic and skips —
it never blocks the event loop.

## Where to see results

- Dashboard `/organization` has a "Recent audits" panel (frontend bead
  in flight)
- Raw query:

  ```sql
  SELECT bead_id, overall, summary, next_step, created_at
    FROM mind.bead_audits
    ORDER BY created_at DESC
    LIMIT 10;
  ```

## Extending

Drop a new `.py` file in `di_core/mind_events/plugins/` that registers a
handler via `@register_handler("bead.closed")`. The plugins module
auto-imports everything in the directory on service boot.

Audit plugins are first-class: governance (slack post to the bead owner),
decay (recompute memory importance after a close), and cross-bead
stitching (link this close to related closed beads) are all just more
plugins on the same hook.
