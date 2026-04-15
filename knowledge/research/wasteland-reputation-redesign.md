# Wasteland Reputation — Why It's Broken + Redesign

**Status:** research only. No schema migration. No code. Sources cited with
concrete paths and table names.

---

## TL;DR

The wasteland stamp pipeline today scores *free-text evidence blobs* against a
rubric the LLM half-sees. It has no persistent link from a stamp back to the
beads it came from, no per-agent aggregation, and it duplicates the code review
that refinery already performs on the MR. The fix is to stop reviewing *code*
in wasteland and start reviewing *outcomes*: bead closure, merge status, stamp
Q/R/C, all rolled up per assignee. `bead_mappings` already exists — it's just
never read.

---

## 1. Current-State Audit

### Tools (all in `deepwork_intelligence/orgs/deepwork/tools/`)

| Tool | Reads | Writes | LLM context seen |
|------|-------|--------|------------------|
| `wasteland_stamp(completion_id)` | `wl_commons.completions` + `wl_commons.wanted` (title, effort, project, 500-char description, evidence blob) | `wl_commons.stamps` (Q/R/C + reasoning), `wl_commons.completions.validated_by/stamp_id`, `wl_commons.wanted.status='completed'` | **Only the evidence string** — plus a regex-extracted signal block (test phrases, SHA hashes, bead IDs, verb counts). No bead bodies. No code diff. No git log. |
| `wasteland_map_beads(rig)` | `<rig>.issues` (last 100, 200-char description), `wl_commons.wanted` open/claimed | `wl_commons.bead_mappings(wasteland_id, bead_id, rig, confidence)` | 100 bead titles + descriptions, matched to open wanted items. **The mapping is persisted but never read by any other tool.** |
| `wasteland_flywheel(rig)` | N/A (orchestrator) | — | Chains `map_beads` → `complete_matched` → `review_all`. |
| `wasteland_stamp` prompt | `agents/wasteland/stamp.py` — `BASE_PROMPT` + `_extract_signals()` | — | Rubric tells the LLM to look for "tests pass", SHA hashes, merge phrases in a string. No structured access to any of those. |

### Data model (wl_commons)

```
wanted        : id, title, project, description, effort_level, status, claimed_by
completions   : id, wanted_id, evidence, completed_by, validated_by, stamp_id
stamps        : id, author, subject, valence(json Q/R/C), context_id, stamp_type, message
bead_mappings : wasteland_id, bead_id, rig, confidence, mapped_at    ← orphan table
```

Per-rig bead data lives in separate Dolt dbs (`villa_ai_planogram`, `gt_monitor`,
`villa_alc_ai`, ...). The MCP server has `dolt.query(<rig>, …)` already — reach
is a config issue, not an access issue.

### What the LLM actually sees at stamp time

From `wasteland_stamp.py` lines 69–85:

```
SELECT c.id, c.evidence, c.completed_by, w.title, w.effort_level, w.project,
       LEFT(w.description, 500) as description
```

Seven fields. `evidence` is a free-text blob written by the completer. That's it.
The stamp prompt (`agents/wasteland/stamp.py:BASE_PROMPT`) then asks the LLM to
guess at correctness, delivery, and creativity from *one string*.

---

## 2. Root Causes — Why Reputation Is "Not Working"

1. **`bead_mappings` is write-only.** `wasteland_map_beads` inserts rows,
   `wasteland_stamp` never reads them. Every stamp is scored in isolation from
   the beads the work actually closed. The join that matters
   (`completion → wanted → bead_mappings → issues`) is never performed.
2. **No structured delivery signal.** Merge success is a hard fact in
   `gastown.merge_requests` but the stamp pipeline never queries it. The LLM is
   asked to grep the evidence string for "merged to main" — a parody of what is
   already a SQL query away.
3. **Wasteland duplicates refinery's job.** Refinery runs the actual code review
   on the MR (`/home/pratham2/gt/deepwork_intelligence/refinery/`). Wasteland's
   Q score is "code correctness from a text blob" which is strictly worse than
   "the MR merged cleanly after refinery review." Two reviewers, one signal,
   wasteland loses.
4. **No per-agent rollup.** `stamps.subject = completed_by` is present but
   nothing aggregates it. There is no `agent_reputation` table, no view, no
   dashboard tile. A polecat with ten 5-star stamps and a polecat with ten
   1-star stamps look identical to the dispatcher.
5. **Stamps score the report writer, not the worker.** `completed_by` is
   whoever filed the completion (often a mayor script, not the polecat). The
   reputation drifts to the reporter rather than the assignee that did the work.
6. **Context starvation at stamp time.** 500 chars of description + evidence
   string. No `close_reason`, no bead notes, no `issue_history` delta, no MR
   SHA. The rubric asks for signals the tool never fetches.

---

## 3. Proposed Redesign

### 3a. Give the MCP server direct bead + MR access (cheap)

MCP already has `dolt.query(db, sql)`. Add two shared helpers:

- `fetch_bead_bundle(rig, bead_id)` → issue row + history deltas + close_reason
  + assignee + most-recent git refs mentioned in notes.
- `fetch_mr_outcome(branch)` → `gastown.merge_requests` status, merged_at,
  merge_sha, refinery verdict.

No new services. No new tools exposed to agents. Internal helpers only.

### 3b. Make every wasteland item carry its beads forward

`wasteland_stamp` should pull the bead bundle for every `bead_mappings` row
attached to the completion's `wanted_id`. Feed the stamp LLM **structured** data:

```
item.title, item.effort
beads[]: {id, title, status, close_reason, assignee, history_delta}
mrs[]:   {branch, status, merged_at, merge_sha}
evidence: <current free-text>   ← demoted to a secondary signal
```

The rubric changes from "read this blob" to "these N beads closed under this
item, M of them shipped, the close_reasons look like X — score accordingly."

### 3c. Split concerns: code review vs. outcome review

| Concern | Owner | Signal produced |
|---------|-------|-----------------|
| Code correctness, style, tests | **Refinery** (MR review, CI) | `merge_requests.status`, `merge_sha` |
| Outcome holds up, bead stayed closed, work shipped | **Wasteland stamp** | Q/R/C |
| Aggregate per-agent reputation | **New: `agent_reputation` view** | rolling scores |

Wasteland Q (Quality) becomes "did the *outcome* hold — bead stayed closed, no
reopens, CI clean post-merge." Wasteland stops judging code. Refinery keeps
judging code. They stop competing.

### 3d. Per-agent reputation as a Dolt view (not a service)

```sql
-- wl_commons.agent_reputation (materialized view, refreshed by cron)
CREATE VIEW agent_reputation AS
SELECT
  i.assignee                                        AS agent,
  COUNT(DISTINCT bm.bead_id)                        AS beads_touched,
  SUM(CASE WHEN i.status='closed'                    THEN 1 ELSE 0 END) AS beads_closed,
  SUM(CASE WHEN i.close_reason LIKE 'no-changes:%'   THEN 1 ELSE 0 END) AS no_change_closes,
  AVG(JSON_EXTRACT(s.valence,'$.quality'))          AS q_avg,
  AVG(JSON_EXTRACT(s.valence,'$.reliability'))      AS r_avg,
  AVG(JSON_EXTRACT(s.valence,'$.creativity'))       AS c_avg,
  SUM(CASE WHEN mr.status='merged' THEN 1 ELSE 0 END) AS mrs_merged,
  SUM(CASE WHEN mr.status='rejected' THEN 1 ELSE 0 END) AS mrs_rejected
FROM  <each rig>.issues         i
JOIN  wl_commons.bead_mappings  bm ON bm.bead_id = i.id
JOIN  wl_commons.completions    c  ON c.wanted_id = bm.wasteland_id
LEFT  JOIN wl_commons.stamps    s  ON s.id = c.stamp_id
LEFT  JOIN gastown.merge_requests mr ON mr.assignee = i.assignee
                                     AND mr.created_at BETWEEN i.created_at AND i.updated_at
GROUP BY i.assignee;
```

(Schema is illustrative; actual rig joins use a UNION across `PROJECT_TO_DB`.)

Reputation score = `w_q·q_avg + w_r·r_avg + w_c·c_avg + w_merge·(mrs_merged/beads_touched) − w_reject·(mrs_rejected/beads_touched) − w_nochange·(no_change_closes/beads_closed)`.

Defaults: `w_q=0.15, w_r=0.2, w_c=0.1, w_merge=0.4, w_reject=0.3, w_nochange=0.15`.

### 3e. Dispatcher uses reputation

`gt sling` and refinery auto-dispatch consult `agent_reputation` to:
- prefer polecats with `q_avg >= 3.5` for `priority=1` work,
- cool down polecats with `no_change_closes/beads_closed > 0.4` (churn filter),
- short-circuit reputation below a floor (e.g. reassign stalled beads).

---

## 4. State Diagram

```
  ┌──────────────┐   bd create           ┌──────────────┐
  │ bead (open)  │──────────────────────▶│ <rig>.issues │
  └──────┬───────┘                        └──────┬───────┘
         │ assigned                              │
         ▼                                       │
  ┌──────────────┐   wasteland_map_beads         │
  │ bead (in_    │──── (LLM semantic match) ────▶│  bead_mappings
  │  progress)   │                               │  (wasteland_id,bead_id)
  └──────┬───────┘                               │
         │ gt done → push branch                 │
         ▼                                       │
  ┌──────────────┐   refinery review             │
  │  MR in queue │──── code review here ────────▶│ gastown.merge_requests
  └──────┬───────┘                               │
         │ merge                                 │
         ▼                                       │
  ┌──────────────┐                               │
  │ bead (closed)│                               │
  └──────┬───────┘                               │
         │ wasteland_complete_matched            │
         ▼                                       │
  ┌──────────────┐   wasteland_stamp             │
  │  completion  │──── OUTCOME review ──────────▶│ stamps (Q/R/C)
  │  (evidence + │     reads bead_mappings,      │
  │   beads[] +  │     reads merge_requests      │
  │   mrs[])     │                               │
  └──────┬───────┘                               │
         │                                       ▼
         │                           ┌─────────────────────────┐
         └──────────────────────────▶│ agent_reputation (view) │
                                     │  per assignee rollup    │
                                     └───────────┬─────────────┘
                                                 │
                                                 ▼
                                     gt sling / refinery dispatch
                                       (prefer high-rep polecats)
```

---

## 5. Migration Path (read-only → read+write → dispatch)

1. **Phase 0 — instrument.** Run `wasteland_map_beads` for every rig on a cron
   (it already persists into `bead_mappings`). No behaviour change.
2. **Phase 1 — enrich stamps.** Modify `wasteland_stamp` to fetch
   `bead_mappings` + bead bundles + MR outcome before calling the LLM. Rubric
   swapped to outcome-focused. Q now measures "did the bead stay closed,"
   not "is the code good."
3. **Phase 2 — aggregate.** Ship `agent_reputation` view. Dashboard tile +
   `di_health` reads from it.
4. **Phase 3 — gate dispatch.** `gt sling` reads reputation when choosing a
   polecat for a new hook.

Each phase is a separate bead; phase 1 is the one that actually fixes "it's
not working."

---

## 6. Open Questions

1. **Multi-rig reputation stitching.** A polecat that works `gt_monitor` today
   and `villa_ai_planogram` tomorrow needs one identity. Current `assignee`
   strings use `<rig>/polecats/<name>`, so the same human polecat class becomes
   N distinct agents. Do we key reputation on `role_class` (mayor/witness/
   polecat) + `name` only, dropping rig? Probably yes.
2. **Stamp attribution.** Should we score the *assignee* (who did the work) or
   `completed_by` (who filed the completion)? Bead beats completion — switch
   primary subject to `issues.assignee`.
3. **Cold start.** New polecat has zero stamps. Default reputation = `q_avg=3,
   confidence=0.1` — treat as "weak prior."
4. **Stamp decay.** Should reputation half-life? A polecat that was bad in
   February and good in April shouldn't be dragged down. Suggest 30-day
   exponential decay on the aggregation query.
5. **Refinery feedback.** Refinery already produces pass/fail verdicts on MRs.
   Should those land as `stamps` with `stamp_type='review'` so `agent_reputation`
   naturally absorbs them? Likely the cleanest integration — one stamps table,
   two authors (`deepwork-intelligence`, `refinery`).
6. **MCP auth.** Giving MCP read access to every rig's Dolt db is fine; giving
   it *write* is not. The redesign keeps writes in `wl_commons.*` only.
