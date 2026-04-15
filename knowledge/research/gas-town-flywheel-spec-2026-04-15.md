# Gas Town Workflow Spec — Epic-First, Event-Driven, Audit-Native

**Status:** Draft v1 — written for overseer review before any implementation.
**Authority:** Overseer directive 2026-04-15 (crew/dashboard session).
**Supersedes:** `town-governance-epic-first-2026-04-15.md` (keeps its canonical flow, replaces the "shared vs local" scoping with an org model) and `wasteland-reputation-redesign.md` (keeps its data-flow observations, upgrades the stamp payload to an audit payload).

---

## 0. Why this doc exists

Two separate failures drove this rewrite:

1. **The flywheel was dead for 10 days** because a vLLM schema bug silently failed every LLM call. Nothing paged anyone. Cron ran, logged 400s, continued. The dashboard kept showing old numbers. This is what "invisible failure in a polling loop" looks like — it always looks like it's working.
2. **Wasteland was reverse-engineering work already done.** An LLM looked at the last 100 closed beads and guessed which ones belonged together. This created duplicates, mis-clusters, and a permanent drift between "what the team did" and "what wasteland says they did."

Both failures are structural, not bugs. This spec fixes the structure.

---

## 1. Principles (non-negotiable)

| Principle | What it means in practice |
|---|---|
| **Events, not polling.** | Every state transition fires a named event. Crons exist only for reconciliation (detect drift between expected and actual), not for driving work. |
| **Declarative, not archaeological.** | Wasteland items are forged *at epic creation*, not reconstructed from closed beads. The audit trail is built as work happens. |
| **Fail loud, fail once.** | Every event handler writes a structured result (`ok`/`err` + reason) to a `pipeline_events` log table. One failed handler → one red light on the dashboard. No silent 400s. |
| **One source of truth per concept.** | Rigs in `gt_collab.rigs`. Epics in per-rig `issues` table (`issue_type='epic'`). Wasteland in `gt_collab.wanted`. Mountain tracking in `hq.convoys`. No second copies. No "also stored in". |
| **Audit ≠ review.** | Refinery does not review code (it's a merge shim with a health check — confirmed 2026-04-15). Polecat self-police is the only current gate. MiniMax audit is therefore the *first* structured quality signal the system produces. |
| **Every epic is audited.** | Not just shared ones. Internal audit builds the reputation substrate even when no external contributor sees the work. |
| **Org ≠ "shared-rig" flag.** | Rigs belong to a user/town. A rig is *in an org* if its owner has joined that org. The org sees the rig's epics at the visibility level the owner grants. No `shared` boolean on the rig itself. |

---

## 2. The Org Model

### Entities

```
Town  (a user's collection of rigs)
  └─ owns → Rig
             └─ produces → Epic
                            └─ contains → Bead(s)
                                           └─ tracked by → Convoy (Mountain)

Org   (a group of Towns)
  └─ has → OrgMember(town_id, role)
             role ∈ { admin, maintainer, contributor }
  └─ sees → Rig(s) contributed by member Towns at visibility level
             visibility ∈ { private, org, public }
```

### Key rules

- **A user's Town is private by default.** Their rigs are private by default.
- **To share a rig into an org**, the Town owner sets `rigs.org_visibility['<org>'] = 'org' | 'public'`. No "shared=true" boolean — visibility is per-org.
- **An org has no rigs of its own.** It is a membership overlay. All work lives in its members' rigs.
- **Roles:**
  - `admin` — add/remove members, change org settings, archive stale rigs.
  - `maintainer` — audit any org-visible epic, publish wasteland items to the org's upstream Dolt remote.
  - `contributor` — claim wasteland items, post new ones from their own rigs.
- **Upstream Dolt remote** is per-org. `deepwork` org uses `dolthub.com/deepwork/gt-collab`. A user's `origin` is their fork; `upstream` is the org's shared branch.

### Schema delta on `rigs` table

Already has: `handle, display_name, dolthub_org, hop_uri, owner_email, trust_level, rig_type, parent_rig`.

Add:
```sql
ALTER TABLE rigs ADD COLUMN org_visibility JSON DEFAULT NULL;
  -- e.g. {"deepwork": "org", "personal": "public"}
ALTER TABLE rigs ADD COLUMN repos JSON DEFAULT NULL;
  -- e.g. [{"platform": "github", "owner": "masti-ai", "repo": "gt_monitor"}]
```

New table:
```sql
CREATE TABLE orgs (
  handle         VARCHAR(64) PRIMARY KEY,
  display_name   VARCHAR(255),
  dolt_upstream  VARCHAR(512),
  created_at     TIMESTAMP
);

CREATE TABLE org_members (
  org            VARCHAR(64),
  town           VARCHAR(255),
  role           VARCHAR(16),     -- admin | maintainer | contributor
  joined_at      TIMESTAMP,
  PRIMARY KEY (org, town),
  FOREIGN KEY (org)  REFERENCES orgs(handle),
  FOREIGN KEY (town) REFERENCES rigs(handle)  -- town's "home rig" is its handle
);
```

---

## 3. The Canonical Flow

```
┌─────────────────────────────────────────────────────────────────┐
│  OVERSEER: "make me X"  (in Claude session, mail, slack, etc.)   │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  MAYOR / CREW: creates Epic (bd create --type=epic)              │
│  Required fields:                                                │
│    title, description, acceptance_criteria,                      │
│    rig, project, visibility (private|org:<handle>|public),       │
│    repos[] (for github mirror), convoy_id (auto on create)       │
└─────────────────────────────────────────────────────────────────┘
                            │
          ┌─────────────────┼─────────────────┐
          ▼                 ▼                 ▼
   ┌────────────┐    ┌───────────┐    ┌──────────────┐
   │ Mountain   │    │ Wasteland │    │ GitHub Issue │
   │ (convoy)   │    │ (wanted)  │    │ (if org vis) │
   └────────────┘    └───────────┘    └──────────────┘
   auto-created      forged on        issued via
   with convoy_id    epic.created     gh api on
                     event            epic.created
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  MAYOR: authors child beads under epic (1:1, no clustering)      │
│  Each bead: parent_epic_id, convoy_id (inherited), repo target   │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  POLECAT: claims bead → works → git commit → gt done             │
│  gt done pushes branch, submits MR bead, nukes sandbox           │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  REFINERY: merges MR to main, closes bead                         │
│  bead.closed event fires                                          │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
            ┌──────────────────────────────────┐
            │  Is this the last open bead       │
            │  in the epic's convoy?            │
            └──────────────────────────────────┘
                  no  │        │  yes
                      ▼        ▼
             update mountain   epic.completed event fires
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  DEEPWORK INTELLIGENCE (MCP, event handler, not cron):           │
│    1. fetch all bead bundles for the epic                        │
│    2. fetch all MR outcomes (merge SHA, refinery verdict)        │
│    3. fetch PR diff stats (LOC, files touched, tests added)      │
│    4. fetch test results from CI artifact (if present)           │
│    5. call MiniMax with the full epic context                    │
│    6. emit AuditResult → stamps table                            │
│    7. close wasteland entry (status='completed')                 │
│    8. comment + close GitHub issue with PR link                  │
│    9. fire epic.audited event                                    │
└─────────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  DOWNSTREAM CONSUMERS (all subscribe to epic.audited):           │
│    - crown_refresh: update contributor tiers                     │
│    - agent_notify: mail each polecat their stamp + remarks       │
│    - skill_tags_update: update per-town skill profile            │
│    - mayor_digest: append to today's feedback doc                │
│    - dashboard push: update live audit feed                      │
└─────────────────────────────────────────────────────────────────┘
```

---

## 4. The Event Bus

### Design

- **Storage:** single table `gt_collab.pipeline_events` — an append-only log.
  ```sql
  CREATE TABLE pipeline_events (
    id           VARCHAR(64) PRIMARY KEY,   -- ulid
    event_type   VARCHAR(64) NOT NULL,      -- e.g. epic.created
    entity_type  VARCHAR(32) NOT NULL,      -- epic | bead | wasteland | ...
    entity_id    VARCHAR(128) NOT NULL,
    payload      JSON NOT NULL,
    emitted_by   VARCHAR(255) NOT NULL,
    emitted_at   TIMESTAMP NOT NULL,
    -- handler tracking:
    handlers     JSON,   -- {"wasteland_forge":{"ok":true,"ms":420},"github_issue":{"err":"403"},...}
    handler_state VARCHAR(16) DEFAULT 'pending',  -- pending | partial | done | failed
    INDEX (event_type), INDEX (entity_id), INDEX (handler_state, emitted_at)
  );
  ```
- **Emission:** anything that mutates state calls `emit_event(type, entity, payload)`. Single Python helper (`mayor/lib/events.py`) and Rust helper (`gt-monitor-executor/src/events.rs`).
- **Dispatch:** a tiny `eventd` process subscribes to `pipeline_events` inserts (Dolt binlog polling, 1s interval — the *only* polling in the system, and its failure mode is "delayed" not "silent"). Fires handlers in parallel, writes result back to `handlers` JSON.
- **Reconciliation cron (once per hour):** scans `handler_state='failed'` OR `handler_state='pending' AND emitted_at < now() - 5 min` → retries or pages.
- **Dashboard:** live tail of `pipeline_events` on Overview. Failed handlers light red.

### The event catalog

| Event | Emitted by | Handlers | Payload |
|---|---|---|---|
| `epic.created` | `bd create --type=epic` post-hook | `wasteland_forge`, `github_issue_create`, `mountain_create` | `{epic_id, rig, visibility, repos}` |
| `epic.claimed` | `gt epic claim <id>` | `wasteland_claim`, `notify_org` (if org-visible) | `{epic_id, claimer}` |
| `bead.created` | `bd create` post-hook | `convoy_add_bead`, `mountain_update` | `{bead_id, parent_epic_id}` |
| `bead.closed` | `bd close` or refinery merge | `mountain_update`, `check_convoy_complete` | `{bead_id, closed_by, merge_sha}` |
| `convoy.complete` | `check_convoy_complete` handler when last bead closes | `emit epic.completed` | `{convoy_id, epic_id}` |
| `epic.completed` | `convoy.complete` handler | `wasteland_audit` | `{epic_id}` |
| `epic.audited` | `wasteland_audit` handler | `crown_refresh`, `agent_notify`, `skill_tags_update`, `github_pr_comment`, `dashboard_push` | `{epic_id, audit_id, stamp_id}` |
| `audit.failed` | `wasteland_audit` handler on error | `mayor_escalate` | `{epic_id, reason}` |

### Why this is different from the cron flywheel

- Every event has a known handler set. A missing handler result is a red light.
- Handler errors are stored with the event, not buried in a log file.
- The dashboard shows *current* state (queue depth, failed handlers, latency) — not a 20-min-stale snapshot.
- Reconciliation is additive (retry), not driving (the event already fired; we're just repairing handler state).

---

## 5. Data Model Changes

### Bead schema additions (per-rig `issues` table)

```sql
ALTER TABLE issues ADD COLUMN parent_epic_id VARCHAR(64);
ALTER TABLE issues ADD COLUMN convoy_id      VARCHAR(64);
ALTER TABLE issues ADD COLUMN visibility     VARCHAR(32) DEFAULT 'private';
ALTER TABLE issues ADD COLUMN repos          JSON;           -- per-bead override of epic's repos
ALTER TABLE issues ADD INDEX (parent_epic_id);
ALTER TABLE issues ADD INDEX (convoy_id);
```

### Wasteland schema additions (`gt_collab.wanted`)

```sql
ALTER TABLE wanted ADD COLUMN epic_id      VARCHAR(64);       -- forged from this epic
ALTER TABLE wanted ADD COLUMN source_town  VARCHAR(255);      -- who forged
ALTER TABLE wanted ADD COLUMN source_rig   VARCHAR(64);
ALTER TABLE wanted ADD COLUMN org          VARCHAR(64);       -- which org sees this
ALTER TABLE wanted ADD COLUMN github_url   VARCHAR(512);      -- mirror issue URL
ALTER TABLE wanted ADD INDEX (epic_id), ADD INDEX (org);
```

### Stamps upgraded to audits (`gt_collab.stamps`)

Current stamp has `valence JSON` (Q/R/C). Extend the JSON schema to hold the full audit payload — no table change needed, just a richer JSON body:

```json
{
  "valence": {"quality": 4, "reliability": 5, "creativity": 3},
  "remarks": [
    "Counter.most_common pattern is idiomatic — keep using this for aggregation.",
    "Evidence text is rich, reviewers have context to score fairly."
  ],
  "anti_patterns": [
    "Hardcoded 'deepwork' as completed_by — data losses attribution. Always resolve from closed_by_session."
  ],
  "coord_score": 0.85,
  "coord_notes": "3 polecats on this epic: kareo (1 bead), rictus (2 beads), mayor (review). Commits are well-sequenced; no overlapping edits to the same file within 30 min.",
  "skill_tags": ["rust-sql", "dolt-schema", "event-plumbing"],
  "acceptance_criteria_met": [
    {"criterion": "Attribution preserved", "met": true, "evidence": "completions.completed_by = 'kareo' in new row c-4f0cb7d6"}
  ],
  "test_evidence": {
    "tests_run": true,
    "tests_passed": true,
    "coverage_delta": "+2 files covered",
    "ci_artifact_url": "..."
  }
}
```

**Key:** the stamp is now the audit. One row per epic. Agents read it and get remarks they can actually act on.

---

## 6. The Audit: What MiniMax Produces

### Input to MiniMax

The `wasteland_audit(epic_id)` handler gathers, in order:

1. Epic fields: title, description, acceptance_criteria, design.
2. All child beads: id, title, description, close_reason, notes, closed_by_session.
3. All MR outcomes: branch, merge_sha, merged_at, refinery verdict (currently "merged:true/false").
4. PR diff stats: files changed, +LOC / −LOC, tests added/removed.
5. CI artifact (when available): test result counts, coverage.
6. Git log of each bead's branch.

Total context: typically 5–20 KB. MiniMax M2.7 has 262 K context window — no squeezing.

### Output schema (Pydantic)

```python
class AuditResult(BaseModel):
    valence: Valence  # Q/R/C
    remarks: list[str]           # learnings, ≤5, each ≤200 chars
    anti_patterns: list[str]     # concrete "don't do this" ≤3
    coord_score: float           # 0-1, multi-polecat coordination
    coord_notes: str             # ≤300 chars, explanation
    skill_tags: list[str]        # ≤5 normalized tags
    acceptance_criteria_met: list[CriterionCheck]
    test_evidence: TestEvidence
    should_reject: bool = False
    reject_reason: str | None = None
```

### Prompt shape (abbreviated)

> You are the Gas Town auditor. You have the full epic, every bead, every PR diff, and test results. Your job is *not* to re-review the code (refinery already merged it). Your job is to:
> 1. Rate the delivery Q/R/C with specific evidence from the artifacts.
> 2. Extract learnings other agents should carry forward (remarks).
> 3. Flag anti-patterns they should stop doing.
> 4. Score coordination: how well did multiple polecats hand off?
> 5. Tag skills demonstrated.
> 6. Check each acceptance criterion against the diff.
> 7. If tests didn't run or CI is red, `should_reject=true`.

### Why this motivates agents

- **Remarks are remembered.** Each agent's profile surfaces their last 10 remarks on `gt prime`. The next polecat on that agent address reads them before starting.
- **Anti-patterns block.** Three anti_patterns on the same theme within 30 days triggers a `reputation_check` flag — that agent only gets offered bronze-tier beads until three clean audits repair it. Real consequence.
- **Skill tags route work.** New epic with `tags=['rust-sql']` preferentially offers to towns with `skill_tags` profile matching.
- **Coord score builds crew rep.** Multi-polecat epics produce a coordination signal that's separate from individual quality — a team that hands off well ranks higher for coordination-heavy epics.
- **5/5 is visible.** The dashboard shows "Recent 5/5 audits" as a leaderboard. Public reputation. Low scores are not publicly shamed (bronze-tier gating is private) but high scores are celebrated.

---

## 7. Mountain

### What it is

A mountain = a convoy of beads that together close an epic. It exists in the `hq` Dolt DB as convoy rows. `gt mountain status --json` lists them. The UI's Mountain tab queries this.

### Why you thought it wasn't working

It does work at the CLI. Confirmed today: 4 active mountains (`hq-cv-n3zjv`, `hq-cv-u9hiv`, `hq-cv-j5ixq`, `hq-cv-364ff`). The dashboard's Mountain tab renders them but the row-count / progress plumbing appears to have a bug. That's a **dashboard bug**, not a workflow bug. Separate ticket.

### What changes under this spec

- Mountain becomes a **first-class concept in the epic**: `epic.convoy_id` is set at creation. No more discovering convoys after the fact.
- Mountain progress = `count(bead WHERE status='closed') / count(bead WHERE parent_epic_id=<epic>)`. Updated on every `bead.closed` event.
- Dashboard subscribes to `pipeline_events` for mountain updates — live, not polled.

---

## 8. Observability (the lesson from the 10-day silent failure)

Every handler returns a structured result. Every failure is visible. The dashboard's **Pipeline Health** strip (new) shows:

| Handler | Last ok | Last fail | p50 latency | Queue depth |
|---|---|---|---|---|
| `wasteland_forge` | 2m ago | — | 180 ms | 0 |
| `wasteland_audit` | 18m ago | — | 4.2 s | 1 |
| `github_issue_create` | 7m ago | **12m ago** (403) | 310 ms | 2 |
| `crown_refresh` | 22m ago | — | 90 ms | 0 |
| `mountain_update` | 0s ago | — | 40 ms | 0 |

Red rows are loud. Queue depth > 5 pages the mayor. A handler that's been `pending` for > 5 minutes pages the mayor.

**The test for "is this working?":** can the overseer look at one dashboard strip and answer "yes, fully" or "no, X is broken because Y" in under 10 seconds? If yes, the system is observable. If it takes log diving, we built another cron.

---

## 9. Migration from Current State

Ordered, each step independently shippable:

1. **Kill the wl_commons bug.** Delete the stale default in `deepwork_intelligence/refinery/rig/config.py`. Confirm DI flywheel writes to `gt_collab`. **30 min.** Unblocks everything else — until this is done, any reputation/audit data lands in a DB nobody else reads.

2. **Add the event log table.** `gt_collab.pipeline_events` per §4. Add `emit_event()` helper in Python + Rust. No handlers yet — just emission. **2 hours.**

3. **Migrate schemas.** Add columns on `rigs`, `issues`, `wanted`. Add `orgs` and `org_members` tables. All additive, no data loss. **1 hour.**

4. **Build `eventd`.** Simplest possible: Python script polling `pipeline_events WHERE handler_state='pending' ORDER BY emitted_at ASC LIMIT 20`, dispatches handlers, writes results back. Systemd unit. **3 hours.**

5. **Replace `wasteland_map_beads` + `wasteland_complete_matched`** with `wasteland_forge` (on `epic.created`) and `wasteland_audit` (on `epic.completed`). Delete the 20-min cron. **4 hours.**

6. **Rewrite `wasteland_stamp` → AuditResult schema.** Migrate the 125 existing stamps to the new JSON shape (preserve Q/R/C, leave remarks/anti_patterns/etc empty). **2 hours.**

7. **Wire `bd` hooks.** `bd create`, `bd close` emit events. `bd create --type=epic` requires `visibility`, `rig`, `repos`. **3 hours.**

8. **Wire `gt done`** → emit `bead.closed` when refinery merges. **1 hour.**

9. **Deploy crown system.** Already coded in `polecats/guzzle/…/wasteland_crown_refresh.py`. Copy to refinery, register as handler for `epic.audited`. **1 hour.**

10. **Dashboard pipeline-health strip.** New component, subscribes to `pipeline_events`. **4 hours.**

11. **Agent notify handler.** On `epic.audited`, mail each contributor their stamp + remarks. **2 hours.**

12. **Mountain UI fix.** Separate. Not blocked by this spec.

**Total:** ~22 hours if linear, ~8 hours parallelized across 3 polecats + me.

---

## 10. What We Delete

Permanently removed once the event model lands:

- `wasteland_map_beads` — LLM bead clustering (produces duplicates + mis-clusters; obsoleted by epic-first).
- `wasteland_complete_matched` — scans for closed bead sets (obsoleted by convoy completion event).
- `wasteland_flywheel` monolith — replaced by event handlers.
- `crons/flywheel-test.sh` — the 20-minute polling script.
- `mayor/scripts/wasteland-dispatch.sh` (already disabled).
- `mayor/scripts/wasteland-reviewer.sh` — replaced by `wasteland_audit` event handler.
- `mayor/scripts/wasteland-completion-sync.py` — replaced by `convoy.complete` handler.

Kept as reconciliation only:
- `wasteland-push.sh` (Dolt sync to DoltHub).
- `wasteland-pull.sh` (Dolt fetch from org upstream).

---

## 11. Overseer Decisions (2026-04-15)

1. **Org naming:** the word is **org**. Our org is **deepwork**.
2. **Repo mapping:** repos are named by project family (`villa`, `villa-data`, `ai`, `deepwork`/`internal`). Canonical path is `github.com/deepwork/<repo>`. A fork by a member becomes `github.com/<username>/<repo>`. A rig's `repos[]` may list any number. A bead MAY target a repo outside its rig's declared set (because forking is a legitimate cross-rig workflow) — but it MUST be tracked: `bead.repo_override` field, and the `epic.audited` handler verifies the target.
3. **GitHub push semantics:** **strict GitHub flow, no exceptions.**
   - Every bead lives on a branch.
   - Every branch lands via PR.
   - Every PR has ≥1 reviewer (refinery or a maintainer) before merge.
   - **Nothing goes to main without a reviewed PR.** Cross-town commits included.
   - Commit message convention: conventional commits (`feat(scope): …`, `fix(scope): …`) with issue/bead ID ref.
   - On `epic.audited`: comment the audit summary on the epic's GitHub issue, then close it. No auto-release.
4. **Reject path:** on `should_reject=true` → mail mayor, flag `epic.status='audit_rejected'`, await manual triage. Do NOT reopen wanted. Confirmed.
5. **Retention:** stamps immutable forever; Dolt handles growth.
6. **Auditor trust model:** use the wasteland tier system (reputation via crowns). Auditor from any org member's MiniMax is accepted; trust weighting = auditor's own crown tier. A diamond-tier auditor's stamp outweighs an iron-tier auditor's. Aggregation for v2.
7. **Cross-rig claim etiquette:** first-come first-served claim. When overseer forges an epic in a shared rig, wasteland items are created immediately. Agent claims lock the item. Other polecats pick the next unclaimed item. No duplicates. When a polecat finishes and goes back for more, claimed items are skipped.
8. **Mountain:** keep. Works at CLI (`gt mountain status --json` → 4 active convoys confirmed 2026-04-15). UI plumbing bug is a separate ticket, not part of this migration.

---

## 12. Acceptance Criteria for "this spec is done"

- Overseer can give one sentence ("make me X") and every downstream artifact (epic, beads, convoy, wasteland, github issue, mountain, audit, stamp, crown, remarks, mail) is produced or updated without further prompting.
- One dashboard strip answers "is the system healthy right now?" in under 10 seconds.
- A polecat reading `gt prime` sees their last audit's remarks inline, before starting new work.
- A failed handler is visible within 1 minute of failure. A stuck handler is visible within 5 minutes of stall.
- The 10-day silent failure cannot happen again. If it does, the root-cause bug is in `eventd` itself (one script to watch) rather than distributed across 7 cron jobs.

---

*End of spec. Read through once, tell me what to change, I'll revise before any code lands.*
