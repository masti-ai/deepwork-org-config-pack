# DeepWorkMind + DeepWork Intelligence — Event-Driven Knowledge Layer (2026-04-15)

## Product positioning (three pillars)

1. **GT Monitor** — standalone monitoring system for Gas Town. Observability, dashboards, alerts, cost ledger.
2. **DeepWorkMind** — standalone memory + knowledge graph + structured file system for agents. Persists learnings, docs, RAG substrate. Has **segmented minds** (marketing mind, infrastructure mind, product mind, etc.) so agents don't drown in irrelevant context.
3. **DeepWork Intelligence (DI)** — standalone layer of **functions** operating on DeepWorkMind's memory. `wasteland_audit`, `epic_propose`, `docs_add`, `pattern_extract`, `worker_skill_profile`, etc. DI is NOT memory — it's the verb layer.

The three work independently, ship independently, integrate cleanly. MCP is just the API/gateway, not the product.

## Core insight (the one killing the current cron)

Cron-based pack updater runs 24/7 even when I'm asleep. That burns tokens + compute for nothing. Real-world signal is event-driven: a memory gets written, an epic closes, a research doc lands. Replace the cron with **hooks that fire only when something meaningful happens**.

## New DI MCP function: `docs_add(data, source_hint, metadata)`

The smart router. Called from every meaningful event. Mind decides where this data belongs:

```
docs_add(data, source_hint, metadata)
  ↓
MiniMax 2.7 classifies:
  ├── org-level knowledge → deepwork-org-config-pack/knowledge/*.md
  ├── rig-level knowledge → <rig>/knowledge/*.md
  ├── README update → <rig>/README.md patch
  ├── Release note → draft release entry queued for next cut
  └── Knowledge-graph node → dolt graph table (edges to related nodes)
```

Metadata fields: `rig`, `epic_id`, `bead_id`, `author`, `event_type`, `timestamp`. The router uses these + content classification to decide destination.

## Event taxonomy (hooks that fire docs_add)

| Event | Source | When |
|---|---|---|
| `MemoryWritten` | `gt remember` | Every time a memory is stored |
| `BeadClosed` | bd post-close hook | Any bead moves to closed |
| `EpicClosed` | bd post-close hook on epic type | Epic finalized (fires `wasteland_audit` first, then `docs_add` with the AuditResult) |
| `ResearchCommitted` | git post-commit hook on `*/research/*.md` | Research doc lands |
| `FormulaAdded` | git post-commit hook on `*/formulas/*` | New formula file |
| `MoleculeAdded` | git post-commit hook on `*/molecules/*` | New molecule |
| `EscalationClosed` | `gt escalate close` | Anti-pattern candidate |
| `PromptRevised` | git post-commit hook on `*.tmpl`, `*.prompt`, CLAUDE.md | Prompt change — log to prompt_revisions table |

Each event handler calls `docs_add()` with structured input. No cron needed for extraction. Cron stays ONLY as daily reconciliation safety net (catches events that didn't fire).

## Segmented minds

DeepWorkMind doesn't serve one giant context. Agents request context by **mind segment**:

```
mcp__deepwork-mind__query(segment="marketing", question="...")
mcp__deepwork-mind__query(segment="infrastructure", question="...")
mcp__deepwork-mind__query(segment="product", question="...")
mcp__deepwork-mind__query(segment="dashboards", question="...")
```

Segments are declared in `deepwork.yaml` — each segment maps to a set of knowledge files + rigs + memory types. When loading an MCP tool call, only the relevant segment is loaded into MiniMax context. Keeps the token cost down + relevance high.

Default segments: `org`, `marketing`, `infrastructure`, `product`, `ops`, `research`. Users can add more via config.

## Daily batch (rolled-up learnings)

Still need ONE daily batch (not 20-min):
- Cluster all `docs_add` events from the last 24h
- Emit a per-segment "what we learned today" file
- Feeds the town daily report + updates the pack commit
- Squashed PR to `masti-ai/deepwork-org-config-pack`

This is where real learnings emerge — they're extracted on-the-fly by hooks, then curated/clustered daily.

## Product-grade install (both DeepWorkMind and DI)

Dedicated skill in `prompts.chat:skill-manager` (or our own skills registry):

`skill: deepwork-mind-install` — runs:
1. Check prerequisites (dolt, python, MiniMax 2.7 or remote fallback)
2. Clone the mind repo + run migrations + seed segments
3. Write `~/.claude/mcp/deepwork-mind.json` (MCP server registration)
4. Write `~/.claude/mcp/deepwork-intelligence.json` (DI functions registration)
5. Self-test: `docs_add("hello world", source_hint="test")` → confirm routed correctly
6. Print MCP server status + how to invoke from Claude Desktop

Must be idempotent + revertable. Works on fresh Linux/macOS box.

## Integration with existing beads

- Supersedes scope of `de-1u7j` (DOCP updater v2) — that one now becomes "migrate pack-update.sh into new event model"
- Extends `de-anyl` (Template CRM) — templates are a segment of the mind
- Extends `de-74ua` (event bus) — same bus, adds new event types
- `de-hogl` (M2.7 harness) is a prerequisite
- `gtm-wtdp` (Claude data sources research) feeds this — JSONL transcripts become a memory source

## Deliverables (per child bead)

1. DI audit + M2.7 capability harness (already `de-hogl`)
2. `docs_add()` MCP function + router
3. Event taxonomy + hook scripts
4. Knowledge-graph dolt schema
5. Segmented-minds config + query routing
6. Daily batch learnings clustering
7. `deepwork-mind-install` skill (plug-and-play)
8. `deepwork-intelligence-install` skill (plug-and-play)
9. Replace pack-update.sh cron → event-driven
10. Prompt-revisions dolt table + git hook (also satisfies standalone prompt-log need from vision doc)

## Out of scope for v1
- Cross-org knowledge graph federation (single-org for now)
- Real-time segment reindexing (daily batch is fine)
- Auto-summarization into RAG embeddings (segments are keyword-retrievable first)
