# Memory Migration: Claude → GT Memory

**Status:** research only. No migration executed. No hook installed.

---

## Problem

Claude Code memories live per-account at:
`~/.claude-accounts/<account>/projects/<project-slug>/memory/*.md`

When the mayor rotates accounts for quota (`pratham` → `pratham_cc1` → `pratham_cc3`),
each account sees only *its own* memory directory. Useful context written by the cc3
account during one 5-hour window is invisible to cc1 the next morning. The GT memory
system (`gt remember` / bead-backed key-value store) is shared across accounts and is
the right place for any memory that should survive an account cycle.

---

## 1. Inventory (Observed 2026-04-14)

Scoped to `-home-pratham2-gt/memory/` on each account (the mayor-level project slug).
Other project slugs exist (gt_monitor, villa_ai_planogram, deacon, etc.) and should be
enumerated by the script during migration.

### `pratham` (primary) — 12 files
- `MEMORY.md` (index)
- `feedback_crew_for_every_task.md` — feedback
- `feedback_no_private_info_in_public.md` — feedback
- `feedback_polecat_minimal.md` — feedback
- `feedback_rig_usage.md` — feedback
- `feedback_wasteland_architecture.md` — feedback  ← **overlaps Part 3 of this bundle**
- `project_autonomous_mayor.md` — project
- `project_deepwork_intelligence.md` — project
- `project_gasclaw_current.md` — project
- `project_gt_upgrade_wasteland.md` — project
- `project_lxc_migration.md` — project
- `project_ulimit_fix.md` — project
- `project_wasteland_config.md` — project
- `reference_dolthub_wasteland.md` — reference

### `pratham_cc1` — 6 files
- `MEMORY.md`
- `feedback_di_mcp_docs.md` — feedback
- `feedback_gc_gt_mixing.md` — feedback
- `project_di_mcp_product_direction.md` — project
- `project_gt_monitor_milestone_2026_04_08.md` — project (stale date-stamped)
- `project_gt_monitor_product_plan.md` — project

### `pratham_cc3` — 21 files
- `MEMORY.md`
- `feedback_ad_pipeline_quality.md` — feedback
- `feedback_content_hot_takes.md` — feedback
- `feedback_gasclaw_llm_setup.md` — feedback
- `feedback_gasclaws_must_deliver.md` — feedback
- `feedback_github_conservative.md` — feedback
- `feedback_interview_content.md` — feedback
- `feedback_keep_blog_repo.md` — feedback
- `feedback_linkedin_approval.md` — feedback
- `feedback_linkedin_org_default.md` — feedback
- `feedback_mayor_dispatch_on_startup.md` — feedback
- `feedback_never_overwrite.md` — feedback
- `feedback_rig_lifecycle.md` — feedback
- `project_business_strategist_role.md` — project
- `project_command_center.md` — project
- `project_command_center_v3.md` — project (supersedes v1)
- `project_expert_router.md` — project
- `project_gascity_sdk_plan.md` — project
- `project_instagram_vertical.md` — project
- `project_process_guardian.md` — project
- `project_trace_extractor.md` — project  ← **overlaps Part 1 of this bundle**
- `reference_cc_session_format.md` — reference
- `reference_masti_ai_github.md` — reference
- `comfyui-setup.md`, `gt-upgrade-v12.md`, `inference-config.md` — untyped (no frontmatter `type:`)

**Total unique files: ~39.** Three accounts' worth of drift. Several obvious duplicates/
near-duplicates (wasteland feedback in pratham; wasteland config; gt upgrade wasteland).
Deduping is required — do not migrate blindly.

Additional slugs that also contain memories (discover during migration, don't hand-list):
`-home-pratham2-gt-mayor`, `-home-pratham2-gt-gt-monitor`, `-home-pratham2-gt-deacon`,
`-home-pratham2-gt-villa-ai-planogram`, etc.

---

## 2. Migration Script Spec

**Not built.** Spec only. Suggested as `scripts/migrate_claude_memory_to_gt.py`.

### Inputs
- `--account <name>` (default: all three)
- `--project-slug <slug>` (default: all `-home-pratham2-gt*` slugs)
- `--dry-run` (print what *would* be migrated, no `gt remember` calls)
- `--dedup-strategy {skip,merge,overwrite}` (default: skip)

### Algorithm
```
for account in accounts:
    for slug in glob(f"{account}/projects/-home-pratham2-gt*"):
        for md in glob(f"{slug}/memory/*.md"):
            if md.name == "MEMORY.md": continue          # index, not content
            frontmatter, body = parse_yaml_frontmatter(md)
            mem_type = frontmatter.get("type", "project") # default
            key      = frontmatter.get("name") \
                       or slug_from_filename(md.name)     # "feedback_rig_usage"
            title    = frontmatter.get("description", "")

            # Dedup: check existing gt memory
            existing = gt_recall(key=key)                  # read via bd query
            if existing:
                match dedup_strategy:
                    skip:      log "skip, key exists"; continue
                    merge:     body = merge(existing.body, body, title)
                    overwrite: pass

            if dry_run:
                print(f"[DRY] gt remember --type {mem_type} --key {key} <body len={len(body)}>")
            else:
                run(["gt", "remember", "--type", mem_type, "--key", key, body])

            archive(md, f"{account}/projects/.migrated/{date}/")  # keep audit trail
```

### Dedup Strategy Details
- **Key collision on same content hash (>90% match)** → `skip`.
- **Partial overlap, same topic** → `merge` via MiniMax `docs_update` (the DI server
  already has this tool — reuse it to generate a coherent merged memory rather than
  concatenating).
- **Conflicting facts** (e.g. `project_command_center.md` vs `project_command_center_v3.md`)
  → manual review; emit to a `conflicts.tsv` and stop.

### Error Handling
- `gt remember` failures (Dolt down) → retry 3× with exponential backoff, then abort
  the run (don't continue — partial migrations leave worse state than no migration).
- Malformed frontmatter → log to `skipped.tsv`, migrate as `type=general`, don't crash.

---

## 3. Hook Design — Prevent Regression

Goal: once migrated, block future drift back into per-account `.claude/.../memory/`.
Two options; recommend **(B) warn-only** to start.

### A. Block (PreToolUse deny)
A `PreToolUse` hook on `Write` and `Edit` that rejects file paths matching
`**/.claude-accounts/**/memory/**` and returns a stderr message telling the model to
use `gt remember`. **Risk:** memory writes are sometimes legitimately useful within a
single session (e.g. Claude's own auto-memory system). Hard block breaks that.

### B. Warn-only (UserPromptSubmit or PostToolUse)
A `PostToolUse` hook on `Write`/`Edit` that inspects `path` — if it matches the memory
glob, injects additionalContext on the *next* user prompt:

> "Heads up: you wrote to `.claude/memory/`. That path is per-account. For cross-
> session / cross-account persistence, use `gt remember --type <t> --key <k>`."

This nudges without breaking existing flows. After a few weeks, flip to (A) if drift
stops.

### Implementation sketch (settings.json)
```jsonc
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit|NotebookEdit",
        "hooks": [{
          "type": "command",
          "command": "python3 ~/gt/hooks/memory_path_warn.py"
        }]
      }
    ]
  }
}
```
`memory_path_warn.py` reads the tool input JSON from stdin, checks `file_path` against
the regex, and if matched, prints an `additionalContext` payload per the
`update-config` skill's contract.

**Install target:** user settings (`~/.claude/settings.json`), not project-level.
Project-level installs miss worktrees spawned fresh. The existing hooks at
`~/gt/hooks/` + gastown SessionStart pattern is the reference.

---

## 4. Deprecation Path

**Recommendation: delete per-account `MEMORY.md` indices. Don't keep a stub.**

Reasons:
1. A stub `MEMORY.md` pointing to `gt recall` still gets loaded into every session
   prompt and consumes cache budget for no information.
2. Claude Code's auto-memory system will recreate `MEMORY.md` if it decides to write
   a new memory — the stub buys nothing durable.
3. Migration should be **one-way**: all future writes go through `gt remember`; old
   files are archived under `.migrated/` for auditability for N days then deleted.

Steps:
1. Run migration with `--dry-run`, review, resolve conflicts.
2. Run for real. Archive originals to `.claude-accounts/<acct>/projects/.migrated/<date>/`.
3. Install the warn-hook (§3B). Leave in place for 2 weeks.
4. If drift stays at 0, promote to block-hook (§3A). Delete `.migrated/` after 30 days.

---

## 5. Open Questions

1. **`gt remember` key namespacing.** Currently `--key` is flat (one global namespace).
   Do we need per-rig / per-role namespacing to avoid key collisions when e.g. three
   different rigs all remember something called `refinery-worktree`? Likely yes; spec
   a `--scope <rig>` flag before migrating at scale.
2. **Per-project Claude memories** (e.g. `-home-pratham2-gt-gt-monitor/memory/`) — are
   those also mayor-relevant? Probably not. Recommend migrating *only* mayor-level
   (`-home-pratham2-gt/memory/`) in round 1; leave per-project memories in place so
   project-local Claude sessions still benefit.
3. **What about MCP `memory_remember` / `memory_recall` tools?** The DI MCP server
   already exposes memory tools. Are those backed by the same Dolt kv as `gt remember`?
   If yes → pick one API as canonical. If no → consolidate first, *then* migrate.
