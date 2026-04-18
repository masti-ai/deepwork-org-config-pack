# Deepwork Mind — Setup & Install

Deepwork Mind is the shared intelligence layer: memory, skills, workflows,
and signals across every agent in the org. This doc covers how a new org
member gets onboarded and what the product surface looks like end-to-end.

## One-line install

```bash
curl -fsSL https://deepwork.art/install-cli.sh | bash
```

What this does:
1. Downloads the `bd` binary (beads — the work graph) from upstream
2. Creates a hermetic Python venv at `~/.deepwork/venv`
3. Installs `deepwork-mind` into that venv
4. Symlinks `mind` into `~/.local/bin`

No system pip, no Node required. Re-run to upgrade.

## Per-project init

Inside any project directory:

```bash
mind init
```

That detects the harness (Claude Code / Codex / Cursor / OpenCode), runs
`bd init --prefix <slug>` in the project, writes `.mind.json`, patches the
harness's MCP config, installs the Deepwork Mind skill cheatsheet into the
harness's skill/rules location, and — on Claude Code — installs a
SessionStart hook that injects project memory on every session start.

Flags:
- `--beads local` (default) vs `--beads skip` — turn off bd install
- `--harness <name>` to override auto-detection
- `--offline` to skip the Deepwork API and use a local UUID for project_id

## What the MCP exposes to the coding agent

Once `mind init` is done, the LLM running in the harness gets these MCP
tools (all auto-registered, all discoverable via standard MCP listing):

- **Memory**: `memory_remember`, `memory_recall`, `memory_edit`,
  `memory_delete`, `memory_merge`
- **Skills**: `skill_search`, `skill_install`, `skill_list`
- **Workflows**: `workflow_run`, `workflow_test`, `workflow_list`,
  `workflow_create` (draft → tested → published lifecycle)
- **Projects**: `project_init`, `project_join`, `project_list`,
  `project_switch`

The **skill cheatsheet** (dropped into the harness's skill location)
teaches the LLM the tool catalog — see `skills/deepwork-mind.md` for the
canonical content.

## Configuration

Per-user config lives at `~/.deepwork/config.yaml`. Env-var overrides
always win. Most important knobs:

| What | Env var | Default |
|---|---|---|
| Dolt host / port | `MIND_DOLT_HOST` / `MIND_DOLT_PORT` | `127.0.0.1:3307` |
| Dolt Hub upstream | `MIND_DOLT_UPSTREAM_URL` | empty (local-only) |
| LLM provider | `DEEPWORK_PROVIDER` | first ready of minimax / anthropic / openai / ollama |
| Enable wasteland handlers | `MIND_ENABLE_WASTELAND` | `1` (disable for non-GT installs) |
| Enable bead audit plugin | `MIND_ENABLE_BEAD_AUDIT` | `1` |

Provider config (OpenCode-compatible schema) lives at
`~/.config/opencode/opencode.json` or `~/.deepwork/config.yaml`. Mind reads
both — users who already configured OpenCode with Kimi / Codex / Claude
keys get them for free in Mind.

## Check install health

```bash
mind doctor     # reports bd, dolt reachability, MCP wiring, hook status
```

If hooks aren't firing, first thing to check: is `~/.claude` pointing at
the account that has `hooks/` populated? See `knowledge/hooks-reference.md`
for the account-sync rule.
