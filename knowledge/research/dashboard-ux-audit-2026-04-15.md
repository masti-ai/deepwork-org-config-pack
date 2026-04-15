# GT Monitor Dashboard — UX Audit & Improvement Plan

Date: 2026-04-15
Auditor: Mayor (via Playwright against live dashboard at localhost:3000)
Screenshots: `.playwright-mcp/audit-*.png` (16 screens)

## TL;DR

The dashboard works but is actively hostile to readers. Three systemic problems:

1. **The chat panel is permanent & giant.** It's anchored to the bottom of every page, currently showing raw Claude Code terminal scrollback (not bubbles), and occupies a huge vertical slice of every single view. You cannot hide it.
2. **Information architecture is deep-but-shallow.** The left nav has 15+ direct click targets at one level. The rig page shows almost nothing. The convoys page shows too much. The git page shows a 400-row commit table. There is no progressive disclosure.
3. **Data is inconsistent across widgets.** Same metric shows different numbers in different places on the same page (agents: 22 vs 27 "alive", beads: 1612 vs 1614, polecat counts in sidebar don't match reality).

Console: 1 error only (favicon 404 — trivial).
Network: all `/v1/*` calls return 200 from `localhost:9090`. No failed requests. Backend is fine.

---

## Per-page findings

### 01. Overview (`#/`)
- Hero banner: "GT Monitor — Degraded — Planning — 6h 37m up"
- "Snapshot updated 1m ago" with auto-refresh badge
- Two rig cards (villa_ai_planogram, gt_monitor) — OK
- "Beads by status" donut showing 1612 total
- "Active agents" card
- Footer stats row: three unlabeled numbers, no column headers
- Chat panel eats the bottom third of the screen at all times

**Problems:** too many unlabeled numbers, no clear "what matters right now" headline, chat eating vertical space.

### 02. Activity / Feed (`#activity/feed`)
- Event stream with cards by type/severity
- Stats bar: 525 events · 11 · 33 · 200 — again unlabeled
- Event cards are scannable — this tab is probably the least broken

**Problems:** unlabeled numbers, filter chips at top are hard to spot.

### 03. Activity / Timeline (`#activity/timeline`)
- Colored lane chart showing events over time by type
- Legend is labeled (good!)
- Actually useful visualization

**Problems:** hover tooltips were not verified; default window is small.

### 04. Activity / Convoys (`#activity/convoys`)
- 60/56/4/149 counts at top
- Small "mountain" badges that are visually cute but information-poor
- Then a MASSIVE table of 149 convoy rows

**Problems:** the table dominates. No filtering visible. No grouping by rig or state.

### 05. Activity / Beads (`#activity/beads`)
- 1614 / 537 / 3 / 18 / 1051 — totals, open, blocked, in-progress, closed? Unlabeled.
- Donut: "By status" with 1614 total
- "Recent beads" list
- Grid of bead cards below

**Problems:** 1614 here vs 1612 on Overview — already inconsistent. No clear entry for searching or slinging.

### 06. Activity / Merge Queue (`#activity/merge-queue`)
- All zeros: `0 / 0 / 0 / 0`
- "Merge queue empty" — but earlier today there were 8 MRs draining
- Grid of refinery cards per rig, most showing DOCKED

**Problems:** no history surfaced. User explicitly asked for 30-day history (gtm-bhs7 is slung but pending). Page is useless the moment the queue is empty.

### 07. Activity / Mail (`#activity/mail`)
- Inbox layout with 60 mails, 21 waiting, 0, 39, 3 archived
- Three-pane look: sidebar + list + detail
- Actually the cleanest page

**Problems:** the "60 mails" vs inbox list showing very few items — count comes from somewhere else than the visible list.

### 08. Activity / Git (`#activity/git`)
- 387 · 42 · 10 · mayor — unlabeled
- Sparkline of commits
- Commit heatmap (truncated / not visible in my screenshot)
- Big 400-row commit table

**Problems:** table is mostly `WIP: checkpoint (auto)` rows — the bloat user complained about is visible right here. Need filtering by "real commits only."

### 09. Activity / Costs (`#activity/costs`)
- Banner: "Planning Mode Active — Budget of <amount>"
- Top card shows budget / used / % saved (huge green progress bar)
- A SECOND cost view below with different numbers and 7 pie charts
- "Top 30 rigs by cost" list

**Problems:** TWO conflicting cost views on the same page with different "used" figures and no explanation. Probably Planning Mode is a budget simulator, but nothing makes that clear.

### 10. Activity / Changelog / Brief (`#activity/brief`)
- Dated summary: "Execution Summary • Rig Health • Governance Updates • WIP list"
- Includes a big table of WIP: checkpoint (auto) commits — the bloat

**Problems:** brief content is flat text, dense, no bead links. Would benefit from the changelog-style per-rig grouping spec'd in gtm-3qv0.

### 11. Activity / Trends (`#activity/trends`)
- Sparklines: tokens by day, cost by day, events/commits by day
- Multiple 30-day charts with x-axis dates

**Problems:** some charts look empty / no data. Sparkline totals need reconciliation against raw source.

### 12. Activity / Plan (`#activity/plan`)
- Planning Mode summary
- "Create Bead" button, Open Beads list
- Huge scrollable list below

**Problems:** basically a clone of the Beads view. Unclear what "Plan" adds.

### 13. Rigs / gt_monitor (`#rigs/gt_monitor`)
- "Operational • 0/0 • 10,000"
- **HUGE empty area below**
- No polecats visible, no recent activity, no MRs, no beads per rig

**Problems:** THIS IS THE WORST PAGE. User explicitly asked "show me what was done in gt_monitor today" and this page, the natural answer, shows ABSOLUTELY NOTHING. Just an empty scroll. 14 polecats are alive in this rig right now.

### 14. Profile (`#profile`)
- Shows user stats, contributor list, recent activity
- Reasonably populated

### 15. Settings (`#settings`)
- Minimal, loaded but couldn't screenshot in time

---

## Systemic problems (in priority order)

### P0-1: The chat panel must not be omnipresent
- Anchored to bottom, always visible, eats the page
- Should be: collapsed rail by default (like Slack or Messenger drawer), opens on demand
- When open, SIDE PANEL (right edge), not bottom band
- Must have a hide toggle

### P0-2: Rig pages are empty
- `#rigs/<rig>` shows one line of stats and nothing else
- Must show: alive polecats (with avatars), today's merged beads, active convoys, recent commits, MR queue for this rig
- This is the single highest-value fix because the user's default question is "what's happening in rig X right now"

### P0-3: Information architecture flattening
- Left nav needs grouping and default collapse
- Proposed top-level: `Overview • Activity • Rigs • Profile • Settings` (5 items)
- Activity expands to: Feed · Convoys · Mountains · Git · Cost · Mail · Changelog
- Drop: Merge Queue (fold into Convoys), Beads (fold into Activity home), Plan (fold into Beads)
- Trends becomes the Cost/Git detail view, not a separate tab

### P0-4: Kill unlabeled numbers
- Every number on the page needs an inline label
- `<n> · <n> · <n>` → `<n> cost · <n> stamps · <n> beads`
- Hover tooltips aren't enough; visible labels are the bar

### P0-5: One source of truth per metric
- Pick ONE bead count, ONE agent count, ONE cost total per day
- Route every widget through the same API
- Reconcile: 1612 vs 1614 beads, 22 vs 27 agents, 5 vs 14 polecats

### P0-6: Planning Mode explanation
- Banner says "Planning Mode Active" but gives no context
- What does it mean? Is spend real or simulated? Why does the second card below show actual spend?
- One-line description in banner + link to docs

### P1-7: WIP commit filter on Git tab
- Default filter = real commits only
- Toggle to show WIP checkpoints
- Separately: fix the root cause (`gtm-mp0g` squash-on-merge)

### P1-8: Chat = real chat, using Claude JSONL transcripts
- Already covered by `gtm-o98d` epic
- Bubble UI, no regex, persistent chat store

### P1-9: Progressive disclosure on heavy tables
- Convoys table: collapse to rig groups, expand on click
- Git table: virtualize + filter by author/date/type
- Beads grid: default filter to "today" or "this week"

### P1-10: Typography & color polish
- Some pages use muted chip colors; others use saturated — inconsistent
- Font weight hierarchy unclear (two-level h1/h2 would help)
- Already have tokens landed (`gtm-okg4`), now need to enforce per-page

---

## Proposed beads (to sling)

1. **P0 Side-panel chat drawer with open/close toggle** (separate from chat-content rebuild in gtm-o98d)
2. **P0 Rig page rebuild: polecats + today's merged + active convoys + MRs** — THE single most important fix
3. **P0 Nav IA flattening to 5 top-level entries with grouped subs**
4. **P0 Label-every-number sweep across all pages**
5. **P0 Single-source-of-truth reconciliation: beads/agents/polecats counts**
6. **P0 Planning Mode explainer banner + cost view disambiguation**
7. **P1 Git tab: default hide WIP commits + filter**
8. **P1 Progressive disclosure on Convoys, Git, Beads tables**
9. **P1 Typography consistency: enforce dw-font tokens per page audit**
10. **P1 Research Vercel/shadcn dashboard patterns for best practices** — user suggested web search for UX skills; will use the `shadcn` and `frontend-design` skills available to the polecat assigned

## UX research sources to inform implementation

- shadcn/ui dashboard patterns (plugin skill available: `vercel-plugin:shadcn`)
- `frontend-design:frontend-design` skill for distinctive aesthetics (avoids generic AI look)
- Tailwind + Radix for accessible primitives
- GitHub's org/repo/project IA pattern (user already approved this mental model)
- Linear for keyboard-first navigation (cmd+k is present but hidden)
- Notion for progressive disclosure on tables

## Verification contract for every fix
- Playwright before/after screenshots
- Network 200 for all `/v1/*` calls
- Console clean (no new errors)
- Visible labels on every number
- No widget shows >1 metric without a legend
