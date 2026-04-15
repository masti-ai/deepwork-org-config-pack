# Gas Town Human Interface — Design Vision (2026-04-15)

Source: brainstorming session between Overseer (Pratham) and Mayor.
Supersedes prior dashboard specs where they conflict.

## Product framing

This is NOT a monitoring dashboard. This is **the human interface layer for Gas Town** — the place humans command a crew of agents and see the work flow. Think IDE, not Grafana. The human is always *steering*, never *authoring*.

## Core principles

### 1. Three levels, each with an implicit chat target

| Level | What you see | Who you talk to |
|---|---|---|
| **Town** | Org-wide reports, trends, cross-rig activity, required actions | Mayor |
| **Rig** | Project detail: crew, work-done timeline, open boards, KPIs | Crew member of that rig |
| **Agent** | What they're working on — board of beads they own/spawned | That agent |

Navigation = drill-down. URL: `/town/...`, `/rigs/<rig>/...`, `/rigs/<rig>/crew/<name>`, `/rigs/<rig>/workers/<name>`.

### 2. Chat is a steering device, not a reading device

- Always present as a collapsible panel (well-animated)
- Target = current page's entity (Town→Mayor, Rig→crew, Agent→that agent)
- Not the primary view. User sees the work first, opens chat only when they want to steer
- Clicking a bead attaches it to the chat as a widget/chip ("selected item")
- Typing with a bead attached mutates that bead in real time
- Simple mutations (single field) = instant + undo toast
- Complex mutations (cascading deps, priority, multi-bead) = agent-mediated propose-confirm

### 3. Humans never create artifacts

- No "Create Bead" / "Create Epic" / "Create Mountain" buttons in the UI
- All artifact creation flows through conversation with Mayor or crew
- The UI shows the artifact AFTER the agent creates it (in planning mode)
- This is why the local model (MiniMax 2.7) matters: bead churn must be free

### 4. Terminology

- **Polecat → Worker** (rename, repo-wide, with deprecation alias)
- Workers do a bead and are done. You don't talk to a worker.
- Crew stays around. You plan with crew. Crew talks back.

### 5. Telemetry over transcripts

Every bead has a **single-line status**: "writing migration for users table", "waiting on dolt query", "needs overseer — question on scope", etc. That line is the main signal. Chat/transcripts are available but not primary.

Status line origin: workers self-report via `bd status-set` on key transitions (tool call, result, turn boundary). Post-tool hook enforces.

### 6. "Needs You" / Required Actions surface

First-class, separate from Mail:
- Permanent button in top bar with red-dot counter
- Opens an action feed: each item = bead + worker question + context + proposed next step
- Inline actions: Approve · Deny · Open in chat · Reassign
- Auto-clear when answered; auto-surface when worker tags `needs-human` or witness detects stall
- OS-level notifications for high-priority

### 7. Agent-level board is a bespoke visualization

NOT a traditional kanban grid. Innovative, animated, interactive:
- Epic as visible unit
- Workers attached to epic rendered as orbiting circles
- Click a worker circle → their 5 assigned beads appear
- Click a bead → status line visible; click attaches to chat as widget

Detailed visual spec is the designer's job. Board = source-of-truth for "what's actually happening in this agent's world."

### 8. Prompt-revision log

Every edit to a wisp/formula/prompt/molecule file must be logged:
- Dolt table `prompt_revisions` with (who, when, file, diff, reason)
- Git hook automates capture; no manual step
- Enables bisecting behavioral regressions against prompt history
- Hypothesis user flagged: recent behavior changes (polecats getting numbered instead of named, refinery not merging as often) may be due to untracked prompt edits

## Supporting UI (lives at Town level)

Always-accessible, always global:
- Required Actions feed
- Cost ledger
- Mail
- Activity feed
- Search (cmd+k)

Cross-cutting rig/agent views (merge queue, git history) collapse into their rig context — not top-level.

## Open decisions (noted, not blocking)

- Animation library: Framer Motion vs CSS transforms vs Web Animations
- How the agent-level innovative board actually renders (designer to explore)
- Whether the dolt prompt-revisions log is per-rig or global
- Exact schema for "structured status line" (freeform string initially, structured later)

## Out of scope for v1

- Humans creating beads directly (forbidden by principle 3)
- Multi-workspace / multi-tenancy (one Gas Town per user)
- Mobile (desktop-first, responsive later)
