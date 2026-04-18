# Hooks Reference — Claude Code (and any harness with hooks)

Deepwork uses Claude Code hooks as the deterministic substrate that makes the
flywheel (memory / docs / release / audit) work across every agent without
relying on the LLM to remember to call tools.

All hook scripts live under `~/.claude-accounts/<account>/hooks/` and are
mirrored into each account so hooks survive an account switch. Paths below
resolve via the `~/.claude` symlink that flips between accounts.

## Active hooks (2026-04-18+)

| Event | Matcher | Script | Purpose |
|---|---|---|---|
| `PreToolUse` | `Write\|Edit` | `readme-guard.sh` | Refuses edits that would nuke a README/CHANGELOG unless the prompt explicitly asked |
| `PostToolUse` | `Bash` | `post-close-emit.sh` | On `bd close` / `bd done` / `gt done` / `gt close`, emits `bead.closed` or `epic.closed` into `pipeline_events` for downstream handlers |
| `PostToolUse` | `Bash` | `epic-mountain-guard.sh` | Nudges if an epic is created without the expected mountain shape (children + blocks) |
| `PostToolUse` | `Write\|Edit` | `memory-mirror.sh` | Auto-captures edited file paths into agent scratch memory for next-session recall |
| `SessionStart` | — | `memory-recall-session-start.sh` | Pulls recent memories for the current project and prepends them to the model's system context |
| `PreCompact` | — | `deepwork_intelligence/agents/handoff/precompact_hook.sh` | Dumps session state + open beads into a handoff bead before context is compacted |

## Account-sync rule

Every account under `~/.claude-accounts/` (pratham, pratham_cc1, pratham_cc3)
must have:
- a `hooks/` directory with **identical script contents**
- a `settings.json` whose `hooks` block references `~/.claude/hooks/*.sh`

When the `~/.claude` symlink flips (account switch), the hook path resolves
to the newly-active account's `hooks/`. Missing hooks in any account
silently kills the flywheel — we saw this on 2026-04-17 (README/release
stopped updating for 4 hours after a switch).

**Audit command:**

```bash
diff <(ls ~/.claude-accounts/pratham_cc1/hooks/) \
     <(ls ~/.claude-accounts/pratham/hooks/) \
     <(ls ~/.claude-accounts/pratham_cc3/hooks/)
```

Must produce empty output.

## The `post-close-emit.sh` contract

The single most important hook — it's the event bus for the flywheel.

On every `bd close <id>` or equivalent command, the hook:
1. Extracts bead IDs from the command line
2. Looks up the bead type via `bd show --json`
3. Calls `mayor.lib.events.emit_event("bead.closed", ...)` or
   `emit_event("epic.closed", ...)` into the shared `pipeline_events` table
4. Disowns the subprocess so the command returns instantly — the hook
   never blocks the agent

Downstream handlers (docs_from_epic, auto-release, bead_audit plugin, etc)
poll `pipeline_events` and react. Without this hook, everything downstream
sits idle. When onboarding a new agent, verify the hook fires by closing a
throwaway bead and checking the `pipeline_events` table for a new row.

## Productized variant (post-v5 config pack)

`di_core/mind_events/` is a harness-agnostic, config-driven version of the
same pipeline. `MIND_ENABLE_WASTELAND=0 MIND_ENABLE_BEAD_AUDIT=1` turns off
Gas-Town-specific handlers while keeping the agentic audit plugin active —
that's what a Deepwork Mind install outside Gas Town ships with.

## Mandate: `gt` only, never `gc`

The deprecated `gc` (gas-city) CLI is not used. If you see `gc prime` or
`gc nudge` anywhere, replace with `gt prime` / `gt nudge`. Locked owner
rule from memory: mixing `gc` and `gt` in the same environment breaks the
hook pipeline.
