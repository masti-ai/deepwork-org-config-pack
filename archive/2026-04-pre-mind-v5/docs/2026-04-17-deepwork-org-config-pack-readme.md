# deepwork-org-config-pack

Centralized configuration pack for Deepwork Labs agent orchestration. This repository contains reusable workflow formulas that define agent behaviors, task patterns, and operational procedures.

## Overview

The config pack provides a library of `.formula.toml` files that encode organizational knowledge, workflow patterns, and agent behavior specifications. These formulas are consumed by Deepwork Mind MCP tools and drive consistent agent execution across the organization.

**Key characteristics:**
- Automated daily updates via scheduled sync
- Version-controlled configuration with changelog tracking
- Modular formula architecture for independent workflow components
- Mandatory reference for all Deepwork Labs agents

## Formula Structure

Formulas are defined in TOML format under `./formulas/`. Each formula encapsulates:

- **Trigger conditions** — when the workflow applies
- **Action sequences** — ordered steps for execution
- **Validation rules** — success criteria and checkpoints
- **Context requirements** — necessary inputs and state

### Formula Categories

| Category | Description | Count |
|----------|-------------|-------|
| Code Review | Review workflows, conflict resolution, checkpoint validation | 4 |
| Patrol/Checkpoint | Monitoring, validation, and status verification | 4 |
| Planning | Plan writing, review, and expansion patterns | 3 |
| Release | Deployment and release management workflows | 1 |
| Workflow | General task orchestration patterns | 4 |
| Expansion | Pattern expansion and interview-style refinement | 2 |
| Utility | Shutdown, handoff, and convoy operations | 4 |

## Available Formulas

### Code Review & Quality

| Formula | Purpose |
|---------|---------|
| `code-review.formula.toml` | Standard code review workflow |
| `mol-polecat-code-review.formula.toml` | Polecat-specific review pattern |
| `mol-polecat-conflict-resolve.formula.toml` | Conflict resolution for polecat operations |
| `mol-dog-checkpoint.formula.toml` | Checkpoint validation workflow |

### Patrol & Monitoring

| Formula | Purpose |
|---------|---------|
| `mol-deacon-patrol.formula.toml` | Deacon patrol pattern |
| `mol-refinery-patrol.formula.toml` | Refinery monitoring workflow |
| `mol-plan-review.formula.toml` | Plan review and validation |
| `mol-dep-propagate.formula.toml` | Dependency propagation checks |

### Planning & Specification

| Formula | Purpose |
|---------|---------|
| `plan-writing-expansion.formula.toml` | Plan writing with expansion |
| `spec-questions-interview-expansion.formula.toml` | Specification interview pattern |
| `spec-workflow.formula.toml` | Specification workflow |
| `beads-creation-expansion.formula.toml` | Creation workflow with expansion |
| `beads-workflow.formula.toml` | Bead-based task workflow |

### Release & Deployment

| Formula | Purpose |
|---------|---------|
| `gastown-release.formula.toml` | Gastown release workflow |

### Utility Operations

| Formula | Purpose |
|---------|---------|
| `mol-shutdown-dance.formula.toml` | Graceful shutdown procedure |
| `mol-graceful-handoff.formula.toml` | Task handoff pattern |
| `mol-convoy-feed.formula.toml` | Convoy feed operation |
| `mol-polecat-commit.formula.toml` | Commit workflow for polecat |

### Specialized Workflows

| Formula | Purpose |
|---------|---------|
| `design.formula.toml` | Design workflow |
| `shiny.formula.toml` | Shiny workflow pattern |
| `shiny-secure.formula.toml` | Secure variant of shiny workflow |
| `towers-of-hanoi-10.formula.toml` | Complex multi-step orchestration |
| `mol-polecat-work-monorepo.formula.toml` | Monorepo work pattern |

## Usage

### For Agents

All Deepwork Labs agents must reference this config pack. The Deepwork Mind MCP tools integrate with these formulas automatically.

```bash
# Clone the config pack
git clone https://github.com/deepwork-labs/deepwork-org-config-pack.git

# Formulas are loaded by MCP tools at runtime
# No manual configuration required for standard agent operations
```

### For Configuration

Reference specific formulas in your agent configuration:

```toml
[agent]
workflow = "code-review"  # References code-review.formula.toml

[agent]
workflow = "mol-polecat-code-review"  # References polecat-specific pattern
```

## Maintenance

### Update Schedule

The config pack receives automated updates on a scheduled basis:
- Updates run multiple times daily
- Changes are tracked in the changelog
- Version history preserved in git

### Changelog

See [CHANGELOG.md](./CHANGELOG.md) for detailed version history and change records.

### Contributing

When adding new formulas:
1. Create the `.formula.toml` file in `./formulas/`
2. Follow the standard formula schema
3. Commit with descriptive message
4. Updates will propagate automatically

## Schema Reference

```toml
[formula]
name = "workflow-name"
version = "1.0.0"
category = "code-review|patrol|planning|release|utility|workflow"

[formula.triggers]
# Conditions that activate this formula

[formula.actions]
# Ordered sequence of steps

[formula.validation]
# Success criteria

[formula.context]
# Required inputs and state
```

## Related Documentation

- [Deepwork Mind MCP Tools Reference](./knowledge/mcp-tools-reference.md) — Mandatory documentation for all agents
- [CHANGELOG.md](./CHANGELOG.md) — Version history and change log

## License

Internal use only — Deepwork Labs organization configuration.