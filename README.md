# Deepwork Org Config Pack

**v5.0.0 — 2026-04-18**

Onboarding substrate for new Deepwork org members and for any Deepwork Mind
install outside Gas Town. If you just joined the org, this is the only
repo you need to skim to be productive.

## What this pack is

- **Deepwork Mind setup** — how to install the `mind` CLI, run `mind init`
  in a project, and get the MCP + hooks wired into your coding agent.
- **Hook baseline** — the Claude Code hook config every agent runs. See
  `knowledge/hooks-reference.md` for the 6 hooks that make the flywheel
  work (memory-recall on session start, post-close-emit on `bd close`,
  memory-mirror on Write/Edit, etc).
- **Foundational knowledge** — conventions, anti-patterns, decisions.
  Stuff you'd otherwise have to learn by getting corrected three times.
- **Core formulas** — starter bead / release / code-review formulas. Most
  Gas-Town-specific formulas moved to `archive/`.
- **Skills** — the three `gt-mesh` starter skills. Product skills live in
  Mind itself; this pack is scaffolding, not the runtime.

## What this pack is NOT

- Not a runtime. Skills / workflows / memories are managed by Deepwork
  Mind (see `knowledge/deepwork-mind-setup.md`). This pack seeds a fresh
  install; it doesn't replace the product.
- Not the operational handbook for Gas Town internal roles
  (mayor / polecat / refinery / witness / deacon). That content moved to
  `archive/2026-04-pre-mind-v5/` — it's alive but not onboarding-relevant.
- Not the MCP tool catalog. The LLM reads that from the skill cheatsheet
  that `mind init` drops into the harness (see `knowledge/deepwork-mind-tools.md`).

## Quick start for a new org member

```bash
# 1. Install the CLI + bd
curl -fsSL https://deepwork.art/install-cli.sh | bash

# 2. In any project you work on:
cd ~/projects/acme-api
mind init

# 3. Open your coding agent. It will discover the Deepwork Mind MCP
#    automatically and start using memory / skill / workflow tools.

# 4. Read knowledge/deepwork-mind-setup.md and
#    knowledge/hooks-reference.md — ~15 minutes, gets you the full model.
```

## What changed in v5 (from v4)

- **Dropped:** 50+ Gas-Town-internal artefacts (mol-polecat-* formulas,
  wasteland hook scripts, mayor/deacon/witness/refinery role files,
  operational knowledge like mail-routing / account-cycling / worker-sla).
  All in `archive/2026-04-pre-mind-v5/`.
- **Added:** `knowledge/deepwork-mind-setup.md`,
  `knowledge/bead-audit-plugin.md`, revised `knowledge/hooks-reference.md`,
  updated `hooks/claude.json` aligned with the 6 hooks actually running
  (stopped using deprecated `gc` commands — mandate is `gt` only).
- **Rewrote:** `pack.yaml` (5.0.0), this README.
- **Bumped:** version 4.0.0 → 5.0.0.

## Directory map

```
.
├── README.md                        (you are here)
├── pack.yaml                        pack manifest
├── STANDARDS.md                     legacy — keep or absorb into knowledge/
├── knowledge/                       14 .md files — read these
├── hooks/claude.json                Claude Code baseline hook config
├── formulas/                        7 core formulas (40+ archived)
├── scripts/                         knowledge evolve + changelog helpers
├── skills/                          3 gt-mesh skills (Mind manages the rest)
├── roles/                           mayor + crew (operational roles archived)
├── rules/                           governance
├── crons/                           worker crons
├── blueprints/                      deepwork-corp blueprint
├── bin/                             cli helpers
├── enforcement/                     governance enforcement
└── archive/2026-04-pre-mind-v5/     everything we trimmed — keep for reference
```

## Contributing back

If you find yourself maintaining a file in `archive/` across multiple
projects, that's the signal it belongs back in the pack. File a bead
describing the use case and propose it for v5.x inclusion.

---

Updated by mayor, 2026-04-18.
