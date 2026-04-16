# Deepwork Mind — MCP Tools Available to All Agents

> **Every agent has `.mcp.json` in their worktree. These tools are available NOW.**

## Memory Tools — USE THESE

| Tool | When to use |
|------|-------------|
| `memory_recall(scope="town")` | On session start — load org decisions, patterns, anti-patterns |
| `memory_recall(scope="rig")` | On session start — load rig-specific context |
| `memory_remember(content, kind, scope, source_bead)` | When you discover something — save IMMEDIATELY, don't wait |
| `memory_forget(memory_id)` | When a memory is outdated |

**Kinds**: pattern, anti-pattern, decision, incident, skill, context
**Scopes**: agent (private), rig (rig team), town (all agents), global (cross-org)

## Wasteland Tools

| Tool | When to use |
|------|-------------|
| `wasteland_status()` | See the org work board |
| `wasteland_stamp(completion_id)` | Score a completion (MANDATORY on session close) |
| `wasteland_claim(wanted_id)` | Claim a work item |

## Docs Tools

| Tool | When to use |
|------|-------------|
| `docs_generate(rig, doc_type)` | Generate readme, changelog, architecture, api-reference |
| `docs_create(rig, title, doc_type, context)` | Create a new doc |
| `docs_index(rig)` | List all docs for a rig |

## Analytics

| Tool | When to use |
|------|-------------|
| `health()` | Check DI system health |
| `analytics_usage()` | See tool call stats |

## Rules

1. **MANDATORY**: Call `memory_recall` at session start (formula step 1.5)
2. **MANDATORY**: Save ≥1 memory via `memory_remember` before exit (formula step 5)
3. **MANDATORY**: Call `wasteland_stamp` before exit (formula step 6)
4. Store learnings CONTINUOUSLY — when you discover something, save it immediately
5. Read memories before you code — prior polecats may have solved your problem
