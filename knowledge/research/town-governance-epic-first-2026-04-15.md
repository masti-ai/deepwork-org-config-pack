# Gas Town Governance Rule: Epic-First Workflow (2026-04-15)

**Source of authority:** Overseer (Pratham) directive 2026-04-15 + crew dashboard session analysis in tmux `gtm-crew-dashboard`.

**Rule:** All work in Gas Town starts with an epic. No agent creates or executes a bead outside an epic. If the epic's scope is `shared-rig` or `town`, it is mirrored to the Wasteland immediately via the Deepwork Intelligence MCP server. To begin work on an epic, a rig/crew must **claim** it (wasteland side-effect: claim ledgers).

## Canonical flow

```
1. Overseer instructs Mayor ("make me X")
2. Mayor creates Epic with required labels:
   - scope:local-rig | scope:shared-rig | scope:town
   - rig:<name>
   - status:draft initially
3. If scope != local → Mayor calls mcp__deepwork-intelligence__wasteland_forge(epic_id)
   - Creates identical-template wasteland issue
   - Other orgs' agents can see + claim
4. Mayor authors child beads under the epic (1:1, no cluster guessing)
5. Convoy groups the beads; Mountain tracks them
6. Assignee (crew or worker) MUST call `gt epic claim <epic-id>` before starting
   - Updates status:claimed + claimer label
   - Emits EpicClaimed event
7. Beads close 1-by-1; mountain advances
8. Last bead close → EpicClosed event fires
   → mcp__deepwork-intelligence__wasteland_audit(epic_id) → stamp with AuditResult
   → github_push_epic(epic_id) → squashed PR to masti-ai/<repo> with template body
9. Wasteland entry sealed = permanent audit archive
```

## Enforcement (HARD)

- `bd create` hook rejects beads without an `epic:<id>` label (exempt: epic itself)
- `gt sling` hook rejects targets if assigned bead's parent epic has status=draft (must be claimed first)
- `refinery merge` blocks PR if the merging bead's epic has no `claimer:` label
- All such rejections emit a human-readable error pointing to this doc
- Exemption: `ops-exempt` label on individual beads for true emergencies only (logged loudly)

## Template CRM

All templates live in `deepwork-org-config-pack/templates/` (or equivalent global path):
- `issue.wasteland.md.tmpl` — wasteland issue body
- `issue.bead.md.tmpl` — bead body
- `pr.epic-close.md.tmpl` — github PR body on epic close
- `commit.normal.tmpl` — normal commit message (non-WIP)
- `commit.wip-checkpoint.tmpl` — WIP checkpoint commit
- `email.handoff.tmpl` — mayor-to-mayor handoff mail (already exists?)
- `email.escalation.tmpl` — escalation mail
- `mail.daily-report.tmpl` — town daily report

Loader: `mayor/scripts/load_template.sh <name> [vars...]` reads template, substitutes variables from beads/epics/wasteland, outputs rendered content. Single source of truth. Overseer edits files in the pack; changes ripple immediately.

Config knob: `~/.gt/templates.json` per-user overrides (optional).

## Deprecated boards to shut down

Mark archived on masti-ai + Deepwork-AI + remove from any cron/mirror/rig-list references:
- `gasclaw` (replaced by local gasclaw-1/2 containers, not a repo workflow)
- `image-studio-murmur` (planned feature, never built, obsolete)
- `gascity-hotel` (OfficeWorld variant, obsolete)
- `command-center` (v1 dashboard, replaced by gt-monitor)

Do NOT delete — archive to preserve history. Remove from mirror cron. Drop from CLAUDE.md references.

## Out of scope for v1
- Cross-org shared epic handoff UX (just emit the wasteland entry; humans negotiate claims)
- Template versioning (single-branch for now)
- Real-time wasteland sync across orgs (polling is fine until inflow picks up)
