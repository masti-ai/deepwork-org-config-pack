# Deepwork Org Config Pack

Shared operational knowledge for Deepwork-AI Gas Town deployments. All agents reference this pack for roles, patterns, anti-patterns, and operational procedures.

**Version:** 3.0.0 (2026-04-01)

## What's Inside

```
deepwork-org-config-pack/
├── pack.yaml                  # Manifest
├── docs/
│   ├── ARCHITECTURE.md        # How Gas Town works
│   ├── GLOSSARY.md            # Terminology
│   ├── RUNBOOKS.md            # Step-by-step procedures
│   └── wasteland/ONBOARDING.md
├── knowledge/
│   ├── anti-patterns.md       # What breaks (learned from incidents)
│   ├── patterns.md            # What works (proven approaches)
│   ├── decisions.md           # Key architectural decisions with reasoning
│   ├── operations.md          # Debug shortcuts, emergency procedures
│   ├── troubleshooting.md     # Common issues and fixes
│   ├── conventions.md         # Naming, config files, CLI commands
│   ├── hooks-reference.md     # Claude Code hook catalog
│   ├── formulas-reference.md  # Formula catalog + gap analysis
│   └── plugins-reference.md   # Plugin catalog
├── roles/
│   ├── mayor.yaml             # Town coordinator
│   ├── deacon.yaml            # Automated patrol
│   ├── witness.yaml           # Per-rig lifecycle
│   ├── refinery.yaml          # Merge processor
│   ├── polecat.yaml           # Disposable worker
│   └── crew.yaml              # Persistent worker
├── crons/
│   └── town-crons.yaml        # Active + recommended cron jobs
├── rules/
│   └── deepwork-governance.yaml
├── blueprints/
│   └── deepwork-corp/blueprint.yaml
└── templates/
    ├── pr-body.md
    └── mesh.yaml.template
```

## Key Changes in v3.0.0

- **Roles rewritten** — Now reflects actual Gas Town agent roles (mayor, deacon, witness, refinery, polecat, crew) instead of generic planner/worker/reviewer
- **Knowledge modernized** — Removed stale mesh/gasclaw references. Added hooks, formulas, and plugins reference catalogs
- **Anti-patterns updated** — Real incidents from 2026-03/04: Dolt crashes, GitHub suspension, ulimit, Docker CPU loops
- **Crons consolidated** — Single town-crons.yaml with active, disabled, and recommended sections
- **Docs added** — ARCHITECTURE, GLOSSARY, RUNBOOKS from the town knowledge system
- **Gap analysis** — Compared against gascity reference implementation, documented missing formulas and exec orders

## Architecture (Quick Reference)

| Component | Purpose |
|-----------|---------|
| **Dolt** (port 3307) | SQL database for beads, mail, agent state |
| **Gitea** (port 3300) | Git hosting — all agent work. GitHub = public mirror only |
| **gt daemon** | Process manager, plugin scheduler |
| **Hooks** | Claude Code lifecycle (SessionStart, Stop, PreCompact, etc.) |
| **Plugins** | 14 deacon patrol tasks on cooldown gates |
| **Crons** | 3 active: thread guardrail, log rotation, knowledge evolution |
| **Formulas** | 6 workflow templates (polecat-base, polecat-commit, do-work, scoped-work, + demos) |

## Using This Pack

Agents read these files for operational context. The knowledge system auto-evolves:
- **Cron** (every 6h) scans closed beads for lessons
- **Plugin** (every 12h) harvests knowledge during deacon patrol
- **Session Stop** hook logs activity to changelog

To capture knowledge manually:
```bash
bash ~/gt/mayor/knowledge/capture.sh <type> "<title>" "<body>" "<source>"
# Types: pattern, anti-pattern, decision, operations
```

## Gap Analysis (vs gascity reference)

Features in the reference we should adopt:
1. **Exec orders** — Shell scripts on cooldown without LLM. We use plugins (need deacon).
2. **mol-polecat-work** — Feature-branch + refinery variant. We only have mol-polecat-commit.
3. **Patrol formulas** — mol-deacon-patrol, mol-witness-patrol, mol-refinery-patrol.
4. **Per-role hook overlays** — Witness gets different hooks than polecat.
5. **Spawn storm detection** — Auto-detect crash-looping beads.
6. **Pack system** — Composable pack.toml with includes, overlays, doctor checks.

See `knowledge/formulas-reference.md` for the full gap analysis.
